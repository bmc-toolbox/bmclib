package lenovo

import (
	"context"
	"testing"
)

// Requirement: Power state read.
func TestPowerStateGet(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	state, err := c.PowerStateGet(context.Background())
	if err != nil {
		t.Fatalf("PowerStateGet: %v", err)
	}
	if state != "On" {
		t.Fatalf("PowerStateGet = %q, want %q", state, "On")
	}
}

// Requirement: Power control via ComputerSystem.Reset.
func TestPowerSet(t *testing.T) {
	t.Run("power on when already on is a no-op success", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)

		ok, err := c.PowerSet(context.Background(), "on")
		if err != nil || !ok {
			t.Fatalf("PowerSet(on) = (%v, %v), want (true, nil)", ok, err)
		}
		// The system fixture is already On, so no reset action is posted.
		if rt := ts.resetType(); rt != "" {
			t.Fatalf("unexpected reset posted: %q", rt)
		}
	})

	t.Run("force off posts ForceOff", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)

		ok, err := c.PowerSet(context.Background(), "off")
		if err != nil || !ok {
			t.Fatalf("PowerSet(off) = (%v, %v), want (true, nil)", ok, err)
		}
		if rt := ts.resetType(); rt != "ForceOff" {
			t.Fatalf("reset type = %q, want %q", rt, "ForceOff")
		}
	})

	t.Run("unsupported state errors", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)

		if _, err := c.PowerSet(context.Background(), "banana"); err == nil {
			t.Fatal("expected an error for an unsupported power state")
		}
	})
}

// Requirement: NMI.
func TestSendNMI(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.SendNMI(context.Background()); err != nil {
		t.Fatalf("SendNMI: %v", err)
	}
	if rt := ts.resetType(); rt != "Nmi" {
		t.Fatalf("reset type = %q, want %q", rt, "Nmi")
	}
}

// Requirement: Boot device override set.
func TestBootDeviceSet(t *testing.T) {
	tests := []struct {
		name       string
		device     string
		persistent bool
		efi        bool
		wantErr    bool
	}{
		{name: "one-time pxe", device: "pxe"},
		{name: "one-time uefi cdrom", device: "cdrom", efi: true},
		// XCC advertises BootSourceOverrideEnabled {Once, Disabled} only — a
		// persistent (Continuous) override is rejected up front.
		{name: "persistent rejected on XCC", device: "disk", persistent: true, efi: true, wantErr: true},
		{name: "unknown device errors", device: "toaster", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := newTestServer(t, testServerOpts{})
			c := ts.openedClient(t)

			ok, err := c.BootDeviceSet(context.Background(), tt.device, tt.persistent, tt.efi)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error for an unknown boot device")
				}
				return
			}
			if err != nil || !ok {
				t.Fatalf("BootDeviceSet = (%v, %v), want (true, nil)", ok, err)
			}
			if !ts.didPatchSystem() {
				t.Fatal("expected the ComputerSystem to be PATCHed")
			}
		})
	}
}

// Requirement: Boot device override read.
func TestBootDeviceOverrideGet(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	override, err := c.BootDeviceOverrideGet(context.Background())
	if err != nil {
		t.Fatalf("BootDeviceOverrideGet: %v", err)
	}
	// The fixture reports target Hdd, enabled Once, mode Legacy.
	if override.IsPersistent {
		t.Errorf("IsPersistent = true, want false (Once)")
	}
	if override.IsEFIBoot {
		t.Errorf("IsEFIBoot = true, want false (Legacy)")
	}
}

// Requirement: Boot progress reporting.
func TestGetBootProgress(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	bp, err := c.GetBootProgress()
	if err != nil {
		t.Fatalf("GetBootProgress: %v", err)
	}
	if bp == nil {
		t.Fatal("GetBootProgress returned nil")
	}
}

// Requirement: BIOS configuration read.
func TestGetBiosConfiguration(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	cfg, err := c.GetBiosConfiguration(context.Background())
	if err != nil {
		t.Fatalf("GetBiosConfiguration: %v", err)
	}
	if got := cfg["BootModes_SystemBootMode"]; got != "LegacyMode" {
		t.Fatalf("BootModes_SystemBootMode = %q, want %q", got, "LegacyMode")
	}
}

// Requirement: BIOS configuration set.
func TestSetBiosConfiguration(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	err := c.SetBiosConfiguration(context.Background(), map[string]string{"BootModes_SystemBootMode": "UEFIMode"})
	if err != nil {
		t.Fatalf("SetBiosConfiguration: %v", err)
	}
	if !ts.didPatchBios() {
		t.Fatal("expected the Bios settings target to be PATCHed")
	}

	// XCC rejects the @Redfish.SettingsApplyTime / ApplyTime annotation that the
	// shared wrapper sends; the lenovo override must PATCH Attributes only.
	body := ts.biosPatchBody
	if body == nil {
		t.Fatal("expected to capture the Bios PATCH body")
	}
	if _, ok := body["@Redfish.SettingsApplyTime"]; ok {
		t.Errorf("Bios PATCH must not send @Redfish.SettingsApplyTime (XCC rejects it); body=%v", body)
	}
	if _, ok := body["Attributes"]; !ok {
		t.Errorf("Bios PATCH must carry an Attributes object; body=%v", body)
	}
}

// Requirement: BIOS configuration reset.
func TestResetBiosConfiguration(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.ResetBiosConfiguration(context.Background()); err != nil {
		t.Fatalf("ResetBiosConfiguration: %v", err)
	}
	if !ts.didResetBios() {
		t.Fatal("expected the Bios.ResetBios action to be posted")
	}
}
