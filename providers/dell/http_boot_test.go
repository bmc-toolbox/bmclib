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

// TestSetHTTPBootURI verifies SetHTTPBootURI PATCHes the HttpDev1Uri BIOS Setup attribute (device
// slot 1, see httpBootDeviceIndex) rather than the generic Redfish ComputerSystem.Boot.HttpBootUri
// property, which Dell's iDRAC doesn't populate at all.
func TestSetHTTPBootURI(t *testing.T) {
	var patched bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Bios","Id":"Bios","Attributes":{},"@Redfish.Settings":{"SettingsObject":{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Bios"}}}`))
		case http.MethodPatch:
			patched = true
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Contains(t, string(body), `"HttpDev1Uri":"http://example.com/ipxe.efi"`)
			w.WriteHeader(http.StatusNoContent)
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

	ok, err := client.SetHTTPBootURI(context.Background(), "http://example.com/ipxe.efi")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, patched, "expected a PATCH to the Bios resource setting HttpDev1Uri")
}
