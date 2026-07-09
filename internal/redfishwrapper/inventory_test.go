package redfishwrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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
		"/redfish/v1/":          "serviceroot.json",
		"/redfish/v1/Systems":   "systems.json",
		"/redfish/v1/Systems/1": "systems_1.json",

		"/redfish/v1/Systems/1/NetworkInterfaces":                                                "smc_aoc/network_interfaces.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1":                               "smc_aoc/network_interface_integrated_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter":                "smc_aoc/network_adapter.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts":   "smc_aoc/network_ports.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts/1": "smc_aoc/network_port_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts/2": "smc_aoc/network_port_2.json",

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
		"/redfish/v1/":          "serviceroot.json",
		"/redfish/v1/Systems":   "systems.json",
		"/redfish/v1/Systems/1": "systems_1.json",
		// NetworkInterfaces hierarchy
		"/redfish/v1/Systems/1/NetworkInterfaces":                                                "smc_aoc/network_interfaces.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1":                               "smc_aoc/network_interface_integrated_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter":                "smc_aoc/network_adapter.json",
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

// TestInventoryCollectDIMMCapacity proves that collectDIMMs must use
// CapacityMiB — the overall module capacity — rather than VolatileSizeMiB.
// VolatileSizeMiB/NonVolatileSizeMiB only apply to memory with a
// volatile/non-volatile split (e.g. NVDIMM); on ordinary DRAM DIMMs (the vast
// majority of servers), VolatileSizeMiB is 0 even though CapacityMiB is
// correctly populated, so using it alone silently produces SizeBytes=0 for
// every normal DIMM.
func TestInventoryCollectDIMMCapacity(t *testing.T) {
	mux := http.NewServeMux()
	for path, fixture := range map[string]string{
		"/redfish/v1/":                       "serviceroot.json",
		"/redfish/v1/Systems":                "systems.json",
		"/redfish/v1/Systems/1":              "systems_1.json",
		"/redfish/v1/Systems/1/Memory":       "smc_dimm/memory.json",
		"/redfish/v1/Systems/1/Memory/DIMM1": "smc_dimm/memory_dimm1.json",
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

	require.Len(t, device.Memory, 1, "expected exactly one DIMM in inventory")
	dimm := device.Memory[0]
	assert.Equal(t, "TESTDIMMSERIAL01", dimm.Serial)
	assert.Equal(t, int64(65536)*1024*1024, dimm.SizeBytes,
		"DIMM SizeBytes should come from CapacityMiB (64GiB), not VolatileSizeMiB (0 for ordinary DRAM)")
}

// TestInventoryCollectDIMMCapacityNVDIMMFallback proves that collectDIMMs
// still falls back to VolatileSizeMiB+NonVolatileSizeMiB when CapacityMiB is
// absent — the original use case those fields exist for (an NVDIMM's
// volatile/non-volatile split).
func TestInventoryCollectDIMMCapacityNVDIMMFallback(t *testing.T) {
	mux := http.NewServeMux()
	for path, fixture := range map[string]string{
		"/redfish/v1/":                       "serviceroot.json",
		"/redfish/v1/Systems":                "systems.json",
		"/redfish/v1/Systems/1":              "systems_1.json",
		"/redfish/v1/Systems/1/Memory":       "smc_nvdimm/memory.json",
		"/redfish/v1/Systems/1/Memory/DIMM1": "smc_nvdimm/memory_dimm1.json",
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

	require.Len(t, device.Memory, 1, "expected exactly one DIMM in inventory")
	dimm := device.Memory[0]
	assert.Equal(t, "TESTNVDIMMSERIAL01", dimm.Serial)
	assert.Equal(t, int64(8192+8192)*1024*1024, dimm.SizeBytes,
		"DIMM SizeBytes should fall back to VolatileSizeMiB+NonVolatileSizeMiB (8GiB+8GiB) when CapacityMiB is absent")
}

// TestInventoryCollectDrivesWithoutCount proves that collectDrives must not
// gate on Storage.DrivesCount ("Drives@odata.count"). Some implementations
// embed the "Drives" link array without ever including a count property, so
// DrivesCount parses to 0 even though real drives are linked — the old code
// skipped the Storage member entirely and every drive silently disappeared
// from the inventory.
func TestInventoryCollectDrivesWithoutCount(t *testing.T) {
	mux := http.NewServeMux()
	for path, fixture := range map[string]string{
		"/redfish/v1/":          "serviceroot.json",
		"/redfish/v1/Systems":   "systems.json",
		"/redfish/v1/Systems/1": "systems_1.json",

		"/redfish/v1/Systems/1/Storage":                                            "smc_nvme_storage/storage.json",
		"/redfish/v1/Systems/1/Storage/NVMeSSD":                                    "smc_nvme_storage/storage_nvmessd.json",
		"/redfish/v1/Chassis/NVMeSSD.0.Group.0.StorageBackplane/Drives/Disk.Bay.2": "smc_nvme_storage/drive_bay2.json",
		"/redfish/v1/Chassis/NVMeSSD.0.Group.0.StorageBackplane/Drives/Disk.Bay.3": "smc_nvme_storage/drive_bay3.json",
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

	require.Len(t, device.Drives, 2, "expected both drives despite Storage member reporting DrivesCount=0")

	serials := map[string]bool{}
	for _, d := range device.Drives {
		serials[d.Serial] = true
		assert.Equal(t, "KIOXIA KCD8XVUG6T40", d.Model)
		assert.Equal(t, int64(6401252745216), d.CapacityBytes)
	}
	assert.True(t, serials["TESTDRIVESERIAL01"], "Disk.Bay.2 missing from inventory")
	assert.True(t, serials["TESTDRIVESERIAL02"], "Disk.Bay.3 missing from inventory")
}

// TestInventoryNICsAdapterIDCollision proves that collectNICs correctly
// attributes MAC addresses when two distinct NetworkAdapters expose NetworkPorts
// with the same local IDs ("1"/"2").
//
// Scenario (matching a real Supermicro AS-1114S-WN10RT-EU with an AOC-S25G-m2S
// add-on card): the onboard NIC and the AOC card both have NetworkPorts locally
// numbered "1" and "2". collectEthernetInfo's ID-based matching is ambiguous
// here — the AOC's port "1" exact-matches system-wide EthernetInterface "1",
// which actually belongs to the onboard NIC, not the AOC. Without the
// NetworkDeviceFunctions-based authoritative MAC resolution, the AOC's ports
// silently inherit the onboard NIC's MAC addresses (and vice versa is masked
// because the onboard port's own AssociatedNetworkAddresses already set the
// correct value first).
//
// This test will FAIL until collectNICs resolves each port's MAC via the
// adapter's NetworkDeviceFunctions rather than matching by ID alone.
func TestInventoryNICsAdapterIDCollision(t *testing.T) {
	mux := http.NewServeMux()
	for path, fixture := range map[string]string{
		"/redfish/v1/":          "serviceroot.json",
		"/redfish/v1/Systems":   "systems.json",
		"/redfish/v1/Systems/1": "systems_1.json",

		"/redfish/v1/Systems/1/NetworkInterfaces":                  "smc_aoc_id_collision/network_interfaces.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1": "smc_aoc_id_collision/network_interface_integrated_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1":       "smc_aoc_id_collision/network_interface_slot_1.json",

		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter":                "smc_aoc_id_collision/network_adapter_onboard.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts":   "smc_aoc_id_collision/network_ports_onboard.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts/1": "smc_aoc_id_collision/network_port_onboard_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Integrated.1/NetworkAdapter/NetworkPorts/2": "smc_aoc_id_collision/network_port_onboard_2.json",

		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter":                          "smc_aoc_id_collision/network_adapter_aoc.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkPorts":             "smc_aoc_id_collision/network_ports_aoc.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkPorts/1":           "smc_aoc_id_collision/network_port_aoc_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkPorts/2":           "smc_aoc_id_collision/network_port_aoc_2.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkDeviceFunctions":   "smc_aoc_id_collision/network_device_functions_aoc.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkDeviceFunctions/1": "smc_aoc_id_collision/network_device_function_aoc_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkDeviceFunctions/2": "smc_aoc_id_collision/network_device_function_aoc_2.json",

		"/redfish/v1/Systems/1/EthernetInterfaces":   "smc_aoc_id_collision/ethernet_interfaces.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/1": "smc_aoc_id_collision/ethernet_1.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/2": "smc_aoc_id_collision/ethernet_2.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/3": "smc_aoc_id_collision/ethernet_3.json",
		"/redfish/v1/Systems/1/EthernetInterfaces/4": "smc_aoc_id_collision/ethernet_4.json",
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

	macsByAdapter := make(map[string][]string)
	for _, nic := range device.NICs {
		for _, p := range nic.NICPorts {
			macsByAdapter[nic.ID] = append(macsByAdapter[nic.ID], strings.ToLower(p.MacAddress))
		}
	}

	assert.ElementsMatch(t, []string{"00:00:5e:00:53:01", "00:00:5e:00:53:02"}, macsByAdapter["NIC.Integrated.1"],
		"onboard NIC ports have the wrong MACs")
	assert.ElementsMatch(t, []string{"00:00:5e:00:53:03", "00:00:5e:00:53:04"}, macsByAdapter["NIC.Slot.1"],
		"AOC NIC ports have the wrong MACs — likely inherited the onboard NIC's MACs due to a NetworkPort ID collision")
}

// TestInventoryNICsDisabledDeviceFunctionMAC proves that networkPortMACs does
// not treat "00:00:00:00:00:00" (a placeholder MAC reported by disabled
// NetworkDeviceFunctions) as an authoritative value. Every other MAC source in
// this file already filters this placeholder out (AssociatedNetworkAddresses,
// the EthernetInterfaces fallback); networkPortMACs must too, since
// collectEthernetInfo unconditionally overwrites nicPort.MacAddress with the
// preferred MAC once one is resolved, which would otherwise zero out a real
// MAC already collected from the port's own AssociatedNetworkAddresses.
func TestInventoryNICsDisabledDeviceFunctionMAC(t *testing.T) {
	mux := http.NewServeMux()
	for path, fixture := range map[string]string{
		"/redfish/v1/":          "serviceroot.json",
		"/redfish/v1/Systems":   "systems.json",
		"/redfish/v1/Systems/1": "systems_1.json",

		"/redfish/v1/Systems/1/NetworkInterfaces":                                                    "smc_aoc_disabled_function/network_interfaces.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1":                                         "smc_aoc_disabled_function/network_interface_slot_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter":                          "smc_aoc_disabled_function/network_adapter.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkPorts":             "smc_aoc_disabled_function/network_ports.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkPorts/1":           "smc_aoc_disabled_function/network_port_1.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkDeviceFunctions":   "smc_aoc_disabled_function/network_device_functions.json",
		"/redfish/v1/Systems/1/NetworkInterfaces/NIC.Slot.1/NetworkAdapter/NetworkDeviceFunctions/1": "smc_aoc_disabled_function/network_device_function_1.json",

		"/redfish/v1/Systems/1/EthernetInterfaces": "smc_aoc_disabled_function/ethernet_interfaces.json",
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

	require.Len(t, device.NICs, 1)
	require.Len(t, device.NICs[0].NICPorts, 1)
	assert.Equal(t, "00:00:5e:00:53:05", strings.ToLower(device.NICs[0].NICPorts[0].MacAddress),
		"port MAC should come from AssociatedNetworkAddresses, not be zeroed out by a disabled NetworkDeviceFunction's placeholder MAC")
}

// TestInventoryCollectBMCNIC verifies that the BMC's own management network
// interface, exposed under the Manager resource's EthernetInterfaces, is
// collected into device.BMC.NIC — distinct from the host NICs collected
// under ComputerSystem. The fixture's MACAddress is the placeholder
// "00:00:00:00:00:00", so this also exercises the fallback to
// PermanentMACAddress.
func TestInventoryCollectBMCNIC(t *testing.T) {
	mux := http.NewServeMux()
	for path, fixture := range map[string]string{
		"/redfish/v1/":          "serviceroot.json",
		"/redfish/v1/Systems":   "systems.json",
		"/redfish/v1/Systems/1": "systems_1.json",

		"/redfish/v1/Managers":                        "managers.json",
		"/redfish/v1/Managers/1":                      "managers_1.json",
		"/redfish/v1/Managers/1/EthernetInterfaces":   "bmc_nic/ethernet_interfaces.json",
		"/redfish/v1/Managers/1/EthernetInterfaces/1": "bmc_nic/ethernet_1.json",
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

	require.NotNil(t, device.BMC)
	require.NotNil(t, device.BMC.NIC, "BMC NIC should be populated from Manager EthernetInterfaces")

	require.Len(t, device.BMC.NIC.NICPorts, 1)
	port := device.BMC.NIC.NICPorts[0]
	assert.Equal(t, "00:00:5e:00:53:10", port.MacAddress,
		"MacAddress should fall back to PermanentMACAddress since MACAddress is the placeholder")
	assert.Equal(t, int64(1_000_000_000), port.SpeedBits)
	assert.True(t, port.AutoNeg)

	// The host's NICs must remain unaffected by the BMC's own NIC.
	for _, nic := range device.NICs {
		for _, p := range nic.NICPorts {
			assert.NotEqual(t, "00:00:5e:00:53:10", p.MacAddress,
				"BMC NIC MAC leaked into host NIC inventory")
		}
	}
}
