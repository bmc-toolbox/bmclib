package redfishwrapper

import (
	"reflect"
	"testing"

	"github.com/bmc-toolbox/common"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventoryCollectNetworkPortInfo(t *testing.T) {
	testAdapter := &schemas.NetworkAdapter{
		Manufacturer: "Acme",
		Model:        "Anvil 3000",
	}
	testNetworkPort := &schemas.NetworkPort{
		Entity:                     schemas.Entity{ID: "NetworkPort-1"},
		VendorID:                   "Vendor-ID",
		PhysicalPortNumber:         "10",
		LinkStatus:                 "Up",
		ActiveLinkTechnology:       "Ethernet",
		CurrentLinkSpeedMbps:       gofish.ToRef(1000),
		AssociatedNetworkAddresses: []string{"98:E7:43:00:01:0A"},
	}
	testFirmwareVersion := "1.2.3"

	wNicPortOnlyAdapter := &common.NICPort{Common: common.Common{Vendor: testAdapter.Manufacturer, Model: testAdapter.Model}}
	wNicPortOnlyNPort := &common.NICPort{
		Common: common.Common{
			Description: testNetworkPort.Description,
			PCIVendorID: testNetworkPort.VendorID,
			Status: &common.Status{
				Health: string(testNetworkPort.Status.Health),
				State:  string(testNetworkPort.Status.State),
			},
		},
		ID:                   testNetworkPort.ID,
		PhysicalID:           testNetworkPort.PhysicalPortNumber,
		LinkStatus:           string(testNetworkPort.LinkStatus),
		ActiveLinkTechnology: string(testNetworkPort.ActiveLinkTechnology),
		SpeedBits:            1000000000,
		MacAddress:           testNetworkPort.AssociatedNetworkAddresses[0],
	}
	wNicPortOnlyFirmware := &common.NICPort{Common: common.Common{Firmware: &common.Firmware{Installed: testFirmwareVersion}}}
	wNicPortFull := &common.NICPort{
		Common: common.Common{
			Description: testNetworkPort.Description,
			Vendor:      testAdapter.Manufacturer,
			Model:       testAdapter.Model,
			PCIVendorID: testNetworkPort.VendorID,
			Firmware:    &common.Firmware{Installed: testFirmwareVersion},
			Status: &common.Status{
				Health: string(testNetworkPort.Status.Health),
				State:  string(testNetworkPort.Status.State),
			},
		},
		ID:                   testNetworkPort.ID,
		PhysicalID:           testNetworkPort.PhysicalPortNumber,
		LinkStatus:           string(testNetworkPort.LinkStatus),
		ActiveLinkTechnology: string(testNetworkPort.ActiveLinkTechnology),
		SpeedBits:            1000000000,
		MacAddress:           testNetworkPort.AssociatedNetworkAddresses[0],
	}

	tests := []struct {
		name          string
		nicPort       *common.NICPort
		adapter       *schemas.NetworkAdapter
		networkPort   *schemas.NetworkPort
		firmware      string
		wantedNicPort *common.NICPort
	}{
		{name: "nil"},
		{name: "empty", nicPort: &common.NICPort{}, wantedNicPort: &common.NICPort{}},
		{
			name:          "only adapter",
			nicPort:       &common.NICPort{},
			adapter:       testAdapter,
			wantedNicPort: wNicPortOnlyAdapter,
		},
		{
			name:          "only network port",
			nicPort:       &common.NICPort{},
			networkPort:   testNetworkPort,
			wantedNicPort: wNicPortOnlyNPort,
		},
		{
			name:          "only firmware",
			nicPort:       &common.NICPort{},
			firmware:      testFirmwareVersion,
			wantedNicPort: wNicPortOnlyFirmware,
		},
		{
			name:          "full",
			nicPort:       &common.NICPort{},
			adapter:       testAdapter,
			networkPort:   testNetworkPort,
			firmware:      testFirmwareVersion,
			wantedNicPort: wNicPortFull,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			c.collectNetworkPortInfo(tt.nicPort, tt.adapter, tt.networkPort, tt.firmware, []*schemas.SoftwareInventory{})
			if !reflect.DeepEqual(tt.nicPort, tt.wantedNicPort) {
				t.Errorf("collectNetworkPortInfo() gotNicPort = %v, want %v", tt.nicPort, tt.wantedNicPort)
			}
		})
	}
}

