package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// PowerSetter sets the power state of a BMC
type PowerSetter interface {
	PowerSet(ctx context.Context, state string) (ok bool, err error)
}

// PowerStateGetter gets the power state of a BMC
type PowerStateGetter interface {
	PowerStateGet(ctx context.Context) (state string, err error)
}

// SetPowerState sets the power state for a BMC, trying all interface implementations passed in
func SetPowerState(ctx context.Context, state string, p []PowerSetter) (ok bool, err error) {
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
	var powerSetter []PowerSetter
	for _, elem := range generic {
		switch p := elem.(type) {
		case PowerSetter:
			powerSetter = append(powerSetter, p)
		default:
			e := fmt.Sprintf("not a PowerSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(powerSetter) == 0 {
		return ok, multierror.Append(err, errors.New("no PowerSetter implementations found"))
	}
	return SetPowerState(ctx, state, powerSetter)
}

// GetPowerState sets the power state for a BMC, trying all interface implementations passed in
func GetPowerState(ctx context.Context, p []PowerStateGetter) (state string, err error) {
	for _, elem := range p {
		if elem != nil {
			state, stateErr := elem.PowerStateGet(ctx)
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
	var powerStateGetter []PowerStateGetter
	for _, elem := range generic {

		switch p := elem.(type) {
		case PowerStateGetter:
			powerStateGetter = append(powerStateGetter, p)
		default:
			e := fmt.Sprintf("not a PowerStateGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(powerStateGetter) == 0 {
		return state, multierror.Append(err, errors.New("no PowerStateGetter implementations found"))
	}
	return GetPowerState(ctx, powerStateGetter)
}
