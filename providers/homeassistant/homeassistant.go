// Package homeassistant implements a bmclib provider for Home Assistant-managed devices.
package homeassistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/providers"
)

const (
	// ProviderName for the HomeAssistant implementation.
	ProviderName = "homeassistant"
	// ProviderProtocol for the HomeAssistant implementation.
	ProviderProtocol = "http"
)

// Features implemented by the HomeAssistant provider.
var Features = registrar.Features{
	providers.FeaturePowerSet,
	providers.FeaturePowerState,
	providers.FeatureBootDeviceSet, // no-op
}

// Config holds the configuration for the HomeAssistant provider.
type Config struct {
	APIURL                     string
	APIToken                   string
	SwitchEntityID             string
	PowerOperationDelaySeconds uint32
	HTTPClient                 *http.Client
	Logger                     logr.Logger
}

// EntityStateResponse holds the state details of a HomeAssistant entity.
type EntityStateResponse struct {
	EntityID     string
	FriendlyName string
	State        string
}

// New returns a new Config containing all the defaults for the HomeAssistant provider.
func New(apiURL, apiToken string) *Config {
	return &Config{
		APIURL:     apiURL,
		APIToken:   apiToken,
		HTTPClient: httpclient.Build(),
		Logger:     logr.Discard(),
	}
}

// Name returns the name of this HomeAssistant provider.
// Implements bmc.Provider interface
func (p *Config) Name() string {
	return ProviderName
}

// Open a connection to Home Assistant, and validate the entity referenced exists.
func (p *Config) Open(ctx context.Context) error {
	p.Logger.Info("homeassistant provider opened")

	entityState, err := p.haGetEntityState(ctx, p.SwitchEntityID)
	if err != nil {
		return fmt.Errorf("failed to get Home Assistant entity state: %w", err)
	}

	p.Logger.Info("Home Assistant entity state", "entity", p.SwitchEntityID, "entityState", entityState)

	return nil
}

func (p *Config) haGetEntityState(ctx context.Context, haEntityID string) (EntityStateResponse, error) {
	stateURL, err := url.JoinPath(p.APIURL, "api", "states", haEntityID)
	if err != nil {
		return EntityStateResponse{}, err
	}
	p.Logger.Info("Testing connection to Home Assistant API", "url", stateURL)

	req, err := http.NewRequestWithContext(ctx, "GET", stateURL, http.NoBody)
	if err != nil {
		return EntityStateResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIToken)
	req.Header.Set("Accept-Encoding", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return EntityStateResponse{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return EntityStateResponse{}, fmt.Errorf("failed to connect to Home Assistant API, status code: %d", resp.StatusCode)
	}
	if resp.ContentLength < 0 {
		return EntityStateResponse{}, fmt.Errorf("invalid content length in response: %d", resp.ContentLength)
	}
	respBuf := new(bytes.Buffer)
	if _, err := io.CopyN(respBuf, resp.Body, resp.ContentLength); err != nil {
		return EntityStateResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}
	p.Logger.Info("Successfully connected to Home Assistant API", "entity", haEntityID, "statusCode", resp.StatusCode, "respBuf", respBuf)

	// Deserialize into a temp struct
	stateResponse := struct {
		State      string            `json:"state"`
		EntityID   string            `json:"entity_id"`
		Attributes map[string]string `json:"attributes"`
	}{}
	if err := json.Unmarshal(respBuf.Bytes(), &stateResponse); err != nil {
		return EntityStateResponse{}, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	// Ensure we have Attributes["friendly_name"] field
	if _, ok := stateResponse.Attributes["friendly_name"]; !ok {
		return EntityStateResponse{}, fmt.Errorf("missing friendly_name attribute in response")
	}
	finalResponse := EntityStateResponse{
		EntityID:     stateResponse.EntityID,
		FriendlyName: stateResponse.Attributes["friendly_name"],
		State:        stateResponse.State,
	}
	return finalResponse, nil
}

// Close a connection to the HomeAssistant consumer.
func (p *Config) Close(_ context.Context) (err error) {
	return nil
}

// PowerStateGet gets the power state of a BMC machine.
func (p *Config) PowerStateGet(ctx context.Context) (state string, err error) {
	entityState, err := p.haGetEntityState(ctx, p.SwitchEntityID)
	if err != nil {
		return "unknown", fmt.Errorf("failed to get Home Assistant entity state: %w", err)
	}
	p.Logger.Info("Home Assistant PowerStateGet", "entity", p.SwitchEntityID, "entityState", entityState)
	return entityState.State, nil
}

// PowerSet sets the power state of a BMC machine.
func (p *Config) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	// Send a POST request to the Home Assistant API to toggle the switch entity

	var service string
	switch state {
	case "on":
		service = "turn_on"
	case "off":
		service = "turn_off"
	default:
		return false, fmt.Errorf("invalid power state: %s", state)
	}

	serviceURL, err := url.JoinPath(p.APIURL, "api", "services", "switch", service)
	if err != nil {
		return false, err
	}
	p.Logger.Info("Setting Home Assistant entity power state", "url", serviceURL, "entity", p.SwitchEntityID, "desiredState", state)
	reqBodyMap := map[string]interface{}{
		"entity_id": p.SwitchEntityID,
	}
	reqBodyBytes, err := json.Marshal(reqBodyMap)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", serviceURL, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+p.APIToken)
	req.Header.Set("Accept-Encoding", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		return false, fmt.Errorf("failed to set power state, status code: %d", resp.StatusCode)
	}

	p.Logger.Info("Successfully set Home Assistant entity power state", "entity", p.SwitchEntityID, "desiredState", state)

	// Sleep for the configured delay to allow the power operation to take effect
	if p.PowerOperationDelaySeconds > 0 {
		p.Logger.Info("Waiting for power operation delay", "seconds", p.PowerOperationDelaySeconds)
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(time.Duration(p.PowerOperationDelaySeconds) * time.Second):
		}
		p.Logger.Info("Power operation delay complete")
	} else {
		p.Logger.Info("No power operation delay configured, proceeding immediately")
	}

	return true, nil
}

// BootDeviceSet is a no-op here.
func (p *Config) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	// fully no-op for now; in the future, some other switch could touch some GPIO which could work with a custom bootloader
	p.Logger.Info("BootDeviceSet is not implemented for Home Assistant provider; no operation performed", "bootDevice", bootDevice, "setPersistent", setPersistent, "efiBoot", efiBoot)
	return true, nil
}
