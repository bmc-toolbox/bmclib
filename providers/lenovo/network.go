package lenovo

import (
	"context"
	"net/url"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/stmcginnis/gofish/schemas"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.NetworkInterfaceGetter = (*Conn)(nil)
	_ bmc.NetworkInterfaceSetter = (*Conn)(nil)
	_ bmc.NetworkProtocolGetter  = (*Conn)(nil)
	_ bmc.NetworkProtocolSetter  = (*Conn)(nil)
)

// BMCNetworkInterfaces returns the BMC (manager) ethernet interfaces.
//
// Implements bmc.NetworkInterfaceGetter.
func (c *Conn) BMCNetworkInterfaces(ctx context.Context) ([]bmc.NetworkInterface, error) {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return nil, err
	}

	ifaces, err := manager.EthernetInterfaces()
	if err != nil {
		return nil, err
	}

	return mapEthernetInterfaces(ifaces), nil
}

// ServerNetworkInterfaces returns the server (system) ethernet interfaces.
//
// Implements bmc.NetworkInterfaceGetter.
func (c *Conn) ServerNetworkInterfaces(ctx context.Context) ([]bmc.NetworkInterface, error) {
	system, err := c.redfishwrapper.System()
	if err != nil {
		return nil, err
	}

	ifaces, err := system.EthernetInterfaces()
	if err != nil {
		return nil, err
	}

	return mapEthernetInterfaces(ifaces), nil
}

// SetBMCNetworkInterface PATCHes a BMC ethernet interface.
//
// Implements bmc.NetworkInterfaceSetter.
func (c *Conn) SetBMCNetworkInterface(ctx context.Context, id string, attrs map[string]any) error {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return err
	}

	target, err := url.JoinPath(manager.ODataID, "EthernetInterfaces/"+id)
	if err != nil {
		return err
	}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, target, attrs, nil))
}

// SetHostInterfaceEnabled enables or disables a host interface.
//
// Implements bmc.NetworkInterfaceSetter.
func (c *Conn) SetHostInterfaceEnabled(ctx context.Context, id string, enabled bool) error {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return err
	}

	target, err := url.JoinPath(manager.ODataID, "HostInterfaces/"+id)
	if err != nil {
		return err
	}
	payload := map[string]any{"InterfaceEnabled": enabled}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, target, payload, nil))
}

// NetworkProtocols returns the manager network services and their state.
//
// Implements bmc.NetworkProtocolGetter.
func (c *Conn) NetworkProtocols(ctx context.Context) ([]bmc.NetworkProtocol, error) {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return nil, err
	}

	path, err := url.JoinPath(manager.ODataID, "NetworkProtocol")
	if err != nil {
		return nil, err
	}

	type protoSetting struct {
		Port            *int `json:"Port"`
		ProtocolEnabled bool `json:"ProtocolEnabled"`
	}
	// Each protocol is a pointer so an absent key (nil) is distinguished from a
	// present one: XCC only publishes the services it supports (e.g. SR630 V2
	// exposes no Telnet), and we must not report phantom disabled services.
	var doc struct {
		HTTP         *protoSetting `json:"HTTP"`
		HTTPS        *protoSetting `json:"HTTPS"`
		SSH          *protoSetting `json:"SSH"`
		IPMI         *protoSetting `json:"IPMI"`
		SNMP         *protoSetting `json:"SNMP"`
		NTP          *protoSetting `json:"NTP"`
		KVMIP        *protoSetting `json:"KVMIP"`
		VirtualMedia *protoSetting `json:"VirtualMedia"`
		SSDP         *protoSetting `json:"SSDP"`
		Telnet       *protoSetting `json:"Telnet"`
	}
	if err := c.getJSON(path, &doc); err != nil {
		return nil, err
	}

	services := []struct {
		name string
		set  *protoSetting
	}{
		{"HTTP", doc.HTTP}, {"HTTPS", doc.HTTPS}, {"SSH", doc.SSH}, {"IPMI", doc.IPMI},
		{"SNMP", doc.SNMP}, {"NTP", doc.NTP}, {"KVMIP", doc.KVMIP},
		{"VirtualMedia", doc.VirtualMedia}, {"SSDP", doc.SSDP}, {"Telnet", doc.Telnet},
	}

	out := make([]bmc.NetworkProtocol, 0, len(services))
	for _, s := range services {
		if s.set == nil {
			continue // protocol not exposed by this XCC
		}
		port := 0
		if s.set.Port != nil {
			port = *s.set.Port
		}
		out = append(out, bmc.NetworkProtocol{Name: s.name, Enabled: s.set.ProtocolEnabled, Port: port})
	}

	return out, nil
}

// SetNetworkProtocols PATCHes the ManagerNetworkProtocol resource.
//
// Implements bmc.NetworkProtocolSetter.
func (c *Conn) SetNetworkProtocols(ctx context.Context, attrs map[string]any) error {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return err
	}

	target, err := url.JoinPath(manager.ODataID, "NetworkProtocol")
	if err != nil {
		return err
	}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, target, attrs, nil))
}

// mapEthernetInterfaces converts gofish ethernet interfaces to the neutral
// bmc.NetworkInterface shape.
func mapEthernetInterfaces(ifaces []*schemas.EthernetInterface) []bmc.NetworkInterface {
	out := make([]bmc.NetworkInterface, 0, len(ifaces))
	for _, e := range ifaces {
		ni := bmc.NetworkInterface{
			ID:         e.ID,
			MACAddress: e.MACAddress,
			HostName:   e.HostName,
			Enabled:    e.InterfaceEnabled,
		}
		for _, a := range e.IPv4Addresses {
			ni.IPv4Addresses = append(ni.IPv4Addresses, a.Address)
		}
		for _, a := range e.IPv6Addresses {
			ni.IPv6Addresses = append(ni.IPv6Addresses, a.Address)
		}
		out = append(out, ni)
	}

	return out
}
