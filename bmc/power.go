package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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

// setPowerState sets the power state for a BMC, trying all interface implementations passed in
func setPowerState(ctx context.Context, timeout time.Duration, state string, p []powerProviders) (ok bool, m Metadata, err error) {
	metadataLocal := Metadata{
		FailedProviderDetail: make(map[string]string),
	}

	for _, elem := range p {
		if elem.powerSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadataLocal, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ok, setErr := elem.powerSetter.PowerSet(ctx, state)
			if setErr != nil {
				err = multierror.Append(err, errors.WithMessagef(setErr, "provider: %v", elem.name))
				metadataLocal.FailedProviderDetail[elem.name] = setErr.Error()
				continue
			}
			if !ok {
				err = multierror.Append(err, fmt.Errorf("provider: %v, failed to set power state", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to set power state"))
}

// SetPowerStateFromInterfaces identifies implementations of the PostStateSetter interface and passes the found implementations to the setPowerState() wrapper.
func SetPowerStateFromInterfaces(ctx context.Context, timeout time.Duration, state string, generic []interface{}) (ok bool, metadata Metadata, err error) {
	metadata = newMetadata()

	powerSetter := make([]powerProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := powerProviders{name: getProviderName(elem)}
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
		return ok, metadata, multierror.Append(err, errors.New("no PowerSetter implementations found"))
	}
	return setPowerState(ctx, timeout, state, powerSetter)
}

// getPowerState gets the power state for a BMC, trying all interface implementations passed in
func getPowerState(ctx context.Context, timeout time.Duration, p []powerProviders) (state string, m Metadata, err error) {
	metadataLocal := Metadata{
		FailedProviderDetail: make(map[string]string),
	}
	for _, elem := range p {
		if elem.powerStateGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return state, metadataLocal, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			state, stateErr := elem.powerStateGetter.PowerStateGet(ctx)
			if stateErr != nil {
				err = multierror.Append(err, errors.WithMessagef(stateErr, "provider: %v", elem.name))
				metadataLocal.FailedProviderDetail[elem.name] = stateErr.Error()
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return state, metadataLocal, nil
		}
	}
	return state, metadataLocal, multierror.Append(err, errors.New("failed to get power state"))
}

// GetPowerStateFromInterfaces identifies implementations of the PostStateGetter interface and passes the found implementations to the getPowerState() wrapper.
func GetPowerStateFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (state string, metadata Metadata, err error) {
	metadata = newMetadata()

	powerStateGetter := make([]powerProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := powerProviders{name: getProviderName(elem)}
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
		return state, metadata, multierror.Append(err, errors.New("no PowerStateGetter implementations found"))
	}
	return getPowerState(ctx, timeout, powerStateGetter)
}
