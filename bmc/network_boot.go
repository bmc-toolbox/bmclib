package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// NetworkBootEnabledSetter enables/disables UEFI HTTP Boot and/or legacy PXE boot capability in
// BIOS/UEFI firmware. Unlike BootDeviceSetter (which selects among existing boot options), this
// creates or removes the boot options themselves. httpEnabled and pxeEnabled are independent: a
// nil pointer leaves that protocol's state untouched, so either, both, or neither may be set.
type NetworkBootEnabledSetter interface {
	SetNetworkBootEnabled(ctx context.Context, httpEnabled, pxeEnabled *bool) (ok bool, err error)
}

// networkBootEnabledProviders is an internal struct to correlate an implementation/provider and its name
type networkBootEnabledProviders struct {
	name                     string
	networkBootEnabledSetter NetworkBootEnabledSetter
}

// setNetworkBootEnabled sets the network boot enabled state.
func setNetworkBootEnabled(ctx context.Context, httpEnabled, pxeEnabled *bool, b []networkBootEnabledProviders) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range b {
		if elem.networkBootEnabledSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ok, setErr := elem.networkBootEnabledSetter.SetNetworkBootEnabled(ctx, httpEnabled, pxeEnabled)
			if setErr != nil {
				err = multierror.Append(err, errors.WithMessagef(setErr, "provider: %v", elem.name))
				continue
			}
			if !ok {
				err = multierror.Append(err, fmt.Errorf("provider: %v, failed to set network boot enabled state", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to set network boot enabled state"))
}

// SetNetworkBootEnabledFromInterfaces identifies implementations of the NetworkBootEnabledSetter interface and passes the found implementations to the setNetworkBootEnabled() wrapper
func SetNetworkBootEnabledFromInterfaces(ctx context.Context, httpEnabled, pxeEnabled *bool, generic []interface{}) (ok bool, metadata Metadata, err error) {
	setters := make([]networkBootEnabledProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := networkBootEnabledProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case NetworkBootEnabledSetter:
			temp.networkBootEnabledSetter = p
			setters = append(setters, temp)
		default:
			e := fmt.Sprintf("not a NetworkBootEnabledSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(setters) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no NetworkBootEnabledSetter implementations found"))
	}
	return setNetworkBootEnabled(ctx, httpEnabled, pxeEnabled, setters)
}
