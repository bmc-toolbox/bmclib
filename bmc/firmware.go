package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// BMCVersionGetter retrieves the current BMC firmware version information
type BMCVersionGetter interface {
	FirmwareVersionBMC(ctx context.Context) (version string, err error)
}

// BMCFirmwareUpdater upgrades the BMC firmware
type BMCFirmwareUpdater interface {
	FirmwareUpdateBMC(ctx context.Context, fileName string) (err error)
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
				version, vErr := elem.FirmwareVersionBMC(ctx)
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
func UpdateBMCFirmware(ctx context.Context, updateFileName string, p []BMCFirmwareUpdater) (err error) {
Loop:
	for _, elem := range p {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem != nil {
				uErr := elem.FirmwareUpdateBMC(ctx, updateFileName)
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

// GetBMCVersionFromInterfaces pass through to library function
func UpdateBMCFirmwareFromInterfaces(ctx context.Context, updateFileName string, generic []interface{}) (err error) {
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

	return UpdateBMCFirmware(ctx, updateFileName, bmcFirmwareUpdater)
}
