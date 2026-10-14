package dell

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dellPendingSettingsCommitted is the exact error shape returned live by a Dell iDRAC (firmware
// 1.5.3, PowerEdge R6715) when a Bios/Settings PATCH is rejected because an earlier PATCH
// already sealed a pending, uncommitted BIOS config job.
const dellPendingSettingsCommitted = `{"error":{"@Message.ExtendedInfo":[{"Message":"Pending configuration values are already committed, unable to perform another set operation.","MessageArgs":["HttpDev1EnDis"],"MessageArgs@odata.count":1,"MessageId":"IDRAC.2.14.SYS011","RelatedProperties":["#/Attributes/HttpDev1EnDis"],"RelatedProperties@odata.count":1,"Resolution":"Wait for the scheduled job to complete or delete the configuration jobs before attempting more set attribute operations.","Severity":"Warning"}],"code":"Base.1.18.GeneralError","message":"A general error has occurred. See ExtendedInfo for more information"}}`

// dellPendingSettingsCommittedOlderRegistry is the exact error shape returned live by a Dell
// iDRAC on a PowerEdge R760xd2 for the same conflict as dellPendingSettingsCommitted, but under
// message registry schema version 2.9 instead of 2.14 - confirmed live to silently disable
// recovery before isDellPendingSettingsMessageID stopped hardcoding the registry version.
const dellPendingSettingsCommittedOlderRegistry = `{"error":{"@Message.ExtendedInfo":[{"Message":"Pending configuration values are already committed, unable to perform another set operation.","MessageArgs":["SecureBootPolicy"],"MessageArgs@odata.count":1,"MessageId":"IDRAC.2.9.SYS011","RelatedProperties":["#/Attributes/SecureBootPolicy"],"RelatedProperties@odata.count":1,"Resolution":"Wait for the scheduled job to complete or delete the configuration jobs before attempting more set attribute operations. If issue still persists perform iDRAC reboot.","Severity":"Warning"}],"code":"Base.1.12.GeneralError","message":"A general error has occurred. See ExtendedInfo for more information"}}`

func TestIsDellPendingSettingsMessageID(t *testing.T) {
	tests := []struct {
		messageID string
		want      bool
	}{
		{"IDRAC.2.14.SYS011", true},
		{"IDRAC.2.9.SYS011", true},
		{"IDRAC.1.5.SYS011", true},
		{"IDRAC.2.14.SYS010", false},
		{"Base.1.12.GeneralError", false},
		{"SYS011", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.messageID, func(t *testing.T) {
			assert.Equal(t, tt.want, isDellPendingSettingsMessageID(tt.messageID))
		})
	}
}

// biosWithSettingsRedirect is a Bios resource whose settings updates are staged via a separate
// pending resource, matching a real Dell iDRAC.
const biosWithSettingsRedirect = `{
	"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
	"Id": "Bios",
	"Attributes": {},
	"@Redfish.Settings": {
		"SettingsObject": {"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios/Settings"},
		"SupportedApplyTimes": ["OnReset"]
	}
}`

// dellJobMux serves Dell's own OEM job resource, which is where the code under test reads whether
// a job has started applying. started is the ActualRunningStartTime to report; status is the HTTP
// status to answer with (200 for a normal answer).
func dellJobMux(mux *http.ServeMux, id, jobState, started string, status int) {
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/Jobs/"+id, func(w http.ResponseWriter, r *http.Request) {
		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		_, _ = w.Write([]byte(`{"Id": "` + id + `", "JobType": "BIOSConfiguration", "JobState": "` + jobState +
			`", "ActualRunningStartTime": "` + started + `"}`))
	})
}

// jobServiceMux registers the standard Redfish JobService discovery chain (ServiceRoot ->
// JobService -> Jobs collection -> job), serving a single pending BIOS config job whose
// completion status is controlled via jobState/actualRunningStartTime, and recording DELETE
// calls into deleteCalls.
func jobServiceMux(mux *http.ServeMux, jobState, actualRunningStartTime string, deleteCalls *int) {
	dellJobMux(mux, "JID_1", jobState, actualRunningStartTime, http.StatusOK)
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Members": [{"@odata.id": "/redfish/v1/JobService/Jobs/JID_1"}], "Members@odata.count": 1}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{
				"Id": "JID_1",
				"@odata.id": "/redfish/v1/JobService/Jobs/JID_1",
				"JobState": "` + jobState + `",
				"Oem": {"Dell": {"JobType": "BIOSConfiguration", "ActualRunningStartTime": "` + actualRunningStartTime + `"}}
			}`))
		case http.MethodDelete:
			*deleteCalls++
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

// jobServiceMuxNameOnly is jobServiceMux but for a job with no Oem.Dell property at all -
// confirmed live on a PowerEdge R760xd2 iDRAC, whose JobService/Jobs collection returns pending
// BIOS config jobs with only Id/Name/JobState, nothing under Oem. name is the job's Name, used
// to check isBIOSConfigJob's fallback signal (dellBIOSConfigJobNamePrefix) independently of the
// Oem-based one jobServiceMux exercises.
func jobServiceMuxNameOnly(mux *http.ServeMux, name, jobState string, deleteCalls *int) {
	dellJobMux(mux, "JID_1", jobState, "", http.StatusOK)
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Members": [{"@odata.id": "/redfish/v1/JobService/Jobs/JID_1"}], "Members@odata.count": 1}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{
				"Id": "JID_1",
				"@odata.id": "/redfish/v1/JobService/Jobs/JID_1",
				"Name": "` + name + `",
				"JobState": "` + jobState + `"
			}`))
		case http.MethodDelete:
			*deleteCalls++
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

