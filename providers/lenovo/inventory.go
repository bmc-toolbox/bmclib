package lenovo

import (
	"context"

	"github.com/bmc-toolbox/common"
)

// Inventory collects the hardware and firmware inventory of the XCC-managed
// system into a *common.Device: system/board identity, memory, processors
// (CPU/GPU), PCIe devices, network adapters, storage controllers/drives, and
// the firmware versions reported by UpdateService/FirmwareInventory.
//
// When the connection's failInventoryOnError is false (the default, set via
// [WithFailInventoryOnError]), a failure reading one sub-resource does not abort
// the whole inventory — the provider returns what it could collect. When true,
// the first sub-resource error is returned.
//
// Implements bmc.InventoryGetter.
func (c *Conn) Inventory(ctx context.Context) (device *common.Device, err error) {
	return c.redfishwrapper.Inventory(ctx, c.failInventoryOnError)
}
