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

// TestSetSecureBoot verifies SetSecureBoot PATCHes Dell's SecureBoot BIOS Setup attribute and
// leaves the generic Redfish ComputerSystem SecureBoot resource untouched, so that the change is
// staged into the single pending BIOS configuration job iDRAC allows rather than sealing it from
// outside.
func TestSetSecureBoot(t *testing.T) {
	for _, tc := range []struct {
		name     string
		enable   bool
		wantBody string
	}{
		{name: "enable", enable: true, wantBody: `"SecureBoot":"Enabled"`},
		{name: "disable", enable: false, wantBody: `"SecureBoot":"Disabled"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var biosPatched bool

			mux := http.NewServeMux()
			mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
			mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
			mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
			mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					_, _ = w.Write([]byte(`{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Bios","Attributes":{}}`))
				case http.MethodPatch:
					biosPatched = true
					body, err := io.ReadAll(r.Body)
					require.NoError(t, err)
					assert.Contains(t, string(body), tc.wantBody)
					w.WriteHeader(http.StatusNoContent)
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			})
			// Served so that a SecureBoot-resource implementation would get all the way to its
			// PATCH rather than failing earlier on the GET: the point of the guard below is that
			// no PATCH is attempted, not that the resource is unreachable.
			mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/SecureBoot", func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodGet:
					_, _ = w.Write([]byte(`{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/SecureBoot","SecureBootEnable":false}`))
				case http.MethodPatch:
					t.Error("SecureBoot resource must not be PATCHed: that seals iDRAC's single pending BIOS config job")
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

			require.NoError(t, client.SetSecureBoot(context.Background(), tc.enable))
			assert.True(t, biosPatched, "expected a PATCH to the Bios resource setting the SecureBoot attribute")
		})
	}
}
