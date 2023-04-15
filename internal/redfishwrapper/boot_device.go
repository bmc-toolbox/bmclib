package redfishwrapper

import (
	"context"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"
)

// Set the boot device for the system.
func (c *Client) SystemBootDeviceSet(ctx context.Context, bootDevice string, setPersistent, _ bool) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	systems, err := c.client.Service.Systems()
	if err != nil {
		return false, err
	}

	for _, system := range systems {
		boot := rf.Boot{}

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

		boot.BootSourceOverrideEnabled = rf.OnceBootSourceOverrideEnabled
		if setPersistent {
			boot.BootSourceOverrideEnabled = rf.ContinuousBootSourceOverrideEnabled
		}

		if err = system.SetBoot(boot); err != nil {
			return false, err
		}
	}

	return true, nil
}
