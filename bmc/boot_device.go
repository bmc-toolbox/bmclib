package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// BootDeviceSetter sets the next boot device for a machine
type BootDeviceSetter interface {
	BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error)
}

// powerProviders is an internal struct to correlate an implementation/provider and its name
type bootDeviceProviders struct {
	name             string
	bootDeviceSetter BootDeviceSetter
}

// SetBootDevice sets the boot device. Next boot only unless setPersistent=true
// if a successfulProviderName is passed in, it will be updated to be the name of the provider that successfully executed
func SetBootDevice(ctx context.Context, bootDevice string, setPersistent, efiBoot bool, b []bootDeviceProviders, successfulProviderName ...*string) (ok bool, err error) {
Loop:
	for _, elem := range b {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem.bootDeviceSetter != nil {
				ok, setErr := elem.bootDeviceSetter.BootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
				if setErr != nil {
					err = multierror.Append(err, setErr)
					continue
				}
				if !ok {
					err = multierror.Append(err, errors.New("failed to set boot device"))
					continue
				}
				if len(successfulProviderName) > 0 && successfulProviderName[0] != nil {
					*successfulProviderName[0] = elem.name
				}
				return ok, nil
			}
		}
	}
	return ok, multierror.Append(err, errors.New("failed to set boot device"))
}

// SetBootDeviceFromInterfaces pass through to library function
// if a successfulProviderName is passed in, it will be updated to be the name of the provider that successfully executed
func SetBootDeviceFromInterfaces(ctx context.Context, bootDevice string, setPersistent, efiBoot bool, generic []interface{}, successfulProviderName ...*string) (ok bool, err error) {
	bdSetters := make([]bootDeviceProviders, 0)
	for _, elem := range generic {
		var temp bootDeviceProviders
		switch p := elem.(type) {
		case Provider:
			temp.name = p.Name()
		}
		switch p := elem.(type) {
		case BootDeviceSetter:
			temp.bootDeviceSetter = p
			bdSetters = append(bdSetters, temp)
		default:
			e := fmt.Sprintf("not a BootDeviceSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(bdSetters) == 0 {
		return ok, multierror.Append(err, errors.New("no BootDeviceSetter implementations found"))
	}
	return SetBootDevice(ctx, bootDevice, setPersistent, efiBoot, bdSetters, successfulProviderName...)
}
