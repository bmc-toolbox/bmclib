package lenovo

import (
	"context"
	"testing"
)

// Requirement: Read ethernet interfaces — BMC NIC config.
func TestBMCNetworkInterfaces(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ifaces, err := c.BMCNetworkInterfaces(context.Background())
	if err != nil {
		t.Fatalf("BMCNetworkInterfaces: %v", err)
	}
	if len(ifaces) != 1 {
		t.Fatalf("got %d BMC interfaces, want 1", len(ifaces))
	}
	got := ifaces[0]
	if got.MACAddress == "" || got.HostName != "XCC-SR650" {
		t.Errorf("unexpected BMC interface: %+v", got)
	}
	if len(got.IPv4Addresses) != 1 || got.IPv4Addresses[0] != "10.0.0.50" {
		t.Errorf("unexpected IPv4: %v", got.IPv4Addresses)
	}
}

// Requirement: Read ethernet interfaces — server NIC config.
func TestServerNetworkInterfaces(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ifaces, err := c.ServerNetworkInterfaces(context.Background())
	if err != nil {
		t.Fatalf("ServerNetworkInterfaces: %v", err)
	}
	if len(ifaces) != 1 || ifaces[0].ID != "NIC.1" {
		t.Fatalf("unexpected server interfaces: %+v", ifaces)
	}
}

// Requirement: Configure BMC ethernet — set a static IPv4 address.
func TestSetBMCNetworkInterface(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	attrs := map[string]any{
		"IPv4StaticAddresses": []map[string]any{
			{"Address": "10.0.0.51", "SubnetMask": "255.255.255.0", "Gateway": "10.0.0.1"},
		},
	}
	if err := c.SetBMCNetworkInterface(context.Background(), "eth0", attrs); err != nil {
		t.Fatalf("SetBMCNetworkInterface: %v", err)
	}
	if !ts.didPatchBMCEth() {
		t.Error("expected a PATCH of the BMC ethernet interface")
	}
}

// Requirement: Configure BMC ethernet — disable the host interface.
func TestSetHostInterfaceEnabled(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.SetHostInterfaceEnabled(context.Background(), "1", false); err != nil {
		t.Fatalf("SetHostInterfaceEnabled: %v", err)
	}
	if !ts.didPatchHostIface() {
		t.Error("expected a PATCH of the host interface")
	}
}

// Requirement: Manager network protocols — read network services.
func TestNetworkProtocols(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	protos, err := c.NetworkProtocols(context.Background())
	if err != nil {
		t.Fatalf("NetworkProtocols: %v", err)
	}

	byName := map[string]struct {
		enabled bool
		port    int
	}{}
	for _, p := range protos {
		byName[p.Name] = struct {
			enabled bool
			port    int
		}{p.Enabled, p.Port}
	}

	if !byName["HTTPS"].enabled || byName["HTTPS"].port != 443 {
		t.Errorf("unexpected HTTPS protocol: %+v", byName["HTTPS"])
	}
	if byName["IPMI"].port != 623 {
		t.Errorf("unexpected IPMI port: %d", byName["IPMI"].port)
	}
	// A protocol absent from the resource (the fixture, like real XCC, has no
	// Telnet) must not be reported as a phantom disabled service.
	if _, ok := byName["Telnet"]; ok {
		t.Errorf("Telnet is absent from the resource but was reported: %+v", protos)
	}
}

// Requirement: Manager network protocols — disable a service.
func TestSetNetworkProtocols(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	attrs := map[string]any{"IPMI": map[string]any{"ProtocolEnabled": false}}
	if err := c.SetNetworkProtocols(context.Background(), attrs); err != nil {
		t.Fatalf("SetNetworkProtocols: %v", err)
	}
	if !ts.didPatchNetProto() {
		t.Error("expected a PATCH of ManagerNetworkProtocol")
	}
}

// Requirement: Read serial interface.
func TestSerialInterfaces(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	ifaces, err := c.SerialInterfaces(context.Background())
	if err != nil {
		t.Fatalf("SerialInterfaces: %v", err)
	}
	if len(ifaces) != 1 {
		t.Fatalf("got %d serial interfaces, want 1", len(ifaces))
	}
	if ifaces[0].BitRate != "115200" || ifaces[0].FlowControl != "None" {
		t.Errorf("unexpected serial config: %+v", ifaces[0])
	}
}

// Requirement: Configure serial interface — change baud rate.
func TestSetSerialInterface(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	attrs := map[string]any{"BitRate": "57600"}
	if err := c.SetSerialInterface(context.Background(), "1", attrs); err != nil {
		t.Fatalf("SetSerialInterface: %v", err)
	}
	if !ts.didPatchSerial() {
		t.Error("expected a PATCH of the serial interface")
	}
}
