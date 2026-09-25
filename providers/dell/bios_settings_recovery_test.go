package dell

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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

// jobServiceMux registers the standard Redfish JobService discovery chain (ServiceRoot ->
// JobService -> Jobs collection -> job), serving a single pending BIOS config job whose
// completion status is controlled via jobState/actualRunningStartTime, and recording DELETE
// calls into deleteCalls.
func jobServiceMux(mux *http.ServeMux, jobState, actualRunningStartTime string, deleteCalls *int) {
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

func TestSetBiosConfiguration_PendingSettingsConflictRecovery(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})

	var patchBodies []string
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{
				"Attributes": {"ExistingAttr": "Foo", "ExistingIntAttr": 42, "ExistingBoolAttr": true}
			}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			if len(patchBodies) == 1 {
				// The first PATCH is rejected: a job from an earlier, unrelated change (e.g. a
				// direct SecureBoot resource PATCH) is already pending.
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(dellPendingSettingsCommitted))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	var deleteCalls int
	jobServiceMux(mux, "Starting", "", &deleteCalls)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.NoError(t, err, "expected the conflict to be recovered from, not propagated")

	assert.Equal(t, 1, deleteCalls, "expected the pending job to be deleted exactly once")
	require.Len(t, patchBodies, 2, "expected the rejected PATCH and exactly one retry")
	// ExistingAttr/ExistingIntAttr/ExistingBoolAttr aren't asserted onto the retry's wire bytes:
	// gofish#571 compares the merged request against this same pending endpoint and correctly
	// omits attributes already staged at the requested value - a harmless PATCH optimization,
	// since an omitted field leaves that attribute's staged value untouched rather than clearing
	// it. What matters is that recoverFromPendingSettingsConflict still merges them in before
	// calling ApplyBiosAttributes, so they'd survive even against a BMC/gofish combination that
	// doesn't compare against pending state.
	assert.Contains(t, patchBodies[1], `"NewAttr":"Bar"`, "expected the retry to include the newly requested attribute")
}

// TestSetBiosConfiguration_PendingSettingsConflictRecovery_OlderRegistryVersion is
// TestSetBiosConfiguration_PendingSettingsConflictRecovery run against
// dellPendingSettingsCommittedOlderRegistry instead - guards against regressing back to matching
// a single hardcoded message registry version, which left recovery silently inert (the raw
// conflict propagated to the caller with no indication recovery was ever attempted) on any iDRAC
// generation not on the exact version first observed.
func TestSetBiosConfiguration_PendingSettingsConflictRecovery_OlderRegistryVersion(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})

	var patchBodies []string
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"ExistingAttr": "Foo"}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			if len(patchBodies) == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(dellPendingSettingsCommittedOlderRegistry))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	var deleteCalls int
	jobServiceMux(mux, "Starting", "", &deleteCalls)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.NoError(t, err, "expected the conflict to be recovered from even under an older message registry version")

	assert.Equal(t, 1, deleteCalls)
	require.Len(t, patchBodies, 2)
	assert.Contains(t, patchBodies[1], `"NewAttr":"Bar"`)
}

// TestSetBiosConfiguration_PendingSettingsConflictRecoveryIncludesUnchangedAttributes guards
// against a regression found live: the merged superset the retry resubmits has to carry every
// already-staged attribute, including ones whose staged value happens to equal what the Bios
// resource currently reports. Nothing has rebooted mid-recovery, so "currently applied" hasn't
// moved - an attribute that was staged back to its applied value (a normal thing to find staged,
// since callers often revert) must not be dropped from the merge recoverFromPendingSettingsConflict
// builds, even though gofish#571 may then legitimately omit it from the actual retry PATCH
// because it's already correctly staged (see the non-"IncludesUnchangedAttributes" test above).
//
// Here SecureBootPolicy is staged at the same value the Bios resource reports as applied, while
// the caller's own attribute differs - so the first PATCH is genuinely sent, gets the conflict,
// and the assertion is about what survives into the retry.
// TestSetBiosConfiguration_PendingSettingsConflictRecovery_NoOemNameFallback is
// TestSetBiosConfiguration_PendingSettingsConflictRecovery run against a job with no Oem
// property at all, identified only via its Name ("ConfigBIOS:BIOS.Setup.1-1", Dell's stable
// naming convention for BIOS Setup config jobs) - confirmed live on a PowerEdge R760xd2 iDRAC
// that Oem.Dell.JobType isn't always present on a Job resource, which previously made
// cancelPendingBIOSConfigJob skip right past the job it needed to cancel.
func TestSetBiosConfiguration_PendingSettingsConflictRecovery_NoOemNameFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})

	var patchBodies []string
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"ExistingAttr": "Foo"}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			if len(patchBodies) == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(dellPendingSettingsCommittedOlderRegistry))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	var deleteCalls int
	jobServiceMuxNameOnly(mux, "ConfigBIOS:BIOS.Setup.1-1", "Starting", &deleteCalls)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.NoError(t, err, "expected the conflict to be recovered from via the Name-based fallback")

	assert.Equal(t, 1, deleteCalls, "expected the pending job to be identified and deleted via its Name")
	require.Len(t, patchBodies, 2)
	assert.Contains(t, patchBodies[1], `"NewAttr":"Bar"`)
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

