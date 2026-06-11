package lenovo

import (
	"context"
	"crypto/x509"
	"net/http"
	"strings"
	"testing"

	"github.com/go-logr/logr"
)

// Requirement: Provider identity and registration.
func TestName(t *testing.T) {
	if ProviderName != "lenovo" {
		t.Fatalf("ProviderName = %q, want %q", ProviderName, "lenovo")
	}
	c := New("127.0.0.1", "u", "p", logr.Discard())
	if got := c.Name(); got != ProviderName {
		t.Fatalf("Name() = %q, want %q", got, ProviderName)
	}
}

// Requirement: Connection lifecycle and session management — Open establishes a
// session.
func TestOpenCreatesSession(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	defer ts.Close()

	c := ts.client(t)
	if err := c.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer c.Close(context.Background())

	if !ts.didCreateSession() {
		t.Fatal("expected a session to be created on Open")
	}
}

// Requirement: Connection lifecycle — Close releases the session.
func TestCloseReleasesSession(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	defer ts.Close()

	c := ts.client(t)
	if err := c.Open(context.Background()); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := c.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if !ts.didDeleteSession() {
		t.Fatal("expected the session to be deleted on Close")
	}
}

// Requirement: Connection lifecycle — Basic auth bypasses session creation.
func TestOpenBasicAuthNoSession(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	defer ts.Close()

	c := ts.client(t, WithUseBasicAuth(true))
	if err := c.Open(context.Background()); err != nil {
		t.Fatalf("Open (basic auth): %v", err)
	}
	defer c.Close(context.Background())

	if ts.didCreateSession() {
		t.Fatal("did not expect a session to be created under basic auth")
	}
}

// Requirement: Connection lifecycle — Open returns a wrapped error on auth
// failure.
func TestOpenAuthFailure(t *testing.T) {
	ts := newTestServer(t, testServerOpts{rejectAuth: true})
	defer ts.Close()

	c := ts.client(t)
	if err := c.Open(context.Background()); err == nil {
		t.Fatal("expected Open to fail when the BMC rejects authentication")
	}
}

// Requirement: Vendor compatibility gating.
func TestCompatible(t *testing.T) {
	tests := []struct {
		name          string
		systemFixture string
		rejectAuth    bool
		unreachable   bool
		want          bool
	}{
		{name: "lenovo device is compatible", systemFixture: "system.lenovo.json", want: true},
		{name: "non-lenovo device is filtered out", systemFixture: "system.nonlenovo.json", want: false},
		{name: "unreachable device is not compatible", unreachable: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTestServer(t, testServerOpts{systemFixture: tt.systemFixture, rejectAuth: tt.rejectAuth})

			c := ts.client(t)

			if tt.unreachable {
				// Shut the server down so the host cannot be reached; Compatible
				// must return false without panicking.
				ts.Close()
			} else {
				defer ts.Close()
			}

			if got := c.Compatible(context.Background()); got != tt.want {
				t.Fatalf("Compatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Requirement: Configuration options — defaults are applied and options
// override them.
func TestConfigDefaultsAndOverrides(t *testing.T) {
	// Defaults.
	cfg := newConfig()
	if cfg.Port != "443" {
		t.Errorf("default Port = %q, want %q", cfg.Port, "443")
	}
	if cfg.HTTPClient == nil {
		t.Error("default HTTPClient is nil, want a default client")
	}
	if cfg.UseBasicAuth {
		t.Error("default UseBasicAuth = true, want false")
	}

	// Overrides.
	hc := &http.Client{}
	pool := x509.NewCertPool()
	cfg = newConfig(
		WithPort("8443"),
		WithHTTPClient(hc),
		WithUseBasicAuth(true),
		WithRootCAs(pool),
		WithSystemName("Self"),
		WithEtagMatchDisabled(true),
		WithVersionsNotCompatible([]string{"1.0.0"}),
	)
	if cfg.Port != "8443" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8443")
	}
	if cfg.HTTPClient != hc {
		t.Error("HTTPClient not overridden")
	}
	if !cfg.UseBasicAuth {
		t.Error("UseBasicAuth not set")
	}
	if cfg.RootCAs != pool {
		t.Error("RootCAs not set")
	}
	if cfg.SystemName != "Self" {
		t.Errorf("SystemName = %q, want %q", cfg.SystemName, "Self")
	}
	if !cfg.DisableEtagMatch {
		t.Error("DisableEtagMatch not set")
	}
	if len(cfg.VersionsNotCompatible) != 1 || cfg.VersionsNotCompatible[0] != "1.0.0" {
		t.Errorf("VersionsNotCompatible = %v, want [1.0.0]", cfg.VersionsNotCompatible)
	}
}

// Requirement: Service Root discovery.
func TestServiceRoot(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	root, err := c.ServiceRoot(context.Background())
	if err != nil {
		t.Fatalf("ServiceRoot: %v", err)
	}

	if root.RedfishVersion != "1.15.0" {
		t.Errorf("RedfishVersion = %q, want %q", root.RedfishVersion, "1.15.0")
	}
	if root.Vendor != "Lenovo" {
		t.Errorf("Vendor = %q, want %q", root.Vendor, "Lenovo")
	}

	links := map[string]string{
		"Systems":       root.Systems.ODataID,
		"Managers":      root.Managers.ODataID,
		"UpdateService": root.UpdateService.ODataID,
	}
	for name, got := range links {
		if got == "" {
			t.Errorf("service root link %s is empty", name)
		}
	}
	if root.Systems.ODataID != "/redfish/v1/Systems" {
		t.Errorf("Systems link = %q, want /redfish/v1/Systems", root.Systems.ODataID)
	}
}

// Requirement: Error mapping including the OEM ExtendedError registry.
func TestRedfishErrorDetail(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string // substring expected in the detail
	}{
		{
			name: "OEM ExtendedError with resolution",
			body: `{"error":{"code":"Base.1.8.GeneralError","message":"A general error has occurred",` +
				`"@Message.ExtendedInfo":[{"MessageId":"Lenovo.ExtendedError.1.2.2.ParamError",` +
				`"Message":"The parameter is invalid.","Resolution":"Correct the parameter and retry."}]}}`,
			want: "Correct the parameter and retry",
		},
		{
			name: "top-level code and message only",
			body: `{"error":{"code":"Base.1.8.MalformedJSON","message":"The request body is malformed."}}`,
			want: "The request body is malformed.",
		},
		{
			name: "non-redfish body returns empty",
			body: `not json at all`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := redfishErrorDetail([]byte(tt.body))
			if tt.want == "" {
				if got != "" {
					t.Fatalf("detail = %q, want empty", got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Fatalf("detail = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

// Requirement: Error mapping — parseRedfishError tolerates a nil response.
func TestParseRedfishErrorNilResponse(t *testing.T) {
	if err := parseRedfishError(nil); err == nil {
		t.Fatal("expected an error for a nil response")
	}
}
