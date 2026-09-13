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
	assert.Contains(t, patchBodies[1], `"ExistingAttr":"Foo"`, "expected the retry to preserve the already-pending string attribute")
	// Numeric and boolean pending attributes must round-trip with their original JSON type, not
	// get coerced to strings - Dell's BIOS attribute registry is strict about attribute types.
	assert.Contains(t, patchBodies[1], `"ExistingIntAttr":42`, "expected the retry to preserve the already-pending attribute's original numeric type, not stringify it")
	assert.Contains(t, patchBodies[1], `"ExistingBoolAttr":true`, "expected the retry to preserve the already-pending attribute's original boolean type, not stringify it")
	assert.Contains(t, patchBodies[1], `"NewAttr":"Bar"`, "expected the retry to include the newly requested attribute")
}

// TestSetBiosConfiguration_PendingSettingsConflictRecoveryIncludesUnchangedAttributes guards
// against a regression found live: the merged superset the retry resubmits has to carry every
// already-staged attribute, including ones whose staged value happens to equal what the Bios
// resource currently reports. Nothing has rebooted mid-recovery, so "currently applied" hasn't
// moved - and an attribute that was staged back to its applied value (a normal thing to find
// staged, since callers often revert) would be dropped from the retry if the merge went through
// a path that diffs against applied state. Recovery uses ApplyBiosAttributesExact so every
// merged attribute is resubmitted regardless.
//
// Here SecureBootPolicy is staged at the same value the Bios resource reports as applied, while
// the caller's own attribute differs - so the first PATCH is genuinely sent, gets the conflict,
// and the assertion is about what survives into the retry.
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
	assert.Contains(t, patchBodies[1], `"SecureBootPolicy":"Standard"`, "expected the retry to preserve the already-staged attribute even though its value equals the applied one")
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
	jobServiceMux(mux, "Running", "2026-09-10T08:05:49", &deleteCalls)

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
	assert.Contains(t, patchBodies[1], `"SecureBootPolicy":"Custom"`, "expected the retry to preserve the already-pending attribute")
	assert.Contains(t, patchBodies[1], `"HttpDev1Uri":"http://example.com/ipxe.efi"`)
}
