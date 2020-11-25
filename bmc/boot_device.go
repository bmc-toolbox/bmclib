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

// SetBootDevice sets the boot device. Next boot only unless setPersistent=true
func SetBootDevice(ctx context.Context, bootDevice string, setPersistent, efiBoot bool, b []BootDeviceSetter) (ok bool, err error) {
	for _, elem := range b {
		if elem != nil {
			ok, setErr := elem.BootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
			if setErr != nil {
				err = multierror.Append(err, setErr)
				continue
			}
			if !ok {
				err = multierror.Append(err, errors.New("failed to set boot device"))
				continue
			}
			return ok, err
		}
	}
	return ok, multierror.Append(err, errors.New("failed to set boot device"))
}

// SetBootDeviceFromInterfaces pass through to library function
func SetBootDeviceFromInterfaces(ctx context.Context, bootDevice string, setPersistent, efiBoot bool, generic []interface{}) (ok bool, err error) {
	var bdSetters []BootDeviceSetter
	for _, elem := range generic {
		switch p := elem.(type) {
		case BootDeviceSetter:
			bdSetters = append(bdSetters, p)
		default:
			e := fmt.Sprintf("not a BootDeviceSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(bdSetters) == 0 {
		return ok, multierror.Append(err, errors.New("no BootDeviceSetter implementations found"))
	}
	return SetBootDevice(ctx, bootDevice, setPersistent, efiBoot, bdSetters)
}
