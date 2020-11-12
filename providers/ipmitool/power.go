package ipmitool

import (
	"context"
	"errors"
	"strings"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
)

// PowerState gets the power state of a BMC machine
func (c *Conn) PowerState(ctx context.Context) (state string, err error) {
	i, err := ipmi.New(c.User, c.Pass, c.Host)
	if err != nil {
		return state, err
	}
	result, err := i.IsOn(ctx)
	if err != nil {
		return state, err
	}
	if result {
		state = "on"
	} else {
		state = "off"
	}
	return state, err
}

// PowerSet sets the power state of a BMC machine
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	i, err := ipmi.New(c.User, c.Pass, c.Host)
	if err != nil {
		return ok, err
	}
	switch strings.ToLower(state) {
	case "on":
		ok, err = i.PowerOn()
	case "off":
		ok, err = i.PowerOff()
	case "cycle":
		ok, err = i.PowerCycle()
	default:
		err = errors.New("unknown state request")
	}

	return ok, err
}
