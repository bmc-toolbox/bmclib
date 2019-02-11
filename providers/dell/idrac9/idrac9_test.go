package idrac9

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/spf13/viper"
)

var (
	mux     *http.ServeMux
	server  *httptest.Server
	Answers = map[string][]byte{
		"/sysmgmt/2012/server/inventory/hardware": []byte(`<?xml version="1.0" ?>
			<Inventory version="2.0">
				<Component Classname="DCIM_ControllerView" Key="RAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180207121335.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:35</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121335.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:35</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ConnectorCount" TYPE="uint32">
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="AlarmState" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Alarm Not present</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RealtimeCapability" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportControllerBootMode" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportEnhancedAutoForeignImport" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxAvailablePCILinkSpeed" TYPE="string">
					 <VALUE>Generation 3</VALUE>
					 <DisplayValue>Generation 3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxPossiblePCILinkSpeed" TYPE="string">
					 <VALUE>Generation 3</VALUE>
					 <DisplayValue>Generation 3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PatrolReadState" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Stopped</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DriverVersion" TYPE="string">
					 <VALUE>--NA--</VALUE>
					 <DisplayValue>--NA--</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CacheSizeInMB" TYPE="uint32">
					 <VALUE>2048</VALUE>
					 <DisplayValue>2048 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportRAID10UnevenSpans" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="T10PICapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlicedVDCapability" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Sliced Virtual Disk creation supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CachecadeCapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Cachecade Virtual Disk not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="KeyID" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="EncryptionCapability" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Local Key Management Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EncryptionMode" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>None</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SecurityStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>Encryption Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					 <VALUE>51866DA0C022D000</VALUE>
					 <DisplayValue>51866DA0C022D000</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>PERC H730P Mini</VALUE>
					 <DisplayValue>PERC H730P Mini</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardSlotType" TYPE="string">
					 <VALUE>Unknown</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardSlotLength" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardDataBusWidth" TYPE="string">
					 <VALUE>Unknown</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardManufacturer" TYPE="string">
					 <VALUE>DELL</VALUE>
					 <DisplayValue>DELL</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>1F48</VALUE>
					 <DisplayValue>1F48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>5D</VALUE>
					 <DisplayValue>5D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>1000</VALUE>
					 <DisplayValue>1000</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Function" TYPE="string">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Device" TYPE="string">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Bus" TYPE="string">
					 <VALUE>3B</VALUE>
					 <DisplayValue>3B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ControllerFirmwareVersion" TYPE="string">
					 <VALUE>25.5.3.0004</VALUE>
					 <DisplayValue>25.5.3.0004</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISlot" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Integrated RAID Controller 1</VALUE>
					 <DisplayValue>Integrated RAID Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>RAID.Integrated.1-1</VALUE>
					 <DisplayValue>RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>RAID.Integrated.1-1</VALUE>
					 <DisplayValue>RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_ControllerView" Key="AHCI.Embedded.2-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030130602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T13:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180123191030.000000+000</VALUE>
					 <DisplayValue>2018-01-23T19:10:30</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ConnectorCount" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="AlarmState" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Alarm Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RealtimeCapability" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Incapable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportControllerBootMode" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not Supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportEnhancedAutoForeignImport" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxAvailablePCILinkSpeed" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxPossiblePCILinkSpeed" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PatrolReadState" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DriverVersion" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CacheSizeInMB" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportRAID10UnevenSpans" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="T10PICapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlicedVDCapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Sliced Virtual Disk creation not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CachecadeCapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Cachecade Virtual Disk not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="KeyID" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="EncryptionCapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>None</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EncryptionMode" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>None</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SecurityStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Encryption Not Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>Lewisburg SATA Controller [AHCI mode]</VALUE>
					 <DisplayValue>Lewisburg SATA Controller [AHCI mode]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardSlotType" TYPE="string">
					 <VALUE>Unknown</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardSlotLength" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardDataBusWidth" TYPE="string">
					 <VALUE>Unknown</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardManufacturer" TYPE="string">
					 <VALUE>DELL</VALUE>
					 <DisplayValue>DELL</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>747</VALUE>
					 <DisplayValue>747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>A182</VALUE>
					 <DisplayValue>A182</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Function" TYPE="string">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Device" TYPE="string">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Bus" TYPE="string">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ControllerFirmwareVersion" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="PCISlot" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RollupStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded AHCI 2</VALUE>
					 <DisplayValue>Embedded AHCI 2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>AHCI.Embedded.2-1</VALUE>
					 <DisplayValue>AHCI.Embedded.2-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>AHCI.Embedded.2-1</VALUE>
					 <DisplayValue>AHCI.Embedded.2-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_ControllerView" Key="AHCI.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030130602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T13:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180123191030.000000+000</VALUE>
					 <DisplayValue>2018-01-23T19:10:30</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ConnectorCount" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="AlarmState" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Alarm Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RealtimeCapability" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Incapable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportControllerBootMode" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not Supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportEnhancedAutoForeignImport" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxAvailablePCILinkSpeed" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxPossiblePCILinkSpeed" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PatrolReadState" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DriverVersion" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CacheSizeInMB" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SupportRAID10UnevenSpans" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="T10PICapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlicedVDCapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Sliced Virtual Disk creation not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CachecadeCapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Cachecade Virtual Disk not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="KeyID" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="EncryptionCapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>None</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EncryptionMode" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>None</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SecurityStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Encryption Not Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>Lewisburg SSATA Controller [AHCI mode]</VALUE>
					 <DisplayValue>Lewisburg SSATA Controller [AHCI mode]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardSlotType" TYPE="string">
					 <VALUE>Unknown</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardSlotLength" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardDataBusWidth" TYPE="string">
					 <VALUE>Unknown</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceCardManufacturer" TYPE="string">
					 <VALUE>DELL</VALUE>
					 <DisplayValue>DELL</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>747</VALUE>
					 <DisplayValue>747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>A1D2</VALUE>
					 <DisplayValue>A1D2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Function" TYPE="string">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Device" TYPE="string">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Bus" TYPE="string">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ControllerFirmwareVersion" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="PCISlot" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RollupStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded AHCI 1</VALUE>
					 <DisplayValue>Embedded AHCI 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>AHCI.Embedded.1-1</VALUE>
					 <DisplayValue>AHCI.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>AHCI.Embedded.1-1</VALUE>
					 <DisplayValue>AHCI.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180123103206.000000+000</VALUE>
					 <DisplayValue>2018-01-23T10:32:06</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Rank" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>Double Rank</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufactureDate" TYPE="string">
					 <VALUE>Mon Aug 28 14:00:00 2017 UTC</VALUE>
					 <DisplayValue>Mon Aug 28 14:00:00 2017 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7AFR4N-VK</VALUE>
					 <DisplayValue>HMA84GR7AFR4N-VK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>11A03F9C</VALUE>
					 <DisplayValue>11A03F9C</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Hynix Semiconductor</VALUE>
					 <DisplayValue>Hynix Semiconductor</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BankLabel" TYPE="string">
					 <VALUE>B</VALUE>
					 <DisplayValue>B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Size" TYPE="uint32">
					 <VALUE>32768</VALUE>
					 <DisplayValue>32768 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentOperatingSpeed" TYPE="uint32">
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Speed" TYPE="uint32">
					 <VALUE>2666</VALUE>
					 <DisplayValue>2666 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM B1</VALUE>
					 <DisplayValue>DIMM B1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.B1</VALUE>
					 <DisplayValue>DIMM.Socket.B1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.B1</VALUE>
					 <DisplayValue>DIMM.Socket.B1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180123103206.000000+000</VALUE>
					 <DisplayValue>2018-01-23T10:32:06</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Rank" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>Double Rank</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufactureDate" TYPE="string">
					 <VALUE>Mon Aug 28 14:00:00 2017 UTC</VALUE>
					 <DisplayValue>Mon Aug 28 14:00:00 2017 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7AFR4N-VK</VALUE>
					 <DisplayValue>HMA84GR7AFR4N-VK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>11A03FA4</VALUE>
					 <DisplayValue>11A03FA4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Hynix Semiconductor</VALUE>
					 <DisplayValue>Hynix Semiconductor</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BankLabel" TYPE="string">
					 <VALUE>B</VALUE>
					 <DisplayValue>B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Size" TYPE="uint32">
					 <VALUE>32768</VALUE>
					 <DisplayValue>32768 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentOperatingSpeed" TYPE="uint32">
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Speed" TYPE="uint32">
					 <VALUE>2666</VALUE>
					 <DisplayValue>2666 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM B2</VALUE>
					 <DisplayValue>DIMM B2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.B2</VALUE>
					 <DisplayValue>DIMM.Socket.B2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.B2</VALUE>
					 <DisplayValue>DIMM.Socket.B2</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180123103206.000000+000</VALUE>
					 <DisplayValue>2018-01-23T10:32:06</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Rank" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>Double Rank</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufactureDate" TYPE="string">
					 <VALUE>Mon Aug 28 14:00:00 2017 UTC</VALUE>
					 <DisplayValue>Mon Aug 28 14:00:00 2017 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7AFR4N-VK</VALUE>
					 <DisplayValue>HMA84GR7AFR4N-VK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>11A03FA2</VALUE>
					 <DisplayValue>11A03FA2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Hynix Semiconductor</VALUE>
					 <DisplayValue>Hynix Semiconductor</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BankLabel" TYPE="string">
					 <VALUE>A</VALUE>
					 <DisplayValue>A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Size" TYPE="uint32">
					 <VALUE>32768</VALUE>
					 <DisplayValue>32768 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentOperatingSpeed" TYPE="uint32">
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Speed" TYPE="uint32">
					 <VALUE>2666</VALUE>
					 <DisplayValue>2666 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM A2</VALUE>
					 <DisplayValue>DIMM A2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.A2</VALUE>
					 <DisplayValue>DIMM.Socket.A2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.A2</VALUE>
					 <DisplayValue>DIMM.Socket.A2</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180123103206.000000+000</VALUE>
					 <DisplayValue>2018-01-23T10:32:06</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Rank" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>Double Rank</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufactureDate" TYPE="string">
					 <VALUE>Mon Aug 28 14:00:00 2017 UTC</VALUE>
					 <DisplayValue>Mon Aug 28 14:00:00 2017 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7AFR4N-VK</VALUE>
					 <DisplayValue>HMA84GR7AFR4N-VK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>11A03FD6</VALUE>
					 <DisplayValue>11A03FD6</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Hynix Semiconductor</VALUE>
					 <DisplayValue>Hynix Semiconductor</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BankLabel" TYPE="string">
					 <VALUE>A</VALUE>
					 <DisplayValue>A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Size" TYPE="uint32">
					 <VALUE>32768</VALUE>
					 <DisplayValue>32768 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentOperatingSpeed" TYPE="uint32">
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Speed" TYPE="uint32">
					 <VALUE>2666</VALUE>
					 <DisplayValue>2666 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM A1</VALUE>
					 <DisplayValue>DIMM A1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.A1</VALUE>
					 <DisplayValue>DIMM.Socket.A1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.A1</VALUE>
					 <DisplayValue>DIMM.Socket.A1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="SMBus.Embedded.3-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Lewisburg SMBus</VALUE>
					 <DisplayValue>Lewisburg SMBus</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0747</VALUE>
					 <DisplayValue>0747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>A1A3</VALUE>
					 <DisplayValue>A1A3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>4</VALUE>
					 <DisplayValue>4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>31</VALUE>
					 <DisplayValue>31</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded SM Bus 3</VALUE>
					 <DisplayValue>Embedded SM Bus 3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>SMBus.Embedded.3-1</VALUE>
					 <DisplayValue>SMBus.Embedded.3-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>SMBus.Embedded.3-1</VALUE>
					 <DisplayValue>SMBus.Embedded.3-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.4-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Sky Lake-E PCI Express Root Port A</VALUE>
					 <DisplayValue>Sky Lake-E PCI Express Root Port A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0000</VALUE>
					 <DisplayValue>0000</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>2030</VALUE>
					 <DisplayValue>2030</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>58</VALUE>
					 <DisplayValue>58</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded P2P Bridge 4-1</VALUE>
					 <DisplayValue>Embedded P2P Bridge 4-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.4-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.4-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.4-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.4-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.2-2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Sky Lake-E PCI Express Root Port A</VALUE>
					 <DisplayValue>Sky Lake-E PCI Express Root Port A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0000</VALUE>
					 <DisplayValue>0000</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>2030</VALUE>
					 <DisplayValue>2030</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>23</VALUE>
					 <DisplayValue>23</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded P2P Bridge 2-2</VALUE>
					 <DisplayValue>Embedded P2P Bridge 2-2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.2-2</VALUE>
					 <DisplayValue>P2PBridge.Embedded.2-2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.2-2</VALUE>
					 <DisplayValue>P2PBridge.Embedded.2-2</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="AHCI.Embedded.2-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Lewisburg SATA Controller [AHCI mode]</VALUE>
					 <DisplayValue>Lewisburg SATA Controller [AHCI mode]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0747</VALUE>
					 <DisplayValue>0747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>A182</VALUE>
					 <DisplayValue>A182</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>23</VALUE>
					 <DisplayValue>23</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded AHCI 2</VALUE>
					 <DisplayValue>Embedded AHCI 2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>AHCI.Embedded.2-1</VALUE>
					 <DisplayValue>AHCI.Embedded.2-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>AHCI.Embedded.2-1</VALUE>
					 <DisplayValue>AHCI.Embedded.2-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.2-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Lewisburg PCI Express Root Port #5</VALUE>
					 <DisplayValue>Lewisburg PCI Express Root Port #5</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0747</VALUE>
					 <DisplayValue>0747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>A194</VALUE>
					 <DisplayValue>A194</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>4</VALUE>
					 <DisplayValue>4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>28</VALUE>
					 <DisplayValue>28</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded P2P Bridge 2-1</VALUE>
					 <DisplayValue>Embedded P2P Bridge 2-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.2-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.2-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.2-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.2-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="Video.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Integrated Matrox G200eW3 Graphics Controller</VALUE>
					 <DisplayValue>Integrated Matrox G200eW3 Graphics Controller</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Matrox Electronics Systems Ltd.</VALUE>
					 <DisplayValue>Matrox Electronics Systems Ltd.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0747</VALUE>
					 <DisplayValue>0747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>0536</VALUE>
					 <DisplayValue>0536</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>102B</VALUE>
					 <DisplayValue>102B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>3</VALUE>
					 <DisplayValue>3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded Video Controller 1</VALUE>
					 <DisplayValue>Embedded Video Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Video.Embedded.1-1</VALUE>
					 <DisplayValue>Video.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Video.Embedded.1-1</VALUE>
					 <DisplayValue>Video.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="ISABridge.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Lewisburg LPC Controller</VALUE>
					 <DisplayValue>Lewisburg LPC Controller</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0747</VALUE>
					 <DisplayValue>0747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>A1C1</VALUE>
					 <DisplayValue>A1C1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>31</VALUE>
					 <DisplayValue>31</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded ISA Bridge 1</VALUE>
					 <DisplayValue>Embedded ISA Bridge 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>ISABridge.Embedded.1-1</VALUE>
					 <DisplayValue>ISABridge.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>ISABridge.Embedded.1-1</VALUE>
					 <DisplayValue>ISABridge.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="AHCI.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Lewisburg SSATA Controller [AHCI mode]</VALUE>
					 <DisplayValue>Lewisburg SSATA Controller [AHCI mode]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0747</VALUE>
					 <DisplayValue>0747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>A1D2</VALUE>
					 <DisplayValue>A1D2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>5</VALUE>
					 <DisplayValue>5</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>17</VALUE>
					 <DisplayValue>17</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded AHCI 1</VALUE>
					 <DisplayValue>Embedded AHCI 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>AHCI.Embedded.1-1</VALUE>
					 <DisplayValue>AHCI.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>AHCI.Embedded.1-1</VALUE>
					 <DisplayValue>AHCI.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="HostBridge.Embedded.1-2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Sky Lake-E DMI3 Registers</VALUE>
					 <DisplayValue>Sky Lake-E DMI3 Registers</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0000</VALUE>
					 <DisplayValue>0000</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>2020</VALUE>
					 <DisplayValue>2020</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded Host Bridge 1</VALUE>
					 <DisplayValue>Embedded Host Bridge 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>HostBridge.Embedded.1-2</VALUE>
					 <DisplayValue>HostBridge.Embedded.1-2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>HostBridge.Embedded.1-2</VALUE>
					 <DisplayValue>HostBridge.Embedded.1-2</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Lewisburg PCI Express Root Port #1</VALUE>
					 <DisplayValue>Lewisburg PCI Express Root Port #1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0747</VALUE>
					 <DisplayValue>0747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>A190</VALUE>
					 <DisplayValue>A190</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>28</VALUE>
					 <DisplayValue>28</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded P2P Bridge 1-1</VALUE>
					 <DisplayValue>Embedded P2P Bridge 1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.1-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.1-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="RAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180207101330.000000+000</VALUE>
					 <DisplayValue>2018-02-07T10:13:30</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>PERC H730P Mini (for blades)</VALUE>
					 <DisplayValue>PERC H730P Mini (for blades)</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>LSI Logic / Symbios Logic</VALUE>
					 <DisplayValue>LSI Logic / Symbios Logic</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>1F48</VALUE>
					 <DisplayValue>1F48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>005D</VALUE>
					 <DisplayValue>005D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>1000</VALUE>
					 <DisplayValue>1000</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>59</VALUE>
					 <DisplayValue>59</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Integrated RAID Controller 1</VALUE>
					 <DisplayValue>Integrated RAID Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>RAID.Integrated.1-1</VALUE>
					 <DisplayValue>RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>RAID.Integrated.1-1</VALUE>
					 <DisplayValue>RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="NIC.Integrated.1-1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030131100.000000+000</VALUE>
					 <DisplayValue>2017-10-30T13:11:00</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>10GbE 2P X520k bNDC</VALUE>
					 <DisplayValue>10GbE 2P X520k bNDC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>1F63</VALUE>
					 <DisplayValue>1F63</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>10F8</VALUE>
					 <DisplayValue>10F8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>24</VALUE>
					 <DisplayValue>24</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Integrated NIC 1 Port 1 Partition 1</VALUE>
					 <DisplayValue>Integrated NIC 1 Port 1 Partition 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>NIC.Integrated.1-1-1</VALUE>
					 <DisplayValue>NIC.Integrated.1-1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>NIC.Integrated.1-1-1</VALUE>
					 <DisplayValue>NIC.Integrated.1-1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="NIC.Integrated.1-2-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171122115214.000000+000</VALUE>
					 <DisplayValue>2017-11-22T11:52:14</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>10GbE 2P X520k bNDC</VALUE>
					 <DisplayValue>10GbE 2P X520k bNDC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>1F63</VALUE>
					 <DisplayValue>1F63</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>10F8</VALUE>
					 <DisplayValue>10F8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>24</VALUE>
					 <DisplayValue>24</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Integrated NIC 1 Port 2 Partition 1</VALUE>
					 <DisplayValue>Integrated NIC 1 Port 2 Partition 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>NIC.Integrated.1-2-1</VALUE>
					 <DisplayValue>NIC.Integrated.1-2-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>NIC.Integrated.1-2-1</VALUE>
					 <DisplayValue>NIC.Integrated.1-2-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_SystemView" Key="System.Embedded.1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180123113206.000000+000</VALUE>
					 <DisplayValue>2018-01-23T11:32:06</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EstimatedExhaustTemperature" TYPE="uint16">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EstimatedSystemAirflow" TYPE="uint16">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SELRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LicensingRollupStatus" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="StorageRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TempStatisticsRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SDCardRollupStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BatteryRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FanRollupStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="IDSDMRollupStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="IntrusionRollupStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VoltRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TempRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PSRollupStatus" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CPURollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PopulatedPCIeSlots" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxPCIeSlots" TYPE="uint32">
					 <VALUE>3</VALUE>
					 <DisplayValue>3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PopulatedCPUSockets" TYPE="uint32">
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxCPUSockets" TYPE="uint32">
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="IsOEMBranded" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BladeGeometry" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>Single width, half height</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CPLDVersion" TYPE="string">
					 <VALUE>1.0.0</VALUE>
					 <DisplayValue>1.0.0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BoardPartNumber" TYPE="string">
					 <VALUE>05YC4PA01</VALUE>
					 <DisplayValue>05YC4PA01</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BoardSerialNumber" TYPE="string">
					 <VALUE>CNWS30079D00BW</VALUE>
					 <DisplayValue>CNWS30079D00BW</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ChassisName" TYPE="string">
					 <VALUE>CMC-H1645M2</VALUE>
					 <DisplayValue>CMC-H1645M2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ServerAllocation" TYPE="uint32">
					 <VALUE>320</VALUE>
					 <DisplayValue>320 Watts</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PowerCap" TYPE="uint32">
					 <VALUE>32767</VALUE>
					 <DisplayValue>32767 Watts</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PowerCapEnabledState" TYPE="uint16">
					 <VALUE>3</VALUE>
					 <DisplayValue>Disabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PowerState" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>On</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BaseBoardChassisSlot" TYPE="string">
					 <VALUE>Slot 02</VALUE>
					 <DisplayValue>Slot 02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PopulatedDIMMSlots" TYPE="uint32">
					 <VALUE>4</VALUE>
					 <DisplayValue>4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryOperationMode" TYPE="string">
					 <VALUE>OptimizerMode</VALUE>
					 <DisplayValue>OptimizerMode</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SysMemFailOverState" TYPE="string">
					 <VALUE>NotInUse</VALUE>
					 <DisplayValue>NotInUse</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SysMemPrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="AssetTag" TYPE="string">
					 <VALUE/>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="HostName" TYPE="string">
					 <VALUE>fdi</VALUE>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="UUID" TYPE="string">
					 <VALUE>4c4c4544-0031-3610-805a-c8c04f344d32</VALUE>
					 <DisplayValue>4c4c4544-0031-3610-805a-c8c04f344d32</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="smbiosGUID" TYPE="string">
					 <VALUE>44454c4c-3100-1036-805a-c8c04f344d32</VALUE>
					 <DisplayValue>44454c4c-3100-1036-805a-c8c04f344d32</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PlatformGUID" TYPE="string">
					 <VALUE>324d344f-c0c8-5a80-3610-00314c4c4544</VALUE>
					 <DisplayValue>324d344f-c0c8-5a80-3610-00314c4c4544</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SystemID" TYPE="uint32">
					 <VALUE>1863</VALUE>
					 <DisplayValue>1863</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BIOSReleaseDate" TYPE="string">
					 <VALUE>12/06/2017</VALUE>
					 <DisplayValue>12/06/2017</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BIOSVersionString" TYPE="string">
					 <VALUE>1.2.71</VALUE>
					 <DisplayValue>1.2.71</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SystemRevision" TYPE="uint16">
					 <VALUE/>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="MaxDIMMSlots" TYPE="uint32">
					 <VALUE>16</VALUE>
					 <DisplayValue>16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SysMemErrorMethodology" TYPE="uint16">
					 <VALUE>6</VALUE>
					 <DisplayValue>Multi-bit ECC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SysMemLocation" TYPE="uint16">
					 <VALUE>3</VALUE>
					 <DisplayValue>System board or motherboard</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SysMemMaxCapacitySize" TYPE="uint64">
					 <VALUE>2097152</VALUE>
					 <DisplayValue>2097152 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SysMemTotalSize" TYPE="uint32">
					 <VALUE>131072</VALUE>
					 <DisplayValue>131072 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ChassisSystemHeight" TYPE="uint16">
					 <VALUE>10</VALUE>
					 <DisplayValue>10U</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NodeID" TYPE="string">
					 <VALUE>H16Z4M2</VALUE>
					 <DisplayValue>H16Z4M2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ChassisModel" TYPE="string">
					 <VALUE>PowerEdge M1000e</VALUE>
					 <DisplayValue>PowerEdge M1000e</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ChassisServiceTag" TYPE="string">
					 <VALUE>H1645M2</VALUE>
					 <DisplayValue>H1645M2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ExpressServiceCode" TYPE="string">
					 <VALUE>37077482522</VALUE>
					 <DisplayValue>37077482522</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ServiceTag" TYPE="string">
					 <VALUE>H16Z4M2</VALUE>
					 <DisplayValue>H16Z4M2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Dell Inc.</VALUE>
					 <DisplayValue>Dell Inc.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>PowerEdge M640</VALUE>
					 <DisplayValue>PowerEdge M640</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LifecycleControllerVersion" TYPE="string">
					 <VALUE>3.15.15.15</VALUE>
					 <DisplayValue>3.15.15.15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CMCIP" TYPE="string">
					 <VALUE>10.193.244.84</VALUE>
					 <DisplayValue>10.193.244.84</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SystemGeneration" TYPE="string">
					 <VALUE>14G Modular</VALUE>
					 <DisplayValue>14G Modular</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>System</VALUE>
					 <DisplayValue>System</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>System.Embedded.1</VALUE>
					 <DisplayValue>System.Embedded.1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>System.Embedded.1</VALUE>
					 <DisplayValue>System.Embedded.1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_NICView" Key="NIC.Integrated.1-1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030141100.000000+000</VALUE>
					 <DisplayValue>2017-10-30T14:11:00</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180123191030.000000+000</VALUE>
					 <DisplayValue>2018-01-23T19:10:30</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Protocol" TYPE="string">
					 <VALUE>NIC</VALUE>
					 <DisplayValue>NIC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MediaType" TYPE="string">
					 <VALUE>KR,KX</VALUE>
					 <DisplayValue>KR,KX</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ReceiveFlowControl" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>On</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TransmitFlowControl" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>On</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="AutoNegotiation" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>Enabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LinkSpeed" TYPE="uint8">
					 <VALUE>5</VALUE>
					 <DisplayValue>10 Gbps</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LinkDuplex" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Full Duplex</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VendorName" TYPE="string">
					 <VALUE>Intel Corp</VALUE>
					 <DisplayValue>Intel Corp</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FCoEWWNN" TYPE="string">
					 <VALUE>24:6e:96:78:33:0d</VALUE>
					 <DisplayValue>24:6e:96:78:33:0d</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentFCOEMACAddress" TYPE="string">
					 <VALUE/>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentiSCSIMACAddress" TYPE="string">
					 <VALUE/>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>1f63</VALUE>
					 <DisplayValue>1f63</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>10f8</VALUE>
					 <DisplayValue>10f8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>24</VALUE>
					 <DisplayValue>24</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
					 <VALUE>24:6E:96:78:33:0C</VALUE>
					 <DisplayValue>24:6E:96:78:33:0C</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentMACAddress" TYPE="string">
					 <VALUE>24:6E:96:78:33:0C</VALUE>
					 <DisplayValue>24:6E:96:78:33:0C</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>Intel(R) Ethernet 10G 2P X520-k bNDC - 24:6E:96:78:33:0C</VALUE>
					 <DisplayValue>Intel(R) Ethernet 10G 2P X520-k bNDC - 24:6E:96:78:33:0C</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VirtWWN" TYPE="string">
					 <VALUE>20:00:24:6E:96:78:33:0D</VALUE>
					 <DisplayValue>20:00:24:6E:96:78:33:0D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VirtWWPN" TYPE="string">
					 <VALUE>20:01:24:6E:96:78:33:0D</VALUE>
					 <DisplayValue>20:01:24:6E:96:78:33:0D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="WWN" TYPE="string">
					 <VALUE>10:00:24:6E:96:78:33:0D</VALUE>
					 <DisplayValue>10:00:24:6E:96:78:33:0D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="WWPN" TYPE="string">
					 <VALUE>20:00:24:6E:96:78:33:0D</VALUE>
					 <DisplayValue>20:00:24:6E:96:78:33:0D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EFIVersion" TYPE="string">
					 <VALUE>5.8.12</VALUE>
					 <DisplayValue>5.8.12</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ControllerBIOSVersion" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="FamilyVersion" TYPE="string">
					 <VALUE>18.0.17</VALUE>
					 <DisplayValue>18.0.17</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MinBandwidth" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxBandwidth" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="iScsiOffloadMode" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Disabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FCoEOffloadMode" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Disabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NicMode" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Disabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Integrated NIC 1 Port 1 Partition 1</VALUE>
					 <DisplayValue>Integrated NIC 1 Port 1 Partition 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>NIC.Integrated.1-1-1</VALUE>
					 <DisplayValue>NIC.Integrated.1-1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>NIC.Integrated.1-1-1</VALUE>
					 <DisplayValue>NIC.Integrated.1-1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_NICView" Key="NIC.Integrated.1-2-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171122125214.000000+000</VALUE>
					 <DisplayValue>2017-11-22T12:52:14</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180123191030.000000+000</VALUE>
					 <DisplayValue>2018-01-23T19:10:30</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Protocol" TYPE="string">
					 <VALUE>NIC</VALUE>
					 <DisplayValue>NIC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MediaType" TYPE="string">
					 <VALUE>KR,KX</VALUE>
					 <DisplayValue>KR,KX</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ReceiveFlowControl" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Off</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TransmitFlowControl" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Off</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="AutoNegotiation" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Disabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LinkSpeed" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LinkDuplex" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VendorName" TYPE="string">
					 <VALUE>Intel Corp</VALUE>
					 <DisplayValue>Intel Corp</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FCoEWWNN" TYPE="string">
					 <VALUE>24:6e:96:78:33:0f</VALUE>
					 <DisplayValue>24:6e:96:78:33:0f</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentFCOEMACAddress" TYPE="string">
					 <VALUE/>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentiSCSIMACAddress" TYPE="string">
					 <VALUE/>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>1f63</VALUE>
					 <DisplayValue>1f63</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>10f8</VALUE>
					 <DisplayValue>10f8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>24</VALUE>
					 <DisplayValue>24</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
					 <VALUE>24:6E:96:78:33:0E</VALUE>
					 <DisplayValue>24:6E:96:78:33:0E</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentMACAddress" TYPE="string">
					 <VALUE>24:6E:96:78:33:0E</VALUE>
					 <DisplayValue>24:6E:96:78:33:0E</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>Intel(R) Ethernet 10G 2P X520-k bNDC - 24:6E:96:78:33:0E</VALUE>
					 <DisplayValue>Intel(R) Ethernet 10G 2P X520-k bNDC - 24:6E:96:78:33:0E</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VirtWWN" TYPE="string">
					 <VALUE>20:00:24:6E:96:78:33:0F</VALUE>
					 <DisplayValue>20:00:24:6E:96:78:33:0F</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VirtWWPN" TYPE="string">
					 <VALUE>20:01:24:6E:96:78:33:0F</VALUE>
					 <DisplayValue>20:01:24:6E:96:78:33:0F</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="WWN" TYPE="string">
					 <VALUE>10:00:24:6E:96:78:33:0F</VALUE>
					 <DisplayValue>10:00:24:6E:96:78:33:0F</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="WWPN" TYPE="string">
					 <VALUE>20:00:24:6E:96:78:33:0F</VALUE>
					 <DisplayValue>20:00:24:6E:96:78:33:0F</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EFIVersion" TYPE="string">
					 <VALUE>5.8.12</VALUE>
					 <DisplayValue>5.8.12</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ControllerBIOSVersion" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="FamilyVersion" TYPE="string">
					 <VALUE>18.0.17</VALUE>
					 <DisplayValue>18.0.17</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MinBandwidth" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxBandwidth" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="iScsiOffloadMode" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Disabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FCoEOffloadMode" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Disabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NicMode" TYPE="uint8">
					 <VALUE>3</VALUE>
					 <DisplayValue>Disabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Integrated NIC 1 Port 2 Partition 1</VALUE>
					 <DisplayValue>Integrated NIC 1 Port 2 Partition 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>NIC.Integrated.1-2-1</VALUE>
					 <DisplayValue>NIC.Integrated.1-2-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>NIC.Integrated.1-2-1</VALUE>
					 <DisplayValue>NIC.Integrated.1-2-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_ControllerBatteryView" Key="Battery.Integrated.1:RAID.Integrated.1-1">
				   <PROPERTY NAME="RAIDState" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Ready</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Battery on Integrated RAID Controller 1</VALUE>
					 <DisplayValue>Battery on Integrated RAID Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Battery.Integrated.1:RAID.Integrated.1-1</VALUE>
					 <DisplayValue>Battery.Integrated.1:RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Battery.Integrated.1:RAID.Integrated.1-1</VALUE>
					 <DisplayValue>Battery.Integrated.1:RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_VideoView" Key="Video.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171030120602.000000+000</VALUE>
					 <DisplayValue>2017-10-30T12:06:02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotType" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotLength" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DataBusWidth" TYPE="string">
					 <VALUE>0002</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Description" TYPE="string">
					 <VALUE>Integrated Matrox G200eW3 Graphics Controller</VALUE>
					 <DisplayValue>Integrated Matrox G200eW3 Graphics Controller</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Matrox Electronics Systems Ltd.</VALUE>
					 <DisplayValue>Matrox Electronics Systems Ltd.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0747</VALUE>
					 <DisplayValue>0747</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>0536</VALUE>
					 <DisplayValue>0536</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>102B</VALUE>
					 <DisplayValue>102B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>3</VALUE>
					 <DisplayValue>3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded Video Controller 1</VALUE>
					 <DisplayValue>Embedded Video Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Video.Embedded.1-1</VALUE>
					 <DisplayValue>Video.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Video.Embedded.1-1</VALUE>
					 <DisplayValue>Video.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PhysicalDiskView" Key="Disk.Bay.0:Enclosure.Internal.0-1:RAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180207111330.000000+000</VALUE>
					 <DisplayValue>2018-02-07T11:13:30</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RAIDType" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SystemEraseCapability" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
					 <VALUE>99</VALUE>
					 <DisplayValue>99%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="OperationPercentComplete" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="OperationName" TYPE="string">
					 <VALUE>None</VALUE>
					 <DisplayValue>None</DisplayValue>
				   </PROPERTY>
				   <PROPERTY.ARRAY NAME="SupportedEncryptionTypes" TYPE="string">
				  <VALUE.ARRAY>
				   <VALUE>None</VALUE>
				   <DisplayValue>None</DisplayValue>
				   </VALUE.ARRAY>
				   </PROPERTY.ARRAY>
				   <PROPERTY NAME="DriveFormFactor" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>2.5 inch</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PPID" TYPE="string">
					 <VALUE>CN09Y3HDSSX0077H005VA00</VALUE>
					 <DisplayValue>CN09Y3HDSSX0077H005VA00</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					 <VALUE>4433221104000000</VALUE>
					 <DisplayValue>4433221104000000</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxCapableSpeed" TYPE="uint32">
					 <VALUE>3</VALUE>
					 <DisplayValue>6Gbs</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="UsedSizeInBytes" TYPE="uint64">
					 <VALUE>0</VALUE>
					 <DisplayValue>0 Bytes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FreeSizeInBytes" TYPE="uint64">
					 <VALUE>3840103415808</VALUE>
					 <DisplayValue>3840103415808 Bytes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MediaType" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>Solid State Drive</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BlockSizeInBytes" TYPE="uint32">
					 <VALUE>512</VALUE>
					 <DisplayValue>512 Bytes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="T10PICapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SecurityState" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PredictiveFailureState" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Smart Alert Absent</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="HotSpareStatus" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>No</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusProtocol" TYPE="uint32">
					 <VALUE>5</VALUE>
					 <DisplayValue>SATA</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>S37MNX0J700554</VALUE>
					 <DisplayValue>S37MNX0J700554</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Revision" TYPE="string">
					 <VALUE>GC57</VALUE>
					 <DisplayValue>GC57</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufacturingYear" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufacturingWeek" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufacturingDay" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>ATA</VALUE>
					 <DisplayValue>ATA</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>MZ7LM3T8HMLP0D3</VALUE>
					 <DisplayValue>MZ7LM3T8HMLP0D3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SizeInBytes" TYPE="uint64">
					 <VALUE>3840103415808</VALUE>
					 <DisplayValue>3840103415808 Bytes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Slot" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Connector" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RaidStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>Ready</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Disk 0 in Backplane 1 of Integrated RAID Controller 1</VALUE>
					 <DisplayValue>Disk 0 in Backplane 1 of Integrated RAID Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Disk.Bay.0:Enclosure.Internal.0-1:RAID.Integrated.1-1</VALUE>
					 <DisplayValue>Disk.Bay.0:Enclosure.Internal.0-1:RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Disk.Bay.0:Enclosure.Internal.0-1:RAID.Integrated.1-1</VALUE>
					 <DisplayValue>Disk.Bay.0:Enclosure.Internal.0-1:RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PhysicalDiskView" Key="Disk.Bay.1:Enclosure.Internal.0-1:RAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180207111330.000000+000</VALUE>
					 <DisplayValue>2018-02-07T11:13:30</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RAIDType" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SystemEraseCapability" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
					 <VALUE>99</VALUE>
					 <DisplayValue>99%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="OperationPercentComplete" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="OperationName" TYPE="string">
					 <VALUE>None</VALUE>
					 <DisplayValue>None</DisplayValue>
				   </PROPERTY>
				   <PROPERTY.ARRAY NAME="SupportedEncryptionTypes" TYPE="string">
				  <VALUE.ARRAY>
				   <VALUE>None</VALUE>
				   <DisplayValue>None</DisplayValue>
				   </VALUE.ARRAY>
				   </PROPERTY.ARRAY>
				   <PROPERTY NAME="DriveFormFactor" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>2.5 inch</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PPID" TYPE="string">
					 <VALUE>CN09Y3HDSSX0077H005YA00</VALUE>
					 <DisplayValue>CN09Y3HDSSX0077H005YA00</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					 <VALUE>4433221105000000</VALUE>
					 <DisplayValue>4433221105000000</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxCapableSpeed" TYPE="uint32">
					 <VALUE>3</VALUE>
					 <DisplayValue>6Gbs</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="UsedSizeInBytes" TYPE="uint64">
					 <VALUE>0</VALUE>
					 <DisplayValue>0 Bytes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FreeSizeInBytes" TYPE="uint64">
					 <VALUE>3840103415808</VALUE>
					 <DisplayValue>3840103415808 Bytes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MediaType" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>Solid State Drive</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BlockSizeInBytes" TYPE="uint32">
					 <VALUE>512</VALUE>
					 <DisplayValue>512 Bytes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="T10PICapability" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not supported</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SecurityState" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Not Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PredictiveFailureState" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>Smart Alert Absent</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="HotSpareStatus" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>No</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusProtocol" TYPE="uint32">
					 <VALUE>5</VALUE>
					 <DisplayValue>SATA</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>S37MNX0J700557</VALUE>
					 <DisplayValue>S37MNX0J700557</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Revision" TYPE="string">
					 <VALUE>GC57</VALUE>
					 <DisplayValue>GC57</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufacturingYear" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufacturingWeek" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ManufacturingDay" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>ATA</VALUE>
					 <DisplayValue>ATA</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>MZ7LM3T8HMLP0D3</VALUE>
					 <DisplayValue>MZ7LM3T8HMLP0D3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SizeInBytes" TYPE="uint64">
					 <VALUE>3840103415808</VALUE>
					 <DisplayValue>3840103415808 Bytes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Slot" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Connector" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RaidStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>Ready</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Disk 1 in Backplane 1 of Integrated RAID Controller 1</VALUE>
					 <DisplayValue>Disk 1 in Backplane 1 of Integrated RAID Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Disk.Bay.1:Enclosure.Internal.0-1:RAID.Integrated.1-1</VALUE>
					 <DisplayValue>Disk.Bay.1:Enclosure.Internal.0-1:RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Disk.Bay.1:Enclosure.Internal.0-1:RAID.Integrated.1-1</VALUE>
					 <DisplayValue>Disk.Bay.1:Enclosure.Internal.0-1:RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_iDRACCardView" Key="iDRAC.Embedded.1-1#IDRACinfo">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180207123526.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:35:26</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DNSDomainName" TYPE="string">
					 <VALUE>machine.example.com</VALUE>
					 <DisplayValue>machine.example.com</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DNSRacName" TYPE="string">
					 <VALUE>spare-H16Z4M2</VALUE>
					 <DisplayValue>spare-H16Z4M2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SOLEnabledState" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Enabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LANEnabledState" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Enabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="URLString" TYPE="string">
					 <VALUE>https://10.193.244.110:443</VALUE>
					 <DisplayValue>https://10.193.244.110:443</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="GUID" TYPE="string">
					 <VALUE>44454c4c-3100-1036-805a-c8c04f344d32</VALUE>
					 <DisplayValue>44454c4c-3100-1036-805a-c8c04f344d32</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
					 <VALUE>50:9a:4c:6e:d7:2c</VALUE>
					 <DisplayValue>50:9a:4c:6e:d7:2c</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductDescription" TYPE="string">
					 <VALUE>This system component provides a complete set of remote management functions for PowerEdge servers</VALUE>
					 <DisplayValue>This system component provides a complete set of remote management functions for PowerEdge servers</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>Enterprise</VALUE>
					 <DisplayValue>Enterprise</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FirmwareVersion" TYPE="string">
					 <VALUE>3.15.15.15</VALUE>
					 <DisplayValue>3.15.15.15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="IPMIVersion" TYPE="string">
					 <VALUE>2.0</VALUE>
					 <DisplayValue>2.0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>iDRAC</VALUE>
					 <DisplayValue>iDRAC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>iDRAC.Embedded.1-1</VALUE>
					 <DisplayValue>iDRAC.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>iDRAC.Embedded.1-1#IDRACinfo</VALUE>
					 <DisplayValue>iDRAC.Embedded.1-1#IDRACinfo</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_EnclosureView" Key="Enclosure.Internal.0-1:RAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180207111330.000000+000</VALUE>
					 <DisplayValue>2018-02-07T11:13:30</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>BP14G+ 0:1</VALUE>
					 <DisplayValue>BP14G+ 0:1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TempProbeCount" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FanCount" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PSUCount" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EMMCount" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SlotCount" TYPE="uint8">
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Version" TYPE="string">
					 <VALUE>4.23</VALUE>
					 <DisplayValue>4.23</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="AssetName" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="AssetTag" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="ServiceTag" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="WiredOrder" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Connector" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="State" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>Ready</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Backplane 1 on Connector 0 of Integrated RAID Controller 1</VALUE>
					 <DisplayValue>Backplane 1 on Connector 0 of Integrated RAID Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Enclosure.Internal.0-1:RAID.Integrated.1-1</VALUE>
					 <DisplayValue>Enclosure.Internal.0-1:RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Enclosure.Internal.0-1:RAID.Integrated.1-1</VALUE>
					 <DisplayValue>Enclosure.Internal.0-1:RAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_CPUView" Key="CPU.Socket.2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180123103206.000000+000</VALUE>
					 <DisplayValue>2018-01-23T10:32:06</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Location" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Internal</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Location" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Internal</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Location" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Internal</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Associativity" TYPE="uint16">
					 <VALUE>6</VALUE>
					 <DisplayValue>Fully Associative</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Associativity" TYPE="uint16">
					 <VALUE>8</VALUE>
					 <DisplayValue>16-way Set-Associative</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Associativity" TYPE="uint16">
					 <VALUE>7</VALUE>
					 <DisplayValue>8-way Set-Associative</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Type" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Unified</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Type" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Unified</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Type" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Unified</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3ErrorMethodology" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Single-bit ECC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2ErrorMethodology" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Single-bit ECC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1ErrorMethodology" TYPE="uint16">
					 <VALUE>4</VALUE>
					 <DisplayValue>Parity</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3SRAMType" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2SRAMType" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1SRAMType" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3WritePolicy" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Write Back</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2WritePolicy" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Write Back</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1WritePolicy" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Write Back</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Level" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>L3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Level" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>L2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Level" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>L1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Size" TYPE="uint32">
					 <VALUE>11264</VALUE>
					 <DisplayValue>11264 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Size" TYPE="uint32">
					 <VALUE>8192</VALUE>
					 <DisplayValue>8192 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Size" TYPE="uint32">
					 <VALUE>512</VALUE>
					 <DisplayValue>512 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>Intel(R) Xeon(R) Silver 4110 CPU @ 2.10GHz</VALUE>
					 <DisplayValue>Intel(R) Xeon(R) Silver 4110 CPU @ 2.10GHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel</VALUE>
					 <DisplayValue>Intel</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TurboModeCapable" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TurboModeEnabled" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ExecuteDisabledCapable" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ExecuteDisabledEnabled" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VirtualizationTechnologyCapable" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VirtualizationTechnologyEnabled" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="HyperThreadingCapable" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="HyperThreadingEnabled" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Characteristics" TYPE="uint32">
					 <VALUE>4</VALUE>
					 <DisplayValue>64-bit Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CPUStatus" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>CPU Enabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Voltage" TYPE="string">
					 <VALUE>1.8</VALUE>
					 <DisplayValue>1.8V</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfProcessorCores" TYPE="uint32">
					 <VALUE>8</VALUE>
					 <DisplayValue>8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfEnabledThreads" TYPE="uint32">
					 <VALUE>16</VALUE>
					 <DisplayValue>16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfEnabledCores" TYPE="uint32">
					 <VALUE>8</VALUE>
					 <DisplayValue>8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ExternalBusClockSpeed" TYPE="uint32">
					 <VALUE>9600</VALUE>
					 <DisplayValue>9600 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxClockSpeed" TYPE="uint32">
					 <VALUE>4000</VALUE>
					 <DisplayValue>4000 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentClockSpeed" TYPE="uint32">
					 <VALUE>2100</VALUE>
					 <DisplayValue>2100 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CPUFamily" TYPE="string">
					 <VALUE>B3</VALUE>
					 <DisplayValue>Intel(R) Xeon(TM)</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>CPU 2</VALUE>
					 <DisplayValue>CPU 2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>CPU.Socket.2</VALUE>
					 <DisplayValue>CPU.Socket.2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>CPU.Socket.2</VALUE>
					 <DisplayValue>CPU.Socket.2</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_CPUView" Key="CPU.Socket.1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20180123103206.000000+000</VALUE>
					 <DisplayValue>2018-01-23T10:32:06</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20180207121329.000000+000</VALUE>
					 <DisplayValue>2018-02-07T12:13:29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Location" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Internal</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Location" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Internal</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Location" TYPE="uint8">
					 <VALUE>0</VALUE>
					 <DisplayValue>Internal</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Associativity" TYPE="uint16">
					 <VALUE>6</VALUE>
					 <DisplayValue>Fully Associative</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Associativity" TYPE="uint16">
					 <VALUE>8</VALUE>
					 <DisplayValue>16-way Set-Associative</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Associativity" TYPE="uint16">
					 <VALUE>7</VALUE>
					 <DisplayValue>8-way Set-Associative</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Type" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Unified</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Type" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Unified</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Type" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Unified</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3ErrorMethodology" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Single-bit ECC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2ErrorMethodology" TYPE="uint16">
					 <VALUE>5</VALUE>
					 <DisplayValue>Single-bit ECC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1ErrorMethodology" TYPE="uint16">
					 <VALUE>4</VALUE>
					 <DisplayValue>Parity</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3SRAMType" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2SRAMType" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1SRAMType" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3WritePolicy" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Write Back</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2WritePolicy" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Write Back</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1WritePolicy" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>Write Back</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Level" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>L3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Level" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>L2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Level" TYPE="uint16">
					 <VALUE>0</VALUE>
					 <DisplayValue>L1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache3Size" TYPE="uint32">
					 <VALUE>11264</VALUE>
					 <DisplayValue>11264 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Size" TYPE="uint32">
					 <VALUE>8192</VALUE>
					 <DisplayValue>8192 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Size" TYPE="uint32">
					 <VALUE>512</VALUE>
					 <DisplayValue>512 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>Intel(R) Xeon(R) Silver 4110 CPU @ 2.10GHz</VALUE>
					 <DisplayValue>Intel(R) Xeon(R) Silver 4110 CPU @ 2.10GHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel</VALUE>
					 <DisplayValue>Intel</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TurboModeCapable" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="TurboModeEnabled" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ExecuteDisabledCapable" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ExecuteDisabledEnabled" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VirtualizationTechnologyCapable" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VirtualizationTechnologyEnabled" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="HyperThreadingCapable" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="HyperThreadingEnabled" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Yes</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Characteristics" TYPE="uint32">
					 <VALUE>4</VALUE>
					 <DisplayValue>64-bit Capable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CPUStatus" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>CPU Enabled</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Voltage" TYPE="string">
					 <VALUE>1.8</VALUE>
					 <DisplayValue>1.8V</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfProcessorCores" TYPE="uint32">
					 <VALUE>8</VALUE>
					 <DisplayValue>8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfEnabledThreads" TYPE="uint32">
					 <VALUE>16</VALUE>
					 <DisplayValue>16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfEnabledCores" TYPE="uint32">
					 <VALUE>8</VALUE>
					 <DisplayValue>8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ExternalBusClockSpeed" TYPE="uint32">
					 <VALUE>9600</VALUE>
					 <DisplayValue>9600 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MaxClockSpeed" TYPE="uint32">
					 <VALUE>4000</VALUE>
					 <DisplayValue>4000 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentClockSpeed" TYPE="uint32">
					 <VALUE>2100</VALUE>
					 <DisplayValue>2100 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CPUFamily" TYPE="string">
					 <VALUE>B3</VALUE>
					 <DisplayValue>Intel(R) Xeon(TM)</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>CPU 1</VALUE>
					 <DisplayValue>CPU 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>CPU.Socket.1</VALUE>
					 <DisplayValue>CPU.Socket.1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>CPU.Socket.1</VALUE>
					 <DisplayValue>CPU.Socket.1</DisplayValue>
				   </PROPERTY>
				</Component>
			</Inventory>
			`),
		"/sysmgmt/2012/server/license": []byte(`{"License":{"AUTO_DISCOVERY":1,"AUTO_UPDATE":1,"AVOTON_4CORE":0,"AVOTON_8CORE":0,"BACKUP_RESTORE":1,"BASIC_REMOTE_INVENTORY_EXPORT":1,"BOOT_CAPTURE":1,"CONNECTION_VIEW":1,"CONSOLE_COLLABORATION":1,"DCS_GUI":0,"DEDICATED_NIC":1,"DEVICE_MONITORING":1,"DHCP_CONFIGURE":1,"DIRECTORY_SERVICES":1,"DYNAMIC_DNS":1,"EMAIL_ALERTING":1,"FULL_UI":1,"GROUP_MANAGER":1,"INBAND_FIRMWARE_UPDATE":1,"IPV6":1,"LAST_CRASH_SCREEN_CAPTURE":1,"LAST_CRASH_VIDEO_CAPTURE":1,"LC_UI":1,"LICENSE_UI":1,"LOCKDOWN_MODE":1,"NTP":1,"OME":1,"OOB":1,"PART_REPLACEMENT":1,"POWER_BUDGETING":1,"POWER_MONITORING":1,"RACADM_CLI":1,"REDFISH_EVENT":1,"REMOTE_ASSET_INVENTORY":1,"REMOTE_CONFIGURATION":1,"REMOTE_FILE_SHARE":1,"REMOTE_FIRWARE_UPDATE":1,"REMOTE_OS_DEPLOYMENT":1,"REMOTE_SYSLOG":1,"RESTORE":1,"SECURITY_LOCKOUT":1,"SMASH_CLP":1,"SNMP":1,"SSH":1,"SSH_PK_AUTHEN":1,"SSO":1,"STORAGE_MONITORING":1,"TELNET":1,"TWO_FACTOR_AUTHEN":1,"UPDATE_FROM_REPO":1,"USC_ASSISTED_OS_DEPLOYEMENT":1,"USC_DEVICE_CONFIGURATION":1,"USC_EMBEDDED_DIAGNOSTICS":1,"USC_FIRMWARE_UPDATE":1,"VCONSOLE":1,"VFOLDER":1,"VIRTUAL_FLASH_PARTITIONS":1,"VIRTUAL_NW_CONSOLE":1,"VMEDIA":1,"WSMAN":1}}`),
		"/sysmgmt/2015/server/sensor/power": []byte(`{
			"root":{
			"powergraphdata":{
			"minPowerWatts":"221", "maxPowerWatts":"368", "minPowerAmps":"0", "maxPowerAmps":"20", "lastHourData":{
			"powerGperiodW":"0", "powerInterval":"5", "startTime":"1518007153", "recordCount":"12", "powerData":{
			"record":[
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"116,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5"
			]
			}
			},
			"lastDayData":{
			"powerGperiodW":"1", "powerInterval":"5", "startTime":"1517924353", "recordCount":"288", "powerData":{
			"record":[
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"116,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"105,0.5",
			"118,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"116,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"113,0.5",
			"115,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"116,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"116,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5"
			]
			}
			},
			"lastWeekData":{
			"powerGperiodW":"1", "powerInterval":"35", "startTime":"1517405653", "recordCount":"288", "powerData":{
			"record":[
			"112,0.5",
			"118,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"115,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"116,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"114,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"114,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"116,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"118,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"117,0.6",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5",
			"112,0.5"
			]
			}
			}
			},
			"powermonitordata":{
			"historicalPeak":{
			"startTime":"Mon Oct 30 15:03:36 2017 ", "peakWattTime":"Mon Oct 30 16:25:14 2017 ", "peakWattValue":"213", "peakAmpsTime":"Mon Oct 30 16:25:14 2017 ", "peakAmpsValue":"1.0" },
			"sysHeadRoom":{
			"instantaneous":"720", "peak":"720" },
			"cumReading":{
			"time":"Mon Oct 30 15:03:36 2017 ", "totalUsage":"123.797" },
			"powerCapacity":"0" ,"historicalTrend":{
			"trendData":[
			{
			"time":"Last Hour", "avgUsage":"112", "maxPeak":"142", "maxPeakTime":"Wed Feb  7 14:21:35 2018 ", "minPeak":"102", "minPeakTime":"Wed Feb  7 14:21:11 2018 " },{
			"time":"Last Day", "avgUsage":"112", "maxPeak":"153", "maxPeakTime":"Tue Feb  6 17:42:11 2018 ", "minPeak":"7", "minPeakTime":"Tue Feb  6 22:38:09 2018 " },{
			"time":"Last Week", "avgUsage":"112", "maxPeak":"159", "maxPeakTime":"Tue Feb  6 03:56:11 2018 ", "minPeak":"7", "minPeakTime":"Thu Feb  1 12:13:21 2018 " }
			]
			},"presentReading":{
			"reading":{
			"probeStatus":"2", "probeName":"System Board Pwr Consumption", "reading":"112", "warningThreshold":"760", "failureThreshold":"840", "maxWarningThresholdSettable":"0" }
			}
			,"psuReading":{
			"reading":[
			{
			"probeName":"System Board Current", "psuAmps":"0.5", "psuVolts":"0" }
			]
			}
			}
			}
			}`),
		"/sysmgmt/2012/server/temperature":     []byte(`{"Statistics":"/sysmgmt/2012/server/temperature/statistics","Temperatures":{"iDRAC.Embedded.1#CPU1Temp":{"max_failure":90,"max_warning":"NA","max_warning_settable":0,"min_failure":3,"min_warning":"NA","min_warning_settable":0,"name":"CPU1 Temp","reading":38,"status":2},"iDRAC.Embedded.1#CPU2Temp":{"max_failure":90,"max_warning":"NA","max_warning_settable":0,"min_failure":3,"min_warning":"NA","min_warning_settable":0,"name":"CPU2 Temp","reading":36,"status":2},"iDRAC.Embedded.1#SystemBoardInletTemp":{"max_failure":47,"max_warning":43,"max_warning_settable":1,"min_failure":-7,"min_warning":3,"min_warning_settable":1,"name":"System Board Inlet Temp","reading":22,"status":2}},"is_fresh_air_compliant":1}`),
		"/sysmgmt/2016/server/extended_health": []byte(`{"healthStatus":[2,2,0,0,0,0,2,0,2,2,2,2,2,2,0,2,2,2]}`),
		"/sysmgmt/2015/bmc/session/logout":     []byte(``),
		"/sysmgmt/2015/bmc/session":            []byte(`{"authResult":0}`),
	}
)