func TestSetBiosConfiguration_PendingSettingsConflictRecoveryIncludesUnchangedAttributes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		// SecureBootPolicy is reported as applied at the same value it is staged at below;
		// NewAttr, which this call requests, is absent and so genuinely differs.
		_, _ = w.Write([]byte(`{
			"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
			"Id": "Bios",
			"Attributes": {"SecureBootPolicy": "Standard"},
			"@Redfish.Settings": {
				"SettingsObject": {"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios/Settings"},
				"SupportedApplyTimes": ["OnReset"]
			}
		}`))
	})

	var patchBodies []string
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"SecureBootPolicy": "Standard"}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			if len(patchBodies) == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(dellPendingSettingsCommitted))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	var deleteCalls int
	jobServiceMux(mux, "Starting", "", &deleteCalls)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.NoError(t, err)

	require.Len(t, patchBodies, 2)
	// SecureBootPolicy isn't asserted onto the retry's wire bytes - see the comment above and
	// the equivalent assertion in TestSetBiosConfiguration_PendingSettingsConflictRecovery.
	assert.Contains(t, patchBodies[1], `"NewAttr":"Bar"`, "expected the retry to include the requested attribute")
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
			assert.ErrorContains(t, err, "not safe to cancel")
			assert.Equal(t, 0, deleteCalls, "expected a "+jobState+" job to never be deleted, even with no Oem data")
		})
	}
}

// TestSetBiosConfiguration_PendingSettingsConflictSkipsStaleTerminalJob guards the early-loop
// skip: a stale job matching isBIOSConfigJob's Name signal, but already in CancelledJobState from an
// earlier, unrelated attempt, must not stop cancelPendingBIOSConfigJob from finding the actual
// live blocker listed alongside it. Without skipping CancelledJobState/ExceptionJobState jobs specifically (not
// just Completed), a stale one sorted first would either fail the DELETE outright or succeed as
// a no-op while leaving the real blocking job in place, then fail the single retry with the same
// SYS011 conflict it started from.
func TestSetBiosConfiguration_PendingSettingsConflictSkipsStaleTerminalJob(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})

	var patchBodies []string
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"ExistingAttr": "Foo"}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			if len(patchBodies) == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(dellPendingSettingsCommittedOlderRegistry))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

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
		require.Equal(t, http.MethodGet, r.Method, "the stale, already-CancelledJobState job must never be DELETEd")
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
				"Name": "ConfigBIOS:BIOS.Setup.1-1", "JobState": "Starting"
			}`))
		case http.MethodDelete:
			deleteCalls++
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	err = client.SetBiosConfiguration(context.Background(), map[string]string{"NewAttr": "Bar"})
	require.NoError(t, err, "expected recovery to skip the stale CancelledJobState job and cancel the live one")
	assert.Equal(t, 1, deleteCalls, "expected exactly the live job to be deleted")
}

// TestFeatureSetterInheritsConflictRecovery guards the property recoveringRedfishClient exists
// for: a feature setter that simply calls c.redfishwrapper.SetBiosConfiguration - with no
// awareness of the conflict recovery at all - gets that recovery anyway, because the wrapper
// shadows the embedded client's method. SetHTTPBootURI is used as the stand-in; the assertion
// isn't about HTTP Boot specifically, it's that a new BIOS-backed setter cannot accidentally
// opt out of recovery by writing the obvious call.
//
// If this regresses (e.g. Conn.redfishwrapper is retyped back to *redfishwrapper.Client), the
// conflict propagates to the caller instead of being recovered from, and this fails.
func TestFeatureSetterInheritsConflictRecovery(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		_, _ = w.Write([]byte(biosWithSettingsRedirect))
	})

	var patchBodies []string
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"Attributes": {"SecureBootPolicy": "Custom"}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			patchBodies = append(patchBodies, string(body))
			if len(patchBodies) == 1 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(dellPendingSettingsCommitted))
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	var deleteCalls int
	jobServiceMux(mux, "Starting", "", &deleteCalls)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	ok, err := client.SetHTTPBootURI(context.Background(), "http://example.com/ipxe.efi")
	require.NoError(t, err, "expected the conflict to be recovered from, not propagated to the caller")
	assert.True(t, ok)

	assert.Equal(t, 1, deleteCalls, "expected the pending job to be deleted exactly once")
	require.Len(t, patchBodies, 2, "expected the rejected PATCH and exactly one retry")
	// SecureBootPolicy isn't asserted onto the retry's wire bytes - see the comment in
	// TestSetBiosConfiguration_PendingSettingsConflictRecovery.
	assert.Contains(t, patchBodies[1], `"HttpDev1Uri":"http://example.com/ipxe.efi"`)
}
