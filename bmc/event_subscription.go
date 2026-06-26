package bmc

import "context"

// EventServiceInfo describes the Redfish EventService configuration.
type EventServiceInfo struct {
	// ServiceEnabled reports whether the event service is enabled.
	ServiceEnabled bool
	// DeliveryRetryAttempts is the number of delivery retry attempts.
	DeliveryRetryAttempts int
	// DeliveryRetryIntervalSeconds is the delivery retry interval.
	DeliveryRetryIntervalSeconds int
	// SSEFilterURI is the Server-Sent-Events stream URI, when supported.
	SSEFilterURI string
}

// EventSubscription describes an event destination (subscription).
type EventSubscription struct {
	// ID is the Redfish subscription Id.
	ID string
	// Destination is the subscription delivery URL.
	Destination string
	// Protocol is the subscription protocol (e.g. "Redfish").
	Protocol string
	// Context is the opaque client context echoed in events.
	Context string
}

// EventSubscriptionRequest describes a subscription to create.
type EventSubscriptionRequest struct {
	// Destination is the delivery URL (required).
	Destination string
	// Protocol is the subscription protocol (defaults to "Redfish" when empty).
	Protocol string
	// Context is an opaque client context.
	Context string
	// RegistryPrefixes optionally filters the message-registry prefixes.
	RegistryPrefixes []string
}

// EventSubscriber is implemented by providers that can read the event service
// and manage event subscriptions.
type EventSubscriber interface {
	// EventService returns the event-service configuration.
	EventService(ctx context.Context) (EventServiceInfo, error)
	// EventSubscriptions lists the event subscriptions.
	EventSubscriptions(ctx context.Context) ([]EventSubscription, error)
	// EventSubscriptionCreate creates a subscription and returns its Id.
	EventSubscriptionCreate(ctx context.Context, req EventSubscriptionRequest) (id string, err error)
	// EventSubscriptionDelete deletes a subscription by Id.
	EventSubscriptionDelete(ctx context.Context, id string) error
	// SubmitTestEvent submits a test event (optionally for a specific MessageId).
	SubmitTestEvent(ctx context.Context, messageID string) error
	// SetEventService PATCHes event-service properties.
	SetEventService(ctx context.Context, attrs map[string]any) error
}
