package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.ManagerEthernetInterfacesGetter = (*Conn)(nil)
	_ bmc.ManagerNetworkProtocolGetter    = (*Conn)(nil)
	_ bmc.ManagerNetworkServiceSetter     = (*Conn)(nil)
	_ bmc.ManagerNTPServersSetter         = (*Conn)(nil)
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

// SetManagerNetworkService enables or disables a management service of the XCC; a non-nil port
// also sets its port.
//
// Implements bmc.ManagerNetworkServiceSetter.
func (c *Conn) SetManagerNetworkService(ctx context.Context, service string, enabled bool, port *int) (err error) {
	return c.redfishwrapper.SetManagerNetworkService(ctx, service, enabled, port)
}

// SetManagerNTPServers replaces the NTP server list of the XCC; an empty list clears it.
//
// Implements bmc.ManagerNTPServersSetter.
func (c *Conn) SetManagerNTPServers(ctx context.Context, servers []string) (err error) {
	return c.redfishwrapper.SetManagerNTPServers(ctx, servers)
}
