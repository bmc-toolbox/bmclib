package lenovo

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/stmcginnis/gofish/schemas"
)

// BootDeviceSet sets the next boot device by writing the ComputerSystem Boot
// override.
//
// bootDevice is a bmclib boot-device name (e.g. "pxe", "disk", "cdrom",
// "bios"). setPersistent selects BootSourceOverrideEnabled "Continuous" when
// true and "Once" otherwise; efiBoot selects BootSourceOverrideMode "UEFI" when
// true and "Legacy" otherwise. Implements bmc.BootDeviceSetter.
//
// XCC quirk: the BootSourceOverrideEnabled property typically advertises only
// {Once, Disabled} — it rejects "Continuous" (a persistent override) with
// PropertyValueNotInList. XCC expresses persistent boot through the BootOrder,
// not the override. So when a persistent override is requested but the system
// does not advertise "Continuous" as allowable, fail with an actionable error
// instead of the opaque dual ("system resource"/"settings resource") 400 the
// shared wrapper would surface.
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	if setPersistent && !c.bootOverrideAllowsContinuous(ctx) {
		return false, fmt.Errorf(
			"XCC does not support a persistent (Continuous) boot override on this system "+
				"(BootSourceOverrideEnabled allows only Once/Disabled); use a one-time override "+
				"(setPersistent=false) and manage persistent boot via the BootOrder: %w",
			errPersistentBootUnsupported)
	}
	return c.redfishwrapper.SystemBootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
}

// bootOverrideAllowsContinuous reports whether the ComputerSystem advertises
// "Continuous" among Boot.BootSourceOverrideEnabled@Redfish.AllowableValues.
// When the system does not advertise the list at all (field absent), it returns
// true so the write is still attempted (no regression on BMCs that accept
// Continuous without advertising it).
func (c *Conn) bootOverrideAllowsContinuous(ctx context.Context) bool {
	sys, err := c.redfishwrapper.System()
	if err != nil {
		return true // can't tell — don't pre-empt the write
	}

	var doc struct {
		Boot struct {
			AllowableEnabled []string `json:"BootSourceOverrideEnabled@Redfish.AllowableValues"`
		} `json:"Boot"`
	}
	if err := c.getJSON(sys.ODataID, &doc); err != nil {
		return true
	}

	if len(doc.Boot.AllowableEnabled) == 0 {
		return true // not advertised — let the write attempt proceed
	}
	for _, v := range doc.Boot.AllowableEnabled {
		if v == "Continuous" {
			return true
		}
	}
	return false
}

// BootDeviceOverrideGet returns the current boot override (target, persistence,
// UEFI/legacy mode) read from the ComputerSystem Boot object.
//
// Implements bmc.BootDeviceOverrideGetter.
func (c *Conn) BootDeviceOverrideGet(ctx context.Context) (override bmc.BootDeviceOverride, err error) {
	return c.redfishwrapper.GetBootDeviceOverride(ctx)
}

// GetBootProgress returns the BootProgress of the managed system.
//
// XCC reports BootProgress on the ComputerSystem (Redfish >= 1.13.0). Following
// the convention of the other vendor providers, the first system's progress is
// returned. Advertised via providers.FeatureBootProgress.
func (c *Conn) GetBootProgress() (*schemas.BootProgress, error) {
	progress, err := c.redfishwrapper.GetBootProgress()
	if err != nil {
		return nil, err
	}

	if len(progress) == 0 {
		return nil, fmt.Errorf("no boot progress reported by the device")
	}

	return progress[0], nil
}
