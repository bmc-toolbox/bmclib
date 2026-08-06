package redfishwrapper

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func biosConfigFromFixture(t *testing.T) map[string]string {
	t.Helper()

	fixturePath := fixturesDir + "/dell/bios.json"
	fh, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("%s, failed to open fixture: %s", err.Error(), fixturePath)
	}

	defer fh.Close()

	b, err := io.ReadAll(fh)
	if err != nil {
		t.Fatalf("%s, failed to read fixture: %s", err.Error(), fixturePath)
	}

	var bios map[string]any
	err = json.Unmarshal([]byte(b), &bios)
	if err != nil {
		t.Fatalf("%s, failed to unmarshal fixture: %s", err.Error(), fixturePath)
	}

	expectedBiosConfig := make(map[string]string)
	for k, v := range bios["Attributes"].(map[string]any) {
		expectedBiosConfig[k] = fmt.Sprintf("%v", v)
	}

	return expectedBiosConfig
}

func TestGetBiosConfiguration(t *testing.T) {
	tests := []struct {
		testName           string
		hfunc              map[string]func(http.ResponseWriter, *http.Request)
		expectedBiosConfig map[string]string
	}{
		{
			"GetBiosConfiguration",
			map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/":                               endpointFunc(t, "/dell/serviceroot.json"),
				"/redfish/v1/Systems":                        endpointFunc(t, "/dell/systems.json"),
				"/redfish/v1/Systems/System.Embedded.1":      endpointFunc(t, "/dell/system.embedded.1.json"),
				"/redfish/v1/Systems/System.Embedded.1/Bios": endpointFunc(t, "/dell/bios.json"),
			},
			biosConfigFromFixture(t),
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			mux := http.NewServeMux()
			handleFunc := tc.hfunc
			for endpoint, handler := range handleFunc {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			biosConfig, err := client.GetBiosConfiguration(ctx)
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedBiosConfig, biosConfig)
		})
	}
}

// biosWithoutSettingsApplyTimes is a Bios resource with no @Redfish.Settings
// block at all - matching a BMC that doesn't declare
// @Redfish.Settings.SupportedApplyTimes and rejects the
// @Redfish.SettingsApplyTime property outright (confirmed live against a
// Supermicro AS-1114S-WN10RT-EU).
const biosWithoutSettingsApplyTimes = `{
	"@odata.type": "#Bios.v1_2_3.Bios",
	"@odata.id": "/redfish/v1/Systems/System.Embedded.1/Bios",
	"Id": "Bios",
	"Name": "BIOS Configuration",
	"AttributeRegistry": "BiosAttributeRegistryU32.v1_0_0",
	"Attributes": {
		"BootModeSelect": "UEFI"
	}
}`

// propertyUnknownSettingsApplyTime is the exact error shape returned live by
// a Supermicro BMC when @Redfish.SettingsApplyTime is included in a PATCH to
// a Bios resource that doesn't support it.
const propertyUnknownSettingsApplyTime = `{"error":{"code":"Base.v1_10_3.GeneralError","message":"A general error has occurred. See ExtendedInfo for more information.","@Message.ExtendedInfo":[{"MessageId":"Base.1.10.PropertyUnknown","Severity":"Warning","Resolution":"Remove the unknown property from the request body and resubmit the request if the operation failed.","Message":"The property @Redfish.SettingsApplyTime is not in the list of valid properties for the resource.","MessageArgs":["@Redfish.SettingsApplyTime"],"RelatedProperties":["@Redfish.SettingsApplyTime"]}]}}`

func TestSetBiosConfiguration_ApplyTimeUnsupportedFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "/dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "/dell/system.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(biosWithoutSettingsApplyTimes))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			if strings.Contains(string(body), "SettingsApplyTime") {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(propertyUnknownSettingsApplyTime))
				return
			}
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

	err = client.SetBiosConfiguration(ctx, map[string]string{"BootModeSelect": "Legacy"})
	assert.NoError(t, err, "expected fallback retry without the apply-time hint to succeed")
}

func TestSetBiosConfiguration_ApplyTimeRejectedWithoutFallbackTrigger(t *testing.T) {
	// A BMC that rejects the PATCH for an unrelated reason should not trigger
	// the apply-time fallback, and the original error should propagate.
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "/dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "/dell/system.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/Bios", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(biosWithoutSettingsApplyTimes))
		case http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"code":"Base.v1_10_3.GeneralError","message":"unrelated failure","@Message.ExtendedInfo":[{"MessageId":"Base.1.10.PropertyValueNotInList","RelatedProperties":["BootModeSelect"]}]}}`))
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

	err = client.SetBiosConfiguration(ctx, map[string]string{"BootModeSelect": "Legacy"})
	assert.ErrorContains(t, err, "PropertyValueNotInList")
}
