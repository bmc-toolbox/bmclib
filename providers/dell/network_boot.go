package dell

import (
	"context"
	"errors"
	"fmt"
)

// networkBootDeviceIndex is the Dell "device" slot (HttpDevN / PxeDevN in the BIOS attribute
// registry) this file manages. Dell models HTTP Boot and PXE as up to 4 and 16 independently
// enable/disable-able per-NIC device slots respectively (HttpDev1..HttpDev4,
// PxeDev1..PxeDev16), each bound to a NIC FQDD via HttpDevNInterface/PxeDevNInterface - unlike
// Supermicro/AMI Aptio's single global attribute pair. There is no field in bmc.NetworkBootConfig
// to select a NIC, so this always targets slot 1, the primary boot device slot - confirmed live
// on a PowerEdge R6715 (iDRAC firmware 1.5.3) where both HttpDev1Interface and PxeDev1Interface
// are already bound to the same NIC (NIC.Slot.5-1-1).
const networkBootDeviceIndex = 1

const (
	attrEnabled  = "Enabled"
	attrDisabled = "Disabled"
)

// SetNetworkBootEnabled enables/disables UEFI HTTP Boot and/or legacy PXE boot capability by
// PATCHing Dell's per-NIC HttpDevNEnDis/PxeDevNEnDis BIOS Setup attributes (see
// networkBootDeviceIndex). httpEnabled and pxeEnabled are independent: a nil pointer leaves that
// protocol's attribute untouched, so either, both, or neither may be set.
func (c *Conn) SetNetworkBootEnabled(ctx context.Context, httpEnabled, pxeEnabled *bool) (ok bool, err error) {
	if httpEnabled == nil && pxeEnabled == nil {
		return false, errors.New("at least one of httpEnabled or pxeEnabled must be set")
	}

	attrs := map[string]string{}
	if httpEnabled != nil {
		attrs[fmt.Sprintf("HttpDev%dEnDis", networkBootDeviceIndex)] = enDis(*httpEnabled)
	}
	if pxeEnabled != nil {
		attrs[fmt.Sprintf("PxeDev%dEnDis", networkBootDeviceIndex)] = enDis(*pxeEnabled)
	}

	if err := c.redfishwrapper.SetBiosConfiguration(ctx, attrs); err != nil {
		return false, err
	}

	return true, nil
}

func enDis(enabled bool) string {
	if enabled {
		return attrEnabled
	}
	return attrDisabled
}
