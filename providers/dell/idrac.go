package dell

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
)

const (
	// ProviderName for the provider Dell implementation
	ProviderName = "dell"
	// ProviderProtocol for the provider Dell implementation
	ProviderProtocol = "redfish"

	redfishV1Prefix           = "/redfish/v1"
	screenshotEndpoint        = "/Dell/Managers/iDRAC.Embedded.1/DellLCService/Actions/DellLCService.ExportServerScreenShot"
	managerAttributesEndpoint = "/Managers/iDRAC.Embedded.1/Attributes"
)

var (
	// Features implemented by dell redfish
	Features = registrar.Features{
		providers.FeatureScreenshot,
		providers.FeaturePowerState,
		providers.FeaturePowerSet,
		providers.FeatureFirmwareInstallSteps,
		providers.FeatureFirmwareUploadInitiateInstall,
		providers.FeatureFirmwareTaskStatus,
		providers.FeatureInventoryRead,
		providers.FeatureBmcReset,
		providers.FeatureGetBiosConfiguration,
		providers.FeatureSetBiosConfiguration,
		providers.FeatureSetBiosConfigurationFromFile,
		providers.FeatureResetBiosConfiguration,
	}

	errManufacturerUnknown = errors.New("error identifying device manufacturer")
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

// Conn details for redfish client
type Conn struct {
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
		redfishwrapper.WithVersionsNotCompatible(defaultConfig.VersionsNotCompatible),
		redfishwrapper.WithBasicAuthEnabled(defaultConfig.UseBasicAuth),
	}

	if defaultConfig.RootCAs != nil {
		rfOpts = append(rfOpts, redfishwrapper.WithSecureTLS(defaultConfig.RootCAs))
	}

	return &Conn{
		Log:            log,
		redfishwrapper: redfishwrapper.NewClient(host, defaultConfig.Port, user, pass, rfOpts...),
	}
}

// Open a connection to a BMC via redfish
func (c *Conn) Open(ctx context.Context) (err error) {
	if err := c.redfishwrapper.Open(ctx); err != nil {
		return err
	}

	// because this uses the redfish interface and the redfish interface
	// is available across various BMC vendors, we verify the device we're connected to is dell.
	if err := c.deviceSupported(); err != nil {
		if er := c.redfishwrapper.Close(ctx); er != nil {
			return fmt.Errorf("%v: %w", err, er)
		}

		return err
	}

	return nil
}

func (c *Conn) deviceSupported() error {
	manufacturer, err := c.deviceManufacturer()
	if err != nil {
		return err
	}

	m := strings.ToLower(manufacturer)
	if !strings.Contains(m, common.VendorDell) {
		return errors.Wrap(bmclibErrs.ErrIncompatibleProvider, m)
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

// GetBiosConfiguration returns the BIOS configuration settings via the BMC
func (c *Conn) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	return c.redfishwrapper.GetBiosConfiguration(ctx)
}

// SetBiosConfiguration sets the BIOS configuration settings via the BMC
func (c *Conn) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	return c.redfishwrapper.SetBiosConfiguration(ctx, biosConfig)
}

// SetBiosConfigurationFromFile sets the bios configuration from a raw vendor config file
func (c *Conn) SetBiosConfigurationFromFile(ctx context.Context, biosConfg string) (err error) {
	configMap := make(map[string]string)
	err = json.Unmarshal([]byte(biosConfg), &configMap)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal config file")
	}

	return c.redfishwrapper.SetBiosConfiguration(ctx, configMap)
}

// ResetBiosConfiguration resets the BIOS configuration settings back to 'factory defaults' via the BMC
func (c *Conn) ResetBiosConfiguration(ctx context.Context) (err error) {
	return c.redfishwrapper.ResetBiosConfiguration(ctx)
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Conn) SendNMI(ctx context.Context) error {
	return c.redfishwrapper.SendNMI(ctx)
}

// deviceManufacturer returns the device manufacturer and model attributes
func (c *Conn) deviceManufacturer() (vendor string, err error) {
	systems, err := c.redfishwrapper.Systems()
	if err != nil {
		return "", errors.Wrap(errManufacturerUnknown, err.Error())
	}

	for _, sys := range systems {
		if sys.Manufacturer != "" {
			return sys.Manufacturer, nil
		}
	}

	return "", errManufacturerUnknown
}

func (c *Conn) Screenshot(ctx context.Context) (image []byte, fileType string, err error) {
	fileType = "png"

	resp, err := c.redfishwrapper.PostWithHeaders(
		ctx,
		redfishV1Prefix+screenshotEndpoint,
		// other FileType parameters are LastCrashScreenshot, Preview
		json.RawMessage(`{"FileType":"ServerScreenShot"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if err != nil {
		return nil, "", errors.Wrap(bmclibErrs.ErrScreenshot, err.Error())
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", errors.Wrap(bmclibErrs.ErrScreenshot, err.Error())
	}

	if resp.StatusCode != 200 {
		return nil, "", errors.Wrap(bmclibErrs.ErrScreenshot, resp.Status)
	}

	data := &struct {
		B64encoded string `json:"ServerScreenshotFile"`
	}{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, "", errors.Wrap(bmclibErrs.ErrScreenshot, err.Error())
	}

	if data.B64encoded == "" {
		return nil, "", errors.Wrap(bmclibErrs.ErrScreenshot, "no screencapture data in response")
	}

	image, err = base64.StdEncoding.DecodeString(data.B64encoded)
	if err != nil {
		return nil, "", errors.Wrap(bmclibErrs.ErrScreenshot, err.Error())
	}

	return image, fileType, nil
}