// TestSetBiosConfiguration_PendingSettingsConflictJobNameNotBIOSConfig guards against the
// Name-based fallback becoming too permissive: a job with neither a matching Oem.Dell.JobType
// nor a "ConfigBIOS:" Name prefix (e.g. a firmware update job) must not be mistaken for the
// blocking BIOS config job and canceled.
func TestSetBiosConfiguration_PendingSettingsConflictJobNameNotBIOSConfig(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"ExistingAttr": "Foo"}}`))
		case http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(dellPendingSettingsCommittedOlderRegistry))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	var deleteCalls int
	jobServiceMuxNameOnly(mux, "FWUpdate:Firmware.1-1", "Running", &deleteCalls)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	assert.ErrorContains(t, err, "no pending BIOS config job found to cancel")
	assert.Equal(t, 0, deleteCalls, "expected an unrelated job to never be deleted")
}

func TestSetBiosConfiguration_PendingSettingsConflictNoPendingJobFound(t *testing.T) {
	// An iDRAC that returns the conflict message ID but whose JobService reports no matching
	// job should surface an error, not silently drop the new attributes.
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"ExistingAttr": "Foo"}}`))
		case http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(dellPendingSettingsCommitted))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	// The job is already Completed, so cancelPendingBIOSConfigJob has nothing to cancel.
	var deleteCalls int
	jobServiceMux(mux, "Completed", "", &deleteCalls)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	assert.ErrorContains(t, err, "no pending BIOS config job found to cancel")
	assert.Equal(t, 0, deleteCalls)
}

