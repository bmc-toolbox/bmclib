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

func networkBootBoolPtr(b bool) *bool { return &b }

// newNetworkBootTestClient wires a mux to a fresh Conn, capturing every PATCH to the Bios
// resource's body into patchedBodies.
func newNetworkBootTestClient(t *testing.T, patchedBodies *[]string) *Conn {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"@odata.id":"/redfish/v1/Systems/System.Embedded.1/Bios","Attributes":{}}`))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			*patchedBodies = append(*patchedBodies, string(body))
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	t.Cleanup(func() { _ = client.Close(context.Background()) })

	return client
}

// TestSetNetworkBootEnabled verifies SetNetworkBootEnabled PATCHes Dell's per-NIC
// HttpDev1EnDis/PxeDev1EnDis BIOS Setup attributes (device slot 1, see networkBootDeviceIndex),
// and that httpEnabled/pxeEnabled are independent: only the non-nil one is included in the PATCH.
func TestSetNetworkBootEnabled(t *testing.T) {
	testCases := map[string]struct {
		httpEnabled *bool
		pxeEnabled  *bool
		want        string
		notWant     string
	}{
		"http only": {httpEnabled: networkBootBoolPtr(true), want: `"HttpDev1EnDis":"Enabled"`, notWant: "PxeDev1EnDis"},
		"pxe only":  {pxeEnabled: networkBootBoolPtr(false), want: `"PxeDev1EnDis":"Disabled"`, notWant: "HttpDev1EnDis"},
		"both, http on pxe off": {
			httpEnabled: networkBootBoolPtr(true),
			pxeEnabled:  networkBootBoolPtr(false),
			want:        `"HttpDev1EnDis":"Enabled"`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var patchedBodies []string
			client := newNetworkBootTestClient(t, &patchedBodies)

			ok, err := client.SetNetworkBootEnabled(context.Background(), tc.httpEnabled, tc.pxeEnabled)
			require.NoError(t, err)
			assert.True(t, ok)
			require.Len(t, patchedBodies, 1)
			assert.Contains(t, patchedBodies[0], tc.want)
			if tc.notWant != "" {
				assert.NotContains(t, patchedBodies[0], tc.notWant)
			}
		})
	}
}

func TestSetNetworkBootEnabled_NeitherSet(t *testing.T) {
	var patchedBodies []string
	client := newNetworkBootTestClient(t, &patchedBodies)

	ok, err := client.SetNetworkBootEnabled(context.Background(), nil, nil)
	assert.False(t, ok)
	assert.ErrorContains(t, err, "at least one of httpEnabled or pxeEnabled must be set")
	assert.Empty(t, patchedBodies, "expected no PATCH when neither httpEnabled nor pxeEnabled is set")
}
