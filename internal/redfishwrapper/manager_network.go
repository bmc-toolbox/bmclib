package redfishwrapper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// GetManagerEthernetInterfaces returns the ethernet interfaces of the first manager that links an
// EthernetInterfaces collection. Errors raised while collecting, including a collection or member
// the BMC advertises but does not serve, are returned to the caller.
func (c *Client) GetManagerEthernetInterfaces(ctx context.Context) ([]bmc.ManagerEthernetInterface, error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range managers {
		if m == nil {
			continue
		}

		members, err := m.EthernetInterfaces()
		if err != nil {
			return nil, fmt.Errorf("manager %s EthernetInterfaces: %w", m.ID, err)
		}
		if members == nil {
			continue
		}

		interfaces := make([]bmc.ManagerEthernetInterface, 0, len(members))
		for _, member := range members {
			if member == nil {
				continue
			}
			interfaces = append(interfaces, toManagerEthernetInterface(member))
		}

		return interfaces, nil
	}

	return nil, bmclibErrs.ErrManagerEthernetInterfacesNotFound
}

// GetManagerNetworkProtocol returns the network protocol resource of the first manager that links
// one: the BMC host name and the management services it reports. Errors raised while collecting
// are returned to the caller.
func (c *Client) GetManagerNetworkProtocol(ctx context.Context) (*bmc.ManagerNetworkProtocol, error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range managers {
		if m == nil {
			continue
		}

		np, err := m.NetworkProtocol()
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return nil, fmt.Errorf("manager %s NetworkProtocol: %w", m.ID, err)
		}
		if np == nil {
			continue
		}

		protocol, err := toManagerNetworkProtocol(np)
		if err != nil {
			return nil, fmt.Errorf("manager %s NetworkProtocol: %w", m.ID, err)
		}

		return protocol, nil
	}

	return nil, bmclibErrs.ErrManagerNetworkProtocolNotFound
}

// toManagerEthernetInterface maps a Redfish EthernetInterface to the bmclib representation.
func toManagerEthernetInterface(eth *schemas.EthernetInterface) bmc.ManagerEthernetInterface {
	out := bmc.ManagerEthernetInterface{
		ID:                  eth.ID,
		Name:                eth.Name,
		MACAddress:          eth.MACAddress,
		PermanentMACAddress: eth.PermanentMACAddress,
		HostName:            eth.HostName,
		FQDN:                eth.FQDN,
		Enabled:             eth.InterfaceEnabled,
		LinkStatus:          string(eth.LinkStatus),
		DHCPv4Enabled:       eth.DHCPv4.DHCPEnabled,
		VLANEnabled:         eth.VLAN.VLANEnable,
		VLANID:              int(eth.VLAN.VLANID),
		Health:              string(eth.Status.Health),
		State:               string(eth.Status.State),
	}
	if eth.SpeedMbps != nil {
		speed := *eth.SpeedMbps
		out.SpeedMbps = &speed
	}
	if eth.MTUSize != nil {
		mtu := *eth.MTUSize
		out.MTUSize = &mtu
	}
	if len(eth.NameServers) > 0 {
		out.NameServers = append([]string(nil), eth.NameServers...)
	}
	for _, a := range eth.IPv4Addresses {
		out.IPv4Addresses = append(out.IPv4Addresses, bmc.ManagerIPv4Address{
			Address:    a.Address,
			SubnetMask: a.SubnetMask,
			Gateway:    a.Gateway,
			Origin:     string(a.AddressOrigin),
		})
	}
	for _, a := range eth.IPv6Addresses {
		out.IPv6Addresses = append(out.IPv6Addresses, bmc.ManagerIPv6Address{
			Address:      a.Address,
			PrefixLength: int(a.PrefixLength),
			Origin:       string(a.AddressOrigin),
		})
	}

	return out
}

// toManagerNetworkProtocol maps a Redfish ManagerNetworkProtocol to the bmclib representation.
//
// gofish decodes every service into a value struct, so a service the BMC does not report and a
// disabled one look alike; the property names present in the resource document decide which
// services are listed, the values come from gofish.
func toManagerNetworkProtocol(np *schemas.ManagerNetworkProtocol) (*bmc.ManagerNetworkProtocol, error) {
	var present map[string]json.RawMessage
	if err := json.Unmarshal(np.RawData, &present); err != nil {
		return nil, fmt.Errorf("decoding resource: %w", err)
	}

	services := []struct {
		name    string
		enabled bool
		port    *uint
	}{
		{"HTTP", np.HTTP.ProtocolEnabled, np.HTTP.Port},
		{"HTTPS", np.HTTPS.ProtocolEnabled, np.HTTPS.Port},
		{"SSH", np.SSH.ProtocolEnabled, np.SSH.Port},
		{"Telnet", np.Telnet.ProtocolEnabled, np.Telnet.Port},
		{"IPMI", np.IPMI.ProtocolEnabled, np.IPMI.Port},
		{"SNMP", np.SNMP.ProtocolEnabled, np.SNMP.Port},
		{"NTP", np.NTP.ProtocolEnabled, np.NTP.Port},
		{"KVMIP", np.KVMIP.ProtocolEnabled, np.KVMIP.Port},
		{"VirtualMedia", np.VirtualMedia.ProtocolEnabled, np.VirtualMedia.Port},
		{"SSDP", np.SSDP.ProtocolEnabled, np.SSDP.Port},
	}

	out := &bmc.ManagerNetworkProtocol{
		HostName: np.HostName,
		FQDN:     np.FQDN,
	}
	if len(np.NTP.NTPServers) > 0 {
		out.NTPServers = append([]string(nil), np.NTP.NTPServers...)
	}
	for _, s := range services {
		raw, ok := present[s.name]
		if !ok || string(raw) == "null" {
			continue
		}
		service := bmc.ManagerNetworkService{Name: s.name, Enabled: s.enabled}
		if s.port != nil {
			port := int(*s.port)
			service.Port = &port
		}
		out.Services = append(out.Services, service)
	}

	return out, nil
}
