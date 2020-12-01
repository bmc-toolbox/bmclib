package ipmitool

import (
	"context"
	"errors"
	"strings"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/go-logr/logr"
)

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context, log logr.Logger) (state string, err error) {
	i, err := ipmi.New(c.User, c.Pass, c.Host+":"+c.Port)
	if err != nil {
		return state, err
	}
	return i.PowerState(ctx)
}

// PowerSet sets the power state of a BMC machine
func (c *Conn) PowerSet(ctx context.Context, log logr.Logger, state string) (ok bool, err error) {
	i, err := ipmi.New(c.User, c.Pass, c.Host+":"+c.Port)
	if err != nil {
		return ok, err
	}

	switch strings.ToLower(state) {
	case "on":
		on, _ := i.IsOn(ctx)
		if on {
			ok = true
		} else {
			ok, err = i.PowerOn(ctx)
		}
	case "off":
		on, _ := i.IsOn(ctx)
		if !on {
			ok = true
		} else {
			ok, err = i.PowerOff(ctx)
		}
	case "soft":
		ok, err = i.PowerSoft(ctx)
	case "reset":
		ok, err = i.PowerReset(ctx)
	case "cycle":
		ok, err = i.PowerCycle(ctx)
	default:
		err = errors.New("requested state type unknown")
	}

	return ok, err
}
