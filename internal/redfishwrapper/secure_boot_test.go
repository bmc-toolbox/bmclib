package redfishwrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDellSecureBootClient(t *testing.T, mux *http.ServeMux) *Client {
	t.Helper()

	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "dell/system.embedded.1.json"))

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

func TestGetSecureBoot(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/SecureBoot", endpointFunc(t, "dell/secureboot.json"))
	client := newDellSecureBootClient(t, mux)

	enabled, err := client.GetSecureBoot(context.Background())
	assert.NoError(t, err)
	assert.True(t, enabled)
}

func TestSetSecureBoot(t *testing.T) {
	var patchAttempts int

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/SecureBoot", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(mustReadFile(t, "dell/secureboot.json"))
		case http.MethodPatch:
			patchAttempts++
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	client := newDellSecureBootClient(t, mux)

	err := client.SetSecureBoot(context.Background(), false)
	assert.NoError(t, err)
	assert.Equal(t, 1, patchAttempts, "expected SetSecureBoot to PATCH the SecureBoot resource once")
}

func TestResetSecureBootKeys(t *testing.T) {
	var resetAttempts int

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/SecureBoot", endpointFunc(t, "dell/secureboot.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/SecureBoot/Actions/SecureBoot.ResetKeys", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		resetAttempts++
		w.WriteHeader(http.StatusNoContent)
	})
	client := newDellSecureBootClient(t, mux)

	err := client.ResetSecureBootKeys(context.Background(), "ResetAllKeysToDefault")
	assert.NoError(t, err)
	assert.Equal(t, 1, resetAttempts, "expected ResetSecureBootKeys to POST to the ResetKeys action once")
}