func setup() (bmc *IDrac9, err error) {
	viper.SetDefault("debug", true)
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	for url := range Answers {
		url := url
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			w.Write(Answers[url])
		})
	}

	bmc, err = New(ip, username, password)
	if err != nil {
		return bmc, err
	}

	return bmc, err
}

func tearDown() {
	server.Close()
}

func TestIDracSerial(t *testing.T) {
	expectedAnswer := "h16z4m2"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Serial()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Serial %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracModel(t *testing.T) {
	expectedAnswer := "PowerEdge M640"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Model()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Model %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracBmcType(t *testing.T) {
	expectedAnswer := "idrac9"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer := bmc.BmcType()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracBmcVersion(t *testing.T) {
	expectedAnswer := "3.15.15.15"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.BmcVersion()
	if err != nil {
		t.Fatalf("Found errors calling bmc.BmcVersion %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracName(t *testing.T) {
	expectedAnswer := "fdi"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Name()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracStatus(t *testing.T) {
	expectedAnswer := "OK"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Status()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Status %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracMemory(t *testing.T) {
	expectedAnswer := 128

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Memory()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Memory %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracCPU(t *testing.T) {
	expectedAnswerCPUType := "intel(r) xeon(r) silver 4110 cpu"
	expectedAnswerCPUCount := 2
	expectedAnswerCore := 8
	expectedAnswerHyperthread := 16

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	cpuType, cpuCount, core, ht, err := bmc.CPU()
	if err != nil {
		t.Fatalf("Found errors calling bmc.CPU %v", err)
	}

	if cpuType != expectedAnswerCPUType {
		t.Errorf("Expected cpuType answer %v: found %v", expectedAnswerCPUType, cpuType)
	}

	if cpuCount != expectedAnswerCPUCount {
		t.Errorf("Expected cpuCount answer %v: found %v", expectedAnswerCPUCount, cpuCount)
	}

	if core != expectedAnswerCore {
		t.Errorf("Expected core answer %v: found %v", expectedAnswerCore, core)
	}

	if ht != expectedAnswerHyperthread {
		t.Errorf("Expected ht answer %v: found %v", expectedAnswerHyperthread, ht)
	}

	tearDown()
}

func TestIDracBiosVersion(t *testing.T) {
	expectedAnswer := "1.2.71"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.BiosVersion()
	if err != nil {
		t.Fatalf("Found errors calling bmc.BiosVersion %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracPowerKW(t *testing.T) {
	expectedAnswer := 0.112

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerKw()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerKW %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracTempC(t *testing.T) {
	expectedAnswer := 22

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.TempC()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Temp %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracNics(t *testing.T) {
	expectedAnswer := []*devices.Nic{
		{
			MacAddress: "24:6e:96:78:33:0c",
			Name:       "Intel(R) Ethernet 10G 2P X520-k bNDC",
			Up:         true,
			Speed:      "10 Gbps",
		},
		{
			MacAddress: "24:6e:96:78:33:0e",
			Name:       "Intel(R) Ethernet 10G 2P X520-k bNDC",
			Up:         false,
			Speed:      "",
		},
		{
			MacAddress: "50:9a:4c:6e:d7:2c",
			Name:       "bmc",
			Up:         false,
			Speed:      "",
		},
	}

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	nics, err := bmc.Nics()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Nics %v", err)
	}

	if len(nics) != len(expectedAnswer) {
		t.Fatalf("Expected %v nics: found %v nics", len(expectedAnswer), len(nics))
	}

	for pos, nic := range nics {
		if nic.MacAddress != expectedAnswer[pos].MacAddress || nic.Name != expectedAnswer[pos].Name || nic.Speed != expectedAnswer[pos].Speed || nic.Up != expectedAnswer[pos].Up {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], nic)
		}
	}

	tearDown()
}

func TestIDracLicense(t *testing.T) {
	expectedName := "Enterprise"
	expectedLicType := "Licensed"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	name, licType, err := bmc.License()
	if err != nil {
		t.Fatalf("Found errors calling bmc.License %v", err)
	}

	if name != expectedName {
		t.Errorf("Expected name %v: found %v", expectedName, name)
	}

	if licType != expectedLicType {
		t.Errorf("Expected name %v: found %v", expectedLicType, licType)
	}

	tearDown()
}

func TestDiskDisks(t *testing.T) {
	expectedAnswer := []*devices.Disk{
		{
			Serial:    "s37mnx0j700554",
			Type:      "SSD",
			Size:      "3576 GB",
			Model:     "mz7lm3t8hmlp0d3",
			Location:  "Disk 0 in Backplane 1 of Integrated RAID Controller 1",
			Status:    "OK",
			FwVersion: "gc57",
		},
		{
			Serial:    "s37mnx0j700557",
			Type:      "SSD",
			Size:      "3576 GB",
			Model:     "mz7lm3t8hmlp0d3",
			Location:  "Disk 1 in Backplane 1 of Integrated RAID Controller 1",
			Status:    "OK",
			FwVersion: "gc57",
		},
	}

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test hpChassissetup %v", err)
	}

	disks, err := bmc.Disks()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Disks %v", err)
	}

	if len(disks) != len(expectedAnswer) {
		t.Fatalf("Expected %v disks: found %v disks", len(expectedAnswer), len(disks))
	}

	for pos, disk := range disks {
		if disk.Serial != expectedAnswer[pos].Serial ||
			disk.Type != expectedAnswer[pos].Type ||
			disk.Size != expectedAnswer[pos].Size ||
			disk.Status != expectedAnswer[pos].Status ||
			disk.Model != expectedAnswer[pos].Model ||
			disk.FwVersion != expectedAnswer[pos].FwVersion ||
			disk.Location != expectedAnswer[pos].Location {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], disk)
		}
	}

	tearDown()
}

// func TestIDracPsu(t *testing.T) {
// 	expectedAnswer := []*devices.Psu{
// 		&devices.Psu{
// 			Serial:     "65kt7j2_PS1",
// 			CapacityKw: 0.75,
// 			Status:     "OK",
// 			PowerKw:    0.0,
// 		},
// 		&devices.Psu{
// 			Serial:     "65kt7j2_PS2",
// 			CapacityKw: 0.75,
// 			Status:     "OK",
// 			PowerKw:    0.0,
// 		},
// 	}

// 	bmc, err := setup()
// 	if err != nil {
// 		t.Fatalf("Found errors during the test hpChassisSetup %v", err)
// 	}

// 	psus, err := bmc.Psus()
// 	if err != nil {
// 		t.Fatalf("Found errors calling chassis.Psus %v", err)
// 	}

// 	if len(psus) != len(expectedAnswer) {
// 		t.Fatalf("Expected %v psus: found %v psus", len(expectedAnswer), len(psus))
// 	}

// 	for pos, psu := range psus {
// 		if psu.Serial != expectedAnswer[pos].Serial || psu.CapacityKw != expectedAnswer[pos].CapacityKw || psu.PowerKw != expectedAnswer[pos].PowerKw || psu.Status != expectedAnswer[pos].Status {
// 			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], psu)
// 		}
// 	}

// 	tearDown()
// }

func TestIDracIsBlade(t *testing.T) {
	expectedAnswer := true

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.IsBlade()
	if err != nil {
		t.Fatalf("Found errors calling bmc.isBlade %v", err)
	}

	if expectedAnswer != answer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracPoweState(t *testing.T) {
	expectedAnswer := "on"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerState()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerState %v", err)
	}

	if expectedAnswer != answer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIDracInterface(t *testing.T) {
	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	_ = devices.Bmc(bmc)
	_ = devices.Configure(bmc)
	tearDown()
}

func TestUpdateCredentials(t *testing.T) {
	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	bmc.UpdateCredentials("newUsername", "newPassword")

	if bmc.username != "newUsername" {
		t.Fatalf("Expected username to be updated to 'newUsername' but is: %s", bmc.username)
	}

	if bmc.password != "newPassword" {
		t.Fatalf("Expected password to be updated to 'newPassword' but is: %s", bmc.password)
	}

	tearDown()
}
