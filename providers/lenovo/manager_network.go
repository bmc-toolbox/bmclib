package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.ManagerEthernetInterfacesGetter = (*Conn)(nil)
	_ bmc.ManagerNetworkProtocolGetter    = (*Conn)(nil)
)

// GetManagerEthernetInterfaces returns the XCC ethernet interfaces.
//
// Implements bmc.ManagerEthernetInterfacesGetter.
func (c *Conn) GetManagerEthernetInterfaces(ctx context.Context) (interfaces []bmc.ManagerEthernetInterface, err error) {
	return c.redfishwrapper.GetManagerEthernetInterfaces(ctx)
}

// GetManagerNetworkProtocol returns the XCC network protocol resource: host name and the
// management services the controller reports.
//
// Implements bmc.ManagerNetworkProtocolGetter.
func (c *Conn) GetManagerNetworkProtocol(ctx context.Context) (protocol *bmc.ManagerNetworkProtocol, err error) {
	return c.redfishwrapper.GetManagerNetworkProtocol(ctx)
}