func TestSetBiosConfiguration_PendingSettingsConflictJobAlreadyRunning(t *testing.T) {
	// A job that has already started applying (ActualRunningStartTime set) must not be
	// canceled - that would be far more disruptive than the conflict it's working around.
	// JobState is "Starting", not "Running", so this exercises the Oem-based
	// ActualRunningStartTime signal specifically, independent of the JobState-based guard
	// TestSetBiosConfiguration_PendingSettingsConflictJobStateRunning covers.
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"ExistingAttr": "Foo"}}`))
		case http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(dellPendingSettingsCommitted))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	var deleteCalls int
	jobServiceMux(mux, "Starting", "2026-09-10T08:05:49", &deleteCalls)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	assert.ErrorContains(t, err, "has already started applying")
	assert.Equal(t, 0, deleteCalls, "expected an already-running job to never be deleted")
}

// TestSetBiosConfiguration_PendingSettingsConflictJobStateNotCancelable is the standard-Redfish
// counterpart to TestSetBiosConfiguration_PendingSettingsConflictJobAlreadyRunning, covering a
// job identified only via its Name (see
// TestSetBiosConfiguration_PendingSettingsConflictRecovery_NoOemNameFallback) - where
// oem.Dell.ActualRunningStartTime is always empty, so it alone could never catch a job that's
// already started. dellJobCancelableStates is an allow-list rather than a deny-list of just
// "Running" specifically because Suspended and Interrupted both mean a job already started
// executing and merely paused (the Redfish spec's own wording: "expected to restart"), and
// Continue means the same thing mid-resume - a deny-list keyed on "Running" alone would let any
// of these through to the DELETE call.
func TestSetBiosConfiguration_PendingSettingsConflictJobStateNotCancelable(t *testing.T) {
	for _, jobState := range []string{"Running", "Suspended", "Interrupted", "Continue", "Stopping", "UserIntervention"} {
		t.Run(jobState, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
			mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
			mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
			mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(biosWithSettingsRedirect))
			})
			mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					_, _ = w.Write([]byte(`{"Attributes": {"ExistingAttr": "Foo"}}`))
				case http.MethodPatch:
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(dellPendingSettingsCommittedOlderRegistry))
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			})

			var deleteCalls int
			jobServiceMuxNameOnly(mux, "ConfigBIOS:BIOS.Setup.1-1", jobState, &deleteCalls)

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
			require.NoError(t, client.Open(context.Background()))
			defer client.Close(context.Background())

			err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
			assert.ErrorContains(t, err, "not known to be safe to delete")
			assert.Equal(t, 0, deleteCalls, "expected a "+jobState+" job to never be deleted, even with no Oem data")
		})
	}
}

// TestSetBiosConfiguration_SkipsStaleTerminalJob guards the early-loop skip: a stale job matching
// the BIOS config Name signal but already finished (canceled here, from an earlier, unrelated
// attempt) must not stop pendingBIOSConfigJob from finding the live job listed alongside it.
// Without skipping such jobs, the stale one sorted first would either fail the DELETE outright or
// succeed as a no-op while leaving the real job holding Bios/Settings locked.
func TestSetBiosConfiguration_SkipsStaleTerminalJob(t *testing.T) {
	var patchBodies []string
	mux := biosSettingsMux(t, `{"ExistingAttr": "Foo"}`, http.StatusNoContent, &patchBodies)

	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"Members": [
				{"@odata.id": "/redfish/v1/JobService/Jobs/JID_stale"},
				{"@odata.id": "/redfish/v1/JobService/Jobs/JID_live"}
			],
			"Members@odata.count": 2
		}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_stale", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method, "the stale, already-finished job must never be DELETEd")
		//nolint:misspell // "Cancelled" is the Redfish spec's actual JobState value
		_, _ = w.Write([]byte(`{
			"Id": "JID_stale", "@odata.id": "/redfish/v1/JobService/Jobs/JID_stale",
			"Name": "ConfigBIOS:BIOS.Setup.1-1", "JobState": "Cancelled"
		}`))
	})
	var deleteCalls int
	mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_live", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{
				"Id": "JID_live", "@odata.id": "/redfish/v1/JobService/Jobs/JID_live",
				"Name": "ConfigBIOS:BIOS.Setup.1-1", "JobState": "Scheduled"
			}`))
		case http.MethodDelete:
			deleteCalls++
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"}),
		"expected the stale finished job to be skipped and the live one canceled")
	assert.Equal(t, 1, deleteCalls, "expected exactly the live job to be deleted")
	require.Len(t, patchBodies, 1)
}

// TestFeatureSetterInheritsPendingJobHandling guards the property recoveringRedfishClient exists
// for: a feature setter that simply calls c.redfishwrapper.SetBiosConfiguration - with no
// awareness of the pending-job handling at all - gets it anyway, because the wrapper shadows the
// embedded client's method. SetHTTPBootURI is used as the stand-in; the assertion isn't about
// HTTP Boot specifically, it's that a new BIOS-backed setter cannot accidentally opt out by
// writing the obvious call.
//
// If this regresses (e.g. Conn.redfishwrapper is retyped back to *redfishwrapper.Client), the
// pending job is never canceled and the write is rejected by iDRAC instead, so this fails.
func TestFeatureSetterInheritsPendingJobHandling(t *testing.T) {
	var patchBodies []string
	// iDRAC rejects any write while a job is holding Bios/Settings, as a real one does.
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})

	var deleteCalls int
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"SecureBootPolicy": "Custom"}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			if deleteCalls == 0 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(dellPendingSettingsCommitted))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	jobServiceMux(mux, "Scheduled", "", &deleteCalls)

	client := newBiosTestClient(t, mux)

	ok, err := client.SetHTTPBootURI(context.Background(), "http://example.com/ipxe.efi")
	require.NoError(t, err, "expected the pending job to be handled, not the conflict propagated to the caller")
	assert.True(t, ok)

	assert.Equal(t, 1, deleteCalls, "expected the pending job to be deleted exactly once")
	require.Len(t, patchBodies, 1, "expected one PATCH, sent only after the job was canceled")
	assert.Contains(t, patchBodies[0], `"HttpDev1Uri":"http://example.com/ipxe.efi"`)
}

// newBiosTestClient opens a Dell client against mux over TLS, for the tests below that all share
// the same ServiceRoot/Systems plumbing.
func newBiosTestClient(t *testing.T, mux *http.ServeMux) *Conn {
	t.Helper()

	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	t.Cleanup(func() { client.Close(context.Background()) })

	return client
}

// biosSettingsMux serves a Bios resource with a settings redirect and no applied attributes, plus
// a Bios/Settings object whose staged attributes are pending (raw JSON object body), and records
// every PATCH body. patchStatus is what a PATCH answers with.
func biosSettingsMux(t *testing.T, pending string, patchStatus int, patchBodies *[]string) *http.ServeMux {
	t.Helper()

	return biosSettingsMuxWithApplied(t, `{}`, pending, patchStatus, patchBodies)
}

