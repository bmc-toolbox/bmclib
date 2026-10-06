package bmc

import "context"

// SerialInterface describes a BMC serial interface.
type SerialInterface struct {
	// ID is the Redfish SerialInterface Id.
	ID string
	// Enabled reports whether the interface is enabled.
	Enabled bool
	// BitRate is the configured baud rate (e.g. "115200").
	BitRate string
	// FlowControl is the flow-control mode (e.g. "None", "Software", "Hardware").
	FlowControl string
	// Parity is the parity mode (e.g. "None", "Even", "Odd").
	Parity string
	// StopBits is the number of stop bits (e.g. "1", "2").
	StopBits string
	// DataBits is the number of data bits (e.g. "8").
	DataBits string
}

// SerialInterfaceGetter is implemented by providers that can read BMC serial
// interfaces.
type SerialInterfaceGetter interface {
	// SerialInterfaces returns the BMC serial interfaces.
	SerialInterfaces(ctx context.Context) ([]SerialInterface, error)
}

// SerialInterfaceSetter is implemented by providers that can configure a BMC
// serial interface.
type SerialInterfaceSetter interface {
	// SetSerialInterface PATCHes the serial interface with id using the given
	// Redfish attributes (e.g. {"BitRate": "115200", "FlowControl": "None"}).
	SetSerialInterface(ctx context.Context, id string, attrs map[string]any) error
}
