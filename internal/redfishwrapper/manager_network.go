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
	np, managerID, err := c.managerNetworkProtocol(ctx)
	if err != nil {
		return nil, err
	}

	protocol, err := toManagerNetworkProtocol(np)
	if err != nil {
		return nil, fmt.Errorf("manager %s NetworkProtocol: %w", managerID, err)
	}

	return protocol, nil
}

// SetManagerNetworkService enables or disables a service of the manager network protocol resource
// and, with a non-nil port, sets its port. Only the named service is sent. A service the resource
// does not report is refused rather than written blindly.
func (c *Client) SetManagerNetworkService(ctx context.Context, service string, enabled bool, port *int) error {
	if port != nil && (*port < 1 || *port > 65535) {
		return fmt.Errorf("service %s: port %d is out of range", service, *port)
	}

	np, managerID, err := c.managerNetworkProtocol(ctx)
	if err != nil {
		return err
	}

	present, err := managerNetworkProtocolProperties(np)
	if err != nil {
		return fmt.Errorf("manager %s NetworkProtocol: %w", managerID, err)
	}

	known := false
	for _, s := range managerNetworkServices(np) {
		if s.name == service {
			known = true
			break
		}
	}
	if !known || !present[service] {
		return fmt.Errorf("%w: %s", bmclibErrs.ErrManagerNetworkServiceNotFound, service)
	}

	settings := map[string]any{"ProtocolEnabled": enabled}
	if port != nil {
		settings["Port"] = *port
	}

	return c.patchManagerNetworkProtocol(ctx, np, map[string]any{service: settings})
}

// managerNetworkProtocol returns the NetworkProtocol resource of the first manager that links one,
// with the ID of that manager. A manager without the link, or whose link answers 404, is skipped.
func (c *Client) managerNetworkProtocol(ctx context.Context) (*schemas.ManagerNetworkProtocol, string, error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return nil, "", err
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
			return nil, "", fmt.Errorf("manager %s NetworkProtocol: %w", m.ID, err)
		}
		if np == nil {
			continue
		}

		return np, m.ID, nil
	}

	return nil, "", bmclibErrs.ErrManagerNetworkProtocolNotFound
}

// patchManagerNetworkProtocol PATCHes the given properties of the NetworkProtocol resource.
//
// gofish's ManagerNetworkProtocol.Update() writes IPv6AddressGenerationMode only, so the service
// properties are sent with an explicit PATCH. As Update() does, it sends the resource ETag in
// If-Match unless ETag matching is disabled for the client.
func (c *Client) patchManagerNetworkProtocol(ctx context.Context, np *schemas.ManagerNetworkProtocol, payload map[string]any) error {
	var headers map[string]string
	if np.ODataEtag != "" && !c.disableEtagMatch {
		headers = map[string]string{"If-Match": np.ODataEtag}
	}

	// gofish returns an error for any non-2xx response, carrying the BMC's error body.
	resp, err := c.PatchWithHeaders(ctx, np.ODataID, payload, headers)
	if err != nil {
		return fmt.Errorf("%w: %s", bmclibErrs.ErrManagerNetworkProtocolSet, err.Error())
	}
	_ = resp.Body.Close()

	return nil
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
	present, err := managerNetworkProtocolProperties(np)
	if err != nil {
		return nil, err
	}

	out := &bmc.ManagerNetworkProtocol{
		HostName: np.HostName,
		FQDN:     np.FQDN,
	}
	if len(np.NTP.NTPServers) > 0 {
		out.NTPServers = append([]string(nil), np.NTP.NTPServers...)
	}
	for _, s := range managerNetworkServices(np) {
		if !present[s.name] {
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

// managerNetworkService is a service of the NetworkProtocol resource as gofish decodes it.
type managerNetworkService struct {
	name    string
	enabled bool
	port    *uint
}

// managerNetworkServices lists the services bmclib handles, in the order they are reported.
func managerNetworkServices(np *schemas.ManagerNetworkProtocol) []managerNetworkService {
	return []managerNetworkService{
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
}

// managerNetworkProtocolProperties reports which properties the resource document carries with a
// non-null value.
func managerNetworkProtocolProperties(np *schemas.ManagerNetworkProtocol) (map[string]bool, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(np.RawData, &document); err != nil {
		return nil, fmt.Errorf("decoding resource: %w", err)
	}

	present := make(map[string]bool, len(document))
	for name, raw := range document {
		present[name] = string(raw) != "null"
	}

	return present, nil
}
