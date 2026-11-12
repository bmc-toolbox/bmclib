package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// ManagerIPv4Address is an IPv4 address configured on a manager ethernet interface.
type ManagerIPv4Address struct {
	Address    string
	SubnetMask string
	Gateway    string
	// Origin is the Redfish AddressOrigin, e.g. "Static", "DHCP".
	Origin string
}

// ManagerIPv6Address is an IPv6 address configured on a manager ethernet interface.
type ManagerIPv6Address struct {
	Address      string
	PrefixLength int
	// Origin is the Redfish AddressOrigin, e.g. "Static", "DHCPv6", "LinkLocal", "SLAAC".
	Origin string
}

// ManagerEthernetInterface describes an ethernet interface of the manager (the BMC's own NIC).
type ManagerEthernetInterface struct {
	// ID is the Redfish EthernetInterface resource ID, e.g. "NIC.1" or "eth0".
	ID   string
	Name string
	// MACAddress is the currently configured MAC address, PermanentMACAddress the factory one.
	MACAddress          string
	PermanentMACAddress string
	HostName            string
	FQDN                string
	// Enabled is the Redfish InterfaceEnabled flag.
	Enabled bool
	// LinkStatus is the Redfish LinkStatus, e.g. "LinkUp", "LinkDown", "NoLink"; empty when not reported.
	LinkStatus string
	// SpeedMbps and MTUSize are nil when the BMC does not report them.
	SpeedMbps *int
	MTUSize   *int
	// IPv4Addresses and IPv6Addresses are the addresses currently assigned to the interface.
	IPv4Addresses []ManagerIPv4Address
	IPv6Addresses []ManagerIPv6Address
	// DHCPv4Enabled reports whether IPv4 addressing is obtained through DHCP.
	DHCPv4Enabled bool
	// VLANEnabled and VLANID describe the VLAN tag of the interface; VLANID is 0 when untagged.
	VLANEnabled bool
	VLANID      int
	// NameServers are the DNS servers in use by the interface.
	NameServers []string
	// Health and State are the Redfish Status fields.
	Health string
	State  string
}

// ManagerNetworkService is one management service (protocol) of the manager network protocol
// resource.
type ManagerNetworkService struct {
	// Name is the Redfish property name of the service: "HTTP", "HTTPS", "SSH", "Telnet", "IPMI",
	// "SNMP", "NTP", "KVMIP", "VirtualMedia" or "SSDP".
	Name string
	// Enabled is the Redfish ProtocolEnabled flag.
	Enabled bool
	// Port is the service port; nil when the BMC does not report one.
	Port *int
}

// ManagerNetworkProtocol describes the manager network protocol resource: the host name of the
// BMC and the management services it exposes.
type ManagerNetworkProtocol struct {
	HostName string
	FQDN     string
	// NTPServers are the NTP servers configured on the BMC.
	NTPServers []string
	// Services lists only the services the BMC reports; a service absent from the resource is not
	// listed, a disabled one is listed with Enabled false.
	Services []ManagerNetworkService
}

// ManagerEthernetInterfacesGetter provides retrieval of the manager (BMC) ethernet interfaces.
type ManagerEthernetInterfacesGetter interface {
	GetManagerEthernetInterfaces(ctx context.Context) (interfaces []ManagerEthernetInterface, err error)
}

// ManagerNetworkProtocolGetter provides retrieval of the manager network protocol resource.
type ManagerNetworkProtocolGetter interface {
	GetManagerNetworkProtocol(ctx context.Context) (protocol *ManagerNetworkProtocol, err error)
}

type managerEthernetInterfacesGetterProvider struct {
	name string
	ManagerEthernetInterfacesGetter
}

type managerNetworkProtocolGetterProvider struct {
	name string
	ManagerNetworkProtocolGetter
}

