package dell

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// httpBootDeviceIndex is the Dell "device" slot (HttpDevN in the BIOS attribute registry) this
// file manages. Dell models HTTP Boot as up to 4 independently configurable per-NIC device
// slots (HttpDev1..HttpDev4), each bound to a NIC FQDD via HttpDevNInterface - unlike
// Supermicro/AMI Aptio's single global attribute. There is no field in bmc.NetworkBootConfig to
// select a NIC, so this always targets slot 1, the primary boot device slot - confirmed live on
// a PowerEdge R6715 (iDRAC firmware 1.5.3) where HttpDev1Interface is already bound to the same
// NIC as PxeDev1Interface.
const httpBootDeviceIndex = 1

// SetHTTPBootURI sets the URI UEFI HTTP Boot fetches its boot image from.
//
// Unlike the Redfish-generic provider, which PATCHes the standard
// ComputerSystem.Boot.HttpBootUri property, Dell's iDRAC does not populate that property at all
// (confirmed live: absent from the Boot object entirely, not just empty, on iDRAC firmware
// 1.5.3). Dell instead models the URI as a BIOS Setup attribute, HttpDevNUri, so this PATCHes
// that attribute via SetBiosConfiguration instead. Independent of the HTTP Boot enable/disable
// toggle: setting the URI does not enable the capability, and enabling the capability does not
// require a URI, matching bmc.NetworkBootConfig's contract.
func (c *Conn) SetHTTPBootURI(ctx context.Context, uri string) (ok bool, err error) {
	attrs := map[string]string{
		fmt.Sprintf("HttpDev%dUri", httpBootDeviceIndex): uri,
	}
	if err := c.redfishwrapper.SetBiosConfiguration(ctx, attrs); err != nil {
		return false, err
	}

	return true, nil
}

// SetHTTPBootTLSMode sets the TLS authentication mode UEFI HTTP Boot uses to connect to the HTTP
// boot server, by PATCHing Dell's per-NIC HttpDevNTlsMode BIOS Setup attribute (see
// httpBootDeviceIndex). Dell's BIOS attribute registry declares this attribute as an
// enumeration offering only "None" and "OneWay", so validating mode here would just duplicate
// what the iDRAC already enforces on the PATCH.
//
// Implements bmc.HTTPBootTLSModeSetter.
func (c *Conn) SetHTTPBootTLSMode(ctx context.Context, mode bmc.HTTPBootTLSMode) (ok bool, err error) {
	attrs := map[string]string{
		fmt.Sprintf("HttpDev%dTlsMode", httpBootDeviceIndex): string(mode),
	}
	if err := c.redfishwrapper.SetBiosConfiguration(ctx, attrs); err != nil {
		return false, err
	}

	return true, nil
}
