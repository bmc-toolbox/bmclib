package dell

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// BootDeviceSet sets the next boot device by writing the ComputerSystem Boot
// override.
//
// bootDevice is a bmclib boot-device name (e.g. "pxe", "disk", "cdrom",
// "bios"). setPersistent selects BootSourceOverrideEnabled "Continuous" when
// true and "Once" otherwise; efiBoot selects BootSourceOverrideMode "UEFI"
// when true and "Legacy" otherwise. Implements bmc.BootDeviceSetter.
//
// On iDRAC10, Boot is read-only on the main ComputerSystem resource and must
// be set via its @Redfish.Settings resource instead; on iDRAC9 and earlier,
// Boot is writable directly and the Settings resource only advertises a
// subset of Boot properties (observed: BootSourceOverrideMode, not Target or
// Enabled). As of gofish v0.22.0, the underlying SetBoot call tries the main
// resource first and falls back to Settings on rejection, which handles both
// generations without a version check here.
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	return c.redfishwrapper.SystemBootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
}

// BootDeviceOverrideGet returns the current boot override (target,
// persistence, UEFI/legacy mode) read from the ComputerSystem Boot object.
//
// Implements bmc.BootDeviceOverrideGetter.
func (c *Conn) BootDeviceOverrideGet(ctx context.Context) (override bmc.BootDeviceOverride, err error) {
	return c.redfishwrapper.GetBootDeviceOverride(ctx)
}
