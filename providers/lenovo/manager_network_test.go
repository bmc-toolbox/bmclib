package lenovo

import (
	"context"
	"errors"
	"testing"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// Requirement: Read ethernet interfaces — the XCC manager NIC is readable.
func TestGetManagerEthernetInterfaces(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	interfaces, err := c.GetManagerEthernetInterfaces(context.Background())
	if err != nil {
		t.Fatalf("GetManagerEthernetInterfaces: %v", err)
	}
	if len(interfaces) != 1 {
		t.Fatalf("got %d interfaces, want 1", len(interfaces))
	}
	eth := interfaces[0]
	if eth.ID != "eth0" || eth.MACAddress != "7C:D3:0A:11:22:33" || eth.HostName != "XCC-SR650" || !eth.Enabled {
		t.Errorf("unexpected interface: %+v", eth)
	}
	if len(eth.IPv4Addresses) != 1 || eth.IPv4Addresses[0].Address != "10.0.0.50" || eth.IPv4Addresses[0].Origin != "Static" {
		t.Errorf("unexpected IPv4 addresses: %+v", eth.IPv4Addresses)
	}
	if len(eth.IPv6Addresses) != 1 || eth.IPv6Addresses[0].Address != "fe80::7ed3:aff:fe11:2233" || eth.IPv6Addresses[0].PrefixLength != 64 {
		t.Errorf("unexpected IPv6 addresses: %+v", eth.IPv6Addresses)
	}
}

// Requirement: Manager network protocols — only the services the XCC exposes are listed.
func TestGetManagerNetworkProtocol(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	protocol, err := c.GetManagerNetworkProtocol(context.Background())
	if err != nil {
		t.Fatalf("GetManagerNetworkProtocol: %v", err)
	}

	want := []string{"HTTP", "HTTPS", "SSH", "IPMI", "SNMP", "NTP", "KVMIP", "VirtualMedia", "SSDP"}
	if len(protocol.Services) != len(want) {
		t.Fatalf("got %d services %+v, want %d", len(protocol.Services), protocol.Services, len(want))
	}
	for i, name := range want {
		if protocol.Services[i].Name != name {
			t.Errorf("service %d = %q, want %q", i, protocol.Services[i].Name, name)
		}
		if name == "Telnet" {
			t.Errorf("Telnet must not be reported: the XCC mock does not expose it")
		}
	}
	for _, s := range protocol.Services {
		if s.Name == "HTTPS" && (!s.Enabled || s.Port == nil || *s.Port != 443) {
			t.Errorf("unexpected HTTPS service: %+v", s)
		}
	}
}

// Requirement: Manager network protocols — a service is switched with a PATCH of that service only.
func TestSetManagerNetworkService(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.SetManagerNetworkService(context.Background(), "SSDP", false, nil); err != nil {
		t.Fatalf("SetManagerNetworkService: %v", err)
	}

	patched, body := ts.networkProtocolPatch()
	if !patched {
		t.Fatal("expected the NetworkProtocol resource to be PATCHed")
	}
	if want := `{"SSDP":{"ProtocolEnabled":false}}`; body != want {
		t.Errorf("PATCH body = %s, want %s", body, want)
	}
}

// Requirement: Manager network protocols — a service the XCC does not expose is not written.
func TestSetManagerNetworkServiceNotExposed(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	err := c.SetManagerNetworkService(context.Background(), "Telnet", true, nil)
	if !errors.Is(err, bmclibErrs.ErrManagerNetworkServiceNotFound) {
		t.Fatalf("err = %v, want ErrManagerNetworkServiceNotFound", err)
	}
	if patched, _ := ts.networkProtocolPatch(); patched {
		t.Fatal("a service the XCC does not expose must not be PATCHed")
	}
}

// Requirement: Manager network protocols — the NTP server list is replaced as a whole.
func TestSetManagerNTPServers(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.SetManagerNTPServers(context.Background(), []string{"10.0.0.1", "10.0.0.2"}); err != nil {
		t.Fatalf("SetManagerNTPServers: %v", err)
	}

	patched, body := ts.networkProtocolPatch()
	if !patched {
		t.Fatal("expected the NetworkProtocol resource to be PATCHed")
	}
	if want := `{"NTP":{"NTPServers":["10.0.0.1","10.0.0.2"]}}`; body != want {
		t.Errorf("PATCH body = %s, want %s", body, want)
	}
}