func getManagerEthernetInterfaces(ctx context.Context, generic []managerEthernetInterfacesGetterProvider) (interfaces []ManagerEthernetInterface, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ManagerEthernetInterfacesGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			interfaces, vErr := elem.GetManagerEthernetInterfaces(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return interfaces, metadata, nil
		}
	}

	return nil, metadata, multierror.Append(err, errors.New("failure to get manager ethernet interfaces"))
}

func getManagerNetworkProtocol(ctx context.Context, generic []managerNetworkProtocolGetterProvider) (protocol *ManagerNetworkProtocol, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ManagerNetworkProtocolGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			protocol, vErr := elem.GetManagerNetworkProtocol(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return protocol, metadata, nil
		}
	}

	return nil, metadata, multierror.Append(err, errors.New("failure to get manager network protocol"))
}

// GetManagerEthernetInterfacesFromInterfaces identifies implementations of the
// ManagerEthernetInterfacesGetter interface and passes them to the getManagerEthernetInterfaces() wrapper.
func GetManagerEthernetInterfacesFromInterfaces(ctx context.Context, generic []interface{}) (interfaces []ManagerEthernetInterface, metadata Metadata, err error) {
	implementations := make([]managerEthernetInterfacesGetterProvider, 0)
	for _, elem := range generic {
		temp := managerEthernetInterfacesGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ManagerEthernetInterfacesGetter:
			temp.ManagerEthernetInterfacesGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ManagerEthernetInterfacesGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return nil, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ManagerEthernetInterfacesGetter implementations found",
			),
		)
	}

	return getManagerEthernetInterfaces(ctx, implementations)
}

// GetManagerNetworkProtocolFromInterfaces identifies implementations of the
// ManagerNetworkProtocolGetter interface and passes them to the getManagerNetworkProtocol() wrapper.
func GetManagerNetworkProtocolFromInterfaces(ctx context.Context, generic []interface{}) (protocol *ManagerNetworkProtocol, metadata Metadata, err error) {
	implementations := make([]managerNetworkProtocolGetterProvider, 0)
	for _, elem := range generic {
		temp := managerNetworkProtocolGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ManagerNetworkProtocolGetter:
			temp.ManagerNetworkProtocolGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ManagerNetworkProtocolGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return nil, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ManagerNetworkProtocolGetter implementations found",
			),
		)
	}

	return getManagerNetworkProtocol(ctx, implementations)
}

// ManagerNetworkServiceSetter enables or disables a management service of the manager and
// optionally changes its port.
type ManagerNetworkServiceSetter interface {
	// SetManagerNetworkService sets the enabled state of the named service; service is a name as
	// reported in ManagerNetworkService.Name. A non-nil port also changes the service port, a nil
	// port leaves it as is.
	//
	// Disabling the service the caller is connected through (e.g. HTTPS) cuts the caller off.
	SetManagerNetworkService(ctx context.Context, service string, enabled bool, port *int) (err error)
}

type managerNetworkServiceSetterProvider struct {
	name string
	ManagerNetworkServiceSetter
}

func setManagerNetworkService(ctx context.Context, generic []managerNetworkServiceSetterProvider, service string, enabled bool, port *int) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ManagerNetworkServiceSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.SetManagerNetworkService(ctx, service, enabled, port)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to set manager network service"))
}

// SetManagerNetworkServiceFromInterfaces identifies implementations of the
// ManagerNetworkServiceSetter interface and passes them to the setManagerNetworkService() wrapper.
func SetManagerNetworkServiceFromInterfaces(ctx context.Context, generic []interface{}, service string, enabled bool, port *int) (metadata Metadata, err error) {
	implementations := make([]managerNetworkServiceSetterProvider, 0)
	for _, elem := range generic {
		temp := managerNetworkServiceSetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ManagerNetworkServiceSetter:
			temp.ManagerNetworkServiceSetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ManagerNetworkServiceSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ManagerNetworkServiceSetter implementations found",
			),
		)
	}

	return setManagerNetworkService(ctx, implementations, service, enabled, port)
}
