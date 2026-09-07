package redfishwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func networkBootBoolPtr(b bool) *bool { return &b }

func TestNetworkBootAttributes(t *testing.T) {
	tests := map[string]struct {
		httpEnabled *bool
		pxeEnabled  *bool
		current     map[string]string
		want        map[string]string
		wantError   bool
	}{
		"enable http only, pxe untouched": {
			httpEnabled: networkBootBoolPtr(true),
			current:     map[string]string{"IPv4HTTPSupport": "Disabled", "IPv4PXESupport": "Enabled"},
			want: map[string]string{
				"NetworkStack":    "Enabled",
				"BootModeSelect":  "UEFI",
				"IPv4HTTPSupport": "Enabled",
				"IPv6HTTPSupport": "Enabled",
			},
		},
		"disable http only, pxe untouched": {
			httpEnabled: networkBootBoolPtr(false),
			current:     map[string]string{"IPv4HTTPSupport": "Enabled"},
			want: map[string]string{
				"IPv4HTTPSupport": "Disabled",
				"IPv6HTTPSupport": "Disabled",
			},
		},
		"enable pxe only, http untouched": {
			pxeEnabled: networkBootBoolPtr(true),
			current:    map[string]string{"IPv4HTTPSupport": "Enabled", "IPv4PXESupport": "Disabled"},
			want: map[string]string{
				"NetworkStack":   "Enabled",
				"BootModeSelect": "UEFI",
				"IPv4PXESupport": "Enabled",
			},
		},
		"disable pxe only, http untouched": {
			pxeEnabled: networkBootBoolPtr(false),
			current:    map[string]string{"IPv4HTTPSupport": "Enabled"},
			want: map[string]string{
				"IPv4PXESupport": "Disabled",
			},
		},
		"enable both http and pxe": {
			httpEnabled: networkBootBoolPtr(true),
			pxeEnabled:  networkBootBoolPtr(true),
			current:     map[string]string{"IPv4HTTPSupport": "Disabled"},
			want: map[string]string{
				"NetworkStack":    "Enabled",
				"BootModeSelect":  "UEFI",
				"IPv4HTTPSupport": "Enabled",
				"IPv6HTTPSupport": "Enabled",
				"IPv4PXESupport":  "Enabled",
			},
		},
		"disable http while enabling pxe": {
			httpEnabled: networkBootBoolPtr(false),
			pxeEnabled:  networkBootBoolPtr(true),
			current:     map[string]string{"IPv4HTTPSupport": "Enabled"},
			want: map[string]string{
				"NetworkStack":    "Enabled",
				"BootModeSelect":  "UEFI",
				"IPv4HTTPSupport": "Disabled",
				"IPv6HTTPSupport": "Disabled",
				"IPv4PXESupport":  "Enabled",
			},
		},
		"unknown fingerprint returns error": {
			httpEnabled: networkBootBoolPtr(true),
			current:     map[string]string{"SomeUnrelatedAttribute": "value"},
			wantError:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := networkBootAttributes(tt.httpEnabled, tt.pxeEnabled, tt.current)
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected nil err, got: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Fatalf("unexpected attributes (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSetNetworkBootEnabled(t *testing.T) {
	var patchBody map[string]string

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "/dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "/dell/system.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{
				"@odata.type": "#Bios.v1_2_3.Bios",
				"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
				"Id": "Bios",
				"Name": "BIOS Configuration",
				"AttributeRegistry": "BiosAttributeRegistryU32.v1_0_0",
				"Attributes": {
					"IPv4HTTPSupport": "Disabled"
				}
			}`))
		case http.MethodPatch:
			var body struct {
				Attributes map[string]string `json:"Attributes"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			patchBody = body.Attributes
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(ctx))
	defer client.Close(ctx)

	ok, err := client.SetNetworkBootEnabled(ctx, networkBootBoolPtr(true), nil)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, map[string]string{
		"NetworkStack":    "Enabled",
		"BootModeSelect":  "UEFI",
		"IPv4HTTPSupport": "Enabled",
		"IPv6HTTPSupport": "Enabled",
	}, patchBody)
}

func TestSetNetworkBootEnabled_NeitherSet(t *testing.T) {
	client := NewClient("unused", "", "", "")

	ok, err := client.SetNetworkBootEnabled(context.Background(), nil, nil)
	assert.False(t, ok)
	assert.ErrorContains(t, err, "at least one of httpEnabled or pxeEnabled must be set")
}
