package redfishwrapper

import (
	"reflect"
	"testing"

	common "github.com/metal-toolbox/bmc-common"
	common2 "github.com/stmcginnis/gofish/common"
	redfish "github.com/stmcginnis/gofish/redfish"
)

func TestInventoryCollectNetworkPortInfo(t *testing.T) {
	testAdapter := &redfish.NetworkAdapter{
		Manufacturer: "Acme",
		Model:        "Anvil 3000",
	}
	testNetworkPort := &redfish.NetworkPort{
		Entity:                     common2.Entity{ID: "NetworkPort-1"},
		Description:                "NetworkPort One",
		VendorID:                   "Vendor-ID",
		PhysicalPortNumber:         "10",
		LinkStatus:                 "Up",
		ActiveLinkTechnology:       "Ethernet",
		CurrentLinkSpeedMbps:       1000,
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
		adapter       *redfish.NetworkAdapter
		networkPort   *redfish.NetworkPort
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
			c.collectNetworkPortInfo(tt.nicPort, tt.adapter, tt.networkPort, tt.firmware, []*redfish.SoftwareInventory{})
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
	testUnmatchingEthList := []*redfish.EthernetInterface{
		{Entity: common2.Entity{ID: "other ID"}},
		{Entity: common2.Entity{ID: "another one"}},
	}
	testMatchingEth := &redfish.EthernetInterface{
		Entity:      common2.Entity{ID: testEthernetID},
		Description: "Ethernet Interface",
		Status: common2.Status{
			Health: "OK",
			State:  "Enabled",
		},
		SpeedMbps:  10000,
		AutoNeg:    true,
		MTUSize:    1500,
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
		MTUSize:    testMatchingEth.MTUSize,
		MacAddress: testMatchingEth.MACAddress,
	}

	tests := []struct {
		name               string
		nicPort            *common.NICPort
		ethernetInterfaces []*redfish.EthernetInterface
		wantedNicPort      *common.NICPort
	}{
		{name: "nil"},
		{name: "empty", nicPort: testNicPort, wantedNicPort: testNicPort},
		{name: "empty ethernet list", nicPort: testNicPort, ethernetInterfaces: []*redfish.EthernetInterface{}, wantedNicPort: testNicPort},
		{name: "unmatching ethernet list", nicPort: testNicPort, ethernetInterfaces: testUnmatchingEthList, wantedNicPort: testNicPort},
		{
			name:               "full",
			nicPort:            testNicPort,
			ethernetInterfaces: testMatchingEthList,
			wantedNicPort:      wNicPortFull},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{}
			c.collectEthernetInfo(tt.nicPort, tt.ethernetInterfaces)
		})
	}
}
