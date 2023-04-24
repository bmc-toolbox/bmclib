package redfishwrapper

import (
	"context"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"
)

// Set the boot device for the system.
func (c *Client) SystemBootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
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
