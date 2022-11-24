package redfishwrapper

import (
	"context"
	"crypto/x509"
	"io"
	"net/http"
	"os"
	"strings"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish"
)

// Client is a redfishwrapper client which wraps the gofish client.
type Client struct {
	host                 string
	port                 string
	user                 string
	pass                 string
	client               *gofish.APIClient
	httpClient           *http.Client
	httpClientSetupFuncs []func(*http.Client)
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

// NewClient returns a redfishwrapper client
func NewClient(host, port, user, pass string, opts ...Option) *Client {
	if !strings.HasPrefix(host, "https://") && !strings.HasPrefix(host, "http://") {
		host = "https://" + host
	}

	client := &Client{
		host: host,
		port: port,
		user: user,
		pass: pass,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// Open sets up a new redfish session.
func (c *Client) Open(ctx context.Context) error {
	config := gofish.ClientConfig{
		Endpoint:   c.host,
		Username:   c.user,
		Password:   c.pass,
		Insecure:   true,
		HTTPClient: c.httpClient,
	}

	if config.HTTPClient == nil {
		var err error
		config.HTTPClient, err = httpclient.Build(c.httpClientSetupFuncs...)
		if err != nil {
			return err
		}
	} else {
		for _, setupFunc := range c.httpClientSetupFuncs {
			setupFunc(config.HTTPClient)
		}
	}

	debug := os.Getenv("DEBUG_BMCLIB")
	if debug == "true" {
		config.DumpWriter = os.Stdout
	}

	var err error
	c.client, err = gofish.ConnectContext(ctx, config)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the redfish session.
func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.client.Service == nil {
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

	_, err := c.client.GetSession()
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
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
