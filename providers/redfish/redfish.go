package redfish

import (
	"context"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/bmclib/v2/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
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
		providers.FeatureInventoryRead,
		providers.FeatureFirmwareInstall,
		providers.FeatureFirmwareInstallStatus,
		providers.FeatureBmcReset,
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

	_, err = c.PowerStateGet(ctx)
	if err != nil {
		c.Log.V(2).WithValues(
			"provider",
			c.Name(),
		).Info("warn", bmclibErrs.ErrCompatibilityCheck.Error(), err.Error())
	}

	return err == nil
}

// DeviceVendorModel returns the device manufacturer and model attributes
func (c *Conn) DeviceVendorModel(ctx context.Context) (vendor, model string, err error) {
	systems, err := c.redfishwrapper.Systems()
	if err != nil {
		return "", "", err
	}

	for _, sys := range systems {
		if !compatibleOdataID(sys.ODataID, systemsOdataIDs) {
			continue
		}

		return sys.Manufacturer, sys.Model, nil
	}

	return vendor, model, bmclibErrs.ErrRedfishSystemOdataID
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
	switch strings.ToLower(state) {
	case "on":
		return c.redfishwrapper.SystemPowerOn(ctx)
	case "off":
		return c.redfishwrapper.SystemForceOff(ctx)
	case "soft":
		return c.redfishwrapper.SystemPowerOff(ctx)
	case "reset":
		return c.redfishwrapper.SystemReset(ctx)
	case "cycle":
		return c.redfishwrapper.SystemPowerCycle(ctx)
	default:
		return false, errors.New("unknown power action")
	}
}
