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

// TestBootDeviceSet_DirectWrite verifies BootDeviceSet against a fixture with
// no @Redfish.Settings (iDRAC 7.x-9.x style): the PATCH must go directly to
// the main ComputerSystem resource and succeed in a single request.
func TestBootDeviceSet_DirectWrite(t *testing.T) {
	var patched bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			patched = true
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), "Pxe")
			w.WriteHeader(http.StatusOK)
			return
		}
		endpointFunc("/systems_embedded.1.json")(w, r)
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	ok, err := client.BootDeviceSet(context.Background(), "pxe", false, true)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, patched, "expected a PATCH to the main ComputerSystem resource")
}

// TestBootDeviceSet_SettingsFallback verifies BootDeviceSet against a fixture
// advertising @Redfish.Settings where the main resource rejects Boot writes
// (iDRAC10 style): the PATCH to the main resource must fail, and BootDeviceSet
// must still succeed by falling back to the Settings resource.
func TestBootDeviceSet_SettingsFallback(t *testing.T) {
	var settingsPatched bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"code":"Base.1.0.GeneralError","message":"Boot is a read-only property"}}`))
			return
		}
		endpointFunc("/systems_embedded_idrac10.1.json")(w, r)
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Settings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			settingsPatched = true
			body, _ := io.ReadAll(r.Body)
			assert.Contains(t, string(body), "Pxe")
			w.WriteHeader(http.StatusOK)
			return
		}
		endpointFunc("/systems_embedded_idrac10.1.json")(w, r)
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	ok, err := client.BootDeviceSet(context.Background(), "pxe", false, true)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, settingsPatched, "expected a fallback PATCH to the Settings resource")
}