func TestInventoryCollectEthernetInfo(t *testing.T) {
	testNicPortID := "test NIC port ID"
	testEthernetID := "test NIC port ID ethernet"
	testNicPort := &common.NICPort{
		ID: testNicPortID,
	}
	testUnmatchingEthList := make([]*schemas.EthernetInterface, 0, 3)
	testUnmatchingEthList = append(testUnmatchingEthList,
		&schemas.EthernetInterface{Entity: schemas.Entity{ID: "other ID"}},
		&schemas.EthernetInterface{Entity: schemas.Entity{ID: "another one"}},
	)
	testMatchingEth := &schemas.EthernetInterface{
		Entity: schemas.Entity{ID: testEthernetID},
		Status: schemas.Status{
			Health: "OK",
			State:  "Enabled",
		},
		SpeedMbps:  gofish.ToRef(10000),
		AutoNeg:    true,
		MTUSize:    gofish.ToRef(1500),
		MACAddress: "f6:a9:26:e3:e6:32",
	}
	testMatchingEthList := append(testUnmatchingEthList, testMatchingEth)

	wNicPortFull := &common.NICPort{
		Common: common.Common{
			Description: testMatchingEth.Description,
			Status: &common.Status{
				Health: string(testMatchingEth.Status.Health),
				State:  string(testMatchingEth.Status.State),
			},
		},
		ID:         testMatchingEth.ID,
		SpeedBits:  10000000000,
		AutoNeg:    testMatchingEth.AutoNeg,
		MTUSize:    gofish.Deref(testMatchingEth.MTUSize),
		MacAddress: testMatchingEth.MACAddress,
	}

	tests := []struct {
		name               string
		nicPort            *common.NICPort
		ethernetInterfaces []*schemas.EthernetInterface
		wantedNicPort      *common.NICPort
	}{
		{name: "nil"},
		{name: "empty", nicPort: testNicPort, wantedNicPort: testNicPort},
		{name: "empty ethernet list", nicPort: testNicPort, ethernetInterfaces: []*schemas.EthernetInterface{}, wantedNicPort: testNicPort},
		{name: "unmatching ethernet list", nicPort: testNicPort, ethernetInterfaces: testUnmatchingEthList, wantedNicPort: testNicPort},
		{
			name:               "full",
			nicPort:            testNicPort,
			ethernetInterfaces: testMatchingEthList,
			wantedNicPort:      wNicPortFull,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			c.collectEthernetInfo(tt.nicPort, tt.ethernetInterfaces, "")
		})
	}
}

func TestInventoryCollectBIOS(t *testing.T) {
	tests := []struct {
		name       string
		sys        *schemas.ComputerSystem
		wantVendor string
	}{
		{
			name:       "vendor populated from system manufacturer",
			sys:        &schemas.ComputerSystem{Manufacturer: "Supermicro", BiosVersion: "1.4"},
			wantVendor: "supermicro",
		},
		{
			name:       "no manufacturer",
			sys:        &schemas.ComputerSystem{BiosVersion: "1.4"},
			wantVendor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			device := common.NewDevice()
			if err := c.collectBIOS(tt.sys, &device, nil); err != nil {
				t.Fatalf("collectBIOS() error = %v", err)
			}
			if device.BIOS.Vendor != tt.wantVendor {
				t.Errorf("collectBIOS() Vendor = %q, want %q", device.BIOS.Vendor, tt.wantVendor)
			}
			if device.BIOS.Firmware.Installed != tt.sys.BiosVersion {
				t.Errorf("collectBIOS() Firmware.Installed = %q, want %q", device.BIOS.Firmware.Installed, tt.sys.BiosVersion)
			}
		})
	}
}

