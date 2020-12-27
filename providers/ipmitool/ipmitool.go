package ipmitool

import (
	"context"
	"errors"
	"strings"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/bmc-toolbox/bmclib/registry"
	"github.com/go-logr/logr"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "ipmitool"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "ipmi"
)

// Conn for Ipmitool connection details
type Conn struct {
	Host string
	Port string
	User string
	Pass string
	Log  logr.Logger
	con  *ipmi.Ipmi
}

func init() {
	registry.Register(ProviderName, ProviderProtocol, func(host, port, user, pass string, log logr.Logger) (interface{}, registry.IsCompatibleFn, error) {
		if port == "" {
			port = "623"
		}
		i, err := ipmi.New(user, pass, host+":"+port)
		conn := &Conn{Host: host, User: user, Pass: pass, Port: port, Log: log, con: i}
		return conn, conn.isCompatible, err
	}, []registry.Feature{
		registry.FeaturePowerSet,
		registry.FeaturePowerState,
		registry.FeatureUserRead,
		registry.FeatureBmcReset,
		registry.FeatureBootDeviceSet,
	})
}

// Open a connection to a BMC
func (c *Conn) Open(ctx context.Context) (err error) {
	return nil
}

// Close a connection to a BMC
func (c *Conn) Close(ctx context.Context) (err error) {
	return nil
}

// Compatible tests whether a BMC is compatible with the ipmitool provider
func (c *Conn) isCompatible(ctx context.Context) bool {
	_, err := c.con.IsOn(ctx)
	return err == nil
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
		on, _ := c.con.IsOn(ctx)
		if on {
			ok = true
		} else {
			ok, err = c.con.PowerOn(ctx)
		}
	case "off":
		on, _ := c.con.IsOn(ctx)
		if !on {
			ok = true
		} else {
			ok, err = c.con.PowerOff(ctx)
		}
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
