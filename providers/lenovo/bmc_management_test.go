package lenovo

import (
	"context"
	"testing"
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
