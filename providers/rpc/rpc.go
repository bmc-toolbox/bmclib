package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/providers"
	hmac "github.com/bmc-toolbox/bmclib/v2/providers/rpc/internal/http"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

// Config defines the configuration for sending rpc notifications.
type Config struct {
	// IncludeAlgoPrefix will prepend the algorithm and an equal sign to the signature. Example: sha256=abc123
	IncludeAlgoPrefix bool
	// Logger is the logger to use for logging.
	Logger logr.Logger
	// LogNotifications determines whether responses from rpc consumer/listeners will be logged or not.
	LogNotifications bool
	// HTTPContentType is the content type to use for the rpc request notification.
	HTTPContentType string
	// HTTPMethod is the HTTP method to use for the rpc request notification.
	HTTPMethod string

	// consumerURL is the URL where a rpc consumer/listener is running and to which we will send notifications.
	consumerURL string
	// host is the BMC ip address or hostname or identifier.
	host string
	// httpClient is the http client used for all methods.
	httpClient *http.Client
	// listenerURL is the URL of the rpc consumer/listener.
	listenerURL *url.URL
	// sig is for adding the signature to the request header.
	sig hmac.Signature
	// timestampHeader is the header name that should contain the timestamp. Example: X-BMCLIB-Timestamp
	timestampHeader string
	// timestampFormat is the time format for the timestamp header.
	timestampFormat string
}

type Secrets map[Algorithm][]string

// Algorithm is the type for HMAC algorithms.
type Algorithm string

const (
	// ProviderName for the Webook implementation.
	ProviderName = "rpc"
	// ProviderProtocol for the rpc implementation.
	ProviderProtocol = "http"

	// defaults
	timestampHeader = "X-BMCLIB-Timestamp"
)

// Features implemented by the AMT provider.
var Features = registrar.Features{
	providers.FeaturePowerSet,
	providers.FeaturePowerState,
	providers.FeatureBootDeviceSet,
}

// New returns a new Config for this rpc provider.
//
// Defaults:
//
//	BaseSignatureHeader: X-BMCLIB-Signature
//	IncludeAlgoHeader: true
//	IncludedPayloadHeaders: []string{"X-BMCLIB-Timestamp"}
//	IncludeAlgoPrefix: true
//	Logger: logr.Discard()
//	LogNotifications: true
//	httpClient: http.DefaultClient
func New(consumerURL string, host string, secrets Secrets) *Config {
	cfg := &Config{
		host:              host,
		consumerURL:       consumerURL,
		IncludeAlgoPrefix: true,
		Logger:            logr.Discard(),
		LogNotifications:  true,
		HTTPContentType:   "application/json",
		HTTPMethod:        http.MethodPost,
		timestampHeader:   timestampHeader,
		timestampFormat:   time.RFC3339,
		httpClient:        http.DefaultClient,
	}

	// create the signature object
	// maybe validate BaseSignatureHeader and that there are secrets?
	cfg.sig = hmac.NewSignature()
	cfg.sig.AppendAlgoToHeader = true
	cfg.sig.IncludedPayloadHeaders = []string{timestampHeader}
	if len(secrets) > 0 {
		cfg.addSecrets(secrets)
	}

	return cfg
}

// Name returns the name of this rpc provider.
// Implements bmc.Provider interface
func (c *Config) Name() string {
	return ProviderName
}

// Open a connection to the rpc consumer.
// For the rpc provider, Open means validating the Config and
// that communication with the rpc consumer can be established.
func (c *Config) Open(ctx context.Context) error {
	// 1. validate consumerURL is a properly formatted URL.
	// 2. validate that we can communicate with the rpc consumer.
	u, err := url.Parse(c.consumerURL)
	if err != nil {
		return err
	}
	c.listenerURL = u
	testReq, err := http.NewRequestWithContext(ctx, c.HTTPMethod, c.listenerURL.String(), nil)
	if err != nil {
		return err
	}
	// test that we can communicate with the rpc consumer.
	// and that it responses with the spec contract (Response{}).
	if _, err := c.httpClient.Do(testReq); err != nil { //nolint:bodyclose // not reading the body
		return err
	}

	return nil
}

// Close a connection to the rpc consumer.
func (c *Config) Close(_ context.Context) (err error) {
	return nil
}

// BootDeviceSet sends a next boot device rpc notification.
func (c *Config) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	p := RequestPayload{
		ID:     int64(time.Now().UnixNano()),
		Host:   c.host,
		Method: BootDeviceMethod,
		Params: BootDeviceParams{
			Device:     bootDevice,
			Persistent: setPersistent,
			EFIBoot:    efiBoot,
		},
	}
	req, err := c.createRequest(ctx, p)
	if err != nil {
		return false, err
	}

	resp, err := c.signAndSend(p, req)
	if err != nil {
		return ok, err
	}
	if resp.Error != nil {
		return ok, fmt.Errorf("error from rpc consumer: %v", resp.Error)
	}

	return true, nil
}

// PowerSet sets the power state of a BMC machine.
func (c *Config) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	switch strings.ToLower(state) {
	case "on", "off", "cycle":
		p := RequestPayload{
			ID:     int64(time.Now().UnixNano()),
			Host:   c.host,
			Method: PowerSetMethod,
			Params: PowerSetParams{
				State: strings.ToLower(state),
			},
		}
		req, err := c.createRequest(ctx, p)
		if err != nil {
			return false, err
		}
		resp, err := c.signAndSend(p, req)
		if err != nil {
			return ok, err
		}
		if resp.Error != nil {
			return ok, fmt.Errorf("error from rpc consumer: %v", resp.Error)
		}

		return true, nil
	}

	return false, errors.New("requested power state is not supported")
}

// PowerStateGet gets the power state of a BMC machine.
func (c *Config) PowerStateGet(ctx context.Context) (state string, err error) {
	p := RequestPayload{
		ID:     int64(time.Now().UnixNano()),
		Host:   c.host,
		Method: PowerGetMethod,
	}
	req, err := c.createRequest(ctx, p)
	if err != nil {
		return "", err
	}
	resp, err := c.signAndSend(p, req)
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", fmt.Errorf("error from rpc consumer: %v", resp.Error)
	}

	return resp.Result.(string), nil
}
