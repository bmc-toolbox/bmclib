package idrac9

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	muxC6420     *http.ServeMux
	serverC6420  *httptest.Server
	AnswersC6420 = map[string][]byte{
		"/sysmgmt/2012/server/inventory/hardware": []byte(`<?xml version="1.0" ?>
		<Inventory version="2.0">
			<Component Classname="DCIM_ControllerView" Key="NonRAID.Mezzanine.1-1">
			   <PROPERTY NAME="PersistentHotspare" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BootVirtualDiskFQDD" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SharedSlotAssignmentAllowed" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ConnectorCount" TYPE="uint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="AlarmState" TYPE="uint8">
				 <VALUE>1</VALUE>
				 <DisplayValue>Alarm Not present</DisplayValue>
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
				 <DisplayValue>Not Supported</DisplayValue>
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
				 <VALUE>54CD98F06C360900</VALUE>
				 <DisplayValue>54CD98F06C360900</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ProductName" TYPE="string">
				 <VALUE>Dell HBA330 Mini</VALUE>
				 <DisplayValue>Dell HBA330 Mini</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceCardSlotType" TYPE="string">
				 <VALUE>PCI Express Gen3</VALUE>
				 <DisplayValue>PCI Express Gen3</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceCardSlotLength" TYPE="uint8">
				 <VALUE>4</VALUE>
				 <DisplayValue>LongLength</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceCardDataBusWidth" TYPE="string">
				 <VALUE>8x or x8</VALUE>
				 <DisplayValue>8x or x8</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceCardManufacturer" TYPE="string">
				 <VALUE>DELL</VALUE>
				 <DisplayValue>DELL</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>1F53</VALUE>
				 <DisplayValue>1F53</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCIDeviceID" TYPE="string">
				 <VALUE>97</VALUE>
				 <DisplayValue>97</DisplayValue>
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
				 <VALUE>16.17.00.03</VALUE>
				 <DisplayValue>16.17.00.03</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISlot" TYPE="uint8">
				 <VALUE/>
				 <DisplayValue/>
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
				 <VALUE>Storage Controller in Mezzanine 1</VALUE>
				 <DisplayValue>Storage Controller in Mezzanine 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>NonRAID.Mezzanine.1-1</VALUE>
				 <DisplayValue>NonRAID.Mezzanine.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>NonRAID.Mezzanine.1-1</VALUE>
				 <DisplayValue>NonRAID.Mezzanine.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191115154259.000000+000</VALUE>
				 <DisplayValue>2019-11-15T15:42:59</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191115154259.000000+000</VALUE>
				 <DisplayValue>2019-11-15T15:42:59</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_ControllerView" Key="AHCI.Embedded.2-1">
			   <PROPERTY NAME="PersistentHotspare" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BootVirtualDiskFQDD" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SharedSlotAssignmentAllowed" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ConnectorCount" TYPE="uint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="AlarmState" TYPE="uint8">
				 <VALUE>1</VALUE>
				 <DisplayValue>Alarm Not present</DisplayValue>
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
				 <DisplayValue>Not Supported</DisplayValue>
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
				 <VALUE>C620 Series Chipset Family SATA Controller [AHCI mode]</VALUE>
				 <DisplayValue>C620 Series Chipset Family SATA Controller [AHCI mode]</DisplayValue>
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
				 <VALUE>757</VALUE>
				 <DisplayValue>757</DisplayValue>
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
				 <VALUE/>
				 <DisplayValue/>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191115141412.000000+000</VALUE>
				 <DisplayValue>2019-11-15T14:14:12</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_ControllerView" Key="AHCI.Embedded.1-1">
			   <PROPERTY NAME="PersistentHotspare" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BootVirtualDiskFQDD" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SharedSlotAssignmentAllowed" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ConnectorCount" TYPE="uint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="AlarmState" TYPE="uint8">
				 <VALUE>1</VALUE>
				 <DisplayValue>Alarm Not present</DisplayValue>
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
				 <DisplayValue>Not Supported</DisplayValue>
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
				 <VALUE>C620 Series Chipset Family SSATA Controller [AHCI mode]</VALUE>
				 <DisplayValue>C620 Series Chipset Family SSATA Controller [AHCI mode]</DisplayValue>
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
				 <VALUE>757</VALUE>
				 <DisplayValue>757</DisplayValue>
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
				 <VALUE/>
				 <DisplayValue/>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191115141412.000000+000</VALUE>
				 <DisplayValue>2019-11-15T14:14:12</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A1">
			   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SystemEraseCapability" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>Not Supported</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CacheSize" TYPE="uint64">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="NonVolatileSize" TYPE="uint64">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="VolatileSize" TYPE="uint64">
				 <VALUE>32768</VALUE>
				 <DisplayValue>32768 MB</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MemoryTechnology" TYPE="uint8">
				 <VALUE>3</VALUE>
				 <DisplayValue>DRAM</DisplayValue>
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
				 <VALUE>Mon Feb 11 13:00:00 2019 UTC</VALUE>
				 <DisplayValue>Mon Feb 11 13:00:00 2019 UTC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Model" TYPE="string">
				 <VALUE>DDR4 DIMM</VALUE>
				 <DisplayValue>DDR4 DIMM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PartNumber" TYPE="string">
				 <VALUE>36ASF4G72PZ-2G9E2</VALUE>
				 <DisplayValue>36ASF4G72PZ-2G9E2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SerialNumber" TYPE="string">
				 <VALUE>20B8682A</VALUE>
				 <DisplayValue>20B8682A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Micron Technology</VALUE>
				 <DisplayValue>Micron Technology</DisplayValue>
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
				 <VALUE>2933</VALUE>
				 <DisplayValue>2933 MHz</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B1">
			   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SystemEraseCapability" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>Not Supported</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CacheSize" TYPE="uint64">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="NonVolatileSize" TYPE="uint64">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="VolatileSize" TYPE="uint64">
				 <VALUE>32768</VALUE>
				 <DisplayValue>32768 MB</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MemoryTechnology" TYPE="uint8">
				 <VALUE>3</VALUE>
				 <DisplayValue>DRAM</DisplayValue>
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
				 <VALUE>Mon Feb 11 13:00:00 2019 UTC</VALUE>
				 <DisplayValue>Mon Feb 11 13:00:00 2019 UTC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Model" TYPE="string">
				 <VALUE>DDR4 DIMM</VALUE>
				 <DisplayValue>DDR4 DIMM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PartNumber" TYPE="string">
				 <VALUE>36ASF4G72PZ-2G9E2</VALUE>
				 <DisplayValue>36ASF4G72PZ-2G9E2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SerialNumber" TYPE="string">
				 <VALUE>20B879FC</VALUE>
				 <DisplayValue>20B879FC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Micron Technology</VALUE>
				 <DisplayValue>Micron Technology</DisplayValue>
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
				 <VALUE>2933</VALUE>
				 <DisplayValue>2933 MHz</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B2">
			   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SystemEraseCapability" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>Not Supported</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CacheSize" TYPE="uint64">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="NonVolatileSize" TYPE="uint64">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="VolatileSize" TYPE="uint64">
				 <VALUE>32768</VALUE>
				 <DisplayValue>32768 MB</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MemoryTechnology" TYPE="uint8">
				 <VALUE>3</VALUE>
				 <DisplayValue>DRAM</DisplayValue>
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
				 <VALUE>Mon Feb 11 13:00:00 2019 UTC</VALUE>
				 <DisplayValue>Mon Feb 11 13:00:00 2019 UTC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Model" TYPE="string">
				 <VALUE>DDR4 DIMM</VALUE>
				 <DisplayValue>DDR4 DIMM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PartNumber" TYPE="string">
				 <VALUE>36ASF4G72PZ-2G9E2</VALUE>
				 <DisplayValue>36ASF4G72PZ-2G9E2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SerialNumber" TYPE="string">
				 <VALUE>20B877AE</VALUE>
				 <DisplayValue>20B877AE</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Micron Technology</VALUE>
				 <DisplayValue>Micron Technology</DisplayValue>
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
				 <VALUE>2933</VALUE>
				 <DisplayValue>2933 MHz</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A2">
			   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SystemEraseCapability" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>Not Supported</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CacheSize" TYPE="uint64">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="NonVolatileSize" TYPE="uint64">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="VolatileSize" TYPE="uint64">
				 <VALUE>32768</VALUE>
				 <DisplayValue>32768 MB</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MemoryTechnology" TYPE="uint8">
				 <VALUE>3</VALUE>
				 <DisplayValue>DRAM</DisplayValue>
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
				 <VALUE>Mon Feb 11 13:00:00 2019 UTC</VALUE>
				 <DisplayValue>Mon Feb 11 13:00:00 2019 UTC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Model" TYPE="string">
				 <VALUE>DDR4 DIMM</VALUE>
				 <DisplayValue>DDR4 DIMM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PartNumber" TYPE="string">
				 <VALUE>36ASF4G72PZ-2G9E2</VALUE>
				 <DisplayValue>36ASF4G72PZ-2G9E2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SerialNumber" TYPE="string">
				 <VALUE>20B87049</VALUE>
				 <DisplayValue>20B87049</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Micron Technology</VALUE>
				 <DisplayValue>Micron Technology</DisplayValue>
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
				 <VALUE>2933</VALUE>
				 <DisplayValue>2933 MHz</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.2-1">
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
				 <VALUE>C620 Series Chipset Family PCI Express Root Port #5</VALUE>
				 <DisplayValue>C620 Series Chipset Family PCI Express Root Port #5</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Intel Corporation</VALUE>
				 <DisplayValue>Intel Corporation</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>0757</VALUE>
				 <DisplayValue>0757</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="AHCI.Embedded.2-1">
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
				 <VALUE>C620 Series Chipset Family SATA Controller [AHCI mode]</VALUE>
				 <DisplayValue>C620 Series Chipset Family SATA Controller [AHCI mode]</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Intel Corporation</VALUE>
				 <DisplayValue>Intel Corporation</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>0757</VALUE>
				 <DisplayValue>0757</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="HostBridge.Embedded.1-1">
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
				 <VALUE>HostBridge.Embedded.1-1</VALUE>
				 <DisplayValue>HostBridge.Embedded.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>HostBridge.Embedded.1-1</VALUE>
				 <DisplayValue>HostBridge.Embedded.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.1-1">
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
				 <VALUE>C620 Series Chipset Family PCI Express Root Port #1</VALUE>
				 <DisplayValue>C620 Series Chipset Family PCI Express Root Port #1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Intel Corporation</VALUE>
				 <DisplayValue>Intel Corporation</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>0757</VALUE>
				 <DisplayValue>0757</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="Video.Embedded.1-1">
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
				 <VALUE>0757</VALUE>
				 <DisplayValue>0757</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="ISABridge.Embedded.1-1">
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
				 <VALUE>C621 Series Chipset LPC/eSPI Controller</VALUE>
				 <DisplayValue>C621 Series Chipset LPC/eSPI Controller</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Intel Corporation</VALUE>
				 <DisplayValue>Intel Corporation</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>0757</VALUE>
				 <DisplayValue>0757</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="AHCI.Embedded.1-1">
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
				 <VALUE>C620 Series Chipset Family SSATA Controller [AHCI mode]</VALUE>
				 <DisplayValue>C620 Series Chipset Family SSATA Controller [AHCI mode]</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Intel Corporation</VALUE>
				 <DisplayValue>Intel Corporation</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>0757</VALUE>
				 <DisplayValue>0757</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="NonRAID.Mezzanine.1-1">
			   <PROPERTY NAME="SlotType" TYPE="string">
				 <VALUE>00B1</VALUE>
				 <DisplayValue>PCI Express Gen 3</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SlotLength" TYPE="string">
				 <VALUE>0004</VALUE>
				 <DisplayValue>Long Length</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DataBusWidth" TYPE="string">
				 <VALUE>000B</VALUE>
				 <DisplayValue>8x or x8</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Description" TYPE="string">
				 <VALUE>HBA330 Mini</VALUE>
				 <DisplayValue>HBA330 Mini</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Broadcom / LSI</VALUE>
				 <DisplayValue>Broadcom / LSI</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>1F53</VALUE>
				 <DisplayValue>1F53</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubVendorID" TYPE="string">
				 <VALUE>1028</VALUE>
				 <DisplayValue>1028</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCIDeviceID" TYPE="string">
				 <VALUE>0097</VALUE>
				 <DisplayValue>0097</DisplayValue>
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
				 <VALUE>Storage Controller in Mezzanine 1</VALUE>
				 <DisplayValue>Storage Controller in Mezzanine 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>NonRAID.Mezzanine.1-1</VALUE>
				 <DisplayValue>NonRAID.Mezzanine.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>NonRAID.Mezzanine.1-1</VALUE>
				 <DisplayValue>NonRAID.Mezzanine.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814100256.000000+000</VALUE>
				 <DisplayValue>2019-08-14T10:02:56</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="NIC.Slot.4-1-1">
			   <PROPERTY NAME="SlotType" TYPE="string">
				 <VALUE>00B1</VALUE>
				 <DisplayValue>PCI Express Gen 3</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SlotLength" TYPE="string">
				 <VALUE>0004</VALUE>
				 <DisplayValue>Long Length</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DataBusWidth" TYPE="string">
				 <VALUE>000D</VALUE>
				 <DisplayValue>16x or x16</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Description" TYPE="string">
				 <VALUE>Ethernet 25G 2P XXV710 Adapter</VALUE>
				 <DisplayValue>Ethernet 25G 2P XXV710 Adapter</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Intel Corporation</VALUE>
				 <DisplayValue>Intel Corporation</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>0009</VALUE>
				 <DisplayValue>0009</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubVendorID" TYPE="string">
				 <VALUE>8086</VALUE>
				 <DisplayValue>8086</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCIDeviceID" TYPE="string">
				 <VALUE>158B</VALUE>
				 <DisplayValue>158B</DisplayValue>
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
				 <VALUE>94</VALUE>
				 <DisplayValue>94</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>NIC in Slot 4 Port 1 Partition 1</VALUE>
				 <DisplayValue>NIC in Slot 4 Port 1 Partition 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>NIC.Slot.4-1-1</VALUE>
				 <DisplayValue>NIC.Slot.4-1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>NIC.Slot.4-1-1</VALUE>
				 <DisplayValue>NIC.Slot.4-1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190830101919.000000+000</VALUE>
				 <DisplayValue>2019-08-30T10:19:19</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="NIC.Slot.4-2-1">
			   <PROPERTY NAME="SlotType" TYPE="string">
				 <VALUE>00B1</VALUE>
				 <DisplayValue>PCI Express Gen 3</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SlotLength" TYPE="string">
				 <VALUE>0004</VALUE>
				 <DisplayValue>Long Length</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DataBusWidth" TYPE="string">
				 <VALUE>000D</VALUE>
				 <DisplayValue>16x or x16</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Description" TYPE="string">
				 <VALUE>Ethernet Network Adapter XXV710</VALUE>
				 <DisplayValue>Ethernet Network Adapter XXV710</DisplayValue>
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
				 <VALUE>158B</VALUE>
				 <DisplayValue>158B</DisplayValue>
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
				 <VALUE>94</VALUE>
				 <DisplayValue>94</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>NIC in Slot 4 Port 2 Partition 1</VALUE>
				 <DisplayValue>NIC in Slot 4 Port 2 Partition 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>NIC.Slot.4-2-1</VALUE>
				 <DisplayValue>NIC.Slot.4-2-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>NIC.Slot.4-2-1</VALUE>
				 <DisplayValue>NIC.Slot.4-2-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190830101919.000000+000</VALUE>
				 <DisplayValue>2019-08-30T10:19:19</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PCIDeviceView" Key="SMBus.Embedded.3-1">
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
				 <VALUE>C620 Series Chipset Family SMBus</VALUE>
				 <DisplayValue>C620 Series Chipset Family SMBus</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Intel Corporation</VALUE>
				 <DisplayValue>Intel Corporation</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
				 <VALUE>0757</VALUE>
				 <DisplayValue>0757</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_SystemView" Key="System.Embedded.1">
			   <PROPERTY NAME="EstimatedExhaustTemperature" TYPE="uint16">
				 <VALUE>255</VALUE>
				 <DisplayValue>Not applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="EstimatedSystemAirflow" TYPE="uint16">
				 <VALUE>255</VALUE>
				 <DisplayValue>Not applicable</DisplayValue>
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
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="TempStatisticsRollupStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SDCardRollupStatus" TYPE="uint32">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentRollupStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BatteryRollupStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FanRollupStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="IDSDMRollupStatus" TYPE="uint32">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="IntrusionRollupStatus" TYPE="uint32">
				 <VALUE/>
				 <DisplayValue/>
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
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CPURollupStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PopulatedPCIeSlots" TYPE="uint32">
				 <VALUE>2</VALUE>
				 <DisplayValue>2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MaxPCIeSlots" TYPE="uint32">
				 <VALUE>4</VALUE>
				 <DisplayValue>4</DisplayValue>
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
				 <DisplayValue>False</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BladeGeometry" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CPLDVersion" TYPE="string">
				 <VALUE>1.0.7</VALUE>
				 <DisplayValue>1.0.7</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BoardPartNumber" TYPE="string">
				 <VALUE>0YTVTTA03</VALUE>
				 <DisplayValue>0YTVTTA03</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BoardSerialNumber" TYPE="string">
				 <VALUE>CNWS30097M005W</VALUE>
				 <DisplayValue>CNWS30097M005W</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ChassisName" TYPE="string">
				 <VALUE>C6400</VALUE>
				 <DisplayValue>C6400</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ServerAllocation" TYPE="uint32">
				 <VALUE/>
				 <DisplayValue/>
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
				 <VALUE>NA</VALUE>
				 <DisplayValue>NA</DisplayValue>
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
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="UUID" TYPE="string">
				 <VALUE>4c4c4544-005a-3510-805a-b8c04f335a32</VALUE>
				 <DisplayValue>4c4c4544-005a-3510-805a-b8c04f335a32</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="smbiosGUID" TYPE="string">
				 <VALUE>44454c4c-5a00-1035-805a-b8c04f335a32</VALUE>
				 <DisplayValue>44454c4c-5a00-1035-805a-b8c04f335a32</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PlatformGUID" TYPE="string">
				 <VALUE>325a334f-c0b8-5a80-3510-005a4c4c4544</VALUE>
				 <DisplayValue>325a334f-c0b8-5a80-3510-005a4c4c4544</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SystemID" TYPE="uint32">
				 <VALUE>1879</VALUE>
				 <DisplayValue>1879</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BIOSReleaseDate" TYPE="string">
				 <VALUE>10/27/2019</VALUE>
				 <DisplayValue>10/27/2019</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BIOSVersionString" TYPE="string">
				 <VALUE>2.4.7</VALUE>
				 <DisplayValue>2.4.7</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SystemRevision" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>I</DisplayValue>
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
				 <VALUE>2</VALUE>
				 <DisplayValue>2 U</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="NodeID" TYPE="string">
				 <VALUE>8Z5Z3Z2</VALUE>
				 <DisplayValue>8Z5Z3Z2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ChassisModel" TYPE="string">
				 <VALUE>PowerEdge C6400</VALUE>
				 <DisplayValue>PowerEdge C6400</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ChassisServiceTag" TYPE="string">
				 <VALUE>9Z5Z3Z2</VALUE>
				 <DisplayValue>9Z5Z3Z2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ExpressServiceCode" TYPE="string">
				 <VALUE>19540611038</VALUE>
				 <DisplayValue>19540611038</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ServiceTag" TYPE="string">
				 <VALUE>8Z5Z3Z2</VALUE>
				 <DisplayValue>8Z5Z3Z2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <VALUE>Dell Inc.</VALUE>
				 <DisplayValue>Dell Inc.</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Model" TYPE="string">
				 <VALUE>PowerEdge C6420</VALUE>
				 <DisplayValue>PowerEdge C6420</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LifecycleControllerVersion" TYPE="string">
				 <VALUE>3.36.36.38</VALUE>
				 <DisplayValue>3.36.36.38</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CMCIP" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SystemGeneration" TYPE="string">
				 <VALUE>14G DCS</VALUE>
				 <DisplayValue>14G DCS</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191114154155.000000+000</VALUE>
				 <DisplayValue>2019-11-14T15:41:55</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_NICView" Key="NIC.Slot.4-1-1">
			   <PROPERTY NAME="Protocol" TYPE="string">
				 <VALUE>NIC</VALUE>
				 <DisplayValue>NIC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MediaType" TYPE="string">
				 <VALUE>SFF_CAGE</VALUE>
				 <DisplayValue>SFF_CAGE</DisplayValue>
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
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SlotType" TYPE="string">
				 <VALUE>00B1</VALUE>
				 <DisplayValue>PCI Express Gen 3</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SlotLength" TYPE="string">
				 <VALUE>0004</VALUE>
				 <DisplayValue>Long Length</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DataBusWidth" TYPE="string">
				 <VALUE>000D</VALUE>
				 <DisplayValue>16x or x16</DisplayValue>
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
				 <VALUE>0009</VALUE>
				 <DisplayValue>0009</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCIDeviceID" TYPE="string">
				 <VALUE>158b</VALUE>
				 <DisplayValue>158b</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubVendorID" TYPE="string">
				 <VALUE>8086</VALUE>
				 <DisplayValue>8086</DisplayValue>
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
				 <VALUE>94</VALUE>
				 <DisplayValue>94</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
				 <VALUE>3C:FD:FE:DF:8E:C0</VALUE>
				 <DisplayValue>3C:FD:FE:DF:8E:C0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentMACAddress" TYPE="string">
				 <VALUE>3C:FD:FE:DF:8E:C0</VALUE>
				 <DisplayValue>3C:FD:FE:DF:8E:C0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ProductName" TYPE="string">
				 <VALUE>Intel(R) Ethernet 25G 2P XXV710 Adapter - 3C:FD:FE:DF:8E:C0</VALUE>
				 <DisplayValue>Intel(R) Ethernet 25G 2P XXV710 Adapter - 3C:FD:FE:DF:8E:C0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VirtWWN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="VirtWWPN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="WWN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="WWPN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="EFIVersion" TYPE="string">
				 <VALUE>3.3.37</VALUE>
				 <DisplayValue>3.3.37</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ControllerBIOSVersion" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="FamilyVersion" TYPE="string">
				 <VALUE>18.8.9</VALUE>
				 <DisplayValue>18.8.9</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MinBandwidth" TYPE="uint16">
				 <VALUE>25</VALUE>
				 <DisplayValue>25</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MaxBandwidth" TYPE="uint16">
				 <VALUE>100</VALUE>
				 <DisplayValue>100</DisplayValue>
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
				 <VALUE>2</VALUE>
				 <DisplayValue>Enabled</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>NIC in Slot 4 Port 1 Partition 1</VALUE>
				 <DisplayValue>NIC in Slot 4 Port 1 Partition 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>NIC.Slot.4-1-1</VALUE>
				 <DisplayValue>NIC.Slot.4-1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>NIC.Slot.4-1-1</VALUE>
				 <DisplayValue>NIC.Slot.4-1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190830101919.000000+000</VALUE>
				 <DisplayValue>2019-08-30T10:19:19</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191115141412.000000+000</VALUE>
				 <DisplayValue>2019-11-15T14:14:12</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_NICView" Key="NIC.Slot.4-2-1">
			   <PROPERTY NAME="Protocol" TYPE="string">
				 <VALUE>NIC</VALUE>
				 <DisplayValue>NIC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MediaType" TYPE="string">
				 <VALUE>SFF_CAGE</VALUE>
				 <DisplayValue>SFF_CAGE</DisplayValue>
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
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SlotType" TYPE="string">
				 <VALUE>00B1</VALUE>
				 <DisplayValue>PCI Express Gen 3</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SlotLength" TYPE="string">
				 <VALUE>0004</VALUE>
				 <DisplayValue>Long Length</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DataBusWidth" TYPE="string">
				 <VALUE>000D</VALUE>
				 <DisplayValue>16x or x16</DisplayValue>
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
				 <VALUE>0000</VALUE>
				 <DisplayValue>0000</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCIDeviceID" TYPE="string">
				 <VALUE>158b</VALUE>
				 <DisplayValue>158b</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubVendorID" TYPE="string">
				 <VALUE>8086</VALUE>
				 <DisplayValue>8086</DisplayValue>
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
				 <VALUE>94</VALUE>
				 <DisplayValue>94</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
				 <VALUE>3C:FD:FE:DF:8E:C1</VALUE>
				 <DisplayValue>3C:FD:FE:DF:8E:C1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentMACAddress" TYPE="string">
				 <VALUE>3C:FD:FE:DF:8E:C1</VALUE>
				 <DisplayValue>3C:FD:FE:DF:8E:C1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ProductName" TYPE="string">
				 <VALUE>Intel(R) Ethernet Network Adapter XXV710 - 3C:FD:FE:DF:8E:C1</VALUE>
				 <DisplayValue>Intel(R) Ethernet Network Adapter XXV710 - 3C:FD:FE:DF:8E:C1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VirtWWN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="VirtWWPN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="WWN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="WWPN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="EFIVersion" TYPE="string">
				 <VALUE>3.3.37</VALUE>
				 <DisplayValue>3.3.37</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ControllerBIOSVersion" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="FamilyVersion" TYPE="string">
				 <VALUE>18.8.9</VALUE>
				 <DisplayValue>18.8.9</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MinBandwidth" TYPE="uint16">
				 <VALUE>25</VALUE>
				 <DisplayValue>25</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MaxBandwidth" TYPE="uint16">
				 <VALUE>100</VALUE>
				 <DisplayValue>100</DisplayValue>
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
				 <VALUE>2</VALUE>
				 <DisplayValue>Enabled</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>NIC in Slot 4 Port 2 Partition 1</VALUE>
				 <DisplayValue>NIC in Slot 4 Port 2 Partition 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>NIC.Slot.4-2-1</VALUE>
				 <DisplayValue>NIC.Slot.4-2-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>NIC.Slot.4-2-1</VALUE>
				 <DisplayValue>NIC.Slot.4-2-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190830101919.000000+000</VALUE>
				 <DisplayValue>2019-08-30T10:19:19</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191115141412.000000+000</VALUE>
				 <DisplayValue>2019-11-15T14:14:12</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_NICView" Key="NIC.Embedded.1-1-1">
			   <PROPERTY NAME="Protocol" TYPE="string">
				 <VALUE>NIC</VALUE>
				 <DisplayValue>NIC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MediaType" TYPE="string">
				 <VALUE>Base T</VALUE>
				 <DisplayValue>Base T</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ReceiveFlowControl" TYPE="uint8">
				 <VALUE>2</VALUE>
				 <DisplayValue>On</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="TransmitFlowControl" TYPE="uint8">
				 <VALUE>3</VALUE>
				 <DisplayValue>Off</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="AutoNegotiation" TYPE="uint8">
				 <VALUE>2</VALUE>
				 <DisplayValue>Enabled</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LinkSpeed" TYPE="uint8">
				 <VALUE>3</VALUE>
				 <DisplayValue>1000 Mbps</DisplayValue>
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
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SlotType" TYPE="string">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="SlotLength" TYPE="string">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="DataBusWidth" TYPE="string">
				 <VALUE/>
				 <DisplayValue/>
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
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="PCIDeviceID" TYPE="string">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="PCISubVendorID" TYPE="string">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="PCIVendorID" TYPE="string">
				 <VALUE/>
				 <DisplayValue/>
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
			   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentMACAddress" TYPE="string">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="ProductName" TYPE="string">
				 <VALUE>I350 GbE Controller</VALUE>
				 <DisplayValue>I350 GbE Controller</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VirtWWN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="VirtWWPN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="WWN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="WWPN" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="EFIVersion" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="ControllerBIOSVersion" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="FamilyVersion" TYPE="string">
				 <DisplayValue/>
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
				 <VALUE>Embedded NIC 1 Port 1 Partition 1</VALUE>
				 <DisplayValue>Embedded NIC 1 Port 1 Partition 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>NIC.Embedded.1-1-1</VALUE>
				 <DisplayValue>NIC.Embedded.1-1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>NIC.Embedded.1-1-1</VALUE>
				 <DisplayValue>NIC.Embedded.1-1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>19700101000000.000000+000</VALUE>
				 <DisplayValue>1970-01-01T00:00:00</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>19700101000000.000000+000</VALUE>
				 <DisplayValue>1970-01-01T00:00:00</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_VideoView" Key="Video.Embedded.1-1">
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
				 <VALUE>0757</VALUE>
				 <DisplayValue>0757</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PowerSupplyView" Key="PSU.Slot.1">
			   <PROPERTY NAME="LineStatus" TYPE="uint8">
				 <VALUE>2</VALUE>
				 <DisplayValue>High line</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="EffectiveCapacity" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PMBusMonitoring" TYPE="uint8">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="Range1MaxInputPower" TYPE="uint32">
				 <VALUE>2260</VALUE>
				 <DisplayValue>2260 Watts</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedMinNumberNeeded" TYPE="uint32">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="Type" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>AC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="TotalOutputPower" TYPE="uint32">
				 <VALUE>2000</VALUE>
				 <DisplayValue>2000 Watts</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InputVoltage" TYPE="uint32">
				 <VALUE>228</VALUE>
				 <DisplayValue>228 Volts</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FirmwareVersion" TYPE="string">
				 <VALUE>0.11.1a</VALUE>
				 <DisplayValue>0.11.1a</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DetailedState" TYPE="string">
				 <VALUE>Presence Detected</VALUE>
				 <DisplayValue>Presence Detected</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="PartNumber" TYPE="string">
				 <VALUE>0J5WMGA02</VALUE>
				 <DisplayValue>0J5WMGA02</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SerialNumber" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="Model" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Power Supply 1</VALUE>
				 <DisplayValue>Power Supply 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>PSU.Slot.1</VALUE>
				 <DisplayValue>PSU.Slot.1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>PSU.Slot.1</VALUE>
				 <DisplayValue>PSU.Slot.1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090336.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:36</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PowerSupplyView" Key="PSU.Slot.2">
			   <PROPERTY NAME="LineStatus" TYPE="uint8">
				 <VALUE>2</VALUE>
				 <DisplayValue>High line</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="EffectiveCapacity" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PMBusMonitoring" TYPE="uint8">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="Range1MaxInputPower" TYPE="uint32">
				 <VALUE>2260</VALUE>
				 <DisplayValue>2260 Watts</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedMinNumberNeeded" TYPE="uint32">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				 <VALUE/>
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="Type" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>AC</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="TotalOutputPower" TYPE="uint32">
				 <VALUE>2000</VALUE>
				 <DisplayValue>2000 Watts</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InputVoltage" TYPE="uint32">
				 <VALUE>230</VALUE>
				 <DisplayValue>230 Volts</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FirmwareVersion" TYPE="string">
				 <VALUE>0.11.1a</VALUE>
				 <DisplayValue>0.11.1a</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DetailedState" TYPE="string">
				 <VALUE>Presence Detected</VALUE>
				 <DisplayValue>Presence Detected</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Manufacturer" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="PartNumber" TYPE="string">
				 <VALUE>0J5WMGA02</VALUE>
				 <DisplayValue>0J5WMGA02</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SerialNumber" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="Model" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Power Supply 2</VALUE>
				 <DisplayValue>Power Supply 2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>PSU.Slot.2</VALUE>
				 <DisplayValue>PSU.Slot.2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>PSU.Slot.2</VALUE>
				 <DisplayValue>PSU.Slot.2</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090336.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:36</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_PhysicalDiskView" Key="Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1">
			   <PROPERTY NAME="ForeignKeyIdentifier" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="RAIDType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>Unknown</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SystemEraseCapability" TYPE="uint8">
				 <VALUE>2</VALUE>
				 <DisplayValue>CryptographicErasePD</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
				 <VALUE>100</VALUE>
				 <DisplayValue>100 %</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="OperationPercentComplete" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>0 %</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="OperationName" TYPE="string">
				 <VALUE>None</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY.ARRAY NAME="SupportedEncryptionTypes" TYPE="string">
			  <VALUE.ARRAY>
			   <VALUE>None</VALUE>
			   <DisplayValue>No encryption supported</DisplayValue>
			   </VALUE.ARRAY>
			   </PROPERTY.ARRAY>
			   <PROPERTY NAME="DriveFormFactor" TYPE="uint8">
				 <VALUE>2</VALUE>
				 <DisplayValue>2.5 inch</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PPID" TYPE="string">
				 <VALUE>TW033R2TPIHIT94K011TA01</VALUE>
				 <DisplayValue>TW033R2TPIHIT94K011TA01</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SASAddress" TYPE="string">
				 <VALUE>4433221104000000</VALUE>
				 <DisplayValue>4433221104000000</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="MaxCapableSpeed" TYPE="uint32">
				 <VALUE>3</VALUE>
				 <DisplayValue>6 Gbps</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UsedSizeInBytes" TYPE="uint64">
				 <VALUE>0</VALUE>
				 <DisplayValue>0 Bytes</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FreeSizeInBytes" TYPE="uint64">
				 <VALUE>0</VALUE>
				 <DisplayValue>0 Bytes</DisplayValue>
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
				 <VALUE>PHYF916300GQ1P9DGN</VALUE>
				 <DisplayValue>PHYF916300GQ1P9DGN</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Revision" TYPE="string">
				 <VALUE>XCV1DL63</VALUE>
				 <DisplayValue>XCV1DL63</DisplayValue>
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
				 <VALUE>INTEL</VALUE>
				 <DisplayValue>INTEL</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Model" TYPE="string">
				 <VALUE>SSDSC2KB019T8R</VALUE>
				 <DisplayValue>SSDSC2KB019T8R</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="SizeInBytes" TYPE="uint64">
				 <VALUE>1920383409664</VALUE>
				 <DisplayValue>1920383409664 Bytes</DisplayValue>
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
				 <VALUE>8</VALUE>
				 <DisplayValue>Non-RAID</DisplayValue>
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
				 <VALUE>Disk 0 in Backplane 1 of Storage Controller in Mezzanine 1</VALUE>
				 <DisplayValue>Disk 0 in Backplane 1 of Storage Controller in Mezzanine 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1</VALUE>
				 <DisplayValue>Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1</VALUE>
				 <DisplayValue>Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191115154259.000000+000</VALUE>
				 <DisplayValue>2019-11-15T15:42:59</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191115154259.000000+000</VALUE>
				 <DisplayValue>2019-11-15T15:42:59</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_iDRACCardView" Key="iDRAC.Embedded.1-1#IDRACinfo">
			   <PROPERTY NAME="DNSDomainName" TYPE="string">
				 <DisplayValue/>
			   </PROPERTY>
			   <PROPERTY NAME="DNSRacName" TYPE="string">
				 <VALUE>iDRAC-8Z5Z3Z2</VALUE>
				 <DisplayValue>iDRAC-8Z5Z3Z2</DisplayValue>
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
				 <VALUE>https://10.193.19.167:443</VALUE>
				 <DisplayValue>https://10.193.19.167:443</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="GUID" TYPE="string">
				 <VALUE>44454c4c-5a00-1035-805a-b8c04f335a32</VALUE>
				 <DisplayValue>44454c4c-5a00-1035-805a-b8c04f335a32</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
				 <VALUE>6c:2b:59:a0:e9:2b</VALUE>
				 <DisplayValue>6c:2b:59:a0:e9:2b</DisplayValue>
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
				 <VALUE>3.36.36.38</VALUE>
				 <DisplayValue>3.36.36.38</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090336.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:36</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_FanView" Key="Fan.Embedded.1A">
			   <PROPERTY NAME="FanType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>NA</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ActiveCooling" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VariableSpeed" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PWM" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentReading" TYPE="sint32">
				 <VALUE>7912</VALUE>
				 <DisplayValue>7912 RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RateUnits" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UnitModifier" TYPE="sint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BaseUnits" TYPE="uint16">
				 <VALUE>19</VALUE>
				 <DisplayValue>RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Fan 1A</VALUE>
				 <DisplayValue>Fan 1A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Fan.Embedded.1A</VALUE>
				 <DisplayValue>Fan.Embedded.1A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Fan.Embedded.1A</VALUE>
				 <DisplayValue>Fan.Embedded.1A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090336.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:36</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_FanView" Key="Fan.Embedded.1B">
			   <PROPERTY NAME="FanType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>NA</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ActiveCooling" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VariableSpeed" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PWM" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentReading" TYPE="sint32">
				 <VALUE>7482</VALUE>
				 <DisplayValue>7482 RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RateUnits" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UnitModifier" TYPE="sint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BaseUnits" TYPE="uint16">
				 <VALUE>19</VALUE>
				 <DisplayValue>RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Fan 1B</VALUE>
				 <DisplayValue>Fan 1B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Fan.Embedded.1B</VALUE>
				 <DisplayValue>Fan.Embedded.1B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Fan.Embedded.1B</VALUE>
				 <DisplayValue>Fan.Embedded.1B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090337.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:37</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_FanView" Key="Fan.Embedded.2A">
			   <PROPERTY NAME="FanType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>NA</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ActiveCooling" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VariableSpeed" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PWM" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentReading" TYPE="sint32">
				 <VALUE>7912</VALUE>
				 <DisplayValue>7912 RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RateUnits" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UnitModifier" TYPE="sint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BaseUnits" TYPE="uint16">
				 <VALUE>19</VALUE>
				 <DisplayValue>RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Fan 2A</VALUE>
				 <DisplayValue>Fan 2A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Fan.Embedded.2A</VALUE>
				 <DisplayValue>Fan.Embedded.2A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Fan.Embedded.2A</VALUE>
				 <DisplayValue>Fan.Embedded.2A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090337.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:37</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_FanView" Key="Fan.Embedded.2B">
			   <PROPERTY NAME="FanType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>NA</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ActiveCooling" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VariableSpeed" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PWM" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentReading" TYPE="sint32">
				 <VALUE>7482</VALUE>
				 <DisplayValue>7482 RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RateUnits" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UnitModifier" TYPE="sint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BaseUnits" TYPE="uint16">
				 <VALUE>19</VALUE>
				 <DisplayValue>RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Fan 2B</VALUE>
				 <DisplayValue>Fan 2B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Fan.Embedded.2B</VALUE>
				 <DisplayValue>Fan.Embedded.2B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Fan.Embedded.2B</VALUE>
				 <DisplayValue>Fan.Embedded.2B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090337.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:37</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_FanView" Key="Fan.Embedded.3A">
			   <PROPERTY NAME="FanType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>NA</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ActiveCooling" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VariableSpeed" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PWM" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentReading" TYPE="sint32">
				 <VALUE>7998</VALUE>
				 <DisplayValue>7998 RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RateUnits" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UnitModifier" TYPE="sint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BaseUnits" TYPE="uint16">
				 <VALUE>19</VALUE>
				 <DisplayValue>RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Fan 3A</VALUE>
				 <DisplayValue>Fan 3A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Fan.Embedded.3A</VALUE>
				 <DisplayValue>Fan.Embedded.3A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Fan.Embedded.3A</VALUE>
				 <DisplayValue>Fan.Embedded.3A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090337.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:37</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_FanView" Key="Fan.Embedded.3B">
			   <PROPERTY NAME="FanType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>NA</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ActiveCooling" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VariableSpeed" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PWM" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentReading" TYPE="sint32">
				 <VALUE>7482</VALUE>
				 <DisplayValue>7482 RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RateUnits" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UnitModifier" TYPE="sint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BaseUnits" TYPE="uint16">
				 <VALUE>19</VALUE>
				 <DisplayValue>RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Fan 3B</VALUE>
				 <DisplayValue>Fan 3B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Fan.Embedded.3B</VALUE>
				 <DisplayValue>Fan.Embedded.3B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Fan.Embedded.3B</VALUE>
				 <DisplayValue>Fan.Embedded.3B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090336.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:36</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_FanView" Key="Fan.Embedded.4A">
			   <PROPERTY NAME="FanType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>NA</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ActiveCooling" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VariableSpeed" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PWM" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentReading" TYPE="sint32">
				 <VALUE>7998</VALUE>
				 <DisplayValue>7998 RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RateUnits" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UnitModifier" TYPE="sint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BaseUnits" TYPE="uint16">
				 <VALUE>19</VALUE>
				 <DisplayValue>RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Fan 4A</VALUE>
				 <DisplayValue>Fan 4A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Fan.Embedded.4A</VALUE>
				 <DisplayValue>Fan.Embedded.4A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Fan.Embedded.4A</VALUE>
				 <DisplayValue>Fan.Embedded.4A</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090336.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:36</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_FanView" Key="Fan.Embedded.4B">
			   <PROPERTY NAME="FanType" TYPE="uint8">
				 <VALUE>0</VALUE>
				 <DisplayValue>NA</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="ActiveCooling" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="VariableSpeed" TYPE="boolean">
				 <VALUE>true</VALUE>
				 <DisplayValue>true</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PWM" TYPE="uint8">
				  <DisplayValue>Not Applicable</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="CurrentReading" TYPE="sint32">
				 <VALUE>7568</VALUE>
				 <DisplayValue>7568 RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="RateUnits" TYPE="uint16">
				 <VALUE>0</VALUE>
				 <DisplayValue>None</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="UnitModifier" TYPE="sint32">
				 <VALUE>0</VALUE>
				 <DisplayValue>0</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="BaseUnits" TYPE="uint16">
				 <VALUE>19</VALUE>
				 <DisplayValue>RPM</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
				 <VALUE>1</VALUE>
				 <DisplayValue>OK</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="DeviceDescription" TYPE="string">
				 <VALUE>Fan 4B</VALUE>
				 <DisplayValue>Fan 4B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Fan.Embedded.4B</VALUE>
				 <DisplayValue>Fan.Embedded.4B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Fan.Embedded.4B</VALUE>
				 <DisplayValue>Fan.Embedded.4B</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191125090337.000000+000</VALUE>
				 <DisplayValue>2019-11-25T09:03:37</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_EnclosureView" Key="Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1">
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
				 <VALUE>6</VALUE>
				 <DisplayValue>6</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Version" TYPE="string">
				 <VALUE>4.27</VALUE>
				 <DisplayValue>4.27</DisplayValue>
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
				 <VALUE>Backplane 1 on Connector 0 of Storage Controller in Mezzanine 1</VALUE>
				 <DisplayValue>Backplane 1 on Connector 0 of Storage Controller in Mezzanine 1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="FQDD" TYPE="string">
				 <VALUE>Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1</VALUE>
				 <DisplayValue>Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="InstanceID" TYPE="string">
				 <VALUE>Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1</VALUE>
				 <DisplayValue>Enclosure.Internal.0-1:NonRAID.Mezzanine.1-1</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20191115154259.000000+000</VALUE>
				 <DisplayValue>2019-11-15T15:42:59</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191115154259.000000+000</VALUE>
				 <DisplayValue>2019-11-15T15:42:59</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_CPUView" Key="CPU.Socket.1">
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
			   <PROPERTY NAME="Cache3InstalledSize" TYPE="uint32">
				 <VALUE>11264</VALUE>
				 <DisplayValue>11264 KB</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Cache2InstalledSize" TYPE="uint32">
				 <VALUE>8192</VALUE>
				 <DisplayValue>8192 KB</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Cache1InstalledSize" TYPE="uint32">
				 <VALUE>512</VALUE>
				 <DisplayValue>512 KB</DisplayValue>
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
				 <DisplayValue>1.8 V</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
			<Component Classname="DCIM_CPUView" Key="CPU.Socket.2">
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
			   <PROPERTY NAME="Cache3InstalledSize" TYPE="uint32">
				 <VALUE>11264</VALUE>
				 <DisplayValue>11264 KB</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Cache2InstalledSize" TYPE="uint32">
				 <VALUE>8192</VALUE>
				 <DisplayValue>8192 KB</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="Cache1InstalledSize" TYPE="uint32">
				 <VALUE>512</VALUE>
				 <DisplayValue>512 KB</DisplayValue>
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
				 <DisplayValue>1.8 V</DisplayValue>
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
			   <PROPERTY NAME="LastUpdateTime" TYPE="string">
				 <VALUE>20190814095230.000000+000</VALUE>
				 <DisplayValue>2019-08-14T09:52:30</DisplayValue>
			   </PROPERTY>
			   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
				 <VALUE>20191125071344.000000+000</VALUE>
				 <DisplayValue>2019-11-25T07:13:44</DisplayValue>
			   </PROPERTY>
			</Component>
		</Inventory>
		`),
		"/sysmgmt/2015/bmc/session/logout":                       []byte(``),
		"/sysmgmt/2015/bmc/session":                              []byte(`{"authResult":0}`),
		"/sysmgmt/2012/server/configgroup/System.ServerTopology": []byte(`{"System.ServerTopology":{"AisleName":"","BladeSlotNumInChassis":"4","DataCenterName":"","RackName":"","RackSlot":"1","RoomName":"","SizeOfManagedSystemInU":"2"}}`),
	}
)

func init() {
	if viper.GetBool("debug") != true {
		viper.SetDefault("debug", true)
	}
}

func setupC6420() (bmc *IDrac9, err error) {
	muxC6420 = http.NewServeMux()
	serverC6420 = httptest.NewTLSServer(muxC6420)
	ip := strings.TrimPrefix(serverC6420.URL, "https://")
	username := "super"
	password := "test"

	for url := range AnswersC6420 {
		url := url
		muxC6420.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(AnswersC6420[url])
		})
	}

	testLogger := logrus.New()
	bmc, err = New(context.TODO(), ip, ip, username, password, logrusr.NewLogger(testLogger))
	if err != nil {
		return bmc, err
	}

	return bmc, err
}

func tearDownC6420() {
	serverC6420.Close()
}

func TestIDracC6420Model(t *testing.T) {
	expectedAnswer := "PowerEdge C6420"

	bmc, err := setupC6420()
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

	tearDownC6420()
}

func TestIDracC6420Slot(t *testing.T) {
	expectedAnswer := 4

	bmc, err := setupC6420()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Slot()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Slot %v", err)
	}

	if expectedAnswer != answer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownC6420()
}
