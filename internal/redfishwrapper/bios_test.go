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
	"testing"

	"github.com/stretchr/testify/assert"
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
	err = json.Unmarshal(b, &bios)
	if err != nil {
		t.Fatalf("%s, failed to unmarshal fixture: %s", err.Error(), fixturePath)
	}

	expectedBiosConfig := make(map[string]string)
	for k, v := range bios["Attributes"].(map[string]any) { // nolint:forcetypeassert
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
