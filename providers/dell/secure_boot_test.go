package dell

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// biosWithSecureBootPolicy is a minimal Bios resource with an OnReset-capable
// @Redfish.Settings block, matching a real iDRAC.
const biosWithSecureBootPolicy = `{
	"@odata.type": "#Bios.v1_1_0.Bios",
	"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
	"Id": "Bios",
	"Name": "BIOS Configuration Current Settings",
	"AttributeRegistry": "BiosAttributeRegistry.v1_0_3",
	"Attributes": {
		"SecureBootPolicy": "%s"
	},
	"@Redfish.Settings": {
		"@odata.type": "#Settings.v1_3_0.Settings",
		"SettingsObject": {
			"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios/Settings"
		},
		"SupportedApplyTimes": ["OnReset"]
	}
}`

// biosWithoutSecureBootPolicy models an older platform generation that
// doesn't expose the attribute.
const biosWithoutSecureBootPolicy = `{
	"@odata.type": "#Bios.v1_1_0.Bios",
	"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
	"Id": "Bios",
	"Name": "BIOS Configuration Current Settings",
	"AttributeRegistry": "BiosAttributeRegistry.v1_0_3",
	"Attributes": {
		"BootMode": "Uefi"
	}
}`

func newSecureBootTestConn(t *testing.T, mux *http.ServeMux) *Conn {
	t.Helper()

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))
	require.NoError(t, client.Open(context.Background()))
	t.Cleanup(func() { _ = client.Close(context.Background()) })

	return client
}

func TestAllowCustomSecureBootKeys_EnableFromStandard(t *testing.T) {
	var settingsPatched bool
	var patchedBody string

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = fmt.Fprintf(w, biosWithSecureBootPolicy, secureBootPolicyStandard)
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case http.MethodPatch:
			settingsPatched = true
			b, _ := io.ReadAll(r.Body)
			patchedBody = string(b)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	client := newSecureBootTestConn(t, mux)

	rebootRequired, err := client.AllowCustomSecureBootKeys(context.Background(), true)
	require.NoError(t, err)
	assert.True(t, rebootRequired)
	assert.True(t, settingsPatched, "expected a BIOS settings job to be scheduled")
	assert.Contains(t, patchedBody, secureBootPolicyCustom)
}

// TestAllowCustomSecureBootKeys_EnableAlreadyCustom verifies the write happens when a different
// value is genuinely pending from an earlier, unrelated call in the same boot cycle, even though
// currently-applied state already matches what's requested (see the doc comment on
// AllowCustomSecureBootKeys) - skipping based on applied state alone would silently leave that
// stale pending value in place.
func TestAllowCustomSecureBootKeys_EnableAlreadyCustom(t *testing.T) {
	var settingsPatched bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = fmt.Fprintf(w, biosWithSecureBootPolicy, secureBootPolicyCustom)
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			// A stale pending value from an earlier, unrelated call - genuinely differs from
			// what's being requested here, even though applied state already matches.
			_, _ = w.Write([]byte(`{"Attributes": {"SecureBootPolicy": "Standard"}}`))
		case http.MethodPatch:
			settingsPatched = true
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	client := newSecureBootTestConn(t, mux)

	rebootRequired, err := client.AllowCustomSecureBootKeys(context.Background(), true)
	require.NoError(t, err)
	assert.True(t, rebootRequired)
	assert.True(t, settingsPatched, "expected the write to happen to correct the stale pending value")
}

// TestAllowCustomSecureBootKeys_DisableAlreadyStandard mirrors
// TestAllowCustomSecureBootKeys_EnableAlreadyCustom in the other direction: the bug is
// symmetric, so both currently-applied values need the same always-reach-the-write behavior.
func TestAllowCustomSecureBootKeys_DisableAlreadyStandard(t *testing.T) {
	var settingsPatched bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = fmt.Fprintf(w, biosWithSecureBootPolicy, secureBootPolicyStandard)
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			// A stale pending value from an earlier, unrelated call - genuinely differs from
			// what's being requested here, even though applied state already matches.
			_, _ = w.Write([]byte(`{"Attributes": {"SecureBootPolicy": "Custom"}}`))
		case http.MethodPatch:
			settingsPatched = true
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	client := newSecureBootTestConn(t, mux)

	rebootRequired, err := client.AllowCustomSecureBootKeys(context.Background(), false)
	require.NoError(t, err)
	assert.True(t, rebootRequired)
	assert.True(t, settingsPatched, "expected the write to happen to correct the stale pending value")
}

func TestAllowCustomSecureBootKeys_Disable(t *testing.T) {
	var settingsPatched bool
	var patchedBody string

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = fmt.Fprintf(w, biosWithSecureBootPolicy, secureBootPolicyCustom)
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios/Settings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case http.MethodPatch:
			settingsPatched = true
			b, _ := io.ReadAll(r.Body)
			patchedBody = string(b)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	client := newSecureBootTestConn(t, mux)

	rebootRequired, err := client.AllowCustomSecureBootKeys(context.Background(), false)
	require.NoError(t, err)
	assert.True(t, rebootRequired)
	assert.True(t, settingsPatched, "expected a BIOS settings job to be scheduled")
	assert.Contains(t, patchedBody, secureBootPolicyStandard)
}

func TestAllowCustomSecureBootKeys_AttributeAbsent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc("/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc("/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc("/systems_embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_, _ = w.Write([]byte(biosWithoutSecureBootPolicy))
	})

	client := newSecureBootTestConn(t, mux)

	rebootRequired, err := client.AllowCustomSecureBootKeys(context.Background(), true)
	require.Error(t, err)
	assert.False(t, rebootRequired)

	var unsupported *bmclibErrs.ErrUnsupportedHardware
	assert.ErrorAs(t, err, &unsupported)
}

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
