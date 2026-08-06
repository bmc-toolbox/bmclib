package bmc

import "context"

// NetworkInterface describes a BMC or server ethernet interface.
type NetworkInterface struct {
	// ID is the Redfish EthernetInterface Id.
	ID string
	// MACAddress is the interface MAC address.
	MACAddress string
	// HostName is the configured hostname (BMC interfaces).
	HostName string
	// IPv4Addresses are the configured IPv4 addresses.
	IPv4Addresses []string
	// IPv6Addresses are the configured IPv6 addresses.
	IPv6Addresses []string
	// Enabled reports whether the interface is enabled.
	Enabled bool
}

// NetworkProtocol describes a manager network service (protocol) and its state.
type NetworkProtocol struct {
	// Name is the service name (e.g. "HTTPS", "SSH", "IPMI", "SNMP", "NTP").
	Name string
	// Enabled reports whether the service is enabled.
	Enabled bool
	// Port is the service port (0 when not reported).
	Port int
}

// NetworkInterfaceGetter is implemented by providers that can read BMC and
// server ethernet interfaces.
type NetworkInterfaceGetter interface {
	// BMCNetworkInterfaces returns the BMC (manager) ethernet interfaces.
	BMCNetworkInterfaces(ctx context.Context) ([]NetworkInterface, error)
	// ServerNetworkInterfaces returns the server (system) ethernet interfaces.
	ServerNetworkInterfaces(ctx context.Context) ([]NetworkInterface, error)
}

// NetworkInterfaceSetter is implemented by providers that can configure BMC
// network interfaces.
type NetworkInterfaceSetter interface {
	// SetBMCNetworkInterface PATCHes the BMC ethernet interface with id using the
	// given Redfish attributes (e.g. {"HostName": "x", "IPv4StaticAddresses": [...],
	// "VLAN": {...}}).
	SetBMCNetworkInterface(ctx context.Context, id string, attrs map[string]any) error
	// SetHostInterfaceEnabled enables or disables the host interface with id.
	SetHostInterfaceEnabled(ctx context.Context, id string, enabled bool) error
}

// NetworkProtocolGetter is implemented by providers that can read the manager
// network protocols/services.
type NetworkProtocolGetter interface {
	// NetworkProtocols returns the manager network services and their state.
	NetworkProtocols(ctx context.Context) ([]NetworkProtocol, error)
}

// NetworkProtocolSetter is implemented by providers that can configure the
// manager network protocols/services.
type NetworkProtocolSetter interface {
	// SetNetworkProtocols PATCHes ManagerNetworkProtocol with the given Redfish
	// attributes (e.g. {"IPMI": {"ProtocolEnabled": false}}).
	SetNetworkProtocols(ctx context.Context, attrs map[string]any) error
}