// biosSettingsMuxWithApplied is biosSettingsMux for a Bios resource that also reports applied
// attributes (raw JSON object body).
func biosSettingsMuxWithApplied(t *testing.T, applied, pending string, patchStatus int, patchBodies *[]string) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
			"Id": "Bios",
			"Attributes": ` + applied + `,
			"@Redfish.Settings": {
				"SettingsObject": {"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios/Settings"},
				"SupportedApplyTimes": ["OnReset"]
			}
		}`))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": ` + pending + `}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			*patchBodies = append(*patchBodies, string(body))
			w.WriteHeader(patchStatus)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return mux
}

// TestSetBiosConfiguration_NothingPendingWritesOnce guards the common case staying cheap and
// untouched: with nothing staged, exactly one PATCH goes out and the job service is never
// consulted (it isn't even served here, so any lookup would fail the test).
func TestSetBiosConfiguration_NothingPendingWritesOnce(t *testing.T) {
	var patchBodies []string
	mux := biosSettingsMux(t, `{}`, http.StatusNoContent, &patchBodies)

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"}))

	require.Len(t, patchBodies, 1, "expected a single PATCH and no recovery")
	assert.Contains(t, patchBodies[0], `"NewAttr":"Bar"`)
}

// TestSetBiosConfiguration_PendingAttributesCancelsAndMergesBeforeWriting is the look-before-
// write path: staged attributes mean a job is holding Bios/Settings, so the job is canceled and
// the merged set goes out as ONE PATCH - the write is never attempted into the lock, so there is
// no rejected PATCH to recover from. "Scheduled" is the JobState a sealed, not-yet-started BIOS
// config job really shows on iDRAC (confirmed live), which gofish's standard JobState set has no
// constant for.
func TestSetBiosConfiguration_PendingAttributesCancelsAndMergesBeforeWriting(t *testing.T) {
	for _, jobState := range []string{"Scheduled", "Scheduling", "New", "Pending", "Starting"} {
		t.Run(jobState, func(t *testing.T) {
			var patchBodies []string
			mux := biosSettingsMux(t, `{"ExistingAttr": "Foo", "ExistingIntAttr": 42}`, http.StatusNoContent, &patchBodies)

			var deleteCalls int
			jobServiceMuxNameOnly(mux, "ConfigBIOS:BIOS.Setup.1-1", jobState, &deleteCalls)

			client := newBiosTestClient(t, mux)

			require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"}))

			assert.Equal(t, 1, deleteCalls, "expected the pending job to be canceled exactly once")
			require.Len(t, patchBodies, 1, "expected exactly one PATCH: no write into the lock, no retry")
			assert.Contains(t, patchBodies[0], `"NewAttr":"Bar"`)
		})
	}
}

// TestSetBiosConfiguration_AlreadyStagedIsANoOp makes a repeated call idempotent: when every
// requested attribute is already staged at the requested value and a live job carries them, the
// job is neither canceled nor rewritten. Requested values are strings, staged ones keep their
// JSON type, so "42"/"true" must compare equal to 42/true.
func TestSetBiosConfiguration_AlreadyStagedIsANoOp(t *testing.T) {
	var patchBodies []string
	mux := biosSettingsMux(t, `{"ExistingAttr": "Foo", "IntAttr": 42, "BoolAttr": true, "Other": "x"}`, http.StatusNoContent, &patchBodies)

	var deleteCalls int
	jobServiceMuxNameOnly(mux, "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", &deleteCalls)

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{
		"ExistingAttr": "Foo", "IntAttr": "42", "BoolAttr": "true",
	}))

	assert.Equal(t, 0, deleteCalls, "expected an already-satisfied request to leave the job alone")
	assert.Empty(t, patchBodies, "expected no PATCH when everything requested is already staged")
}

// TestSetBiosConfiguration_PartiallyStagedStillMerges is the other side of the no-op: one
// attribute already staged but another not still has to cancel and rewrite.
func TestSetBiosConfiguration_PartiallyStagedStillMerges(t *testing.T) {
	var patchBodies []string
	mux := biosSettingsMux(t, `{"ExistingAttr": "Foo"}`, http.StatusNoContent, &patchBodies)

	var deleteCalls int
	jobServiceMuxNameOnly(mux, "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", &deleteCalls)

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{
		"ExistingAttr": "Foo", "NewAttr": "Bar",
	}))

	assert.Equal(t, 1, deleteCalls)
	require.Len(t, patchBodies, 1)
	assert.Contains(t, patchBodies[0], `"NewAttr":"Bar"`)
}

// TestSetBiosConfiguration_StagedWithoutJobWritesWithoutCanceling covers attributes staged with
// no live job behind them (e.g. staged without an apply time): there is nothing to cancel, so the
// merged set is simply written, which also creates the job that was missing.
func TestSetBiosConfiguration_StagedWithoutJobWritesWithoutCanceling(t *testing.T) {
	var patchBodies []string
	mux := biosSettingsMux(t, `{"ExistingAttr": "Foo"}`, http.StatusNoContent, &patchBodies)

	var deleteCalls int
	jobServiceMux(mux, "Completed", "", &deleteCalls)

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"}))

	assert.Equal(t, 0, deleteCalls, "expected nothing to cancel")
	require.Len(t, patchBodies, 1)
	assert.Contains(t, patchBodies[0], `"NewAttr":"Bar"`)
}

// TestSetBiosConfiguration_UnreadablePendingSettingsFallsThroughToWrite guards against the look
// ever making a write that used to work fail: a Bios resource with no @Redfish.Settings redirect
// can't be inspected, and the write must go ahead exactly as before.
func TestSetBiosConfiguration_UnreadablePendingSettingsFallsThroughToWrite(t *testing.T) {
	mux := http.NewServeMux()
	var patchBodies []string
	// With no @Redfish.Settings redirect, the attributes are PATCHed to the Bios resource itself.
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios", "Id": "Bios", "Attributes": {}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"}))
	require.Len(t, patchBodies, 1, "expected the write to go ahead despite the look finding no settings redirect")
	assert.Contains(t, patchBodies[0], `"NewAttr":"Bar"`)
}

func TestAlreadyHolds(t *testing.T) {
	current := biosAttributes{
		applied: map[string]any{"S": "Disabled", "A": "On", "OnlyApplied": json.Number("7")},
		pending: map[string]any{"S": "Enabled", "I": json.Number("42"), "F": json.Number("1.5"), "B": false, "Big": json.Number("9007199254740993")},
	}

	tests := []struct {
		name string
		want map[string]string
		ok   bool
	}{
		{"all staged and equal", map[string]string{"S": "Enabled", "I": "42", "F": "1.5", "B": "false"}, true},
		{"subset of staged", map[string]string{"S": "Enabled"}, true},
		{"different staged string", map[string]string{"S": "Disabled"}, false},
		{"staged wins over applied", map[string]string{"S": "Disabled"}, false},
		{"not staged but equal to applied", map[string]string{"A": "On"}, true},
		{"not staged, applied number equal", map[string]string{"OnlyApplied": "7"}, true},
		{"not staged and different from applied", map[string]string{"A": "Off"}, false},
		{"mix of staged and applied", map[string]string{"S": "Enabled", "A": "On"}, true},
		{"different number", map[string]string{"I": "43"}, false},
		{"absent everywhere", map[string]string{"Missing": "x"}, false},
		{"bool spelled differently", map[string]string{"B": "False"}, false},
		{"large integer exact", map[string]string{"Big": "9007199254740993"}, true},
		{"large integer off by one", map[string]string{"Big": "9007199254740992"}, false},
		{"enum casing differs", map[string]string{"S": "enabled"}, false},
		{"empty request", map[string]string{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.ok, alreadyHolds(current, tt.want))
		})
	}
}

// TestSetBiosConfiguration_ValueAlreadyAppliedLeavesThePendingJobAlone: an attribute whose
// requested value is the one already applied, and that nothing is staged for, is satisfied. With a
// job pending for something else, repeating such a call must not delete and recreate that job.
func TestSetBiosConfiguration_ValueAlreadyAppliedLeavesThePendingJobAlone(t *testing.T) {
	var patchBodies, deleted []string
	mux := biosSettingsMuxWithApplied(t, `{"AppliedAttr": "On"}`, `{"StagedAttr": "Foo"}`, http.StatusNoContent, &patchBodies)
	jobsMux(mux, []testJob{{"JID_1", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""}}, status(http.StatusOK), &deleted)

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"AppliedAttr": "On"}))
	assert.Empty(t, deleted, "expected the pending job to be left alone")
	assert.Empty(t, patchBodies)
}

// TestSetBiosConfiguration_ValueDifferingFromAppliedStillMerges is the other side: when the
// requested value differs from the applied one, the job is replaced.
func TestSetBiosConfiguration_ValueDifferingFromAppliedStillMerges(t *testing.T) {
	var patchBodies, deleted []string
	mux := biosSettingsMuxWithApplied(t, `{"AppliedAttr": "On"}`, `{"StagedAttr": "Foo"}`, http.StatusNoContent, &patchBodies)
	jobsMux(mux, []testJob{{"JID_1", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""}}, status(http.StatusOK), &deleted)

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"AppliedAttr": "Off"}))
	assert.Equal(t, []string{"JID_1"}, deleted)
	require.Len(t, patchBodies, 1)
	assert.Contains(t, patchBodies[0], `"AppliedAttr":"Off"`)
}

// status returns a delete-status function that answers code for every job.
func status(code int) func(string) int { return func(string) int { return code } }

type testJob struct{ id, name, state, started string }

// jobsMux serves the JobService discovery chain with the given jobs. A DELETE of a job is
// recorded in deleted and answers deleteStatus(id).
func jobsMux(mux *http.ServeMux, jobs []testJob, deleteStatus func(id string) int, deleted *[]string) {
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
	})

	// Like iDRAC, answer $expand with the members inline.
	var links, expanded []string
	for _, j := range jobs {
		links = append(links, `{"@odata.id": "/redfish/v1/JobService/Jobs/`+j.id+`"}`)
		expanded = append(expanded, `{"Id": "`+j.id+`", "@odata.id": "/redfish/v1/JobService/Jobs/`+j.id+
			`", "Name": "`+j.name+`", "JobState": "`+j.state+`"}`)
	}
	mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
		members := links
		if r.URL.Query().Has("$expand") {
			members = expanded
		}
		_, _ = w.Write([]byte(`{"Members": [` + strings.Join(members, ",") + `], "Members@odata.count": ` + strconv.Itoa(len(jobs)) + `}`))
	})

	for _, j := range jobs {
		dellJobMux(mux, j.id, j.state, j.started, http.StatusOK)
		mux.HandleFunc("/redfish/v1/JobService/Jobs/"+j.id, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				_, _ = w.Write([]byte(`{"Id": "` + j.id + `", "@odata.id": "/redfish/v1/JobService/Jobs/` + j.id +
					`", "Name": "` + j.name + `", "JobState": "` + j.state + `"}`))
			case http.MethodDelete:
				*deleted = append(*deleted, j.id)
				w.WriteHeader(deleteStatus(j.id))
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		})
	}
}

