package goipmi

import (
	"context"
	"errors"

	"github.com/bmc-toolbox/bmclib/providers"
	"github.com/gebn/bmc"
	"github.com/gebn/bmc/pkg/ipmi"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/jacobweinstock/registrar"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "goipmi"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "ipmi"
)

var (
	// Features implemented by goipmi
	Features = registrar.Features{
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
	}
)

type Conn struct {
	Host      string
	Port      string
	User      string
	Pass      string
	Log       logr.Logger
	transport bmc.SessionlessTransport
	conn      bmc.Session
}

func (c *Conn) Open(ctx context.Context) error {
	machine, err := bmc.Dial(ctx, c.Host)
	if err != nil {
		return err
	}
	c.transport = machine

	sess, err := machine.NewSession(ctx, &bmc.SessionOpts{
		Username:          c.User,
		Password:          []byte(c.Pass),
		MaxPrivilegeLevel: ipmi.PrivilegeLevelAdministrator,
	})
	if err != nil {
		return err
	}
	c.conn = sess
	return nil
}

func (c *Conn) Close(ctx context.Context) error {
	var err error
	if cErr := c.conn.Close(ctx); cErr != nil {
		err = multierror.Append(err, cErr)
	}
	if cErr := c.transport.Close(); cErr != nil {
		err = multierror.Append(err, cErr)
	}

	return err
}

// Compatible tests whether a BMC is compatible with the ipmitool provider
func (c *Conn) Compatible(ctx context.Context) bool {
	err := c.Open(ctx)
	if err != nil {
		c.Log.V(0).Error(err, "error checking compatibility opening connection")
		return false
	}
	err = c.Close(ctx)
	if err != nil {
		c.Log.V(0).Error(err, "error checking compatibility closing connection")
	}
	return err == nil
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.status(ctx)
}

func (c *Conn) PowerSet(ctx context.Context, action string) (ok bool, err error) {
	return doIpmiAction(ctx, action, c)
}

func doIpmiAction(ctx context.Context, action string, pwr *Conn) (ok bool, err error) {
	switch action {
	case "on":
		return pwr.on(ctx)
	case "soft":
		return pwr.off(ctx)
	case "reset":
		return pwr.reset(ctx)
	case "off":
		return pwr.hardoff(ctx)
	case "cycle":
		return pwr.cycle(ctx)
	default:
		return false, errors.New("requested state type unknown")
	}
}

func (c *Conn) on(ctx context.Context) (bool, error) {
	if err := c.conn.ChassisControl(ctx, ipmi.ChassisControlPowerOn); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Conn) off(ctx context.Context) (bool, error) {
	if err := c.conn.ChassisControl(ctx, ipmi.ChassisControlSoftPowerOff); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Conn) status(ctx context.Context) (string, error) {
	result := "off"
	status, err := c.conn.GetChassisStatus(ctx)
	if err != nil {
		return "", err
	}
	if status.PoweredOn {
		result = "on"
	}
	return result, nil
}

func (c *Conn) reset(ctx context.Context) (bool, error) {
	if err := c.conn.ChassisControl(ctx, ipmi.ChassisControlHardReset); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Conn) hardoff(ctx context.Context) (bool, error) {
	if err := c.conn.ChassisControl(ctx, ipmi.ChassisControlPowerOff); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Conn) cycle(ctx context.Context) (bool, error) {
	if err := c.conn.ChassisControl(ctx, ipmi.ChassisControlPowerCycle); err != nil {
		return false, err
	}
	return true, nil
}
