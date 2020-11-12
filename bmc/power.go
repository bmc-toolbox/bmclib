package bmc

import (
	"context"
	"errors"

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
	return ok, multierror.Append(err, errors.New("failed to set power state"))
}

// GetPowerState sets the power state for a BMC, trying all interface implementations passed in
func GetPowerState(ctx context.Context, p []PowerStateSetter) (state string, err error) {
	for _, elem := range p {
		state, stateErr := elem.PowerState(ctx)
		if stateErr != nil {
			err = multierror.Append(err, stateErr)
			continue
		}
		return state, err
	}

	return state, multierror.Append(err, errors.New("failed to get power state"))
}
