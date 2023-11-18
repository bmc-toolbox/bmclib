package redfishwrapper

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"
)

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

		switch bootDevice {
		case "bios":
			boot.BootSourceOverrideTarget = rf.BiosSetupBootSourceOverrideTarget
		case "cdrom":
			boot.BootSourceOverrideTarget = rf.CdBootSourceOverrideTarget
		case "diag":
			boot.BootSourceOverrideTarget = rf.DiagsBootSourceOverrideTarget
		case "floppy":
			boot.BootSourceOverrideTarget = rf.FloppyBootSourceOverrideTarget
		case "disk":
			boot.BootSourceOverrideTarget = rf.HddBootSourceOverrideTarget
		case "none":
			boot.BootSourceOverrideTarget = rf.NoneBootSourceOverrideTarget
		case "pxe":
			boot.BootSourceOverrideTarget = rf.PxeBootSourceOverrideTarget
		case "remote_drive":
			boot.BootSourceOverrideTarget = rf.RemoteDriveBootSourceOverrideTarget
		case "sd_card":
			boot.BootSourceOverrideTarget = rf.SDCardBootSourceOverrideTarget
		case "usb":
			boot.BootSourceOverrideTarget = rf.UsbBootSourceOverrideTarget
		case "utilities":
			boot.BootSourceOverrideTarget = rf.UtilitiesBootSourceOverrideTarget
		default:
			return false, errors.New("invalid boot device")
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

// bootTargetToDevice tries to convert the redfish boot target to a bmclib supported device string.
// if the target is unknown or unsupported, then a string of the target is returned.
func bootTargetToDevice(target rf.BootSourceOverrideTarget) (device string) {
	switch target {
	case rf.BiosSetupBootSourceOverrideTarget:
		device = "bios"
	case rf.CdBootSourceOverrideTarget:
		device = "cdrom"
	case rf.DiagsBootSourceOverrideTarget:
		device = "diag"
	case rf.FloppyBootSourceOverrideTarget:
		device = "floppy"
	case rf.HddBootSourceOverrideTarget:
		device = "disk"
	case rf.NoneBootSourceOverrideTarget:
		device = "none"
	case rf.PxeBootSourceOverrideTarget:
		device = "pxe"
	case rf.RemoteDriveBootSourceOverrideTarget:
		device = "remote_drive"
	case rf.SDCardBootSourceOverrideTarget:
		device = "sd_card"
	case rf.UsbBootSourceOverrideTarget:
		device = "usb"
	case rf.UtilitiesBootSourceOverrideTarget:
		device = "utilities"
	default:
		device = string(target)
	}

	return device
}

// GetBootDeviceOverride returns the current boot override settings
func (c *Client) GetBootDeviceOverride(_ context.Context) (*bmc.BootDeviceOverride, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	systems, err := c.client.Service.Systems()
	if err != nil {
		return nil, err
	}

	for _, system := range systems {
		if system == nil {
			continue
		}

		boot := system.Boot
		override := &bmc.BootDeviceOverride{
			IsPersistent: boot.BootSourceOverrideEnabled == rf.ContinuousBootSourceOverrideEnabled,
			IsEFIBoot:    boot.BootSourceOverrideMode == rf.UEFIBootSourceOverrideMode,
			Device:       bootTargetToDevice(boot.BootSourceOverrideTarget),
		}

		return override, nil
	}

	return nil, bmclibErrs.ErrNoSystemsAvailable
}
