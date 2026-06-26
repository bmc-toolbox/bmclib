package lenovo

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// Requirement: Aggregate hardware and firmware inventory + partial-failure
// tolerance (default failInventoryOnError=false).
func TestInventory(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	device, err := c.Inventory(context.Background())
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if device == nil {
		t.Fatal("Inventory returned a nil device")
	}
}

// Requirement: Partial-failure tolerance — fail-fast mode surfaces the error.
//
// With FailInventoryOnError set and the mock not serving UpdateService, the
// first sub-resource read fails and Inventory returns an error.
func TestInventoryFailFast(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t, WithFailInventoryOnError(true))

	if _, err := c.Inventory(context.Background()); err == nil {
		t.Fatal("expected Inventory to fail fast when a sub-resource is unavailable")
	}
}

// Requirement: Storage pool and controller reads.
func TestStorageControllers(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	controllers, err := c.StorageControllers(context.Background())
	if err != nil {
		t.Fatalf("StorageControllers: %v", err)
	}
	if len(controllers) != 1 {
		t.Fatalf("got %d controllers, want 1", len(controllers))
	}
	got := controllers[0]
	if got.ID != "RAID_Slot1" {
		t.Errorf("controller ID = %q, want %q", got.ID, "RAID_Slot1")
	}
	if got.DriveCount != 2 {
		t.Errorf("DriveCount = %d, want 2", got.DriveCount)
	}
	if got.Model == "" {
		t.Error("controller Model is empty")
	}
}

// Requirement: List volumes of a controller.
func TestVolumes(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	volumes, err := c.Volumes(context.Background(), "RAID_Slot1")
	if err != nil {
		t.Fatalf("Volumes: %v", err)
	}
	if len(volumes) != 1 {
		t.Fatalf("got %d volumes, want 1", len(volumes))
	}
	if volumes[0].RAIDType != "RAID1" {
		t.Errorf("RAIDType = %q, want %q", volumes[0].RAIDType, "RAID1")
	}
	if volumes[0].CapacityBytes != 999653638144 {
		t.Errorf("CapacityBytes = %d, want 999653638144", volumes[0].CapacityBytes)
	}
}

// Requirement: Volume creation.
func TestVolumeCreate(t *testing.T) {
	t.Run("create returns the new volume id", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)

		id, err := c.VolumeCreate(context.Background(), "RAID_Slot1", bmcVolumeReq())
		if err != nil {
			t.Fatalf("VolumeCreate: %v", err)
		}
		if id != "2" {
			t.Errorf("new volume id = %q, want %q", id, "2")
		}
		if !ts.didCreateVolume() {
			t.Fatal("expected a POST to the Volumes collection")
		}
	})

	t.Run("unknown controller errors", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)

		if _, err := c.VolumeCreate(context.Background(), "NoSuchController", bmcVolumeReq()); err == nil {
			t.Fatal("expected an error for an unknown controller")
		}
	})
}

// Requirement: Volume initialize, update, delete.
func TestVolumeLifecycleActions(t *testing.T) {
	t.Run("initialize", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.VolumeInitialize(context.Background(), "RAID_Slot1", "1", "Fast"); err != nil {
			t.Fatalf("VolumeInitialize: %v", err)
		}
		if !ts.didInitializeVolume() {
			t.Fatal("expected the Volume.Initialize action to be posted")
		}
	})

	t.Run("update", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.VolumeUpdate(context.Background(), "RAID_Slot1", "1", map[string]any{"Encrypted": true}); err != nil {
			t.Fatalf("VolumeUpdate: %v", err)
		}
		if !ts.didUpdateVolume() {
			t.Fatal("expected the Volume to be PATCHed")
		}
	})

	t.Run("delete", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.VolumeDelete(context.Background(), "RAID_Slot1", "1"); err != nil {
			t.Fatalf("VolumeDelete: %v", err)
		}
		if !ts.didDeleteVolume() {
			t.Fatal("expected the Volume to be DELETEd")
		}
	})

	t.Run("unknown volume errors", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.VolumeDelete(context.Background(), "RAID_Slot1", "999"); err == nil {
			t.Fatal("expected an error for an unknown volume")
		}
	})
}

func bmcVolumeReq() bmc.VolumeCreateRequest {
	return bmc.VolumeCreateRequest{
		Name:          "newvol",
		RAIDType:      "RAID0",
		CapacityBytes: 1000000000,
		Drives:        []string{"/redfish/v1/Systems/1/Storage/RAID_Slot1/Drives/Disk.0"},
	}
}
