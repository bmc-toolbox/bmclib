package lenovo

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// Requirement: Read event service.
func TestEventService(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	info, err := c.EventService(context.Background())
	if err != nil {
		t.Fatalf("EventService: %v", err)
	}
	if !info.ServiceEnabled || info.DeliveryRetryAttempts != 3 {
		t.Errorf("unexpected event service: %+v", info)
	}
	if info.SSEFilterURI == "" {
		t.Error("expected an SSE URI")
	}
}

// Requirement: List subscriptions.
func TestEventSubscriptions(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	subs, err := c.EventSubscriptions(context.Background())
	if err != nil {
		t.Fatalf("EventSubscriptions: %v", err)
	}
	if len(subs) != 1 || subs[0].Destination == "" {
		t.Fatalf("unexpected subscriptions: %+v", subs)
	}
}

// Requirement: Create a webhook subscription.
func TestEventSubscriptionCreate(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	id, err := c.EventSubscriptionCreate(context.Background(), bmc.EventSubscriptionRequest{
		Destination: "https://listener.example.com/hook",
		Context:     "ctx",
	})
	if err != nil {
		t.Fatalf("EventSubscriptionCreate: %v", err)
	}
	if id != "2" {
		t.Errorf("subscription id = %q, want %q", id, "2")
	}
	if !ts.didCreateSubscription() {
		t.Error("expected a POST to the subscriptions collection")
	}
}

// Requirement: Delete a subscription.
func TestEventSubscriptionDelete(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.EventSubscriptionDelete(context.Background(), "1"); err != nil {
		t.Fatalf("EventSubscriptionDelete: %v", err)
	}
	if !ts.didDeleteSubscription() {
		t.Error("expected a DELETE of the subscription")
	}
}

// Requirement: Submit a test event.
func TestSubmitTestEvent(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.SubmitTestEvent(context.Background(), "Base.1.0.TestMessage"); err != nil {
		t.Fatalf("SubmitTestEvent: %v", err)
	}
	if !ts.didSubmitTestEvent() {
		t.Error("expected a SubmitTestEvent action")
	}
}

// Requirement: Configure event service.
func TestSetEventService(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.SetEventService(context.Background(), map[string]any{"DeliveryRetryAttempts": 5}); err != nil {
		t.Fatalf("SetEventService: %v", err)
	}
	if !ts.didPatchEventService() {
		t.Error("expected a PATCH of the EventService")
	}
}

// Requirement: Read telemetry service.
func TestTelemetryService(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	info, err := c.TelemetryService(context.Background())
	if err != nil {
		t.Fatalf("TelemetryService: %v", err)
	}
	if !info.ServiceEnabled || info.MaxReports != 10 {
		t.Errorf("unexpected telemetry service: %+v", info)
	}
}

// Requirement: Read a metric report.
func TestMetricReport(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ids, err := c.MetricReports(context.Background())
	if err != nil {
		t.Fatalf("MetricReports: %v", err)
	}
	if len(ids) != 1 || ids[0] != "PowerMetrics" {
		t.Fatalf("unexpected report ids: %v", ids)
	}

	report, err := c.MetricReport(context.Background(), "PowerMetrics")
	if err != nil {
		t.Fatalf("MetricReport: %v", err)
	}
	// XCC identifies each reading by MetricProperty (MetricId is typically empty).
	if len(report.Values) != 2 ||
		report.Values[0].MetricProperty != "/redfish/v1/Chassis/1/Power#/PowerControl/0/PowerMetrics/MaxConsumedWatts" ||
		report.Values[0].Value != "287" {
		t.Errorf("unexpected report values: %+v", report.Values)
	}
}

// Requirement: List metric definitions.
func TestMetricDefinitions(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	defs, err := c.MetricDefinitions(context.Background())
	if err != nil {
		t.Fatalf("MetricDefinitions: %v", err)
	}
	if len(defs) != 2 {
		t.Fatalf("got %d metric definitions, want 2", len(defs))
	}
}

// Requirement: Submit a test metric report.
func TestSubmitTestMetricReport(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.SubmitTestMetricReport(context.Background(), "PowerMetrics"); err != nil {
		t.Fatalf("SubmitTestMetricReport: %v", err)
	}
	if !ts.didSubmitTestMetric() {
		t.Error("expected a SubmitTestMetricReport action")
	}
}
