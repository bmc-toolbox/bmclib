package redfishwrapper

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish"
	"golang.org/x/exp/slices"
)

// Client is a redfishwrapper client which wraps the gofish client.
type Client struct {
	host                  string
	port                  string
	user                  string
	pass                  string
	basicAuth             bool
	versionsNotCompatible []string // a slice of redfish versions to ignore as incompatible
	client                *gofish.APIClient
	httpClient            *http.Client
	httpClientSetupFuncs  []func(*http.Client)
}

// Option is a function applied to a *Conn
type Option func(*Client)

// WithHTTPClient returns an option that sets an HTTP client for the connecion
func WithHTTPClient(cli *http.Client) Option {
	return func(c *Client) {
		c.httpClient = cli
	}
}

// WithSecureTLS returns an option that enables secure TLS with an optional cert pool.
func WithSecureTLS(rootCAs *x509.CertPool) Option {
	return func(c *Client) {
		c.httpClientSetupFuncs = append(c.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

// WithVersionsNotCompatible returns an option that sets the redfish versions to ignore as incompatible.
//
// The version string value must match the value returned by
// curl -k  "https://10.247.133.39/redfish/v1" | jq .RedfishVersion
func WithVersionsNotCompatible(versions []string) Option {
	return func(c *Client) {
		c.versionsNotCompatible = append(c.versionsNotCompatible, versions...)
	}
}

// WithBasicAuthEnabled sets Basic Auth on the Gofish driver.
func WithBasicAuthEnabled(e bool) Option {
	return func(c *Client) {
		c.basicAuth = e
	}
}

// NewClient returns a redfishwrapper client
func NewClient(host, port, user, pass string, opts ...Option) *Client {
	if !strings.HasPrefix(host, "https://") && !strings.HasPrefix(host, "http://") {
		host = "https://" + host
	}

	client := &Client{
		host:                  host,
		port:                  port,
		user:                  user,
		pass:                  pass,
		versionsNotCompatible: []string{},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Open sets up a new redfish session.
func (c *Client) Open(ctx context.Context) error {
	endpoint := c.host
	if c.port != "" {
		endpoint = c.host + ":" + c.port
	}

	config := gofish.ClientConfig{
		Endpoint:   endpoint,
		Username:   c.user,
		Password:   c.pass,
		Insecure:   true,
		HTTPClient: c.httpClient,
		BasicAuth:  c.basicAuth,
	}

	if config.HTTPClient == nil {
		config.HTTPClient = httpclient.Build(c.httpClientSetupFuncs...)
	} else {
		for _, setupFunc := range c.httpClientSetupFuncs {
			setupFunc(config.HTTPClient)
		}
	}

	debug := os.Getenv("DEBUG_BMCLIB")
	if debug == "true" {
		config.DumpWriter = os.Stdout
	}

	if tm := getTimeout(ctx); tm != 0 {
		config.HTTPClient.Timeout = tm
	}
	var err error
	c.client, err = gofish.Connect(config)

	return err
}

func getTimeout(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0
	}

	return time.Until(deadline)
}

// Close closes the redfish session.
func (c *Client) Close(ctx context.Context) error {
	if c.client == nil || c.client.Service == nil {
		return nil
	}

	c.client.Logout()

	return nil
}

// SessionActive returns an error if a redfish session is not active.
func (c *Client) SessionActive() error {
	if c.client == nil || c.client.Service == nil {
		return bmclibErrs.ErrNotAuthenticated
	}

	// With basic auth enabled theres no session to be checked.
	if c.basicAuth {
		return nil
	}

	_, err := c.client.GetSession()
	if err != nil {
		fmt.Println("here")
		return err
	}

	return nil
}

// RunRawRequestWithHeaders wraps the gofish client method RunRawRequestWithHeaders
func (c *Client) RunRawRequestWithHeaders(method, url string, payloadBuffer io.ReadSeeker, contentType string, customHeaders map[string]string) (*http.Response, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.RunRawRequestWithHeaders(method, url, payloadBuffer, contentType, customHeaders)
}

func (c *Client) Delete(url string) (*http.Response, error) {
	return c.client.Delete(url)
}

func (c *Client) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

// VersionCompatible compares the redfish version reported by the BMC with the blacklist if specified.
func (c *Client) VersionCompatible() bool {
	if len(c.versionsNotCompatible) == 0 {
		return true
	}

	if err := c.SessionActive(); err != nil {
		return false
	}

	return !slices.Contains(c.versionsNotCompatible, c.client.Service.RedfishVersion)
}

func (c *Client) PostWithHeaders(ctx context.Context, url string, payload interface{}, headers map[string]string) (*http.Response, error) {
	return c.client.PostWithHeaders(url, payload, headers)
}

func (c *Client) PatchWithHeaders(ctx context.Context, url string, payload interface{}, headers map[string]string) (*http.Response, error) {
	return c.client.PatchWithHeaders(url, payload, headers)
}
