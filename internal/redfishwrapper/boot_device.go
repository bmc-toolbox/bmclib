package redfishwrapper

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/schemas"
)

type bootDeviceMapping struct {
	BootDeviceType bmc.BootDeviceType
	RedFishTarget  schemas.BootSource
}

var bootDeviceTypeMappings = []bootDeviceMapping{
	{
		BootDeviceType: bmc.BootDeviceTypeBIOS,
		RedFishTarget:  schemas.BiosSetupBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeCDROM,
		RedFishTarget:  schemas.CdBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeDiag,
		RedFishTarget:  schemas.DiagsBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeFloppy,
		RedFishTarget:  schemas.FloppyBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeDisk,
		RedFishTarget:  schemas.HddBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeNone,
		RedFishTarget:  schemas.NoneBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypePXE,
		RedFishTarget:  schemas.PxeBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeRemoteDrive,
		RedFishTarget:  schemas.RemoteDriveBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeSDCard,
		RedFishTarget:  schemas.SDCardBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeUSB,
		RedFishTarget:  schemas.UsbBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeUtil,
		RedFishTarget:  schemas.UtilitiesBootSource,
	},
	{
		BootDeviceType: bmc.BootDeviceUefiHTTP,
		RedFishTarget:  schemas.UefiHTTPBootSource,
	},
}

// bootDeviceStringToTarget gets the RedFish BootSource that corresponds to the given device string,
// or an error if the device is not a RedFish BootSource.
func bootDeviceStringToTarget(device string) (schemas.BootSource, error) {
	for _, bootDevice := range bootDeviceTypeMappings {
		if string(bootDevice.BootDeviceType) == device {
			return bootDevice.RedFishTarget, nil
		}
	}
	return "", errors.New("invalid boot device")
}

// bootTargetToBootDeviceType converts the redfish boot target to a bmc.BootDeviceType.
// if the target is unknown or unsupported, then an error is returned.
func bootTargetToBootDeviceType(target schemas.BootSource) (bmc.BootDeviceType, error) {
	for _, bootDevice := range bootDeviceTypeMappings {
		if bootDevice.RedFishTarget == target {
			return bootDevice.BootDeviceType, nil
		}
	}
	return "", errors.New("invalid boot device")
}

// SystemBootDeviceSet set the boot device for the system.
func (c *Client) SystemBootDeviceSet(_ context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return false, err
	}

	boot := system.Boot

	boot.BootSourceOverrideTarget, err = bootDeviceStringToTarget(bootDevice)
	if err != nil {
		return false, err
	}

	if setPersistent {
		boot.BootSourceOverrideEnabled = schemas.ContinuousBootSourceOverrideEnabled
	} else {
		boot.BootSourceOverrideEnabled = schemas.OnceBootSourceOverrideEnabled
	}

	if efiBoot {
		boot.BootSourceOverrideMode = schemas.UEFIBootSourceOverrideMode
	} else {
		boot.BootSourceOverrideMode = schemas.LegacyBootSourceOverrideMode
	}

	if err = system.SetBoot(&boot); err != nil {
		// Some redfish implementations don't like all the fields we're setting so we
		// try again here with a minimal set of fields. This has shown to work with the
		// Redfish implementation on HP DL160 Gen10.
		secondTry := schemas.Boot{}
		secondTry.BootSourceOverrideTarget = boot.BootSourceOverrideTarget
		secondTry.BootSourceOverrideEnabled = boot.BootSourceOverrideEnabled
		if err = system.SetBoot(&secondTry); err != nil {
			return false, err
		}
	}

	return true, nil
}

// GetBootDeviceOverride returns the current boot override settings
func (c *Client) GetBootDeviceOverride(_ context.Context) (override bmc.BootDeviceOverride, err error) {
	if err := c.SessionActive(); err != nil {
		return override, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return override, err
	}

	boot := system.Boot
	bootDevice, err := bootTargetToBootDeviceType(boot.BootSourceOverrideTarget)
	if err != nil {
		return override, err
	}

	override = bmc.BootDeviceOverride{
		IsPersistent: boot.BootSourceOverrideEnabled == schemas.ContinuousBootSourceOverrideEnabled,
		IsEFIBoot:    boot.BootSourceOverrideMode == schemas.UEFIBootSourceOverrideMode,
		Device:       bootDevice,
	}

	return override, nil
}
