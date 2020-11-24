package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// PowerStateSetter get power state and set power state
type PowerStateSetter interface {
	PowerState(ctx context.Context) (state string, err error)
	PowerSet(ctx context.Context, state string) (ok bool, err error)
}

// SetPowerState sets the power state for a BMC, trying all interface implementations passed in
func SetPowerState(ctx context.Context, state string, p []PowerStateSetter) (ok bool, err error) {
	for _, elem := range p {
		if elem != nil {
			ok, setErr := elem.PowerSet(ctx, state)
			if setErr != nil {
				err = multierror.Append(err, setErr)
				continue
			}
			if !ok {
				err = multierror.Append(err, errors.New("failed to set power state"))
				continue
			}
			return ok, err
		}
	}
	return ok, multierror.Append(err, errors.New("failed to set power state"))
}

// SetPowerStateFromInterfaces pass through to library function
func SetPowerStateFromInterfaces(ctx context.Context, state string, generic []interface{}) (ok bool, err error) {
	var powerStateSetters []PowerStateSetter
	for _, elem := range generic {
		switch p := elem.(type) {
		case PowerStateSetter:
			powerStateSetters = append(powerStateSetters, p)
		default:
		}
	}
	if len(powerStateSetters) == 0 {
		return ok, errors.New("no PowerStateSetter implementations found")
	}
	return SetPowerState(ctx, state, powerStateSetters)
}

// GetPowerState sets the power state for a BMC, trying all interface implementations passed in
func GetPowerState(ctx context.Context, p []PowerStateSetter) (state string, err error) {
	for _, elem := range p {
		if elem != nil {
			state, stateErr := elem.PowerState(ctx)
			if stateErr != nil {
				err = multierror.Append(err, stateErr)
				continue
			}
			return state, err
		}
	}

	return state, multierror.Append(err, errors.New("failed to get power state"))
}

// GetPowerStateFromInterfaces pass through to library function
func GetPowerStateFromInterfaces(ctx context.Context, generic []interface{}) (state string, err error) {
	var powerStateSetters []PowerStateSetter
	for _, elem := range generic {
		switch p := elem.(type) {
		case PowerStateSetter:
			powerStateSetters = append(powerStateSetters, p)
		default:
			e := fmt.Sprintf("not a PowerStateSetter implementation: %T", p)
			err = multierror.Append(errors.New(e))
		}
	}
	if len(powerStateSetters) == 0 {
		return state, multierror.Append(err, errors.New("no PowerStateSetter implementations found"))
	}
	return GetPowerState(ctx, powerStateSetters)
}