// stagedSettingsMux is biosSettingsMux for a BMC where deleting the pending job discards what it
// staged, as iDRAC does: once any job has been deleted, Bios/Settings reads back empty. patchStatus
// answers each PATCH in order, the last one repeating.
func stagedSettingsMux(t *testing.T, staged string, deleted *[]string, patchStatus []int, patchBodies *[]string) *http.ServeMux {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if len(*deleted) > 0 {
				_, _ = w.Write([]byte(`{"Attributes": {}}`))
				return
			}
			_, _ = w.Write([]byte(`{"Attributes": ` + staged + `}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			*patchBodies = append(*patchBodies, string(body))
			status := patchStatus[len(patchStatus)-1]
			if len(*patchBodies) <= len(patchStatus) {
				status = patchStatus[len(*patchBodies)-1]
			}
			w.WriteHeader(status)
			if status >= 400 {
				_, _ = w.Write([]byte(`{"error":{"code":"Base.1.18.GeneralError","message":"rejected","@Message.ExtendedInfo":[{"MessageId":"IDRAC.2.9.SYS402","Message":"rejected"}]}}`))
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return mux
}

// TestSetBiosConfiguration_FailedWriteRestoresStagedAttributes: once the pending job is deleted,
// its staged attributes exist nowhere but in memory. If the merged write is then rejected, they
// are staged again, so the earlier caller's change is not lost along with the failed one.
func TestSetBiosConfiguration_FailedWriteRestoresStagedAttributes(t *testing.T) {
	var deleted, patchBodies []string
	mux := stagedSettingsMux(t, `{"ExistingAttr": "Foo"}`, &deleted, []int{http.StatusBadRequest, http.StatusNoContent}, &patchBodies)
	jobsMux(mux, []testJob{{"JID_1", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""}}, status(http.StatusOK), &deleted)

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "were re-staged")
	assert.ErrorContains(t, err, "ExistingAttr")

	assert.Equal(t, []string{"JID_1"}, deleted)
	require.Len(t, patchBodies, 2, "expected the rejected merged write, then the re-staging")
	assert.Contains(t, patchBodies[0], `"NewAttr":"Bar"`)
	assert.Contains(t, patchBodies[1], `"ExistingAttr":"Foo"`)
	assert.NotContains(t, patchBodies[1], "NewAttr", "expected only the previously staged attributes to be restored")
}

// TestSetBiosConfiguration_FailedWriteAndFailedRestoreSaysSo: when re-staging fails as well, the
// error must say that the earlier attributes are gone, and which.
func TestSetBiosConfiguration_FailedWriteAndFailedRestoreSaysSo(t *testing.T) {
	var deleted, patchBodies []string
	mux := stagedSettingsMux(t, `{"ExistingAttr": "Foo"}`, &deleted, []int{http.StatusBadRequest}, &patchBodies)
	jobsMux(mux, []testJob{{"JID_1", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""}}, status(http.StatusOK), &deleted)

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "ExistingAttr")
	assert.ErrorContains(t, err, "could not be re-staged either")
	assert.Len(t, patchBodies, 2)
}

// TestSetBiosConfiguration_FailedDeleteWritesNothing: if the job cannot be deleted, nothing is
// written, and the error names the job.
func TestSetBiosConfiguration_FailedDeleteWritesNothing(t *testing.T) {
	var deleted, patchBodies []string
	mux := stagedSettingsMux(t, `{"ExistingAttr": "Foo"}`, &[]string{}, []int{http.StatusNoContent}, &patchBodies)
	jobsMux(mux, []testJob{{"JID_1", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""}}, status(http.StatusInternalServerError), &deleted)

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "deleting job JID_1")
	assert.Empty(t, patchBodies)
}

// TestSetBiosConfiguration_DeletesEveryLiveBIOSJobAndSkipsFinishedOnes: finished jobs (including
// Dell's own Failed and CompletedWithErrors states) are ignored, and every live BIOS config job is
// deleted, whatever order the job service lists them in.
func TestSetBiosConfiguration_DeletesEveryLiveBIOSJobAndSkipsFinishedOnes(t *testing.T) {
	var deleted, patchBodies []string
	mux := stagedSettingsMux(t, `{"ExistingAttr": "Foo"}`, &deleted, []int{http.StatusNoContent}, &patchBodies)
	jobsMux(mux, []testJob{
		{"JID_failed", "ConfigBIOS:BIOS.Setup.1-1", "Failed", ""},
		{"JID_errors", "ConfigBIOS:BIOS.Setup.1-1", "CompletedWithErrors", ""},
		{"JID_live1", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""},
		{"JID_other", "Firmware Update: BIOS", "Scheduled", ""},
		{"JID_live2", "ConfigBIOS:BIOS.Setup.1-1", "Starting", ""},
	}, status(http.StatusOK), &deleted)

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"}))
	assert.ElementsMatch(t, []string{"JID_live1", "JID_live2"}, deleted)
	require.Len(t, patchBodies, 1)
}

// TestSetBiosConfiguration_RefusesWhenAnyBIOSJobHasStarted: a started job anywhere in the list
// stops the whole operation before anything is deleted.
func TestSetBiosConfiguration_RefusesWhenAnyBIOSJobHasStarted(t *testing.T) {
	var deleted, patchBodies []string
	mux := stagedSettingsMux(t, `{"ExistingAttr": "Foo"}`, &[]string{}, []int{http.StatusNoContent}, &patchBodies)
	jobsMux(mux, []testJob{
		{"JID_live1", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""},
		{"JID_running", "ConfigBIOS:BIOS.Setup.1-1", "Running", ""},
	}, status(http.StatusOK), &deleted)

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	assert.ErrorContains(t, err, "JID_running")
	assert.Empty(t, deleted, "expected nothing to be deleted while a job is running")
	assert.Empty(t, patchBodies)
}

// TestSetBiosConfiguration_LargeStagedIntegerSurvivesTheMerge: a staged number above 2^53 must be
// resubmitted exactly; decoding it into a float64 would silently change it.
func TestSetBiosConfiguration_LargeStagedIntegerSurvivesTheMerge(t *testing.T) {
	var deleted, patchBodies []string
	mux := stagedSettingsMux(t, `{"BigAttr": 9007199254740993}`, &deleted, []int{http.StatusNoContent}, &patchBodies)
	jobsMux(mux, []testJob{{"JID_1", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""}}, status(http.StatusOK), &deleted)

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"}))
	require.Len(t, patchBodies, 1)
	assert.Contains(t, patchBodies[0], `"BigAttr":9007199254740993`)
}

// TestSetBiosConfiguration_TransientReadErrorIsReturned: a failing read of the pending settings
// is reported as such instead of being papered over by a write that would then hit a confusing
// conflict, and nothing is written.
func TestSetBiosConfiguration_TransientReadErrorIsReturned(t *testing.T) {
	var patched bool
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			patched = true
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	assert.ErrorContains(t, err, "reading pending BIOS settings")
	assert.False(t, patched)
}

// TestSetBiosConfiguration_JobListFailureIsReported: attributes are pending but the jobs cannot
// be listed, so there is nothing safe to do and nothing is written.
func TestSetBiosConfiguration_JobListFailureIsReported(t *testing.T) {
	var patchBodies []string
	mux := biosSettingsMux(t, `{"ExistingAttr": "Foo"}`, http.StatusNoContent, &patchBodies)
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	assert.ErrorContains(t, err, "listing jobs")
	assert.Empty(t, patchBodies)
}

// TestSetBiosConfiguration_RefusesWhenDellJobResourceShowsItStarted: on some iDRAC generations the
// JobService resource has no Oem block, so whether a job started applying is only visible in
// Dell's own job resource. A job in a not-yet-started state that has an ActualRunningStartTime
// there must not be deleted.
func TestSetBiosConfiguration_RefusesWhenDellJobResourceShowsItStarted(t *testing.T) {
	var deleted, patchBodies []string
	mux := stagedSettingsMux(t, `{"ExistingAttr": "Foo"}`, &[]string{}, []int{http.StatusNoContent}, &patchBodies)
	jobsMux(mux, []testJob{{"JID_1", "ConfigBIOS:BIOS.Setup.1-1", "Starting", "2026-09-10T08:05:49"}}, status(http.StatusOK), &deleted)

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	assert.ErrorContains(t, err, "already started applying")
	assert.Empty(t, deleted)
	assert.Empty(t, patchBodies)
}

// TestSetBiosConfiguration_RefusesWhenStartedCannotBeConfirmed: if Dell's job resource cannot be
// read, nothing proves the job has not started, so it is not deleted.
func TestSetBiosConfiguration_RefusesWhenStartedCannotBeConfirmed(t *testing.T) {
	var patchBodies []string
	mux := stagedSettingsMux(t, `{"ExistingAttr": "Foo"}`, &[]string{}, []int{http.StatusNoContent}, &patchBodies)
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Members": [{"@odata.id": "/redfish/v1/JobService/Jobs/JID_1"}], "Members@odata.count": 1}`))
	})
	var deleted bool
	mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleted = true
		}
		_, _ = w.Write([]byte(`{"Id": "JID_1", "@odata.id": "/redfish/v1/JobService/Jobs/JID_1", "Name": "ConfigBIOS:BIOS.Setup.1-1", "JobState": "Scheduled"}`))
	})
	dellJobMux(mux, "JID_1", "Scheduled", "", http.StatusNotFound)

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	assert.ErrorContains(t, err, "checking whether BIOS config job JID_1 has started")
	assert.False(t, deleted)
	assert.Empty(t, patchBodies)
}

