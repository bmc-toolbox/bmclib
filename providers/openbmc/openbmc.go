package openbmc

import (
	"bytes"
	"context"
	"crypto/x509"
	"io"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/internal/httpclient"
	"github.com/metal-toolbox/bmclib/internal/redfishwrapper"
	"github.com/metal-toolbox/bmclib/providers"
	"github.com/pkg/errors"
)

const (
	// ProviderName for the OpenBMC provider implementation
	ProviderName = "openbmc"
	// ProviderProtocol for the OpenBMC provider implementation
	ProviderProtocol = "redfish"
)

var (
	// Features implemented by dell redfish
	Features = registrar.Features{
		providers.FeaturePowerState,
		providers.FeaturePowerSet,
		providers.FeatureBmcReset,
		providers.FeatureFirmwareInstallSteps,
		providers.FeatureFirmwareUploadInitiateInstall,
		providers.FeatureFirmwareTaskStatus,
		providers.FeatureInventoryRead,
	}

	errNotOpenBMCDevice = errors.New("not an OpenBMC device")
)

type Config struct {
	HttpClient            *http.Client
	Port                  string
	VersionsNotCompatible []string
	RootCAs               *x509.CertPool
	UseBasicAuth          bool
}

// Option for setting optional Client values
type Option func(*Config)

func WithHttpClient(httpClient *http.Client) Option {
	return func(c *Config) {
		c.HttpClient = httpClient
	}
}

func WithPort(port string) Option {
	return func(c *Config) {
		c.Port = port
	}
}

func WithRootCAs(rootCAs *x509.CertPool) Option {
	return func(c *Config) {
		c.RootCAs = rootCAs
	}
}

func WithUseBasicAuth(useBasicAuth bool) Option {
	return func(c *Config) {
		c.UseBasicAuth = useBasicAuth
	}
}

// Conn details for redfish client
type Conn struct {
	host           string
	httpClient     *http.Client
	redfishwrapper *redfishwrapper.Client
	Log            logr.Logger
}

// New returns connection with a redfish client initialized
func New(host, user, pass string, log logr.Logger, opts ...Option) *Conn {
	defaultConfig := &Config{
		HttpClient:            httpclient.Build(),
		Port:                  "443",
		VersionsNotCompatible: []string{},
	}

	for _, opt := range opts {
		opt(defaultConfig)
	}

	rfOpts := []redfishwrapper.Option{
		redfishwrapper.WithHTTPClient(defaultConfig.HttpClient),
		redfishwrapper.WithBasicAuthEnabled(defaultConfig.UseBasicAuth),
		redfishwrapper.WithEtagMatchDisabled(true),
	}

	if defaultConfig.RootCAs != nil {
		rfOpts = append(rfOpts, redfishwrapper.WithSecureTLS(defaultConfig.RootCAs))
	}

	return &Conn{
		host:           host,
		httpClient:     defaultConfig.HttpClient,
		Log:            log,
		redfishwrapper: redfishwrapper.NewClient(host, defaultConfig.Port, user, pass, rfOpts...),
	}
}

// Open a connection to a BMC via redfish
func (c *Conn) Open(ctx context.Context) (err error) {
	if err := c.deviceSupported(ctx); err != nil {
		return err
	}

	if err := c.redfishwrapper.Open(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Conn) deviceSupported(ctx context.Context) error {
	var host = c.host
	if !strings.HasPrefix(host, "https://") && !strings.HasPrefix(host, "http://") {
		host = "https://" + host
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, host, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if !bytes.Contains(b, []byte(`OpenBMC`)) {
		return errNotOpenBMCDevice
	}

	return nil
}

// Close a connection to a BMC via redfish
func (c *Conn) Close(ctx context.Context) error {
	return c.redfishwrapper.Close(ctx)
}

// Name returns the client provider name.
func (c *Conn) Name() string {
	return ProviderName
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.redfishwrapper.SystemPowerStatus(ctx)
}

// PowerSet sets the power state of a server
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	return c.redfishwrapper.PowerSet(ctx, state)
}

// Inventory collects hardware inventory and install firmware information
func (c *Conn) Inventory(ctx context.Context) (device *common.Device, err error) {
	return c.redfishwrapper.Inventory(ctx, false)
}

// BmcReset power cycles the BMC
func (c *Conn) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	return c.redfishwrapper.BMCReset(ctx, resetType)
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Conn) SendNMI(ctx context.Context) error {
	return c.redfishwrapper.SendNMI(ctx)
}
