package idrac8

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
	answers = map[string][]byte{
		"/sysmgmt/2012/server/inventory/hardware": []byte(`
			<?xml version="1.0" ?>
			<Inventory version="2.0">
				<Component Classname="DCIM_ControllerView" Key="NonRAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170614154516.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:45:16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20170614154516.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:45:16</DisplayValue>
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
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					 <VALUE>51866DA089D35F00</VALUE>
					 <DisplayValue>51866DA089D35F00</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>Dell HBA330 Mini</VALUE>
					 <DisplayValue>Dell HBA330 Mini</DisplayValue>
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
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ControllerFirmwareVersion" TYPE="string">
					 <VALUE>13.17.03.00</VALUE>
					 <DisplayValue>13.17.03.00</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISlot" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
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
					 <VALUE>Integrated Storage Controller 1</VALUE>
					 <DisplayValue>Integrated Storage Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_ControllerView" Key="AHCI.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170311001148.000000+000</VALUE>
					 <DisplayValue>2017-03-11T00:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20170614154350.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:43:50</DisplayValue>
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
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>C610/X99 series chipset sSATA Controller [AHCI mode]</VALUE>
					 <DisplayValue>C610/X99 series chipset sSATA Controller [AHCI mode]</DisplayValue>
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
					 <VALUE>601</VALUE>
					 <DisplayValue>601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D62</VALUE>
					 <DisplayValue>8D62</DisplayValue>
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
				<Component Classname="DCIM_ControllerView" Key="AHCI.Embedded.2-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170311001148.000000+000</VALUE>
					 <DisplayValue>2017-03-11T00:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20170614154350.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:43:50</DisplayValue>
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
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					  <DisplayValue>Not Applicable</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>C610/X99 series chipset 6-Port SATA Controller [AHCI mode]</VALUE>
					 <DisplayValue>C610/X99 series chipset 6-Port SATA Controller [AHCI mode]</DisplayValue>
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
					 <VALUE>601</VALUE>
					 <DisplayValue>601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D02</VALUE>
					 <DisplayValue>8D02</DisplayValue>
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
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FDF8</VALUE>
					 <DisplayValue>2876FDF8</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
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
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FD8E</VALUE>
					 <DisplayValue>2876FD8E</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
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
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A3">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FECB</VALUE>
					 <DisplayValue>2876FECB</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM A3</VALUE>
					 <DisplayValue>DIMM A3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.A3</VALUE>
					 <DisplayValue>DIMM.Socket.A3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.A3</VALUE>
					 <DisplayValue>DIMM.Socket.A3</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A4">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FCDD</VALUE>
					 <DisplayValue>2876FCDD</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM A4</VALUE>
					 <DisplayValue>DIMM A4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.A4</VALUE>
					 <DisplayValue>DIMM.Socket.A4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.A4</VALUE>
					 <DisplayValue>DIMM.Socket.A4</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A5">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FCE3</VALUE>
					 <DisplayValue>2876FCE3</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM A5</VALUE>
					 <DisplayValue>DIMM A5</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.A5</VALUE>
					 <DisplayValue>DIMM.Socket.A5</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.A5</VALUE>
					 <DisplayValue>DIMM.Socket.A5</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.A6">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FED5</VALUE>
					 <DisplayValue>2876FED5</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM A6</VALUE>
					 <DisplayValue>DIMM A6</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.A6</VALUE>
					 <DisplayValue>DIMM.Socket.A6</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.A6</VALUE>
					 <DisplayValue>DIMM.Socket.A6</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FD6C</VALUE>
					 <DisplayValue>2876FD6C</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
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
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FF2D</VALUE>
					 <DisplayValue>2876FF2D</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
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
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B3">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FECF</VALUE>
					 <DisplayValue>2876FECF</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM B3</VALUE>
					 <DisplayValue>DIMM B3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.B3</VALUE>
					 <DisplayValue>DIMM.Socket.B3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.B3</VALUE>
					 <DisplayValue>DIMM.Socket.B3</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B4">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FEE5</VALUE>
					 <DisplayValue>2876FEE5</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM B4</VALUE>
					 <DisplayValue>DIMM B4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.B4</VALUE>
					 <DisplayValue>DIMM.Socket.B4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.B4</VALUE>
					 <DisplayValue>DIMM.Socket.B4</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B5">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FE5A</VALUE>
					 <DisplayValue>2876FE5A</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM B5</VALUE>
					 <DisplayValue>DIMM B5</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.B5</VALUE>
					 <DisplayValue>DIMM.Socket.B5</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.B5</VALUE>
					 <DisplayValue>DIMM.Socket.B5</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_MemoryView" Key="DIMM.Socket.B6">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Mon Dec 26 06:00:00 2016 UTC</VALUE>
					 <DisplayValue>Mon Dec 26 06:00:00 2016 UTC</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>DDR4 DIMM</VALUE>
					 <DisplayValue>DDR4 DIMM</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>HMA84GR7MFR4N-UH</VALUE>
					 <DisplayValue>HMA84GR7MFR4N-UH</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>2876FCFE</VALUE>
					 <DisplayValue>2876FCFE</DisplayValue>
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
					 <VALUE>2400</VALUE>
					 <DisplayValue>2400 MHz</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MemoryType" TYPE="uint16">
					 <VALUE>26</VALUE>
					 <DisplayValue>DDR-4</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>DIMM B6</VALUE>
					 <DisplayValue>DIMM B6</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>DIMM.Socket.B6</VALUE>
					 <DisplayValue>DIMM.Socket.B6</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>DIMM.Socket.B6</VALUE>
					 <DisplayValue>DIMM.Socket.B6</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="NonRAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170530155243.000000+000</VALUE>
					 <DisplayValue>2017-05-30T15:52:43</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>SAS3008 PCI-Express Fusion-MPT SAS-3</VALUE>
					 <DisplayValue>SAS3008 PCI-Express Fusion-MPT SAS-3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>LSI Logic / Symbios Logic</VALUE>
					 <DisplayValue>LSI Logic / Symbios Logic</DisplayValue>
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
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Integrated Storage Controller 1</VALUE>
					 <DisplayValue>Integrated Storage Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="NIC.Integrated.1-1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170502182207.000000+000</VALUE>
					 <DisplayValue>2017-05-02T18:22:07</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>MT27710 Family [ConnectX-4 Lx]</VALUE>
					 <DisplayValue>MT27710 Family [ConnectX-4 Lx]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Mellanox Technologies</VALUE>
					 <DisplayValue>Mellanox Technologies</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0025</VALUE>
					 <DisplayValue>0025</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>15B3</VALUE>
					 <DisplayValue>15B3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>1015</VALUE>
					 <DisplayValue>1015</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>15B3</VALUE>
					 <DisplayValue>15B3</DisplayValue>
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
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
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
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>MT27710 Family [ConnectX-4 Lx]</VALUE>
					 <DisplayValue>MT27710 Family [ConnectX-4 Lx]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Mellanox Technologies</VALUE>
					 <DisplayValue>Mellanox Technologies</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0025</VALUE>
					 <DisplayValue>0025</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>15B3</VALUE>
					 <DisplayValue>15B3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>1015</VALUE>
					 <DisplayValue>1015</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>15B3</VALUE>
					 <DisplayValue>15B3</DisplayValue>
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
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
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
				<Component Classname="DCIM_PCIDeviceView" Key="Video.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170530155243.000000+000</VALUE>
					 <DisplayValue>2017-05-30T15:52:43</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>G200eR2</VALUE>
					 <DisplayValue>G200eR2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Matrox Electronics Systems Ltd.</VALUE>
					 <DisplayValue>Matrox Electronics Systems Ltd.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>0534</VALUE>
					 <DisplayValue>0534</DisplayValue>
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
					 <VALUE>9</VALUE>
					 <DisplayValue>9</DisplayValue>
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
				<Component Classname="DCIM_PCIDeviceView" Key="AHCI.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>C610/X99 series chipset sSATA Controller [AHCI mode]</VALUE>
					 <DisplayValue>C610/X99 series chipset sSATA Controller [AHCI mode]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D62</VALUE>
					 <DisplayValue>8D62</DisplayValue>
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
					 <VALUE>23</VALUE>
					 <DisplayValue>23</DisplayValue>
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
				<Component Classname="DCIM_PCIDeviceView" Key="HostBridge.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Terminator 2x/i</VALUE>
					 <DisplayValue>Terminator 2x/i</DisplayValue>
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
					 <VALUE>6F00</VALUE>
					 <DisplayValue>6F00</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170530155243.000000+000</VALUE>
					 <DisplayValue>2017-05-30T15:52:43</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>SH7758 PCIe Switch [PS]</VALUE>
					 <DisplayValue>SH7758 PCIe Switch [PS]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Renesas Technology Corp.</VALUE>
					 <DisplayValue>Renesas Technology Corp.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>001D</VALUE>
					 <DisplayValue>001D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1912</VALUE>
					 <DisplayValue>1912</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>001D</VALUE>
					 <DisplayValue>001D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>1912</VALUE>
					 <DisplayValue>1912</DisplayValue>
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
					 <VALUE>6</VALUE>
					 <DisplayValue>6</DisplayValue>
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
				<Component Classname="DCIM_PCIDeviceView" Key="USBEHCI.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>C610/X99 series chipset USB Enhanced Host Controller #1</VALUE>
					 <DisplayValue>C610/X99 series chipset USB Enhanced Host Controller #1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D26</VALUE>
					 <DisplayValue>8D26</DisplayValue>
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
					 <VALUE>29</VALUE>
					 <DisplayValue>29</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded USB EHCI 1</VALUE>
					 <DisplayValue>Embedded USB EHCI 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>USBEHCI.Embedded.1-1</VALUE>
					 <DisplayValue>USBEHCI.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>USBEHCI.Embedded.1-1</VALUE>
					 <DisplayValue>USBEHCI.Embedded.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="ISABridge.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>C610/X99 series chipset LPC Controller</VALUE>
					 <DisplayValue>C610/X99 series chipset LPC Controller</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D44</VALUE>
					 <DisplayValue>8D44</DisplayValue>
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
				<Component Classname="DCIM_PCIDeviceView" Key="AHCI.Embedded.2-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>C610/X99 series chipset 6-Port SATA Controller [AHCI mode]</VALUE>
					 <DisplayValue>C610/X99 series chipset 6-Port SATA Controller [AHCI mode]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D02</VALUE>
					 <DisplayValue>8D02</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>49</VALUE>
					 <DisplayValue>49</DisplayValue>
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
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Terminator 2x/i</VALUE>
					 <DisplayValue>Terminator 2x/i</DisplayValue>
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
					 <VALUE>6F02</VALUE>
					 <DisplayValue>6F02</DisplayValue>
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
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
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
				<Component Classname="DCIM_PCIDeviceView" Key="USBEHCI.Embedded.2-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>C610/X99 series chipset USB Enhanced Host Controller #2</VALUE>
					 <DisplayValue>C610/X99 series chipset USB Enhanced Host Controller #2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D2D</VALUE>
					 <DisplayValue>8D2D</DisplayValue>
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
					 <VALUE>26</VALUE>
					 <DisplayValue>26</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded USB EHCI 2</VALUE>
					 <DisplayValue>Embedded USB EHCI 2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>USBEHCI.Embedded.2-1</VALUE>
					 <DisplayValue>USBEHCI.Embedded.2-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>USBEHCI.Embedded.2-1</VALUE>
					 <DisplayValue>USBEHCI.Embedded.2-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.2-2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170530155243.000000+000</VALUE>
					 <DisplayValue>2017-05-30T15:52:43</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>SH7758 PCIe Switch [PS]</VALUE>
					 <DisplayValue>SH7758 PCIe Switch [PS]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Renesas Technology Corp.</VALUE>
					 <DisplayValue>Renesas Technology Corp.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>001D</VALUE>
					 <DisplayValue>001D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1912</VALUE>
					 <DisplayValue>1912</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>001D</VALUE>
					 <DisplayValue>001D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>1912</VALUE>
					 <DisplayValue>1912</DisplayValue>
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
					 <VALUE>7</VALUE>
					 <DisplayValue>7</DisplayValue>
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
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.4-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Terminator 2x/i</VALUE>
					 <DisplayValue>Terminator 2x/i</DisplayValue>
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
					 <VALUE>6F04</VALUE>
					 <DisplayValue>6F04</DisplayValue>
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
					 <VALUE>2</VALUE>
					 <DisplayValue>2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
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
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.4-2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170530155243.000000+000</VALUE>
					 <DisplayValue>2017-05-30T15:52:43</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>SH7758 PCIe-PCI Bridge [PPB]</VALUE>
					 <DisplayValue>SH7758 PCIe-PCI Bridge [PPB]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Renesas Technology Corp.</VALUE>
					 <DisplayValue>Renesas Technology Corp.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>001A</VALUE>
					 <DisplayValue>001A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1912</VALUE>
					 <DisplayValue>1912</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>001A</VALUE>
					 <DisplayValue>001A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>1912</VALUE>
					 <DisplayValue>1912</DisplayValue>
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
					 <VALUE>8</VALUE>
					 <DisplayValue>8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded P2P Bridge 4-2</VALUE>
					 <DisplayValue>Embedded P2P Bridge 4-2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.4-2</VALUE>
					 <DisplayValue>P2PBridge.Embedded.4-2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.4-2</VALUE>
					 <DisplayValue>P2PBridge.Embedded.4-2</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.8-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>Terminator 2x/i</VALUE>
					 <DisplayValue>Terminator 2x/i</DisplayValue>
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
					 <VALUE>6F08</VALUE>
					 <DisplayValue>6F08</DisplayValue>
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
					 <VALUE>3</VALUE>
					 <DisplayValue>3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded P2P Bridge 8-1</VALUE>
					 <DisplayValue>Embedded P2P Bridge 8-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.8-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.8-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.8-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.8-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.5-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>C610/X99 series chipset PCI Express Root Port #1</VALUE>
					 <DisplayValue>C610/X99 series chipset PCI Express Root Port #1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D10</VALUE>
					 <DisplayValue>8D10</DisplayValue>
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
					 <VALUE>Embedded P2P Bridge 5-1</VALUE>
					 <DisplayValue>Embedded P2P Bridge 5-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.5-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.5-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.5-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.5-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.12-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>C610/X99 series chipset PCI Express Root Port #8</VALUE>
					 <DisplayValue>C610/X99 series chipset PCI Express Root Port #8</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Intel Corporation</VALUE>
					 <DisplayValue>Intel Corporation</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>8D1E</VALUE>
					 <DisplayValue>8D1E</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>8086</VALUE>
					 <DisplayValue>8086</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>7</VALUE>
					 <DisplayValue>7</DisplayValue>
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
					 <VALUE>Embedded P2P Bridge 12-1</VALUE>
					 <DisplayValue>Embedded P2P Bridge 12-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.12-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.12-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.12-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.12-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PCIDeviceView" Key="P2PBridge.Embedded.3-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170530155243.000000+000</VALUE>
					 <DisplayValue>2017-05-30T15:52:43</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>SH7758 PCIe Switch [PS]</VALUE>
					 <DisplayValue>SH7758 PCIe Switch [PS]</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Renesas Technology Corp.</VALUE>
					 <DisplayValue>Renesas Technology Corp.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>001D</VALUE>
					 <DisplayValue>001D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1912</VALUE>
					 <DisplayValue>1912</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>001D</VALUE>
					 <DisplayValue>001D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>1912</VALUE>
					 <DisplayValue>1912</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FunctionNumber" TYPE="uint32">
					 <VALUE>0</VALUE>
					 <DisplayValue>0</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceNumber" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BusNumber" TYPE="uint32">
					 <VALUE>7</VALUE>
					 <DisplayValue>7</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>Embedded P2P Bridge 3-1</VALUE>
					 <DisplayValue>Embedded P2P Bridge 3-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>P2PBridge.Embedded.3-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.3-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>P2PBridge.Embedded.3-1</VALUE>
					 <DisplayValue>P2PBridge.Embedded.3-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_SystemView" Key="System.Embedded.1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170530155243.000000+000</VALUE>
					 <DisplayValue>2017-05-30T15:52:43</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EstimatedExhaustTemperature" TYPE="uint16">
					 <VALUE>41</VALUE>
					 <DisplayValue>41 Degree C</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="EstimatedSystemAirflow" TYPE="uint16">
					 <VALUE>15</VALUE>
					 <DisplayValue>15 CFM</DisplayValue>
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
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="IntrusionRollupStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
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
				   <PROPERTY NAME="PrimaryStatus" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>OK</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BladeGeometry" TYPE="uint16">
					 <VALUE>255</VALUE>
					 <DisplayValue>Unknown</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CPLDVersion" TYPE="string">
					 <VALUE>1.0.1</VALUE>
					 <DisplayValue>1.0.1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BoardPartNumber" TYPE="string">
					 <VALUE>02C2CPA04</VALUE>
					 <DisplayValue>02C2CPA04</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BoardSerialNumber" TYPE="string">
					 <VALUE>CN747517270290</VALUE>
					 <DisplayValue>CN747517270290</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ChassisName" TYPE="string">
					 <VALUE>Main System Chassis</VALUE>
					 <DisplayValue>Main System Chassis</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ServerAllocation" TYPE="uint32">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="PowerCap" TYPE="uint32">
					 <VALUE>177</VALUE>
					 <DisplayValue>177 Watts</DisplayValue>
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
					 <VALUE>12</VALUE>
					 <DisplayValue>12</DisplayValue>
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
					 <VALUE>machine.example.com</VALUE>
					 <DisplayValue>machine.example.com</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="UUID" TYPE="string">
					 <VALUE>4c4c4544-0035-4b10-8054-b6c04f374a32</VALUE>
					 <DisplayValue>4c4c4544-0035-4b10-8054-b6c04f374a32</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="smbiosGUID" TYPE="string">
					 <VALUE>44454c4c-3500-104b-8054-b6c04f374a32</VALUE>
					 <DisplayValue>44454c4c-3500-104b-8054-b6c04f374a32</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PlatformGUID" TYPE="string">
					 <VALUE>324a374f-c0b6-5480-4b10-00354c4c4544</VALUE>
					 <DisplayValue>324a374f-c0b6-5480-4b10-00354c4c4544</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SystemID" TYPE="uint32">
					 <VALUE>1537</VALUE>
					 <DisplayValue>1537</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BIOSReleaseDate" TYPE="string">
					 <VALUE>01/17/2017</VALUE>
					 <DisplayValue>01/17/2017</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="BIOSVersionString" TYPE="string">
					 <VALUE>2.4.3</VALUE>
					 <DisplayValue>2.4.3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SystemRevision" TYPE="uint16">
					 <VALUE/>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="MaxDIMMSlots" TYPE="uint32">
					 <VALUE>24</VALUE>
					 <DisplayValue>24</DisplayValue>
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
					 <VALUE>3145728</VALUE>
					 <DisplayValue>3145728 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SysMemTotalSize" TYPE="uint32">
					 <VALUE>393216</VALUE>
					 <DisplayValue>393216 MB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ChassisSystemHeight" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>1U</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NodeID" TYPE="string">
					 <VALUE>65KT7J2</VALUE>
					 <DisplayValue>65KT7J2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ChassisModel" TYPE="string">
					 <VALUE/>
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="ChassisServiceTag" TYPE="string">
					 <VALUE>65KT7J2</VALUE>
					 <DisplayValue>65KT7J2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ExpressServiceCode" TYPE="string">
					 <VALUE>13397979998</VALUE>
					 <DisplayValue>13397979998</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ServiceTag" TYPE="string">
					 <VALUE>65KT7J2</VALUE>
					 <DisplayValue>65KT7J2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Dell Inc.</VALUE>
					 <DisplayValue>Dell Inc.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>PowerEdge R630</VALUE>
					 <DisplayValue>PowerEdge R630</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LifecycleControllerVersion" TYPE="string">
					 <VALUE>2.41.40.40</VALUE>
					 <DisplayValue>2.41.40.40</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CMCIP" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="SystemGeneration" TYPE="string">
					 <VALUE>13G Monolithic</VALUE>
					 <DisplayValue>13G Monolithic</DisplayValue>
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
					 <VALUE>20170502192207.000000+000</VALUE>
					 <DisplayValue>2017-05-02T19:22:07</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20170614154350.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:43:50</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Protocol" TYPE="string">
					 <VALUE>NIC,RDMA</VALUE>
					 <DisplayValue>NIC,RDMA</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MediaType" TYPE="string">
					 <VALUE>SR,SFP,SFP_PLUS,DCA</VALUE>
					 <DisplayValue>SR,SFP,SFP_PLUS,DCA</DisplayValue>
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
					 <VALUE>9</VALUE>
					 <DisplayValue>10 Gbps</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LinkDuplex" TYPE="uint8">
					 <VALUE>1</VALUE>
					 <DisplayValue>Full Duplex</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="VendorName" TYPE="string">
					 <VALUE>Mellanox Technologies, Inc.</VALUE>
					 <DisplayValue>Mellanox Technologies, Inc.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FCoEWWNN" TYPE="string">
					 <DisplayValue/>
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
					 <VALUE>0025</VALUE>
					 <DisplayValue>0025</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>1015</VALUE>
					 <DisplayValue>1015</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>15b3</VALUE>
					 <DisplayValue>15b3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>15b3</VALUE>
					 <DisplayValue>15b3</DisplayValue>
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
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
					 <VALUE>24:8A:07:5A:9E:8C</VALUE>
					 <DisplayValue>24:8A:07:5A:9E:8C</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentMACAddress" TYPE="string">
					 <VALUE>24:8A:07:5A:9E:8C</VALUE>
					 <DisplayValue>24:8A:07:5A:9E:8C</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>Mellanox ConnectX-4 LX 25GbE SFP Rack NDC - 24:8A:07:5A:9E:8C</VALUE>
					 <DisplayValue>Mellanox ConnectX-4 LX 25GbE SFP Rack NDC - 24:8A:07:5A:9E:8C</DisplayValue>
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
					 <VALUE>14.9.50</VALUE>
					 <DisplayValue>14.9.50</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ControllerBIOSVersion" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="FamilyVersion" TYPE="string">
					 <VALUE>14.14.23.20</VALUE>
					 <DisplayValue>14.14.23.20</DisplayValue>
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
					 <VALUE>20170311001148.000000+000</VALUE>
					 <DisplayValue>2017-03-11T00:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20170614154350.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:43:50</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Protocol" TYPE="string">
					 <VALUE>NIC,RDMA</VALUE>
					 <DisplayValue>NIC,RDMA</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="MediaType" TYPE="string">
					 <VALUE>SR,SFP,SFP_PLUS,DCA</VALUE>
					 <DisplayValue>SR,SFP,SFP_PLUS,DCA</DisplayValue>
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
					 <VALUE>Mellanox Technologies, Inc.</VALUE>
					 <DisplayValue>Mellanox Technologies, Inc.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FCoEWWNN" TYPE="string">
					 <DisplayValue/>
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
					 <VALUE>0025</VALUE>
					 <DisplayValue>0025</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>1015</VALUE>
					 <DisplayValue>1015</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>15b3</VALUE>
					 <DisplayValue>15b3</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIVendorID" TYPE="string">
					 <VALUE>15b3</VALUE>
					 <DisplayValue>15b3</DisplayValue>
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
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
					 <VALUE>24:8A:07:5A:9E:8D</VALUE>
					 <DisplayValue>24:8A:07:5A:9E:8D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentMACAddress" TYPE="string">
					 <VALUE>24:8A:07:5A:9E:8D</VALUE>
					 <DisplayValue>24:8A:07:5A:9E:8D</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>Mellanox ConnectX-4 LX 25GbE SFP Rack NDC - 24:8A:07:5A:9E:8D</VALUE>
					 <DisplayValue>Mellanox ConnectX-4 LX 25GbE SFP Rack NDC - 24:8A:07:5A:9E:8D</DisplayValue>
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
					 <VALUE>14.9.50</VALUE>
					 <DisplayValue>14.9.50</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ControllerBIOSVersion" TYPE="string">
					 <DisplayValue/>
				   </PROPERTY>
				   <PROPERTY NAME="FamilyVersion" TYPE="string">
					 <VALUE>14.14.23.20</VALUE>
					 <DisplayValue>14.14.23.20</DisplayValue>
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
				<Component Classname="DCIM_VideoView" Key="Video.Embedded.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170530155243.000000+000</VALUE>
					 <DisplayValue>2017-05-30T15:52:43</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>G200eR2</VALUE>
					 <DisplayValue>G200eR2</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Matrox Electronics Systems Ltd.</VALUE>
					 <DisplayValue>Matrox Electronics Systems Ltd.</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubDeviceID" TYPE="string">
					 <VALUE>0601</VALUE>
					 <DisplayValue>0601</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCISubVendorID" TYPE="string">
					 <VALUE>1028</VALUE>
					 <DisplayValue>1028</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PCIDeviceID" TYPE="string">
					 <VALUE>0534</VALUE>
					 <DisplayValue>0534</DisplayValue>
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
					 <VALUE>9</VALUE>
					 <DisplayValue>9</DisplayValue>
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
				<Component Classname="DCIM_PowerSupplyView" Key="PSU.Slot.1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200514.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:14</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PMBusMonitoring" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Range1MaxInputPower" TYPE="uint32">
					 <VALUE>900</VALUE>
					 <DisplayValue>900</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedMinNumberNeeded" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY.ARRAY NAME="RedTypeOfSet" TYPE="uint16">
				  <VALUE.ARRAY>
				   <VALUE>2</VALUE>
				   <VALUE>32768</VALUE>
				   <DisplayValue>N+1</DisplayValue>
				   <DisplayValue>Unknown</DisplayValue>
				   </VALUE.ARRAY>
				   </PROPERTY.ARRAY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>750</VALUE>
					 <DisplayValue>750 Watts</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InputVoltage" TYPE="uint32">
					 <VALUE>236</VALUE>
					 <DisplayValue>236 Volts</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FirmwareVersion" TYPE="string">
					 <VALUE>00.16.4F</VALUE>
					 <DisplayValue>00.16.4F</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DetailedState" TYPE="string">
					 <VALUE>Presence Detected</VALUE>
					 <DisplayValue>Presence Detected</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Dell</VALUE>
					 <DisplayValue>Dell</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>0Y9VFCA01</VALUE>
					 <DisplayValue>0Y9VFCA01</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>PHARP00719027Q</VALUE>
					 <DisplayValue>PHARP00719027Q</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>PWR SPLY,750W,RDNT,ARTESYN    </VALUE>
					 <DisplayValue>PWR SPLY,750W,RDNT,ARTESYN    </DisplayValue>
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
				</Component>
				<Component Classname="DCIM_PowerSupplyView" Key="PSU.Slot.2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200514.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:14</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PMBusMonitoring" TYPE="uint16">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Range1MaxInputPower" TYPE="uint32">
					 <VALUE>900</VALUE>
					 <DisplayValue>900</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedMinNumberNeeded" TYPE="uint32">
					 <VALUE>1</VALUE>
					 <DisplayValue>1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY.ARRAY NAME="RedTypeOfSet" TYPE="uint16">
				  <VALUE.ARRAY>
				   <VALUE>2</VALUE>
				   <VALUE>32768</VALUE>
				   <DisplayValue>N+1</DisplayValue>
				   <DisplayValue>Unknown</DisplayValue>
				   </VALUE.ARRAY>
				   </PROPERTY.ARRAY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>750</VALUE>
					 <DisplayValue>750 Watts</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InputVoltage" TYPE="uint32">
					 <VALUE>234</VALUE>
					 <DisplayValue>234 Volts</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FirmwareVersion" TYPE="string">
					 <VALUE>00.16.4F</VALUE>
					 <DisplayValue>00.16.4F</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DetailedState" TYPE="string">
					 <VALUE>Presence Detected</VALUE>
					 <DisplayValue>Presence Detected</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Manufacturer" TYPE="string">
					 <VALUE>Dell</VALUE>
					 <DisplayValue>Dell</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PartNumber" TYPE="string">
					 <VALUE>0Y9VFCA01</VALUE>
					 <DisplayValue>0Y9VFCA01</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SerialNumber" TYPE="string">
					 <VALUE>PHARP00719027P</VALUE>
					 <DisplayValue>PHARP00719027P</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>PWR SPLY,750W,RDNT,ARTESYN    </VALUE>
					 <DisplayValue>PWR SPLY,750W,RDNT,ARTESYN    </DisplayValue>
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
				</Component>
				<Component Classname="DCIM_PhysicalDiskView" Key="Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170614154516.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:45:16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20170614154516.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:45:16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
					 <VALUE>100</VALUE>
					 <DisplayValue>100%</DisplayValue>
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
					 <VALUE>TW0R3J3YITT0072K03OLA00</VALUE>
					 <DisplayValue>TW0R3J3YITT0072K03OLA00</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					 <VALUE>500056B34231ABC0</VALUE>
					 <DisplayValue>500056B34231ABC0</DisplayValue>
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
					 <VALUE>PHDV707000D51P6EGN</VALUE>
					 <DisplayValue>PHDV707000D51P6EGN</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Revision" TYPE="string">
					 <VALUE>N201DL42</VALUE>
					 <DisplayValue>N201DL42</DisplayValue>
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
					 <VALUE>SSDSC2BB016T7R</VALUE>
					 <DisplayValue>SSDSC2BB016T7R</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SizeInBytes" TYPE="uint64">
					 <VALUE>1600321314304</VALUE>
					 <DisplayValue>1600321314304 Bytes</DisplayValue>
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
					 <VALUE>Disk 0 in Backplane 1 of Integrated Storage Controller 1</VALUE>
					 <DisplayValue>Disk 0 in Backplane 1 of Integrated Storage Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>Disk.Bay.0:Enclosure.Internal.0-1:NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_PhysicalDiskView" Key="Disk.Bay.1:Enclosure.Internal.0-1:NonRAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170614154516.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:45:16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20170614154516.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:45:16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RemainingRatedWriteEndurance" TYPE="uint8">
					 <VALUE>100</VALUE>
					 <DisplayValue>100%</DisplayValue>
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
					 <VALUE>TW0R3J3YITT0072K03R6A00</VALUE>
					 <DisplayValue>TW0R3J3YITT0072K03R6A00</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SASAddress" TYPE="string">
					 <VALUE>500056B34231ABC1</VALUE>
					 <DisplayValue>500056B34231ABC1</DisplayValue>
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
					 <VALUE>PHDV707000FX1P6EGN</VALUE>
					 <DisplayValue>PHDV707000FX1P6EGN</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Revision" TYPE="string">
					 <VALUE>N201DL42</VALUE>
					 <DisplayValue>N201DL42</DisplayValue>
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
					 <VALUE>SSDSC2BB016T7R</VALUE>
					 <DisplayValue>SSDSC2BB016T7R</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="SizeInBytes" TYPE="uint64">
					 <VALUE>1600321314304</VALUE>
					 <DisplayValue>1600321314304 Bytes</DisplayValue>
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
					 <VALUE>Disk 1 in Backplane 1 of Integrated Storage Controller 1</VALUE>
					 <DisplayValue>Disk 1 in Backplane 1 of Integrated Storage Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Disk.Bay.1:Enclosure.Internal.0-1:NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>Disk.Bay.1:Enclosure.Internal.0-1:NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Disk.Bay.1:Enclosure.Internal.0-1:NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>Disk.Bay.1:Enclosure.Internal.0-1:NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_iDRACCardView" Key="iDRAC.Embedded.1-1#IDRACinfo">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200516.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DNSDomainName" TYPE="string">
					 <VALUE>example.com</VALUE>
					 <DisplayValue>example.com</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DNSRacName" TYPE="string">
					 <VALUE>spare-65KT7J2</VALUE>
					 <DisplayValue>spare-65KT7J2</DisplayValue>
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
					 <VALUE>https://10.221.172.67:443</VALUE>
					 <DisplayValue>https://10.221.172.67:443</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="GUID" TYPE="string">
					 <VALUE>324a374f-c0b6-5480-4b10-00354c4c4544</VALUE>
					 <DisplayValue>324a374f-c0b6-5480-4b10-00354c4c4544</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="PermanentMACAddress" TYPE="string">
					 <VALUE>84:7b:eb:f7:82:5a</VALUE>
					 <DisplayValue>84:7b:eb:f7:82:5a</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductDescription" TYPE="string">
					 <VALUE>This system component provides a complete set of remote management functions for Dell PowerEdge servers</VALUE>
					 <DisplayValue>This system component provides a complete set of remote management functions for Dell PowerEdge servers</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>Enterprise</VALUE>
					 <DisplayValue>Enterprise</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FirmwareVersion" TYPE="string">
					 <VALUE>2.41.40.40</VALUE>
					 <DisplayValue>2.41.40.40</DisplayValue>
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
				<Component Classname="DCIM_VFlashView" Key="Disk.vFlashCard.1">
				   <PROPERTY NAME="ComponentName" TYPE="string">
					 <VALUE>No SD Card</VALUE>
					 <DisplayValue>No SD Card</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="DeviceDescription" TYPE="string">
					 <VALUE>vFlash Card</VALUE>
					 <DisplayValue>vFlash Card</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Disk.vFlashCard.1</VALUE>
					 <DisplayValue>Disk.vFlashCard.1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Disk.vFlashCard.1</VALUE>
					 <DisplayValue>Disk.vFlashCard.1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.1A">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200517.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:17</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>10</VALUE>
					 <DisplayValue>10%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>3480</VALUE>
					 <DisplayValue>3480 RPM</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.2A">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200517.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:17</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>15</VALUE>
					 <DisplayValue>15%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>4440</VALUE>
					 <DisplayValue>4440 RPM</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.3A">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200517.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:17</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>15</VALUE>
					 <DisplayValue>15%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>4440</VALUE>
					 <DisplayValue>4440 RPM</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.4A">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200517.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:17</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>10</VALUE>
					 <DisplayValue>10%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>3720</VALUE>
					 <DisplayValue>3720 RPM</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.5A">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200518.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:18</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>15</VALUE>
					 <DisplayValue>15%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>4680</VALUE>
					 <DisplayValue>4680 RPM</DisplayValue>
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
					 <VALUE>Fan 5A</VALUE>
					 <DisplayValue>Fan 5A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Fan.Embedded.5A</VALUE>
					 <DisplayValue>Fan.Embedded.5A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Fan.Embedded.5A</VALUE>
					 <DisplayValue>Fan.Embedded.5A</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.6A">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200518.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:18</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>15</VALUE>
					 <DisplayValue>15%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>4440</VALUE>
					 <DisplayValue>4440 RPM</DisplayValue>
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
					 <VALUE>Fan 6A</VALUE>
					 <DisplayValue>Fan 6A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Fan.Embedded.6A</VALUE>
					 <DisplayValue>Fan.Embedded.6A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Fan.Embedded.6A</VALUE>
					 <DisplayValue>Fan.Embedded.6A</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.7A">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200518.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:18</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>10</VALUE>
					 <DisplayValue>10%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>3480</VALUE>
					 <DisplayValue>3480 RPM</DisplayValue>
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
					 <VALUE>Fan 7A</VALUE>
					 <DisplayValue>Fan 7A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Fan.Embedded.7A</VALUE>
					 <DisplayValue>Fan.Embedded.7A</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Fan.Embedded.7A</VALUE>
					 <DisplayValue>Fan.Embedded.7A</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.1B">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200518.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:18</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>10</VALUE>
					 <DisplayValue>10%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>3240</VALUE>
					 <DisplayValue>3240 RPM</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.2B">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200518.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:18</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>15</VALUE>
					 <DisplayValue>15%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>4200</VALUE>
					 <DisplayValue>4200 RPM</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.3B">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200518.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:18</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>15</VALUE>
					 <DisplayValue>15%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>4440</VALUE>
					 <DisplayValue>4440 RPM</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.4B">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200518.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:18</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>10</VALUE>
					 <DisplayValue>10%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>3240</VALUE>
					 <DisplayValue>3240 RPM</DisplayValue>
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
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.5B">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200519.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:19</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>15</VALUE>
					 <DisplayValue>15%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>4440</VALUE>
					 <DisplayValue>4440 RPM</DisplayValue>
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
					 <VALUE>Fan 5B</VALUE>
					 <DisplayValue>Fan 5B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Fan.Embedded.5B</VALUE>
					 <DisplayValue>Fan.Embedded.5B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Fan.Embedded.5B</VALUE>
					 <DisplayValue>Fan.Embedded.5B</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.6B">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200519.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:19</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>15</VALUE>
					 <DisplayValue>15%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>4440</VALUE>
					 <DisplayValue>4440 RPM</DisplayValue>
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
					 <VALUE>Fan 6B</VALUE>
					 <DisplayValue>Fan 6B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Fan.Embedded.6B</VALUE>
					 <DisplayValue>Fan.Embedded.6B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Fan.Embedded.6B</VALUE>
					 <DisplayValue>Fan.Embedded.6B</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_FanView" Key="Fan.Embedded.7B">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20171107200519.000000+000</VALUE>
					 <DisplayValue>2017-11-07T20:05:19</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="RedundancyStatus" TYPE="uint16">
					 <VALUE>2</VALUE>
					 <DisplayValue>Fully Redundant</DisplayValue>
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
					 <VALUE>10</VALUE>
					 <DisplayValue>10%</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="CurrentReading" TYPE="sint32">
					 <VALUE>3240</VALUE>
					 <DisplayValue>3240 RPM</DisplayValue>
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
					 <VALUE>Fan 7B</VALUE>
					 <DisplayValue>Fan 7B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Fan.Embedded.7B</VALUE>
					 <DisplayValue>Fan.Embedded.7B</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Fan.Embedded.7B</VALUE>
					 <DisplayValue>Fan.Embedded.7B</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_EnclosureView" Key="Enclosure.Internal.0-1:NonRAID.Integrated.1-1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170614154516.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:45:16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20170614154516.000000+000</VALUE>
					 <DisplayValue>2017-06-14T15:45:16</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="ProductName" TYPE="string">
					 <VALUE>BP13G+EXP 0:1</VALUE>
					 <DisplayValue>BP13G+EXP 0:1</DisplayValue>
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
					 <VALUE>10</VALUE>
					 <DisplayValue>10</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Version" TYPE="string">
					 <VALUE>3.32</VALUE>
					 <DisplayValue>3.32</DisplayValue>
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
					 <VALUE>Backplane 1 on Connector 0 of Integrated Storage Controller 1</VALUE>
					 <DisplayValue>Backplane 1 on Connector 0 of Integrated Storage Controller 1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="FQDD" TYPE="string">
					 <VALUE>Enclosure.Internal.0-1:NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>Enclosure.Internal.0-1:NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="InstanceID" TYPE="string">
					 <VALUE>Enclosure.Internal.0-1:NonRAID.Integrated.1-1</VALUE>
					 <DisplayValue>Enclosure.Internal.0-1:NonRAID.Integrated.1-1</DisplayValue>
				   </PROPERTY>
				</Component>
				<Component Classname="DCIM_CPUView" Key="CPU.Socket.1">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>14</VALUE>
					 <DisplayValue>20-way Set-Associative</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Associativity" TYPE="uint16">
					 <VALUE>7</VALUE>
					 <DisplayValue>8-way Set-Associative</DisplayValue>
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
					 <VALUE>35840</VALUE>
					 <DisplayValue>35840 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Size" TYPE="uint32">
					 <VALUE>3584</VALUE>
					 <DisplayValue>3584 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Size" TYPE="uint32">
					 <VALUE>896</VALUE>
					 <DisplayValue>896 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>Intel(R) Xeon(R) CPU E5-2690 v4 @ 2.60GHz</VALUE>
					 <DisplayValue>Intel(R) Xeon(R) CPU E5-2690 v4 @ 2.60GHz</DisplayValue>
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
					 <VALUE>1.3</VALUE>
					 <DisplayValue>1.3V</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfProcessorCores" TYPE="uint32">
					 <VALUE>14</VALUE>
					 <DisplayValue>14</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfEnabledThreads" TYPE="uint32">
					 <VALUE>28</VALUE>
					 <DisplayValue>28</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfEnabledCores" TYPE="uint32">
					 <VALUE>14</VALUE>
					 <DisplayValue>14</DisplayValue>
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
					 <VALUE>2600</VALUE>
					 <DisplayValue>2600 MHz</DisplayValue>
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
				<Component Classname="DCIM_CPUView" Key="CPU.Socket.2">
				   <PROPERTY NAME="LastUpdateTime" TYPE="string">
					 <VALUE>20170310231148.000000+000</VALUE>
					 <DisplayValue>2017-03-10T23:11:48</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="LastSystemInventoryTime" TYPE="string">
					 <VALUE>20171019194115.000000+000</VALUE>
					 <DisplayValue>2017-10-19T19:41:15</DisplayValue>
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
					 <VALUE>14</VALUE>
					 <DisplayValue>20-way Set-Associative</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Associativity" TYPE="uint16">
					 <VALUE>7</VALUE>
					 <DisplayValue>8-way Set-Associative</DisplayValue>
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
					 <VALUE>35840</VALUE>
					 <DisplayValue>35840 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache2Size" TYPE="uint32">
					 <VALUE>3584</VALUE>
					 <DisplayValue>3584 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Cache1Size" TYPE="uint32">
					 <VALUE>896</VALUE>
					 <DisplayValue>896 KB</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="Model" TYPE="string">
					 <VALUE>Intel(R) Xeon(R) CPU E5-2690 v4 @ 2.60GHz</VALUE>
					 <DisplayValue>Intel(R) Xeon(R) CPU E5-2690 v4 @ 2.60GHz</DisplayValue>
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
					 <VALUE>1.3</VALUE>
					 <DisplayValue>1.3V</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfProcessorCores" TYPE="uint32">
					 <VALUE>14</VALUE>
					 <DisplayValue>14</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfEnabledThreads" TYPE="uint32">
					 <VALUE>28</VALUE>
					 <DisplayValue>28</DisplayValue>
				   </PROPERTY>
				   <PROPERTY NAME="NumberOfEnabledCores" TYPE="uint32">
					 <VALUE>14</VALUE>
					 <DisplayValue>14</DisplayValue>
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
					 <VALUE>2600</VALUE>
					 <DisplayValue>2600 MHz</DisplayValue>
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
			</Inventory>			
		`),
		"/data": []byte(`<root><powerOn></powerOn><powermonitordata><historicalPeak><startTime>Fri Mar 10 12:07:06 2017
			</startTime>
			<peakWattTime>Wed Jun 14 05:43:02 2017
			</peakWattTime>
			<peakWattValue>2175</peakWattValue>
			<peakAmpsTime>Wed Jun 14 05:43:03 2017
			</peakAmpsTime>
			<peakAmpsValue>1.8</peakAmpsValue>
			</historicalPeak><sysHeadRoom><instantaneous>724</instantaneous>
			<peak>0</peak>
			</sysHeadRoom><cumReading><time>Fri Mar 10 12:07:06 2017
			</time>
			<totalUsage>741.054</totalUsage>
			</cumReading><powerCapacity>900</powerCapacity>
			<historicalTrend><trendData><time>Last Hour</time>
			<avgUsage>175</avgUsage>
			<maxPeak>212</maxPeak>
			<maxPeakTime>Tue Nov  7 09:34:40 2017
			</maxPeakTime>
			<minPeak>163</minPeak>
			<minPeakTime>Tue Nov  7 08:51:50 2017
			</minPeakTime>
			</trendData><trendData><time>Last Day</time>
			<avgUsage>176</avgUsage>
			<maxPeak>238</maxPeak>
			<maxPeakTime>Mon Nov  6 22:04:58 2017
			</maxPeakTime>
			<minPeak>157</minPeak>
			<minPeakTime>Mon Nov  6 14:49:28 2017
			</minPeakTime>
			</trendData><trendData><time>Last Week</time>
			<avgUsage>174</avgUsage>
			<maxPeak>238</maxPeak>
			<maxPeakTime>Mon Nov  6 22:04:58 2017
			</maxPeakTime>
			<minPeak>157</minPeak>
			<minPeakTime>Wed Nov  1 10:30:36 2017
			</minPeakTime>
			</trendData></historicalTrend><presentReading><reading><probeStatus>2</probeStatus>
			<probeName>System Board Pwr Consumption</probeName>
			<reading>168</reading>
			<warningThreshold>896</warningThreshold>
			<failureThreshold>980</failureThreshold>
			</reading></presentReading><psuReading><reading><probeName>PS1</probeName>
			<psuAmps>0.4</psuAmps>
			<psuVolts>236</psuVolts>
			</reading><reading><probeName>PS2</probeName>
			<psuAmps>0.4</psuAmps>
			<psuVolts>234</psuVolts>
			</reading></psuReading></powermonitordata><powerOn></powerOn><psSensorList><sensor><sensorHealth>2</sensorHealth>
			<name>PS1 Status</name>
			<sensorStatus>1</sensorStatus>
			<inputWattage>900</inputWattage>
			<maxWattage>750</maxWattage>
			<type>AC</type>
			<partno>0Y9VFCA01</partno>
			<fwVersion>00.16.4F</fwVersion>
			</sensor><sensor><sensorHealth>2</sensorHealth>
			<name>PS2 Status</name>
			<sensorStatus>1</sensorStatus>
			<inputWattage>900</inputWattage>
			<maxWattage>750</maxWattage>
			<type>AC</type>
			<partno>0Y9VFCA01</partno>
			<fwVersion>00.16.4F</fwVersion>
			</sensor></psSensorList><status>ok</status>
			</root>`),
		"/sysmgmt/2012/server/license":         []byte(`{"License":{"AUTO_DISCOVERY":1,"BACKUP_RESTORE":1,"BASIC_REMOTE_INVENTORY_EXPORT":1,"BOOT_CAPTURE":1,"CONSOLE_COLLABORATION":1,"DEDICATED_NIC":1,"DEVICE_MONITORING":1,"DIRECTORY_SERVICES":1,"DYNAMIC_DNS":1,"EMAIL_ALERTING":1,"FULL_UI":1,"INBAND_FIRMWARE_UPDATE":1,"IPV6":1,"LAST_CRASH_SCREEN_CAPTURE":1,"LAST_CRASH_VIDEO_CAPTURE":1,"LICENSE_UI":1,"NTP":1,"PART_REPLACEMENT":1,"POWER_BUDGETING":1,"POWER_MONITORING":1,"RACADM_CLI":1,"REMOTE_ASSET_INVENTORY":1,"REMOTE_CONFIGURATION":1,"REMOTE_FILE_SHARE":1,"REMOTE_FIRWARE_UPDATE":1,"REMOTE_OS_DEPLOYMENT":1,"REMOTE_SYSLOG":1,"SECURITY_LOCKOUT":1,"SMASH_CLP":1,"SNMP":1,"SSH":1,"SSH_PK_AUTHEN":1,"SSO":1,"STORAGE_MONITORING":1,"TELNET":1,"TWO_FACTOR_AUTHEN":1,"USC_ASSISTED_OS_DEPLOYEMENT":1,"USC_DEVICE_CONFIGURATION":1,"USC_EMBEDDED_DIAGNOSTICS":1,"USC_FIRMWARE_UPDATE":1,"VCONSOLE":1,"VFOLDER":1,"VIRTUAL_FLASH_PARTITIONS":1,"VIRTUAL_NW_CONSOLE":1,"VMEDIA":1,"WSMAN":1}}`),
		"/sysmgmt/2012/server/processor":       []byte(`{"Processor":{"D2||CPU.Socket.1":{"brand":"Intel(R) Xeon(R) CPU E5-2690 v4 @ 2.60GHz","cache":"/sysmgmt/2012/server/cache?processor=D2||CPU.Socket.1","core_count":14,"current_speed":2600,"device_description":"CPU 1","executeDisable":[{"capable":1,"enabled":1}],"hyperThreading":[{"capable":1,"enabled":1}],"name":"[CPU1]","state":3,"status":2,"turboMode":[{"capable":1,"enabled":1}],"version":"Model 79 Stepping 1","virtualizationTech":[{"capable":1,"enabled":1}]},"D2||CPU.Socket.2":{"brand":"Intel(R) Xeon(R) CPU E5-2690 v4 @ 2.60GHz","cache":"/sysmgmt/2012/server/cache?processor=D2||CPU.Socket.2","core_count":14,"current_speed":2600,"device_description":"CPU 2","executeDisable":[{"capable":1,"enabled":1}],"hyperThreading":[{"capable":1,"enabled":1}],"name":"[CPU2]","state":3,"status":2,"turboMode":[{"capable":1,"enabled":1}],"version":"Model 79 Stepping 1","virtualizationTech":[{"capable":1,"enabled":1}]}}}`),
		"/sysmgmt/2012/server/temperature":     []byte(`{"Statistics":"/sysmgmt/2012/server/temperature/statistics","Temperatures":{"iDRAC.Embedded.1#CPU1Temp":{"max_failure":103,"max_warning":98,"max_warning_settable":0,"min_failure":3,"min_warning":8,"min_warning_settable":0,"name":"CPU1 Temp","reading":46,"sensor_status":2},"iDRAC.Embedded.1#CPU2Temp":{"max_failure":103,"max_warning":98,"max_warning_settable":0,"min_failure":3,"min_warning":8,"min_warning_settable":0,"name":"CPU2 Temp","reading":41,"sensor_status":2},"iDRAC.Embedded.1#SystemBoardInletTemp":{"max_failure":47,"max_warning":42,"max_warning_settable":1,"min_failure":-7,"min_warning":3,"min_warning_settable":1,"name":"System Board Inlet Temp","reading":19,"sensor_status":2}},"is_fresh_air_compliant":1}`),
		"/data/logout":                         []byte(``),
		"/data/login":                          []byte(`<?xml version="1.0" encoding="UTF-8"?> <root> <status>ok</status> <authResult>0</authResult> <forwardUrl>index.html?ST1=3fd2ec4d84e406f348972d2fb5a1cdd2,ST2=a47f9a0ea441fdd5bf59c63f902c03d2</forwardUrl> </root>`),
		"/sysmgmt/2016/server/extended_health": []byte(`{"healthStatus":[2,2,0,0,0,0,2,0,2,2,2,2,2,2,0,2,2]}`),
	}
)

