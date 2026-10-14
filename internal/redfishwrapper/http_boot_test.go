package redfishwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newDellClientForBootTest is like newDellSecureBootClient but leaves the
// System.Embedded.1 resource unregistered so the caller can install its own
// GET/PATCH handler for it.
func newDellClientForBootTest(t *testing.T, mux *http.ServeMux) *Client {
	t.Helper()

	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "dell/systems.json"))

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	ctx := context.Background()
	require.NoError(t, client.Open(ctx))
	t.Cleanup(func() { _ = client.Close(ctx) })

	return client
}

func TestSetHTTPBootURI(t *testing.T) {
	var patchAttempts int
	var patchBody map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(mustReadFile(t, "dell/system.embedded.1.json"))
		case http.MethodPatch:
			patchAttempts++
			require.NoError(t, json.NewDecoder(r.Body).Decode(&patchBody))
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	client := newDellClientForBootTest(t, mux)

	ok, err := client.SetHTTPBootURI(context.Background(), "http://boot.example.com/ipxe/boot.efi")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 1, patchAttempts, "expected SetHTTPBootURI to PATCH the System resource once")

	boot, ok := patchBody["Boot"].(map[string]any)
	require.True(t, ok, "expected PATCH body to contain a Boot object")
	assert.Equal(t, "http://boot.example.com/ipxe/boot.efi", boot["HttpBootUri"])
	assert.Len(t, boot, 1, "expected only HttpBootUri to be sent, not other Boot fields such as BootSourceOverrideTarget")
}
