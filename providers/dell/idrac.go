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

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/bmclib/v2/providers"
	"github.com/bmc-toolbox/common"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
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
	}
)

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
	manufacturer, err := c.deviceManufacturer(ctx)
	if err != nil {
		return err
	}

	if !strings.Contains(strings.ToLower(manufacturer), common.VendorDell) {
		return bmclibErrs.ErrIncompatibleProvider
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

// deviceManufacturer returns the device manufacturer and model attributes
func (c *Conn) deviceManufacturer(ctx context.Context) (vendor string, err error) {
	errManufacturerUnknown := errors.New("error identifying device manufacturer")

	systems, err := c.redfishwrapper.Systems()
	if err != nil {
		fmt.Println(err.Error())
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
