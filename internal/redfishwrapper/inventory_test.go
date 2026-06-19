package redfishwrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInventoryNICsHasPrefixIDCollision proves that collectEthernetInfo's
// strings.HasPrefix match is broken when numeric IDs share a common prefix.
//
// Scenario: NetworkPort "1" (onboard) and EthernetInterface "10" (AOC).
// HasPrefix("10", "1") is true, so when the BMC returns EthernetInterfaces in
// order ["10", "1", "2"], port "1" steals the AOC MAC from interface "10".
// The AOC MAC then lands in discoveredMACs via the main loop, so the fallback
// skips interface "10" entirely — the AOC NIC silently disappears.
//
// This test will FAIL until HasPrefix is replaced with an exact-match or a
// suffix-aware comparison.
func TestInventoryNICsHasPrefixIDCollision(t *testing.T) {
	mux := http.NewServeMux()
	for path, fixture := range map[string]string{
		"/redfish/v1/":      "serviceroot.json",
		"/redfish/v1/Systems":   "systems.json",
		"/redfish/v1/Systems/1": "systems_1.json",

		"/redfish/v1/Systems/1/NetworkInterfaces":                                                        "smc_aoc/network_interfaces.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1":                                      "smc_aoc/network_interface_integrated_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter":                       "smc_aoc/network_adapter.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts":          "smc_aoc/network_ports.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts/1":        "smc_aoc/network_port_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts/2":        "smc_aoc/network_port_2.json",

		// EthernetInterfaces returned in order [10, 1, 2] — AOC "10" is first,
		// causing HasPrefix("10","1") to fire before the correct match on "1".
		"/redfish/v1/Systems/1/EthernetInterfaces":    "smc_aoc_prefix/ethernet_interfaces.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/10": "smc_aoc_prefix/ethernet_10.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/1":  "smc_aoc/ethernet_1.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/2":  "smc_aoc/ethernet_2.json",
	} {
		mux.HandleFunc(path, endpointFunc(t, fixture))
	}

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(u.Hostname(), u.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	device, err := client.Inventory(context.Background(), false)
	require.NoError(t, err)
	require.NotNil(t, device)

	// No port on the onboard NIC should carry the AOC MAC.
	// With the HasPrefix bug, EthernetInterface "10" matches port "1" first
	// (HasPrefix("10","1")==true) and overwrites both its ID and MAC with the
	// AOC's values — the onboard port ends up with MAC 3c:ec:ef:20:30:01.
	for _, nic := range device.NICs {
		if nic.ID != "NIC.Integrated.1" {
			continue
		}
		for _, p := range nic.NICPorts {
			assert.NotEqual(t, "3c:ec:ef:20:30:01", p.MacAddress,
				"onboard NIC port has AOC MAC — HasPrefix collision with EthernetInterface ID '10'")
		}
	}

	// The AOC port must appear in the inventory as a fallback NIC.
	// With the bug, the AOC MAC ends up on the onboard port (above), so
	// discoveredMACs already contains it and the fallback skips it.
	aocFound := false
	for _, nic := range device.NICs {
		if nic.ID == "NIC.Integrated.1" {
			continue // skip onboard NIC
		}
		for _, p := range nic.NICPorts {
			if p.MacAddress == "3c:ec:ef:20:30:01" {
				aocFound = true
			}
		}
	}
	assert.True(t, aocFound, "AOC NIC (EthernetInterface ID '10') missing from inventory due to HasPrefix collision")
}

// TestInventoryNICsWithAOCFallback verifies that EthernetInterfaces-only ports
// (e.g. Supermicro AOC add-on cards absent from NetworkInterfaces) are included
// in the inventory.
//
// Fixture layout:
//   - NetworkInterfaces: 1 onboard NIC with 2 ports (port IDs "1" and "2")
//   - EthernetInterfaces: 4 entries — IDs "1" and "2" enrich the onboard ports
//     via collectEthernetInfo (HasPrefix match); IDs "3" and "4" are AOC ports
//     that do NOT match any network port and must be added by the fallback.
func TestInventoryNICsWithAOCFallback(t *testing.T) {
	mux := http.NewServeMux()
	for path, fixture := range map[string]string{
		// gofish bootstrap
		"/redfish/v1/":      "serviceroot.json",
		"/redfish/v1/Systems":   "systems.json",
		"/redfish/v1/Systems/1": "systems_1.json",
		// NetworkInterfaces hierarchy
		"/redfish/v1/Systems/1/NetworkInterfaces":                                         "smc_aoc/network_interfaces.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1":                       "smc_aoc/network_interface_integrated_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter":        "smc_aoc/network_adapter.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts":   "smc_aoc/network_ports.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts/1": "smc_aoc/network_port_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts/2": "smc_aoc/network_port_2.json",
		// EthernetInterfaces (4 entries: 2 onboard + 2 AOC)
		"/redfish/v1/Systems/1/EthernetInterfaces":   "smc_aoc/ethernet_interfaces.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/1": "smc_aoc/ethernet_1.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/2": "smc_aoc/ethernet_2.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/3": "smc_aoc/ethernet_3.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/4": "smc_aoc/ethernet_4.json",
	} {
		mux.HandleFunc(path, endpointFunc(t, fixture))
	}

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(u.Hostname(), u.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(context.Background()))
	defer client.Close(context.Background())

	device, err := client.Inventory(context.Background(), false)
	require.NoError(t, err)
	require.NotNil(t, device)

	// Collect all NIC port MACs for easy assertion.
	portMACs := make(map[string]bool)
	totalPorts := 0
	for _, nic := range device.NICs {
		for _, p := range nic.NICPorts {
			portMACs[p.MacAddress] = true
			totalPorts++
		}
	}

	assert.Equal(t, 4, totalPorts, "expected 4 NIC ports (2 onboard + 2 AOC)")
	assert.True(t, portMACs["3c:ec:ef:10:20:01"], "onboard port 1 MAC missing")
	assert.True(t, portMACs["3c:ec:ef:10:20:02"], "onboard port 2 MAC missing")
	assert.True(t, portMACs["3c:ec:ef:20:30:01"], "AOC port 1 MAC missing")
	assert.True(t, portMACs["3c:ec:ef:20:30:02"], "AOC port 2 MAC missing")

	// AOC ports should carry the speed from the EthernetInterface (25 Gbps).
	for _, nic := range device.NICs {
		for _, p := range nic.NICPorts {
			if p.MacAddress == "3c:ec:ef:20:30:01" || p.MacAddress == "3c:ec:ef:20:30:02" {
				assert.Equal(t, int64(25_000_000_000), p.SpeedBits,
					"AOC port %s should have 25 Gbps", p.MacAddress)
			}
		}
	}
}
