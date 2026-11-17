package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

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

// powerStateGetters returns the PowerStateGetter implementations found in
// generic, and an error for each element that is not one.
func powerStateGetters(generic []interface{}) (p []powerProviders, err error) {
	p = make([]powerProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		switch g := elem.(type) {
		case PowerStateGetter:
			p = append(p, powerProviders{name: getProviderName(elem), powerStateGetter: g})
		default:
			e := fmt.Sprintf("not a PowerStateGetter implementation: %T", g)
			err = multierror.Append(err, errors.New(e))
		}
	}

	return p, err
}

// getPowerStateWith gets the power state for a BMC, trying all interface
// implementations passed in. Each provider's raw power state is converted with
// parse; a parse error makes the provider count as failed, so the next
// provider is tried.
func getPowerStateWith[T any](ctx context.Context, timeout time.Duration, p []powerProviders, parse func(raw string) (T, error)) (state T, m Metadata, err error) {
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
			raw, stateErr := elem.powerStateGetter.PowerStateGet(ctx)
			cancel()
			var parsed T
			if stateErr == nil {
				parsed, stateErr = parse(raw)
			}
			if stateErr != nil {
				err = multierror.Append(err, errors.WithMessagef(stateErr, "provider: %v", elem.name))
				metadataLocal.FailedProviderDetail[elem.name] = stateErr.Error()
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return parsed, metadataLocal, nil
		}
	}
	return state, metadataLocal, multierror.Append(err, errors.New("failed to get power state"))
}
