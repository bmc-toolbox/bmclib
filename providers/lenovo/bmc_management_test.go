package lenovo

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// Requirement: BMC reset.
func TestBmcReset(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ok, err := c.BmcReset(context.Background(), "GracefulRestart")
	if err != nil || !ok {
		t.Fatalf("BmcReset = (%v, %v), want (true, nil)", ok, err)
	}
	if !ts.didBmcReset() {
		t.Error("expected a Manager.Reset action")
	}
}

// Requirement: Reset to factory defaults.
func TestResetToFactoryDefaults(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.ResetToFactoryDefaults(context.Background(), "ResetAll"); err != nil {
		t.Fatalf("ResetToFactoryDefaults: %v", err)
	}
	if !ts.didFactoryReset() {
		t.Error("expected a Manager.ResetToDefaults action")
	}
}

// Requirement: License management — read installed licenses.
func TestLicenses(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	licenses, err := c.Licenses(context.Background())
	if err != nil {
		t.Fatalf("Licenses: %v", err)
	}
	if len(licenses) != 1 {
		t.Fatalf("got %d licenses, want 1", len(licenses))
	}
	if licenses[0].ID != "XCC_Advanced" || licenses[0].LicenseType != "Production" || !licenses[0].Removable {
		t.Errorf("unexpected license: %+v", licenses[0])
	}
}

// Requirement: License management — an absent LicenseService (404 on firmware
// levels without it) is reported as "no licenses", not an error.
func TestLicensesNotFound(t *testing.T) {
	ts := newTestServer(t, testServerOpts{licenseServiceNotFound: true})
	c := ts.openedClient(t)

	licenses, err := c.Licenses(context.Background())
	if err != nil {
		t.Fatalf("Licenses on a box without LicenseService should not error, got: %v", err)
	}
	if len(licenses) != 0 {
		t.Fatalf("got %d licenses, want 0", len(licenses))
	}
}

// Requirement: License management — install a license.
func TestLicenseInstall(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.LicenseInstall(context.Background(), "QUJDREVGRw=="); err != nil {
		t.Fatalf("LicenseInstall: %v", err)
	}
	if !ts.didInstallLicense() {
		t.Error("expected a POST to the License collection")
	}
}

// Requirement: License management — delete a license.
func TestLicenseDelete(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.LicenseDelete(context.Background(), "XCC_Advanced"); err != nil {
		t.Fatalf("LicenseDelete: %v", err)
	}
	if !ts.didDeleteLicense() {
		t.Error("expected a DELETE of the license")
	}
}

// Requirement: Secure Key Lifecycle — read properties.
func TestGetSecureKeyLifecycle(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	cfg, err := c.GetSecureKeyLifecycle(context.Background())
	if err != nil {
		t.Fatalf("GetSecureKeyLifecycle: %v", err)
	}
	if cfg.DeviceGroup != "TKLM_DEV_GROUP" {
		t.Errorf("DeviceGroup = %q, want %q", cfg.DeviceGroup, "TKLM_DEV_GROUP")
	}
	if len(cfg.KeyRepoServers) != 2 || cfg.KeyRepoServers[0].Port != 5696 {
		t.Errorf("unexpected key repo servers: %+v", cfg.KeyRepoServers)
	}
}

// Requirement: Secure Key Lifecycle — update key servers.
func TestSetSecureKeyRepoServers(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	servers := []bmc.SecureKeyRepoServer{{HostName: "10.0.0.20", Port: 5696}}
	if err := c.SetSecureKeyRepoServers(context.Background(), servers); err != nil {
		t.Fatalf("SetSecureKeyRepoServers: %v", err)
	}
	if !ts.didPatchSKLM() {
		t.Error("expected a PATCH of the SecureKeyLifecycleService")
	}
}
