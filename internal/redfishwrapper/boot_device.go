package redfishwrapper

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"
)

type bootDeviceType struct {
	Name          string
	RedFishTarget rf.BootSourceOverrideTarget
}

var bootDeviceTypes = []bootDeviceType{
	{
		Name:          "bios",
		RedFishTarget: rf.BiosSetupBootSourceOverrideTarget,
	},
	{
		Name:          "cdrom",
		RedFishTarget: rf.CdBootSourceOverrideTarget,
	},
	{
		Name:          "diag",
		RedFishTarget: rf.DiagsBootSourceOverrideTarget,
	},
	{
		Name:          "floppy",
		RedFishTarget: rf.FloppyBootSourceOverrideTarget,
	},
	{
		Name:          "disk",
		RedFishTarget: rf.HddBootSourceOverrideTarget,
	},
	{
		Name:          "none",
		RedFishTarget: rf.NoneBootSourceOverrideTarget,
	},
	{
		Name:          "pxe",
		RedFishTarget: rf.PxeBootSourceOverrideTarget,
	},
	{
		Name:          "remote_drive",
		RedFishTarget: rf.RemoteDriveBootSourceOverrideTarget,
	},
	{
		Name:          "sd_card",
		RedFishTarget: rf.SDCardBootSourceOverrideTarget,
	},
	{
		Name:          "usb",
		RedFishTarget: rf.UsbBootSourceOverrideTarget,
	},
	{
		Name:          "utilities",
		RedFishTarget: rf.UtilitiesBootSourceOverrideTarget,
	},
}

// bootDeviceToTarget gets the RedFish BootSourceOverrideTarget that corresponds to the given device,
// or an error if the device is not a RedFish BootSourceOverrideTarget.
func bootDeviceToTarget(device string) (rf.BootSourceOverrideTarget, error) {
	for _, bootDevice := range bootDeviceTypes {
		if bootDevice.Name == device {
			return bootDevice.RedFishTarget, nil
		}
	}
	return "", errors.New("invalid boot device")
}

// bootTargetToDevice converts the redfish boot target to a bmclib supported device string.
// if the target is unknown or unsupported, then an error is returned.
func bootTargetToDevice(target rf.BootSourceOverrideTarget) (string, error) {
	for _, bootDevice := range bootDeviceTypes {
		if bootDevice.RedFishTarget == target {
			return bootDevice.Name, nil
		}
	}
	return "", errors.New("invalid boot device")
}

// SystemBootDeviceSet set the boot device for the system.
func (c *Client) SystemBootDeviceSet(_ context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	systems, err := c.client.Service.Systems()
	if err != nil {
		return false, err
	}

	for _, system := range systems {
		boot := system.Boot

		boot.BootSourceOverrideTarget, err = bootDeviceToTarget(bootDevice)
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

	systems, err := c.client.Service.Systems()
	if err != nil {
		return override, err
	}

	for _, system := range systems {
		if system == nil {
			continue
		}

		boot := system.Boot
		bootDevice, err := bootTargetToDevice(boot.BootSourceOverrideTarget)
		if err != nil {
			bootDevice = string(boot.BootSourceOverrideTarget)
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
