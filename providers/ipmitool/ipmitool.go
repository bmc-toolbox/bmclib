package ipmitool

import (
	"context"
	"errors"
	"strings"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/bmc-toolbox/bmclib/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "ipmitool"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "ipmi"
)

// Features implemented by ipmitool
var Features = registrar.Features{
	providers.FeaturePowerSet,
	providers.FeaturePowerState,
	providers.FeatureUserRead,
	providers.FeatureBmcReset,
	providers.FeatureBootDeviceSet,
}

// Conn for Ipmitool connection details
type Conn struct {
	Host string
	Port string
	User string
	Pass string
	Log  logr.Logger
	con  *ipmi.Ipmi
}

// Open a connection to a BMC
func (c *Conn) Open(ctx context.Context) (err error) {
	c.con, err = ipmi.New(c.User, c.Pass, c.Host+":"+c.Port)
	return err
}

// Close a connection to a BMC
func (c *Conn) Close(ctx context.Context) (err error) {
	return nil
}

// Compatible tests whether a BMC is compatible with the ipmitool provider
func (c *Conn) Compatible(ctx context.Context) bool {
	err := c.Open(ctx)
	if err != nil {
		c.Log.V(0).Error(err, "error checking compatibility opening connection")
		return false
	}
	defer c.Close(ctx)
	_, err = c.con.PowerState(ctx)
	if err != nil {
		c.Log.V(0).Error(err, "error checking compatibility through power status")
	}
	return err == nil
}

func (c *Conn) Name() string {
	return ProviderName
}

// BootDeviceSet sets the next boot device with options
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	return c.con.BootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
}

// BmcReset will reset a BMC
func (c *Conn) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	return c.con.PowerResetBmc(ctx, resetType)
}

// UserRead list all users
func (c *Conn) UserRead(ctx context.Context) (users []map[string]string, err error) {
	return c.con.ReadUsers(ctx)
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.con.PowerState(ctx)
}

// PowerSet sets the power state of a BMC machine
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	switch strings.ToLower(state) {
	case "on":
		on, errOn := c.con.IsOn(ctx)
		if errOn != nil || !on {
			ok, err = c.con.PowerOn(ctx)
		} else {
			ok = true
		}
	case "off":
		ok, err = c.con.PowerOff(ctx)
	case "soft":
		ok, err = c.con.PowerSoft(ctx)
	case "reset":
		ok, err = c.con.PowerReset(ctx)
	case "cycle":
		ok, err = c.con.PowerCycle(ctx)
	default:
		err = errors.New("requested state type unknown")
	}

	return ok, err
}
