package redfishwrapper

import (
	"context"

	"github.com/metal-toolbox/bmclib/bmc"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"
)

type bootDeviceMapping struct {
	BootDeviceType bmc.BootDeviceType
	RedFishTarget  rf.BootSourceOverrideTarget
}

var bootDeviceTypeMappings = []bootDeviceMapping{
	{
		BootDeviceType: bmc.BootDeviceTypeBIOS,
		RedFishTarget:  rf.BiosSetupBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeCDROM,
		RedFishTarget:  rf.CdBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeDiag,
		RedFishTarget:  rf.DiagsBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeFloppy,
		RedFishTarget:  rf.FloppyBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeDisk,
		RedFishTarget:  rf.HddBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeNone,
		RedFishTarget:  rf.NoneBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypePXE,
		RedFishTarget:  rf.PxeBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeRemoteDrive,
		RedFishTarget:  rf.RemoteDriveBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeSDCard,
		RedFishTarget:  rf.SDCardBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeUSB,
		RedFishTarget:  rf.UsbBootSourceOverrideTarget,
	},
	{
		BootDeviceType: bmc.BootDeviceTypeUtil,
		RedFishTarget:  rf.UtilitiesBootSourceOverrideTarget,
	},
}

// bootDeviceStringToTarget gets the RedFish BootSourceOverrideTarget that corresponds to the given device string,
// or an error if the device is not a RedFish BootSourceOverrideTarget.
func bootDeviceStringToTarget(device string) (rf.BootSourceOverrideTarget, error) {
	for _, bootDevice := range bootDeviceTypeMappings {
		if string(bootDevice.BootDeviceType) == device {
			return bootDevice.RedFishTarget, nil
		}
	}
	return "", errors.New("invalid boot device")
}

// bootTargetToBootDeviceType converts the redfish boot target to a bmc.BootDeviceType.
// if the target is unknown or unsupported, then an error is returned.
func bootTargetToBootDeviceType(target rf.BootSourceOverrideTarget) (bmc.BootDeviceType, error) {
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

	systems, err := c.Systems()
	if err != nil {
		return false, err
	}

	for _, system := range systems {
		boot := system.Boot

		boot.BootSourceOverrideTarget, err = bootDeviceStringToTarget(bootDevice)
		if err != nil {
			return false, err
		}

		if setPersistent {
			boot.BootSourceOverrideEnabled = rf.ContinuousBootSourceOverrideEnabled
		} else {
			boot.BootSourceOverrideEnabled = rf.OnceBootSourceOverrideEnabled
		}

		if efiBoot {
			boot.BootSourceOverrideMode = rf.UEFIBootSourceOverrideMode
		} else {
			boot.BootSourceOverrideMode = rf.LegacyBootSourceOverrideMode
		}

		if err = system.SetBoot(boot); err != nil {
			// Some redfish implementations don't like all the fields we're setting so we
			// try again here with a minimal set of fields. This has shown to work with the
			// Redfish implementation on HP DL160 Gen10.
			secondTry := rf.Boot{}
			secondTry.BootSourceOverrideTarget = boot.BootSourceOverrideTarget
			secondTry.BootSourceOverrideEnabled = boot.BootSourceOverrideEnabled
			if err = system.SetBoot(secondTry); err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

// GetBootDeviceOverride returns the current boot override settings
func (c *Client) GetBootDeviceOverride(_ context.Context) (override bmc.BootDeviceOverride, err error) {
	if err := c.SessionActive(); err != nil {
		return override, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	systems, err := c.Systems()
	if err != nil {
		return override, err
	}

	for _, system := range systems {
		if system == nil {
			continue
		}

		boot := system.Boot
		bootDevice, err := bootTargetToBootDeviceType(boot.BootSourceOverrideTarget)
		if err != nil {
			return override, err
		}

		override = bmc.BootDeviceOverride{
			IsPersistent: boot.BootSourceOverrideEnabled == rf.ContinuousBootSourceOverrideEnabled,
			IsEFIBoot:    boot.BootSourceOverrideMode == rf.UEFIBootSourceOverrideMode,
			Device:       bootDevice,
		}

		return override, nil
	}

	return override, bmclibErrs.ErrRedfishNoSystems
}
