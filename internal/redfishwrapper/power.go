package redfishwrapper

import (
	"context"
	"fmt"
	"strings"
	"time"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
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

	managers, err := c.Managers(ctx)
	if err != nil {
		return false, err
	}

	for _, manager := range managers {
		err = manager.Reset(rf.ResetType(resetType))
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// SystemPowerOn powers on the system.
func (c *Client) SystemPowerOn(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	ss, err := c.Systems()
	if err != nil {
		return false, err
	}

	for _, system := range ss {
		if system.PowerState == rf.OnPowerState {
			break
		}

		system.DisableEtagMatch(c.disableEtagMatch)

		err = system.Reset(rf.OnResetType)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

// SystemPowerOff powers off the system.
func (c *Client) SystemPowerOff(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	ss, err := c.Systems()
	if err != nil {
		return false, err
	}

	for _, system := range ss {
		if system.PowerState == rf.OffPowerState {
			break
		}

		system.DisableEtagMatch(c.disableEtagMatch)

		err = system.Reset(rf.GracefulShutdownResetType)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

// SystemReset power cycles the system.
func (c *Client) SystemReset(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	ss, err := c.Systems()
	if err != nil {
		return false, err
	}

	for _, system := range ss {
		system.DisableEtagMatch(c.disableEtagMatch)
		err = system.Reset(rf.PowerCycleResetType)
		if err != nil {

			_, _ = c.SystemPowerOff(ctx)

			for wait := 1; wait < 10; wait++ {
				status, _ := c.SystemPowerStatus(ctx)
				if status == "off" {
					break
				}
				time.Sleep(1 * time.Second)
			}

			_, errMsg := c.SystemPowerOn(ctx)

			return true, errMsg
		}
	}
	return true, nil
}

// SystemPowerCycle power cycles the system.
func (c *Client) SystemPowerCycle(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	ss, err := c.Systems()
	if err != nil {
		return false, err
	}

	res, err := c.SystemPowerStatus(ctx)
	if err != nil {
		return false, fmt.Errorf("power cycle failed: unable to get current state")
	}

	if strings.ToLower(res) == "off" {
		return false, fmt.Errorf("power cycle failed: Command not supported in present state: %v", res)
	}

	for _, system := range ss {
		system.DisableEtagMatch(c.disableEtagMatch)

		err = system.Reset(rf.ForceRestartResetType)
		if err != nil {
			return false, errors.WithMessage(err, "power cycle failed")
		}
	}

	return true, nil
}

// SystemPowerStatus returns the system power state.
func (c *Client) SystemPowerStatus(ctx context.Context) (result string, err error) {
	if err := c.SessionActive(); err != nil {
		return result, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	ss, err := c.Systems()
	if err != nil {
		return "", err
	}

	for _, system := range ss {
		return string(system.PowerState), nil
	}

	return "", errors.New("unable to retrieve status")
}

// SystemForceOff powers off the system, without waiting for the OS to shutdown.
func (c *Client) SystemForceOff(ctx context.Context) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	ss, err := c.Systems()
	if err != nil {
		return false, err
	}

	for _, system := range ss {
		if system.PowerState == rf.OffPowerState {
			break
		}

		system.DisableEtagMatch(c.disableEtagMatch)

		err = system.Reset(rf.ForceOffResetType)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Client) SendNMI(_ context.Context) error {
	if err := c.SessionActive(); err != nil {
		return errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	ss, err := c.Systems()
	if err != nil {
		return err
	}

	for _, system := range ss {
		if err = system.Reset(rf.NmiResetType); err != nil {
			return err
		}
	}

	return nil
}
