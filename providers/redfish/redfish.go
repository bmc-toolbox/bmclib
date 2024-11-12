package redfish

import (
	"context"
	"crypto/x509"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/internal/httpclient"
	"github.com/metal-toolbox/bmclib/internal/redfishwrapper"
	"github.com/metal-toolbox/bmclib/providers"

	"github.com/metal-toolbox/bmclib/bmc"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "gofish"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "redfish"
)

var (
	// Features implemented by gofish
	Features = registrar.Features{
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
		providers.FeatureUserCreate,
		providers.FeatureUserUpdate,
		providers.FeatureUserDelete,
		providers.FeatureBootDeviceSet,
		providers.FeatureVirtualMedia,
		providers.FeatureInventoryRead,
		providers.FeatureBmcReset,
		providers.FeatureClearSystemEventLog,
		providers.FeatureGetBiosConfiguration,
		providers.FeatureSetBiosConfiguration,
		providers.FeatureResetBiosConfiguration,
	}
)

// Conn details for redfish client
type Conn struct {
	redfishwrapper       *redfishwrapper.Client
	failInventoryOnError bool
	Log                  logr.Logger
}

type Config struct {
	HttpClient *http.Client
	Port       string
	// VersionsNotCompatible	is the list of incompatible redfish versions.
	//
	// With this option set, The bmclib.Registry.FilterForCompatible(ctx) method will not proceed on
	// devices with the given redfish version(s).
	VersionsNotCompatible []string
	RootCAs               *x509.CertPool
	UseBasicAuth          bool
	// DisableEtagMatch disables the If-Match Etag header from being included by the Gofish driver.
	DisableEtagMatch bool
	SystemName       string
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

func WithVersionsNotCompatible(versionsNotCompatible []string) Option {
	return func(c *Config) {
		c.VersionsNotCompatible = versionsNotCompatible
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

func WithSystemName(name string) Option {
	return func(c *Config) {
		c.SystemName = name
	}
}

// WithEtagMatchDisabled disables the If-Match Etag header from being included by the Gofish driver.
//
// As of the current implementation this disables the header for POST/PATCH requests to the System entity endpoints.
func WithEtagMatchDisabled(d bool) Option {
	return func(c *Config) {
		c.DisableEtagMatch = d
	}
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
		redfishwrapper.WithVersionsNotCompatible(defaultConfig.VersionsNotCompatible),
		redfishwrapper.WithEtagMatchDisabled(defaultConfig.DisableEtagMatch),
		redfishwrapper.WithBasicAuthEnabled(defaultConfig.UseBasicAuth),
		redfishwrapper.WithSystemName(defaultConfig.SystemName),
	}

	if defaultConfig.RootCAs != nil {
		rfOpts = append(rfOpts, redfishwrapper.WithSecureTLS(defaultConfig.RootCAs))
	}

	return &Conn{
		Log:                  log,
		failInventoryOnError: false,
		redfishwrapper:       redfishwrapper.NewClient(host, defaultConfig.Port, user, pass, rfOpts...),
	}
}

// Open a connection to a BMC via redfish
func (c *Conn) Open(ctx context.Context) (err error) {
	return c.redfishwrapper.Open(ctx)
}

// Close a connection to a BMC via redfish
func (c *Conn) Close(ctx context.Context) error {
	return c.redfishwrapper.Close(ctx)
}

// Name returns the client provider name.
func (c *Conn) Name() string {
	return ProviderName
}

// Compatible tests whether a BMC is compatible with the gofish provider
func (c *Conn) Compatible(ctx context.Context) bool {
	err := c.Open(ctx)
	if err != nil {
		c.Log.V(2).WithValues(
			"provider",
			c.Name(),
		).Info("warn", bmclibErrs.ErrCompatibilityCheck.Error(), err.Error())

		return false
	}
	defer c.Close(ctx)

	if !c.redfishwrapper.VersionCompatible() {
		c.Log.V(2).WithValues(
			"provider",
			c.Name(),
		).Info("info", bmclibErrs.ErrCompatibilityCheck.Error(), "incompatible redfish version")

		return false
	}

	_, err = c.PowerStateGet(ctx)
	if err != nil {
		c.Log.V(2).WithValues(
			"provider",
			c.Name(),
		).Info("warn", bmclibErrs.ErrCompatibilityCheck.Error(), err.Error())
	}

	return err == nil
}

// BmcReset power cycles the BMC
func (c *Conn) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	return c.redfishwrapper.BMCReset(ctx, resetType)
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.redfishwrapper.SystemPowerStatus(ctx)
}

// PowerSet sets the power state of a server
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	return c.redfishwrapper.PowerSet(ctx, state)
}

// BootDeviceSet sets the boot device
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	return c.redfishwrapper.SystemBootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
}

// BootDeviceOverrideGet gets the boot override device information
func (c *Conn) BootDeviceOverrideGet(ctx context.Context) (bmc.BootDeviceOverride, error) {
	return c.redfishwrapper.GetBootDeviceOverride(ctx)
}

// SetVirtualMedia sets the virtual media
func (c *Conn) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (ok bool, err error) {
	return c.redfishwrapper.SetVirtualMedia(ctx, kind, mediaURL)
}

// Inventory collects hardware inventory and install firmware information
func (c *Conn) Inventory(ctx context.Context) (device *common.Device, err error) {
	return c.redfishwrapper.Inventory(ctx, c.failInventoryOnError)
}

// GetBiosConfiguration return bios configuration
func (c *Conn) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	return c.redfishwrapper.GetBiosConfiguration(ctx)
}

// SetBiosConfiguration set bios configuration
func (c *Conn) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	return c.redfishwrapper.SetBiosConfiguration(ctx, biosConfig)
}

// ResetBiosConfiguration set bios configuration
func (c *Conn) ResetBiosConfiguration(ctx context.Context) (err error) {
	return c.redfishwrapper.ResetBiosConfiguration(ctx)
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Conn) SendNMI(ctx context.Context) error {
	return c.redfishwrapper.SendNMI(ctx)
}