// TestSetBiosConfiguration_FailedSecondDeleteRestoresStagedAttributes: when several jobs are
// pending and a later delete fails after an earlier one succeeded, what the deleted job staged is
// gone, so it is staged again and the error says so.
func TestSetBiosConfiguration_FailedSecondDeleteRestoresStagedAttributes(t *testing.T) {
	var deleted, patchBodies []string
	mux := stagedSettingsMux(t, `{"ExistingAttr": "Foo"}`, &deleted, []int{http.StatusNoContent}, &patchBodies)
	jobsMux(mux, []testJob{
		{"JID_a", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""},
		{"JID_b", "ConfigBIOS:BIOS.Setup.1-1", "Scheduled", ""},
	}, func(id string) int {
		if id == "JID_b" {
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}, &deleted)

	client := newBiosTestClient(t, mux)

	err := client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.Error(t, err)
	assert.ErrorContains(t, err, "deleting job JID_b")
	assert.ErrorContains(t, err, "were re-staged")
	require.Len(t, patchBodies, 1, "expected only the re-staging write, not the merged one")
	assert.Contains(t, patchBodies[0], `"ExistingAttr":"Foo"`)
}

// TestSetBiosConfiguration_SettingsObjectIsTheBiosResourceItself: when writes go to the Bios
// resource itself, its attributes are the applied ones, not pending ones, so they say nothing
// about a pending job. The write goes ahead without listing jobs.
func TestSetBiosConfiguration_SettingsObjectIsTheBiosResourceItself(t *testing.T) {
	var patchBodies []string
	var jobsListed bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{
				"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
				"Id": "Bios",
				"Attributes": {"AppliedAttr": "Foo"},
				"@Redfish.Settings": {"SettingsObject": {"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios"}}
			}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		jobsListed = true
		w.WriteHeader(http.StatusInternalServerError)
	})

	client := newBiosTestClient(t, mux)

	require.NoError(t, client.SetBiosConfiguration(context.Background(), map[string]string{"AppliedAttr": "Bar"}))
	require.Len(t, patchBodies, 1)
	assert.Contains(t, patchBodies[0], `"AppliedAttr":"Bar"`)
	assert.False(t, jobsListed, "expected no job listing when nothing can be pending")
}
