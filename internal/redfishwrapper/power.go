package redfishwrapper

import (
	"context"
	"fmt"
	"strings"
	"time"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"
)

// PowerSet sets the power state of a server
func (c *Client) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	// TODO: create consts for the state values
	switch strings.ToLower(state) {
	case "on":
		return c.SystemPowerOn(ctx)
	case "off":
		return c.SystemForceOff(ctx)
	case "soft":
		return c.SystemPowerOff(ctx)
	case "reset":
		return c.SystemReset(ctx)
	case "cycle":
		return c.SystemPowerCycle(ctx)
	default:
		return false, errors.New("unknown power action")
	}
}

// BMCReset powercycles the BMC.
func (c *Client) BMCReset(ctx context.Context, resetType string) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	manager, err := c.Manager(ctx)
	if err != nil {
		return false, err
	}

	if err = manager.Reset(rf.ResetType(resetType)); err != nil {
		return false, err
	}

	return true, nil
}

// SystemPowerOn powers on the system.
func (c *Client) SystemPowerOn(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return false, err
	}

	if system.PowerState == rf.OnPowerState {
		return true, nil
	}

	system.DisableEtagMatch(c.disableEtagMatch)
	if err = system.Reset(rf.OnResetType); err != nil {
		return false, err
	}

	return true, nil
}

// SystemPowerOff powers off the system.
func (c *Client) SystemPowerOff(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return false, err
	}

	if system.PowerState == rf.OffPowerState {
		return true, nil
	}

	system.DisableEtagMatch(c.disableEtagMatch)

	if err = system.Reset(rf.GracefulShutdownResetType); err != nil {
		return false, err
	}

	return false, nil
}

// SystemReset power cycles the system.
func (c *Client) SystemReset(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return false, err
	}

	system.DisableEtagMatch(c.disableEtagMatch)
	if err = system.Reset(rf.PowerCycleResetType); err != nil {
		_, _ = c.SystemPowerOff(ctx)

		for wait := 1; wait < 10; wait++ {
			status, _ := c.SystemPowerStatus(ctx)
			if status == "off" {
				break
			}
			time.Sleep(1 * time.Second)
		}

		return c.SystemPowerOn(ctx)
	}

	return true, nil
}

// SystemPowerCycle power cycles the system.
func (c *Client) SystemPowerCycle(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return false, err
	}

	if system.PowerState == rf.OffPowerState {
		return false, fmt.Errorf("power cycle failed: Command not supported in present state: %v", system.PowerState)
	}

	system.DisableEtagMatch(c.disableEtagMatch)

	if err = system.Reset(rf.ForceRestartResetType); err != nil {
		return false, errors.WithMessage(err, "power cycle failed")
	}

	return true, nil
}

// SystemPowerStatus returns the system power state.
func (c *Client) SystemPowerStatus(ctx context.Context) (result string, err error) {
	if err := c.SessionActive(); err != nil {
		return result, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return "", err
	}

	return string(system.PowerState), nil
}

// SystemForceOff powers off the system, without waiting for the OS to shutdown.
func (c *Client) SystemForceOff(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return false, err
	}

	if system.PowerState == rf.OffPowerState {
		return true, nil
	}

	system.DisableEtagMatch(c.disableEtagMatch)

	if err = system.Reset(rf.ForceOffResetType); err != nil {
		return false, err
	}

	return true, nil
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Client) SendNMI(_ context.Context) error {
	if err := c.SessionActive(); err != nil {
		return errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return err
	}

	return system.Reset(rf.NmiResetType)
}
