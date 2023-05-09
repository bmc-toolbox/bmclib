package dell

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/bmclib/v2/providers"
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
	screenshotEndpoint        = "/Dell/Managers/iDRAC.Embedded.1/DellLCService/Actions/DellLCService.ExportServerScreenshot"
	managerAttributesEndpoint = "/Managers/iDRAC.Embedded.1/Attributes"
)

var (
	// Features implemented by dell redfish
	Features = registrar.Features{
		providers.FeatureScreenshot,
	}
)

// Conn details for redfish client
type Conn struct {
	redfishwrapper *redfishwrapper.Client
	Log            logr.Logger
}

// New returns connection with a redfish client initialized
func New(host, port, user, pass string, log logr.Logger, opts ...redfishwrapper.Option) *Conn {
	return &Conn{
		Log:            log,
		redfishwrapper: redfishwrapper.NewClient(host, port, user, pass, opts...),
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

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.redfishwrapper.SystemPowerStatus(ctx)
}

func (c *Conn) Screenshot(ctx context.Context) (image []byte, fileType string, err error) {
	fileType = "png"

	resp, err := c.redfishwrapper.PostWithHeaders(
		ctx,
		redfishV1Prefix+screenshotEndpoint,
		// other FileType parameters are LastCrashScreenshot, Preview
		json.RawMessage(`{"FileType":"ServerScreenshot"}`),
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
