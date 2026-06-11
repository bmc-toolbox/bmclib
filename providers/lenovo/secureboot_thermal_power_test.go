package lenovo

import (
	"context"
	"testing"
)

// Requirement: Secure Boot management — read state.
func TestGetSecureBoot(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	state, err := c.GetSecureBoot(context.Background())
	if err != nil {
		t.Fatalf("GetSecureBoot: %v", err)
	}
	if !state.Enabled {
		t.Errorf("Enabled = false, want true")
	}
	if state.Mode != "SetupMode" {
		t.Errorf("Mode = %q, want %q", state.Mode, "SetupMode")
	}
	if state.CurrentBoot != "Disabled" {
		t.Errorf("CurrentBoot = %q, want %q", state.CurrentBoot, "Disabled")
	}
}

// Requirement: Secure Boot management — enable/disable.
func TestSetSecureBoot(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.SetSecureBoot(context.Background(), true); err != nil {
		t.Fatalf("SetSecureBoot: %v", err)
	}
	if !ts.didPatchSecureBoot() {
		t.Fatal("expected the SecureBoot resource to be PATCHed")
	}
}

// Requirement: Secure Boot management — reset keys.
func TestResetSecureBootKeys(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.ResetSecureBootKeys(context.Background(), "ResetAllKeysToDefault"); err != nil {
		t.Fatalf("ResetSecureBootKeys: %v", err)
	}
	if !ts.didResetSecureBootKeys() {
		t.Fatal("expected the SecureBoot.ResetKeys action to be posted")
	}
}

// Requirement: Thermal readings.
func TestThermal(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	reading, err := c.Thermal(context.Background())
	if err != nil {
		t.Fatalf("Thermal: %v", err)
	}
	if len(reading.Temperatures) != 1 {
		t.Fatalf("got %d temperatures, want 1", len(reading.Temperatures))
	}
	if reading.Temperatures[0].Name != "Ambient Temp" || reading.Temperatures[0].ReadingCelsius != 23 {
		t.Errorf("unexpected temperature: %+v", reading.Temperatures[0])
	}
	if len(reading.Fans) != 1 {
		t.Fatalf("got %d fans, want 1", len(reading.Fans))
	}
	if reading.Fans[0].Reading != 6840 {
		t.Errorf("fan reading = %d, want 6840", reading.Fans[0].Reading)
	}
}

// Requirement: Power metrics read.
func TestReadPower(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	info, err := c.ReadPower(context.Background())
	if err != nil {
		t.Fatalf("ReadPower: %v", err)
	}
	if info.ConsumedWatts != 287 {
		t.Errorf("ConsumedWatts = %v, want 287", info.ConsumedWatts)
	}
	if info.LimitInWatts != nil {
		t.Errorf("LimitInWatts = %v, want nil (no cap)", *info.LimitInWatts)
	}
	if len(info.PowerSupplies) != 2 {
		t.Fatalf("got %d power supplies, want 2", len(info.PowerSupplies))
	}
}

// Requirement: Power capping — set a cap.
func TestSetPowerCap(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	limit := 1200.0
	if err := c.SetPowerCap(context.Background(), &limit); err != nil {
		t.Fatalf("SetPowerCap: %v", err)
	}
	if !ts.didPatchPower() {
		t.Fatal("expected the Power resource to be PATCHed")
	}

	// nil limit disables capping; should also PATCH successfully.
	if err := c.SetPowerCap(context.Background(), nil); err != nil {
		t.Fatalf("SetPowerCap(nil): %v", err)
	}
}
