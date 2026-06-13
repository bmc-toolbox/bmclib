package lenovo

import (
	"context"
	"strings"
	"testing"
)

// Requirement: Read the System Event Log.
func TestGetSystemEventLog(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	entries, err := c.GetSystemEventLog(context.Background())
	if err != nil {
		t.Fatalf("GetSystemEventLog: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least one SEL entry")
	}
	// Each row is [id, created, description, message].
	if len(entries[0]) != 4 {
		t.Fatalf("row width = %d, want 4", len(entries[0]))
	}
}

// Requirement: Read the raw System Event Log.
func TestGetSystemEventLogRaw(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	raw, err := c.GetSystemEventLogRaw(context.Background())
	if err != nil {
		t.Fatalf("GetSystemEventLogRaw: %v", err)
	}
	if !strings.Contains(raw, "System boot completed") {
		t.Errorf("raw SEL did not contain the expected entry: %s", raw)
	}
}

// Requirement: Clear the System Event Log.
func TestClearSystemEventLog(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.ClearSystemEventLog(context.Background()); err != nil {
		t.Fatalf("ClearSystemEventLog: %v", err)
	}
	if !ts.didClearSEL() {
		t.Error("expected a LogService.ClearLog action on the chassis SEL")
	}
}

// Requirement: Additional XCC log types — read the audit log.
func TestEventLogAudit(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	entries, err := c.EventLog(context.Background(), LogServiceAudit)
	if err != nil {
		t.Fatalf("EventLog(AuditLog): %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d audit entries, want 1", len(entries))
	}
	if !strings.Contains(entries[0][3], "logged in") {
		t.Errorf("unexpected audit message: %q", entries[0][3])
	}
}

// Requirement: Unknown/absent service errors clearly.
func TestEventLogUnknown(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if _, err := c.EventLog(context.Background(), "NoSuchLog"); err == nil {
		t.Fatal("expected an error for an absent log service")
	}
}
