package lenovo

import (
	"context"
	"testing"
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

// Requirement: Selecting the system by name keeps the BMC in the inventory.
//
// The XCC manager has no ManagerForServers link, so the BMC must be found
// without matching the system name through it.
func TestInventoryWithSystemName(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t, WithSystemName("System"))

	device, err := c.Inventory(context.Background())
	if err != nil {
		t.Fatalf("Inventory: %v", err)
	}
	if device.BMC == nil || device.BMC.ID != "1" {
		t.Fatalf("BMC = %+v, want ID 1", device.BMC)
	}
}
