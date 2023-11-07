package redfishwrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithVersionsNotCompatible(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name     string
		versions []string
	}{
		{
			"no versions",
			[]string{},
		},
		{
			"with versions",
			[]string{"1.2.3", "4.5.6"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithVersionsNotCompatible(tt.versions))
			assert.Equal(t, tt.versions, client.versionsNotCompatible)
		})
	}
}

func TestWithBasicAuthEnabled(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			"disabled",
			false,
		},
		{
			"enabled",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithBasicAuthEnabled(tt.enabled))
			assert.Equal(t, tt.enabled, client.basicAuth)
		})
	}
}

func TestWithEtagMatchDisabled(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name     string
		disabled bool
	}{
		{
			"disabled",
			true,
		},
		{
			"enabled",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithEtagMatchDisabled(tt.disabled))
			assert.Equal(t, tt.disabled, client.disableEtagMatch)
		})
	}
}

const (
	fixturesDir = "./fixtures"
)

func TestManagerOdataID(t *testing.T) {
	tests := map[string]struct {
		hfunc  map[string]func(http.ResponseWriter, *http.Request)
		expect string
		err    error
	}{
		"happy case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				// service root
				"/redfish/v1/":           endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Systems":    endpointFunc(t, "systems.json"),
				"/redfish/v1/Managers":   endpointFunc(t, "managers.json"),
				"/redfish/v1/Managers/1": endpointFunc(t, "managers_1.json"),
			},
			expect: "/redfish/v1/Managers/1",
			err:    nil,
		},
		"failure case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/": endpointFunc(t, "/serviceroot_no_manager.json"),
			},
			expect: "",
			err:    ErrManagerID,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

			//os.Setenv("DEBUG_BMCLIB", "true")
			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "")

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			got, err := client.ManagerOdataID(ctx)
			if err != nil {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.expect, got)

			client.Close(context.Background())
		})
	}
}

func TestSystemsBIOSOdataID(t *testing.T) {
	tests := map[string]struct {
		hfunc  map[string]func(http.ResponseWriter, *http.Request)
		expect string
		err    error
	}{
		"happy case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				// service root
				"/redfish/v1/":               endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Systems":        endpointFunc(t, "systems.json"),
				"/redfish/v1/Systems/1":      endpointFunc(t, "systems_1.json"),
				"/redfish/v1/Systems/1/Bios": endpointFunc(t, "systems_bios.json"),
			},
			expect: "/redfish/v1/Systems/1/Bios",
			err:    nil,
		},
		"failure case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/": endpointFunc(t, "serviceroot.json"),
			},
			expect: "",
			err:    ErrBIOSID,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
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

			//os.Setenv("DEBUG_BMCLIB", "true")
			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "")

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			got, err := client.SystemsBIOSOdataID(ctx)
			if err != nil {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.expect, got)

			client.Close(context.Background())
		})
	}
}