func TestInventoryCollectMainboard(t *testing.T) {
	tests := []struct {
		name         string
		chassis      *schemas.Chassis
		wantedDevice func() *common.Device
	}{
		{
			name:    "nil chassis",
			chassis: nil,
			wantedDevice: func() *common.Device {
				d := common.NewDevice()
				return &d
			},
		},
		{
			name:    "chassis without a board model",
			chassis: &schemas.Chassis{Manufacturer: "Supermicro"},
			wantedDevice: func() *common.Device {
				d := common.NewDevice()
				return &d
			},
		},
		{
			// Real-world Supermicro shape: the board model/manufacturer are on
			// Chassis.Model/Manufacturer, while PartNumber/SerialNumber
			// describe the case rather than the board itself. The board's own
			// serial is filled in later, by whichever provider serves the request.
			name: "supermicro chassis",
			chassis: &schemas.Chassis{
				Entity:       schemas.Entity{ID: "1"},
				Manufacturer: "Supermicro",
				Model:        "H12SSW-NTR",
				SerialNumber: "TESTCHASSISSERIAL01",
				PartNumber:   "CSE-116TS-RWNBP2-N10-2",
				Status:       schemas.Status{State: "Enabled", Health: "OK"},
			},
			wantedDevice: func() *common.Device {
				d := common.NewDevice()
				d.Mainboard = &common.Mainboard{
					Common: common.Common{
						Vendor:   "supermicro",
						Model:    "H12SSW-NTR",
						Status:   &common.Status{State: "Enabled", Health: "OK"},
						Firmware: &common.Firmware{},
					},
					PhysicalID: "1",
				}
				return &d
			},
		},
		{
			name: "dell chassis",
			chassis: &schemas.Chassis{
				Entity:       schemas.Entity{ID: "1"},
				Manufacturer: "Dell Inc.",
				Model:        "0ABC123",
				Status:       schemas.Status{State: "Enabled", Health: "OK"},
			},
			wantedDevice: func() *common.Device {
				d := common.NewDevice()
				d.Mainboard = &common.Mainboard{
					Common: common.Common{
						Vendor:   "dell",
						Model:    "0ABC123",
						Status:   &common.Status{State: "Enabled", Health: "OK"},
						Firmware: &common.Firmware{},
					},
					PhysicalID: "1",
				}
				return &d
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			device := common.NewDevice()
			err := c.collectMainboard(tt.chassis, &device, nil)
			if err != nil {
				t.Fatalf("collectMainboard() error = %v", err)
			}
			if !reflect.DeepEqual(device.Mainboard, tt.wantedDevice().Mainboard) {
				t.Errorf("collectMainboard() gotMainboard = %+v, want %+v", device.Mainboard, tt.wantedDevice().Mainboard)
			}
		})
	}
}

// TestInventoryCollectMainboardDoesNotOverwrite proves that collectMainboard
// only populates device.Mainboard from the first chassis that provides a
// model. chassisAttributes calls this once per compatible Chassis member, and
// some vendors (e.g. Dell) have more than one compatible Chassis ID; since
// device.Mainboard is a single field, not a slice, a second, unrelated
// chassis (e.g. a drive enclosure/backplane with its own Model set) must not
// silently overwrite the motherboard entry already found.
func TestInventoryCollectMainboardDoesNotOverwrite(t *testing.T) {
	c := Client{}
	device := common.NewDevice()

	motherboard := &schemas.Chassis{
		Entity:       schemas.Entity{ID: "System.Embedded.1"},
		Manufacturer: "Dell Inc.",
		Model:        "0ABC123",
	}
	backplane := &schemas.Chassis{
		Entity:       schemas.Entity{ID: "Enclosure.Internal.0-1"},
		Manufacturer: "Dell Inc.",
		Model:        "BP-UNRELATED-SKU",
	}

	require.NoError(t, c.collectMainboard(motherboard, &device, nil))
	require.NoError(t, c.collectMainboard(backplane, &device, nil))

	require.NotNil(t, device.Mainboard)
	assert.Equal(t, "0ABC123", device.Mainboard.Model,
		"the second chassis's model overwrote the first chassis's mainboard entry")
}
