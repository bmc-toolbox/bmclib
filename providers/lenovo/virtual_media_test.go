package lenovo

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Requirement: Insert virtual media by URL + slot selected by kind.
//
// XCC mounts via PATCH on the VirtualMedia resource (it does not expose the
// InsertMedia action), so the provider must PATCH the matching CD slot.
func TestSetVirtualMediaInsert(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ok, err := c.SetVirtualMedia(context.Background(), "CD", "http://192.168.1.2/Core-current.iso")
	if err != nil || !ok {
		t.Fatalf("SetVirtualMedia(insert) = (%v, %v), want (true, nil)", ok, err)
	}
	if !ts.didInsertCD() {
		t.Error("expected a PATCH inserting media into the CD slot")
	}
}

// Requirement: Eject virtual media — ejecting an empty slot is an idempotent
// success (nothing to PATCH).
func TestSetVirtualMediaEjectIdempotent(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	// The CD slot fixture is not inserted; ejecting it should succeed without
	// erroring.
	ok, err := c.SetVirtualMedia(context.Background(), "CD", "")
	if err != nil || !ok {
		t.Fatalf("SetVirtualMedia(eject) = (%v, %v), want (true, nil)", ok, err)
	}
}

// Re-mounting into a slot that already holds media must eject it first (so the
// device ends with a single mount, not a second instance in another slot).
func TestSetVirtualMediaReplacesExisting(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	// The Floppy slot fixture is already occupied (Inserted + Image).
	ok, err := c.SetVirtualMedia(context.Background(), "Floppy", "http://192.168.1.2/new.img")
	if err != nil || !ok {
		t.Fatalf("SetVirtualMedia(replace) = (%v, %v), want (true, nil)", ok, err)
	}

	ops := ts.vmOperations()
	want := []string{"Floppy:eject", "Floppy:insert"}
	if len(ops) != len(want) {
		t.Fatalf("vm ops = %v, want %v (eject before insert, same slot)", ops, want)
	}
	for i := range want {
		if ops[i] != want[i] {
			t.Fatalf("vm ops = %v, want %v", ops, want)
		}
	}
}

// Requirement: Slot selection by kind (Floppy vs CD) — and nondeterministic
// member ordering is handled by selecting on MediaTypes, not index.
func TestSetVirtualMediaInvalidKind(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if _, err := c.SetVirtualMedia(context.Background(), "Banana", "http://x/y.iso"); err == nil {
		t.Fatal("expected an error for an invalid media kind")
	}
}

// Requirement: Floppy image upload/eject — unsupported hardware errors clearly.
func TestMountFloppyImageUnsupported(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	err := c.MountFloppyImage(context.Background(), strings.NewReader("floppy-bytes"))
	if err == nil {
		t.Fatal("expected MountFloppyImage to return an error on XCC")
	}
	if !errors.Is(err, errFloppyUploadUnsupported) {
		t.Errorf("error = %v, want errFloppyUploadUnsupported", err)
	}
}

// Requirement: Floppy eject is supported (ejects the URL-mounted Floppy media).
func TestUnmountFloppyImage(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.UnmountFloppyImage(context.Background()); err != nil {
		t.Fatalf("UnmountFloppyImage: %v", err)
	}
	if !ts.didEjectFloppy() {
		t.Error("expected a PATCH ejecting the Floppy slot")
	}
}
