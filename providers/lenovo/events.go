package lenovo

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

const (
	eventServiceURI       = "/redfish/v1/EventService"
	eventSubscriptionsURI = eventServiceURI + "/Subscriptions"
	submitTestEventURI    = eventServiceURI + "/Actions/EventService.SubmitTestEvent"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.EventSubscriber = (*Conn)(nil)

// EventService returns the XCC event-service configuration.
//
// Implements bmc.EventSubscriber.
func (c *Conn) EventService(ctx context.Context) (bmc.EventServiceInfo, error) {
	var doc struct {
		ServiceEnabled               bool   `json:"ServiceEnabled"`
		DeliveryRetryAttempts        int    `json:"DeliveryRetryAttempts"`
		DeliveryRetryIntervalSeconds int    `json:"DeliveryRetryIntervalSeconds"`
		ServerSentEventURI           string `json:"ServerSentEventUri"`
	}
	if err := c.getJSON(eventServiceURI, &doc); err != nil {
		return bmc.EventServiceInfo{}, err
	}

	return bmc.EventServiceInfo{
		ServiceEnabled:               doc.ServiceEnabled,
		DeliveryRetryAttempts:        doc.DeliveryRetryAttempts,
		DeliveryRetryIntervalSeconds: doc.DeliveryRetryIntervalSeconds,
		SSEFilterURI:                 doc.ServerSentEventURI,
	}, nil
}

// EventSubscriptions lists the XCC event subscriptions.
//
// Implements bmc.EventSubscriber.
func (c *Conn) EventSubscriptions(ctx context.Context) ([]bmc.EventSubscription, error) {
	members, err := c.collectionMembers(eventSubscriptionsURI)
	if err != nil {
		return nil, err
	}

	out := make([]bmc.EventSubscription, 0, len(members))
	for _, m := range members {
		var sub struct {
			ID          string `json:"Id"`
			Destination string `json:"Destination"`
			Protocol    string `json:"Protocol"`
			Context     string `json:"Context"`
		}
		if err := c.getJSON(m.ODataID, &sub); err != nil {
			return nil, err
		}
		out = append(out, bmc.EventSubscription{
			ID:          sub.ID,
			Destination: sub.Destination,
			Protocol:    sub.Protocol,
			Context:     sub.Context,
		})
	}

	return out, nil
}

// EventSubscriptionCreate creates a subscription and returns its Id.
//
// Implements bmc.EventSubscriber.
func (c *Conn) EventSubscriptionCreate(ctx context.Context, req bmc.EventSubscriptionRequest) (string, error) {
	protocol := req.Protocol
	if protocol == "" {
		protocol = "Redfish"
	}

	payload := map[string]any{
		"Destination": req.Destination,
		"Protocol":    protocol,
	}
	if req.Context != "" {
		payload["Context"] = req.Context
	}
	if len(req.RegistryPrefixes) > 0 {
		payload["RegistryPrefixes"] = req.RegistryPrefixes
	}

	resp, err := c.redfishwrapper.PostWithHeaders(ctx, eventSubscriptionsURI, payload, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", parseRedfishError(resp)
	}

	if loc := resp.Header.Get("Location"); loc != "" {
		return path.Base(strings.TrimRight(loc, "/")), nil
	}

	return "", nil
}

// EventSubscriptionDelete deletes a subscription by Id.
//
// Implements bmc.EventSubscriber.
func (c *Conn) EventSubscriptionDelete(ctx context.Context, id string) error {
	target, err := url.JoinPath(eventSubscriptionsURI, id)
	if err != nil {
		return err
	}
	return checkResponse(c.redfishwrapper.Delete(target))
}

// SubmitTestEvent submits a test event via EventService.SubmitTestEvent.
//
// Implements bmc.EventSubscriber.
func (c *Conn) SubmitTestEvent(ctx context.Context, messageID string) error {
	payload := map[string]any{}
	if messageID != "" {
		payload["MessageId"] = messageID
	}

	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, submitTestEventURI, payload, nil))
}

// SetEventService PATCHes the event-service properties.
//
// Implements bmc.EventSubscriber.
func (c *Conn) SetEventService(ctx context.Context, attrs map[string]any) error {
	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, eventServiceURI, attrs, nil))
}
