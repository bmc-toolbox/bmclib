package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

type BootDeviceType string

const (
	BootDeviceTypeBIOS        BootDeviceType = "bios"
	BootDeviceTypeCDROM       BootDeviceType = "cdrom"
	BootDeviceTypeDiag        BootDeviceType = "diag"
	BootDeviceTypeFloppy      BootDeviceType = "floppy"
	BootDeviceTypeDisk        BootDeviceType = "disk"
	BootDeviceTypeNone        BootDeviceType = "none"
	BootDeviceTypePXE         BootDeviceType = "pxe"
	BootDeviceTypeRemoteDrive BootDeviceType = "remote_drive"
	BootDeviceTypeSDCard      BootDeviceType = "sd_card"
	BootDeviceTypeUSB         BootDeviceType = "usb"
	BootDeviceTypeUtil        BootDeviceType = "utilities"
)

// BootDeviceSetter sets the next boot device for a machine
type BootDeviceSetter interface {
	BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error)
}

// BootDeviceOverrideGetter gets boot override settings for a machine
type BootDeviceOverrideGetter interface {
	BootDeviceOverrideGet(ctx context.Context) (override BootDeviceOverride, err error)
}

// bootDeviceProviders is an internal struct to correlate an implementation/provider and its name
type bootDeviceProviders struct {
	name             string
	bootDeviceSetter BootDeviceSetter
}

// bootOverrideProvider is an internal struct to correlate an implementation/provider and its name
type bootOverrideProvider struct {
	name               string
	bootOverrideGetter BootDeviceOverrideGetter
}

type BootDeviceOverride struct {
	IsPersistent bool
	IsEFIBoot    bool
	Device       BootDeviceType
}

// setBootDevice sets the next boot device.
//
// setPersistent persists the next boot device.
// efiBoot sets up the device to boot off UEFI instead of legacy.
func setBootDevice(ctx context.Context, timeout time.Duration, bootDevice string, setPersistent, efiBoot bool, b []bootDeviceProviders) (ok bool, metadata Metadata, err error) {
	metadataLocal := newMetadata()

	for _, elem := range b {
		if elem.bootDeviceSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ok, setErr := elem.bootDeviceSetter.BootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
			if setErr != nil {
				err = multierror.Append(err, errors.WithMessagef(setErr, "provider: %v", elem.name))
				metadataLocal.FailedProviderDetail[elem.name] = setErr.Error()
				continue
			}
			if !ok {
				err = multierror.Append(err, fmt.Errorf("provider: %v, failed to set boot device", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to set boot device"))
}

// SetBootDeviceFromInterfaces identifies implementations of the BootDeviceSetter interface and passes the found implementations to the setBootDevice() wrapper
func SetBootDeviceFromInterfaces(ctx context.Context, timeout time.Duration, bootDevice string, setPersistent, efiBoot bool, generic []interface{}) (ok bool, metadata Metadata, err error) {
	bdSetters := make([]bootDeviceProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := bootDeviceProviders{name: getProviderName(elem)}
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
		return ok, metadata, multierror.Append(err, errors.New("no BootDeviceSetter implementations found"))
	}
	return setBootDevice(ctx, timeout, bootDevice, setPersistent, efiBoot, bdSetters)
}

// getBootDeviceOverride gets the boot device override settings for the given provider,
// and updates the given metadata with provider attempts and errors.
func getBootDeviceOverride(
	ctx context.Context,
	timeout time.Duration,
	provider *bootOverrideProvider,
	metadata *Metadata,
) (override BootDeviceOverride, ok bool, err error) {
	select {
	case <-ctx.Done():
		err = multierror.Append(err, ctx.Err())
		return override, ok, err
	default:
		metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, provider.name)
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		override, err = provider.bootOverrideGetter.BootDeviceOverrideGet(ctx)
		if err != nil {
			metadata.FailedProviderDetail[provider.name] = err.Error()
			return override, ok, nil
		}

		metadata.SuccessfulProvider = provider.name
		return override, true, nil
	}
}

// GetBootDeviceOverrideFromInterface will get boot device override settings from the first successful
// call to a BootDeviceOverrideGetter in the array of providers.
func GetBootDeviceOverrideFromInterface(
	ctx context.Context,
	timeout time.Duration,
	providers []interface{},
) (override BootDeviceOverride, metadata Metadata, err error) {
	metadata = newMetadata()

	for _, elem := range providers {
		if elem == nil {
			continue
		}
		switch p := elem.(type) {
		case BootDeviceOverrideGetter:
			provider := &bootOverrideProvider{name: getProviderName(elem), bootOverrideGetter: p}
			override, ok, getErr := getBootDeviceOverride(ctx, timeout, provider, &metadata)
			if getErr != nil || ok {
				return override, metadata, getErr
			}
		default:
			e := fmt.Errorf("not a BootDeviceOverrideGetter implementation: %T", p)
			err = multierror.Append(err, e)
		}
	}

	if len(metadata.ProvidersAttempted) == 0 {
		err = multierror.Append(err, errors.New("no BootDeviceOverrideGetter implementations found"))
	} else {
		err = multierror.Append(err, errors.New("failed to get boot device override settings"))
	}

	return override, metadata, err
}
