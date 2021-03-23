package bmc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/hashicorp/go-multierror"
)

// BMCVersionGetter retrieves the current BMC firmware version information
type BMCVersionGetter interface {
	GetBMCVersion(ctx context.Context) (version string, err error)
}

// BMCFirmwareUpdater upgrades the BMC firmware
type BMCFirmwareUpdater interface {
	FirmwareUpdateBMC(ctx context.Context, fileReader io.Reader, fileSize int64) (err error)
}

// BIOSVersionGetter retrieves the current BIOS firmware version information
type BIOSVersionGetter interface {
	GetBIOSVersion(ctx context.Context) (version string, err error)
}

// BIOSFirmwareUpdater upgrades the BIOS firmware
type BIOSFirmwareUpdater interface {
	FirmwareUpdateBIOS(ctx context.Context, fileReader io.Reader, fileSize int64) (err error)
}

// GetBMCVersion returns the BMC firmware version, trying all interface implementations passed in
func GetBMCVersion(ctx context.Context, p []BMCVersionGetter) (version string, err error) {
Loop:
	for _, elem := range p {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem != nil {
				version, vErr := elem.GetBMCVersion(ctx)
				if vErr != nil {
					err = multierror.Append(err, vErr)
					continue
				}
				return version, nil
			}
		}
	}

	return version, multierror.Append(err, errors.New("failed to get BMC version"))
}

// GetBMCVersionFromInterfaces pass through to library function
func GetBMCVersionFromInterfaces(ctx context.Context, generic []interface{}) (version string, err error) {
	bmcVersionGetter := make([]BMCVersionGetter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case BMCVersionGetter:
			bmcVersionGetter = append(bmcVersionGetter, p)
		default:
			e := fmt.Sprintf("not a BMCVersionGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(bmcVersionGetter) == 0 {
		return version, multierror.Append(err, errors.New("no BMCVersionGetter implementations found"))
	}

	return GetBMCVersion(ctx, bmcVersionGetter)
}

// UpdateBMCFirmware upgrades the BMC firmware, trying all interface implementations passed ini
func UpdateBMCFirmware(ctx context.Context, fileReader io.Reader, fileSize int64, p []BMCFirmwareUpdater) (err error) {
Loop:
	for _, elem := range p {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem != nil {
				uErr := elem.FirmwareUpdateBMC(ctx, fileReader, fileSize)
				if uErr != nil {
					err = multierror.Append(err, uErr)
					continue
				}
				return nil
			}
		}
	}

	return multierror.Append(err, errors.New("failed to update BMC firmware"))

}

// UpdateBMCFirmwareFromInterfaces pass through to library function
func UpdateBMCFirmwareFromInterfaces(ctx context.Context, fileReader io.Reader, fileSize int64, generic []interface{}) (err error) {
	bmcFirmwareUpdater := make([]BMCFirmwareUpdater, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case BMCFirmwareUpdater:
			bmcFirmwareUpdater = append(bmcFirmwareUpdater, p)
		default:
			e := fmt.Sprintf("not a BMCFirmwareUpdater implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(bmcFirmwareUpdater) == 0 {
		return multierror.Append(err, errors.New("no BMCFirmwareUpdater implementations found"))
	}

	return UpdateBMCFirmware(ctx, fileReader, fileSize, bmcFirmwareUpdater)
}

// GetBIOSVersion returns the BMC firmware version, trying all interface implementations passed in
func GetBIOSVersion(ctx context.Context, p []BIOSVersionGetter) (version string, err error) {
Loop:
	for _, elem := range p {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem != nil {
				version, vErr := elem.GetBIOSVersion(ctx)
				if vErr != nil {
					err = multierror.Append(err, vErr)
					continue
				}
				return version, nil
			}
		}
	}

	return version, multierror.Append(err, errors.New("failed to get BIOS version"))
}

// GetBIOSVersionFromInterfaces pass through to library function
func GetBIOSVersionFromInterfaces(ctx context.Context, generic []interface{}) (version string, err error) {
	biosVersionGetter := make([]BIOSVersionGetter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case BIOSVersionGetter:
			biosVersionGetter = append(biosVersionGetter, p)
		default:
			e := fmt.Sprintf("not a BIOSVersionGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(biosVersionGetter) == 0 {
		return version, multierror.Append(err, errors.New("no BIOSVersionGetter implementations found"))
	}

	return GetBIOSVersion(ctx, biosVersionGetter)
}

// UpdateBIOSFirmware upgrades the BIOS firmware, trying all interface implementations passed ini
func UpdateBIOSFirmware(ctx context.Context, fileReader io.Reader, fileSize int64, p []BIOSFirmwareUpdater) (err error) {
Loop:
	for _, elem := range p {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem != nil {
				uErr := elem.FirmwareUpdateBIOS(ctx, fileReader, fileSize)
				if uErr != nil {
					err = multierror.Append(err, uErr)
					continue
				}
				return nil
			}
		}
	}

	return multierror.Append(err, errors.New("failed to update BIOS firmware"))

}

// GetBMCVersionFromInterfaces pass through to library function
func UpdateBIOSFirmwareFromInterfaces(ctx context.Context, fileReader io.Reader, fileSize int64, generic []interface{}) (err error) {
	biosFirmwareUpdater := make([]BIOSFirmwareUpdater, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case BIOSFirmwareUpdater:
			biosFirmwareUpdater = append(biosFirmwareUpdater, p)
		default:
			e := fmt.Sprintf("not a BIOSFirmwareUpdater implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(biosFirmwareUpdater) == 0 {
		return multierror.Append(err, errors.New("no BIOSFirmwareUpdater implementations found"))
	}

	return UpdateBIOSFirmware(ctx, fileReader, fileSize, biosFirmwareUpdater)
}
