package rpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

// Config defines the configuration for sending rpc notifications.
type Config struct {
	// Host is the BMC ip address or hostname or identifier.
	Host string
	// ConsumerURL is the URL where a rpc consumer/listener is running and to which we will send notifications.
	ConsumerURL string
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

	// httpClient is the http client used for all methods.
	httpClient *http.Client
	// listenerURL is the URL of the rpc consumer/listener.
	listenerURL *url.URL
	// sig is for adding the signature to the request header.
	sig signature
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
	// SHA256 is the SHA256 algorithm.
	SHA256 Algorithm = "sha256"
	// SHA256Short is the short version of the SHA256 algorithm.
	SHA256Short Algorithm = "256"
	// SHA512 is the SHA512 algorithm.
	SHA512 Algorithm = "sha512"
	// SHA512Short is the short version of the SHA512 algorithm.
	SHA512Short Algorithm = "512"

	// defaults
	timestampHeader = "X-BMCLIB-Timestamp"
	signatureHeader = "X-BMCLIB-Signature"
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
		Host:              host,
		ConsumerURL:       consumerURL,
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
	cfg.sig = newSignature()
	cfg.sig.HeaderName = signatureHeader
	cfg.sig.AppendAlgoToHeader = true
	cfg.sig.IncludedPayloadHeaders = []string{timestampHeader}
	if len(secrets) > 0 {
		cfg.AddSecrets(secrets)
	}

	return cfg
}

// SetBaseSignatureHeader sets the header name that should contain the signature(s). Example: X-BMCLIB-Signature
func (c *Config) SetBaseSignatureHeader(header string) {
	c.sig.HeaderName = header
}

// SetIncludedPayloadHeaders are headers whose values will be included in the signature payload. Example: X-BMCLIB-Timestamp
func (c *Config) SetIncludedPayloadHeaders(headers []string) {
	c.sig.IncludedPayloadHeaders = append(headers, c.timestampHeader)
}

// IncludeAlgoHeader determines whether to append the algorithm to the signature header or not.
// Example: X-BMCLIB-Signature becomes X-BMCLIB-Signature-256
// When set to false, a header will be added for each algorithm. Example: X-BMCLIB-Signature-256 and X-BMCLIB-Signature-512
func (c *Config) SetIncludeAlgoHeader(include bool) {
	c.sig.AppendAlgoToHeader = include
}

// remove an element at index i from a slice of strings.
func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	s[len(s)-1] = ""
	return s[:len(s)-1]
}

// SetTimestampHeader sets the header name that should contain the timestamp. Example: X-BMCLIB-Timestamp
func (c *Config) SetTimestampHeader(header string) {
	// update c.IncludedPayloadHeaders with timestamp header
	// remove old timestamp header from c.IncludedPayloadHeaders
	c.sig.IncludedPayloadHeaders = append(c.sig.IncludedPayloadHeaders, header)
	sort.Strings(c.sig.IncludedPayloadHeaders)
	idx := sort.SearchStrings(c.sig.IncludedPayloadHeaders, timestampHeader)
	c.sig.IncludedPayloadHeaders = remove(c.sig.IncludedPayloadHeaders, idx)
	c.timestampHeader = header
}

// AddSecrets adds secrets to the Config.
func (c *Config) AddSecrets(smap map[Algorithm][]string) {
	for algo, secrets := range smap {
		switch algo {
		case SHA256, SHA256Short:
			c.sig.HMAC.Hashes = mergeHashes(c.sig.HMAC.Hashes, newSHA256(secrets...))
		case SHA512, SHA512Short:
			c.sig.HMAC.Hashes = mergeHashes(c.sig.HMAC.Hashes, newSHA512(secrets...))
		}
	}
}

func (c *Config) AddTLSCert(cert []byte) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS12,
		},
	}
	c.httpClient = &http.Client{Transport: tp}
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

	u, err := url.Parse(c.ConsumerURL)
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
		Host:   c.Host,
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
			Host:   c.Host,
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
		Host:   c.Host,
		Method: PowerGetMethod,
		Params: PowerGetParams{
			GetState: true,
		},
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

func requestKVS(req *http.Request) []interface{} {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		// TODO(jacobweinstock): either log the error or change the func signature to return it
		return nil
	}
	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	var p RequestPayload
	if err := json.Unmarshal(reqBody, &p); err != nil {
		// TODO(jacobweinstock): either log the error or change the func signature to return it
		return nil
	}

	return []interface{}{
		"requestBody", p,
		"requestHeaders", req.Header,
		"requestURL", req.URL.String(),
		"requestMethod", req.Method,
	}
}

func responseKVS(resp *http.Response) []interface{} {
	reqBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	var p map[string]interface{}
	if err := json.Unmarshal(reqBody, &p); err != nil {
		return nil
	}

	return []interface{}{
		"statusCode", resp.StatusCode,
		"responseBody", p,
		"responseHeaders", resp.Header,
	}
}

func (c *Config) createRequest(ctx context.Context, p RequestPayload) (*http.Request, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, c.HTTPMethod, c.listenerURL.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", c.HTTPContentType)
	req.Header.Add(c.timestampHeader, time.Now().Format(c.timestampFormat))

	return req, nil
}

func (c *Config) signAndSend(p RequestPayload, req *http.Request) (*ResponsePayload, error) {
	if err := c.sig.AddSignature(req); err != nil {
		return nil, err
	}
	// have to copy the body out before sending the request.
	kvs := requestKVS(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if c.LogNotifications {
			kvs = append(kvs, responseKVS(resp)...)
			kvs = append(kvs, []interface{}{"host", c.Host, "method", p.Method, "params", p.Params, "consumerURL", c.ConsumerURL}...)
			c.Logger.Info("sent rpc notification", kvs...)
		}
	}()
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	res := &ResponsePayload{}
	if err := json.Unmarshal(bodyBytes, res); err != nil {
		example, _ := json.Marshal(ResponsePayload{ID: 123, Host: c.Host, Error: &ResponseError{Code: 1, Message: "error message"}})
		return nil, fmt.Errorf("failed to parse response: got: %q, error: %w, response json spec: %v", string(bodyBytes), err, string(example))
	}

	return res, nil
}