func setup() (bmc *IDrac8, err error) {
	viper.SetDefault("debug", true)
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	for url := range answers {
		url := url
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			w.Write(answers[url])
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
	expectedAnswer := "65kt7j2"

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
	expectedAnswer := "PowerEdge R630"

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
	expectedAnswer := "idrac8"

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
	expectedAnswer := "2.41.40.40"

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
	expectedAnswer := "machine.example.com"
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
	expectedAnswer := 384

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
	expectedAnswerCPUType := "intel(r) xeon(r) cpu e5-2690 v4"
	expectedAnswerCPUCount := 2
	expectedAnswerCore := 14
	expectedAnswerHyperthread := 28

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
	expectedAnswer := "2.4.3"

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
	expectedAnswer := 0.168

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
	expectedAnswer := 19

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
		&devices.Nic{
			MacAddress: "24:8a:07:5a:9e:8c",
			Name:       "Mellanox ConnectX-4 LX 25GbE SFP Rack NDC",
			Up:         true,
			Speed:      "10 Gbps",
		},
		&devices.Nic{
			MacAddress: "24:8a:07:5a:9e:8d",
			Name:       "Mellanox ConnectX-4 LX 25GbE SFP Rack NDC",
			Up:         false,
			Speed:      "",
		},
		&devices.Nic{
			MacAddress: "84:7b:eb:f7:82:5a",
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

func TestIDracPsu(t *testing.T) {
	expectedAnswer := []*devices.Psu{
		&devices.Psu{
			Serial:     "65kt7j2_PS1",
			CapacityKw: 0.75,
			Status:     "OK",
			PowerKw:    0.0,
		},
		&devices.Psu{
			Serial:     "65kt7j2_PS2",
			CapacityKw: 0.75,
			Status:     "OK",
			PowerKw:    0.0,
		},
	}

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test hpChassisSetup %v", err)
	}

	psus, err := bmc.Psus()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Psus %v", err)
	}

	if len(psus) != len(expectedAnswer) {
		t.Fatalf("Expected %v psus: found %v psus", len(expectedAnswer), len(psus))
	}

	for pos, psu := range psus {
		if psu.Serial != expectedAnswer[pos].Serial || psu.CapacityKw != expectedAnswer[pos].CapacityKw || psu.PowerKw != expectedAnswer[pos].PowerKw || psu.Status != expectedAnswer[pos].Status {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], psu)
		}
	}

	tearDown()
}

func TestIDracIsBlade(t *testing.T) {
	expectedAnswer := false

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

func TestDiskDisks(t *testing.T) {
	expectedAnswer := []*devices.Disk{
		&devices.Disk{
			Serial:    "phdv707000d51p6egn",
			Type:      "SSD",
			Size:      "1490 GB",
			Model:     "ssdsc2bb016t7r",
			Location:  "Disk 0 in Backplane 1 of Integrated Storage Controller 1",
			Status:    "OK",
			FwVersion: "n201dl42",
		},
		&devices.Disk{
			Serial:    "phdv707000fx1p6egn",
			Type:      "SSD",
			Size:      "1490 GB",
			Model:     "ssdsc2bb016t7r",
			Location:  "Disk 1 in Backplane 1 of Integrated Storage Controller 1",
			Status:    "OK",
			FwVersion: "n201dl42",
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
