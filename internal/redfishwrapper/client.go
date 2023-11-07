package redfishwrapper

import (
	"context"
	"crypto/x509"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"
	"golang.org/x/exp/slices"
)

var (
	ErrManagerID = errors.New("error identifying Manager Odata ID")
	ErrBIOSID    = errors.New("error identifying System BIOS Odata ID")
)

// Client is a redfishwrapper client which wraps the gofish client.
type Client struct {
	host                  string
	port                  string
	user                  string
	pass                  string
	basicAuth             bool
	disableEtagMatch      bool
	versionsNotCompatible []string // a slice of redfish versions to ignore as incompatible
	client                *gofish.APIClient
	httpClient            *http.Client
	httpClientSetupFuncs  []func(*http.Client)
	logger                logr.Logger
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

// WithEtagMatchDisabled disables the If-Match Etag header from being included by the Gofish driver.
//
// As of the current implementation this disables the header for POST/PATCH requests to the System entity endpoints.
func WithEtagMatchDisabled(d bool) Option {
	return func(c *Client) {
		c.disableEtagMatch = d
	}
}

// WithLogger sets the logger on the redfish wrapper client
func WithLogger(l *logr.Logger) Option {
	return func(c *Client) {
		if l == nil {
			c.logger = logr.Discard()

			return
		}

		c.logger = *l
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
		return err
	}

	return nil
}

// Overrides the HTTP client timeout
func (c *Client) SetHttpClientTimeout(t time.Duration) {
	c.client.HTTPClient.Timeout = t
}

// retrieve the current HTTP client timeout
func (c *Client) HttpClientTimeout() time.Duration {
	return c.client.HTTPClient.Timeout
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

func (c *Client) Tasks(ctx context.Context) ([]*redfish.Task, error) {
	return c.client.Service.Tasks()
}

func (c *Client) ManagerOdataID(ctx context.Context) (string, error) {
	managers, err := c.client.Service.Managers()
	if err != nil {
		return "", errors.Wrap(ErrManagerID, err.Error())
	}

	for _, m := range managers {
		if m.ID != "" {
			return m.ODataID, nil
		}
	}

	return "", ErrManagerID
}

func (c *Client) SystemsBIOSOdataID(ctx context.Context) (string, error) {
	systems, err := c.client.Service.Systems()
	if err != nil {
		return "", errors.Wrap(ErrBIOSID, err.Error())
	}

	for _, s := range systems {
		bios, err := s.Bios()
		if err != nil {
			return "", errors.Wrap(ErrBIOSID, err.Error())
		}

		if bios == nil {
			return "", ErrBIOSID
		}

		if bios.ID != "" {
			return bios.ODataID, nil
		}
	}

	return "", ErrBIOSID
}

// DeviceVendorModel returns the device manufacturer and model attributes
func (c *Client) DeviceVendorModel(ctx context.Context) (vendor, model string, err error) {
	systems, err := c.client.Service.Systems()
	if err != nil {
		return "", "", err
	}

	for _, sys := range systems {
		return sys.Manufacturer, sys.Model, nil
	}

	return vendor, model, bmclibErrs.ErrSystemVendorModel
}
