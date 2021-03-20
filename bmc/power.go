package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// PowerSetter sets the power state of a BMC
type PowerSetter interface {
	// PowerSet sets the power state of a Machine through a BMC.
	// While the state's accepted are ultimately up to what the implementation
	// expects, implementations should generally try to provide support for the following
	// states, modeled after the functionality available in `ipmitool chassis power`.
	//
	// "on": Power up chassis. should not error if the machine is already on
	// "off": Hard powers down chassis. should not error if the machine is already off
	// "soft": Initiate a soft-shutdown of OS via ACPI.
	// "reset": soft down and then power on. simulates a reboot from the host OS.
	// "cycle": hard power down followed by a power on. simulates pressing a power button
	// to turn the machine off then pressing the button again to turn it on.
	PowerSet(ctx context.Context, state string) (ok bool, err error)
}

// PowerStateGetter gets the power state of a BMC
type PowerStateGetter interface {
	PowerStateGet(ctx context.Context) (state string, err error)
}

// powerProviders is an internal struct to correlate an implementation/provider and its name
type powerProviders struct {
	name             string
	powerStateGetter PowerStateGetter
	powerSetter      PowerSetter
}

// SetPowerState sets the power state for a BMC, trying all interface implementations passed in
// if a successfulProviderName is passed in, it will be updated to be the name of the provider that successfully executed
func SetPowerState(ctx context.Context, state string, p []powerProviders, successfulProviderName ...*string) (ok bool, err error) {
Loop:
	for _, elem := range p {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem.powerSetter != nil {
				ok, setErr := elem.powerSetter.PowerSet(ctx, state)
				if setErr != nil {
					err = multierror.Append(err, setErr)
					continue
				}
				if !ok {
					err = multierror.Append(err, errors.New("failed to set power state"))
					continue
				}
				if len(successfulProviderName) > 0 && successfulProviderName[0] != nil {
					*successfulProviderName[0] = elem.name
				}
				return ok, nil
			}
		}
	}
	return ok, multierror.Append(err, errors.New("failed to set power state"))
}

// SetPowerStateFromInterfaces pass through to library function
// if a successfulProviderName is passed in, it will be updated to be the name of the provider that successfully executed
func SetPowerStateFromInterfaces(ctx context.Context, state string, generic []interface{}, successfulProviderName ...*string) (ok bool, err error) {
	powerSetter := make([]powerProviders, 0)
	for _, elem := range generic {
		var temp powerProviders
		switch p := elem.(type) {
		case Provider:
			temp.name = p.Name()
		}
		switch p := elem.(type) {
		case PowerSetter:
			temp.powerSetter = p
			powerSetter = append(powerSetter, temp)
		default:
			e := fmt.Sprintf("not a PowerSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(powerSetter) == 0 {
		return ok, multierror.Append(err, errors.New("no PowerSetter implementations found"))
	}
	return SetPowerState(ctx, state, powerSetter, successfulProviderName...)
}

// GetPowerState sets the power state for a BMC, trying all interface implementations passed in
// if a successfulProviderName is passed in, it will be updated to be the name of the provider that successfully executed
func GetPowerState(ctx context.Context, p []powerProviders, successfulProviderName ...*string) (state string, err error) {
Loop:
	for _, elem := range p {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem.powerStateGetter != nil {
				state, stateErr := elem.powerStateGetter.PowerStateGet(ctx)
				if stateErr != nil {
					err = multierror.Append(err, stateErr)
					continue
				}
				if len(successfulProviderName) > 0 && successfulProviderName[0] != nil {
					*successfulProviderName[0] = elem.name
				}
				return state, nil
			}
		}
	}
	return state, multierror.Append(err, errors.New("failed to get power state"))
}

// GetPowerStateFromInterfaces pass through to library function
// if a successfulProviderName is passed in, it will be updated to be the name of the provider that successfully executed
func GetPowerStateFromInterfaces(ctx context.Context, generic []interface{}, successfulProviderName ...*string) (state string, err error) {
	powerStateGetter := make([]powerProviders, 0)
	for _, elem := range generic {
		var temp powerProviders
		switch p := elem.(type) {
		case Provider:
			temp.name = p.Name()
		}
		switch p := elem.(type) {
		case PowerStateGetter:
			temp.powerStateGetter = p
			powerStateGetter = append(powerStateGetter, temp)
		default:
			e := fmt.Sprintf("not a PowerStateGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(powerStateGetter) == 0 {
		return state, multierror.Append(err, errors.New("no PowerStateGetter implementations found"))
	}
	return GetPowerState(ctx, powerStateGetter, successfulProviderName...)
}
