package redfishwrapper

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystemBootDeviceSet_DetectsSupermicroSilentModeDowngrade guards against
// a real bug seen on a Supermicro AS-1015CS-TNR-EU (AMI-based Redfish,
// RedfishVersion 1.22.2): SystemBootDeviceSet("pxe", false, true) requests a
// one-time UEFI PXE override (BootSourceOverrideTarget=Pxe,
// BootSourceOverrideEnabled=Once, BootSourceOverrideMode=UEFI in a single
// PATCH, exactly as the function under test builds it), the BMC responds
// with 200 OK, but a subsequent GET shows BootSourceOverrideMode reverted to
// "Legacy" (and Enabled to "Disabled") - only BootSourceOverrideTarget stuck
// as requested.
//
// This matters because this system also has CSMSupport (Legacy/CSM boot)
// disabled in its BIOS - it's UEFI-only - so a boot override that silently
// downgrades to Legacy mode is not just wrong bookkeeping, it can never
// actually result in a boot from the requested device. SystemBootDeviceSet
// now re-fetches and verifies the write instead of trusting the PATCH's 200
// OK, so callers (e.g. Tinkerbell's rufio controller) get ok=false and a
// descriptive error instead of a false ok=true.
//
// The fixtures here are real captures from the affected hardware - see
// fixtures/supermicro/README.md.
func TestSystemBootDeviceSet_DetectsSupermicroSilentModeDowngrade(t *testing.T) {
	var gotPatchBody []byte

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/supermicro/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "/supermicro/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// The real BMC never reflects the requested change - every GET,
			// before or after the PATCH, returns this same Legacy/Disabled
			// state.
			w.Write(mustReadFile(t, "/supermicro/system.1.json"))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("reading PATCH body: %s", err)
				return
			}
			gotPatchBody = body
			w.Write(mustReadFile(t, "/supermicro/patch-success.json"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(ctx))

	// Request a one-time UEFI PXE boot override - this is exactly what
	// Tinkerbell's rufio controller sends for a Hardware's bootDevice
	// action with device: "pxe", efiBoot: true (no "persistent" field, so
	// setPersistent defaults to false).
	ok, err := client.SystemBootDeviceSet(ctx, "pxe", false, true)
	assert.Contains(t, string(gotPatchBody), `"BootSourceOverrideMode":"UEFI"`,
		"the PATCH we send does correctly ask for UEFI mode")

	require.Error(t, err, "the BMC's 200 OK alone must not be enough to report success")
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "was not applied",
		"error should say the write wasn't actually applied, not just surface a generic failure")
}

// TestSystemBootDeviceSet_HappyPath is the counterpart to
// TestSystemBootDeviceSet_DetectsSupermicroSilentModeDowngrade: a BMC that
// actually honors the request should still report ok=true, not a false
// negative from the new verification step.
func TestSystemBootDeviceSet_HappyPath(t *testing.T) {
	var patched bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/supermicro/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "/supermicro/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if patched {
				w.Write(mustReadFile(t, "/supermicro/system.1.after-correct-set.json"))
			} else {
				w.Write(mustReadFile(t, "/supermicro/system.1.json"))
			}
		case http.MethodPatch:
			patched = true
			w.Write(mustReadFile(t, "/supermicro/patch-success.json"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(ctx))

	ok, err := client.SystemBootDeviceSet(ctx, "pxe", false, true)
	require.NoError(t, err)
	assert.True(t, ok)
}
