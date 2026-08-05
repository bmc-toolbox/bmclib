package redfishwrapper

import (
	"math"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
)

// defines various inventory collection helper methods

// collectEnclosure collects Enclosure information
func (c *Client) collectEnclosure(ch *schemas.Chassis, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	e := &common.Enclosure{
		Common: common.Common{
			Description: ch.Description,
			Vendor:      common.FormatVendorName(ch.Manufacturer),
			Model:       ch.Model,
			Serial:      ch.SerialNumber,
			Status: &common.Status{
				Health: string(ch.Status.Health),
				State:  string(ch.Status.State),
			},
			Firmware: &common.Firmware{},
		},

		ID:          ch.ID,
		ChassisType: string(ch.ChassisType),
	}

	if e.Model == "" && ch.PartNumber != "" {
		e.Model = ch.PartNumber
	}

	// include additional firmware attributes from redfish firmware inventory
	c.firmwareAttributes(common.SlugEnclosure, e.ID, e.Firmware, softwareInventory)

	device.Enclosures = append(device.Enclosures, e)

	return nil
}

// collectMainboard collects motherboard information.
//
// Some vendors (e.g. Supermicro) don't expose a dedicated motherboard
// resource: the Chassis resource's Model is the board's own model, while
// PartNumber/SerialNumber describe the case rather than the board itself.
// Vendor-specific providers may fill in additional Mainboard attributes
// (e.g. the board's own serial number) after Inventory returns.
//
// chassisAttributes calls this once per compatible Chassis member, and some
// vendors (e.g. Dell) have more than one compatible Chassis ID. device.Mainboard
// is a single field, not a slice, so only populate it from the first chassis
// that provides a model — don't let an unrelated chassis (e.g. a drive
// enclosure/backplane with its own Model set) silently overwrite it later.
func (c *Client) collectMainboard(ch *schemas.Chassis, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	if ch == nil || ch.Model == "" || (device.Mainboard != nil && device.Mainboard.Model != "") {
		return nil
	}

	m := &common.Mainboard{
		Common: common.Common{
			Vendor: common.FormatVendorName(ch.Manufacturer),
			Model:  ch.Model,
			Status: &common.Status{
				Health: string(ch.Status.Health),
				State:  string(ch.Status.State),
			},
			Firmware: &common.Firmware{},
		},

		PhysicalID: ch.ID,
	}

	// include additional firmware attributes from redfish firmware inventory
	c.firmwareAttributes(common.SlugMainboard, "", m.Firmware, softwareInventory)

	device.Mainboard = m

	return nil
}

// collectPSUs collects Power Supply Unit component information
func (c *Client) collectPSUs(ch *schemas.Chassis, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	power, err := ch.Power()
	if err != nil {
		return err
	}

	if power == nil {
		return nil
	}

	for i := range power.PowerSupplies {
		psu := &power.PowerSupplies[i]
		p := &common.PSU{
			Common: common.Common{
				Description: psu.Name,
				Vendor:      common.FormatVendorName(psu.Manufacturer),
				Model:       psu.Model,
				Serial:      psu.SerialNumber,

				Status: &common.Status{
					Health: string(psu.Status.Health),
					State:  string(psu.Status.State),
				},
				Firmware: &common.Firmware{
					Installed: psu.FirmwareVersion,
				},
			},

			ID:                 psu.ID,
			PowerCapacityWatts: int64(gofish.Deref(psu.PowerCapacityWatts)),
		}

		// include additional firmware attributes from redfish firmware inventory
		c.firmwareAttributes(common.SlugPSU, psu.ID, p.Firmware, softwareInventory)

		device.PSUs = append(device.PSUs, p)
	}
	return nil
}

// collectTPMs collects Trusted Platform Module component information
func (c *Client) collectTPMs(sys *schemas.ComputerSystem, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	modules := sys.TrustedModules //nolint:staticcheck // TrustedModules is deprecated but still read for backward compatibility
	for i := range modules {
		module := &modules[i]
		tpm := &common.TPM{
			Common: common.Common{
				Firmware: &common.Firmware{
					Installed: module.FirmwareVersion,
				},
				Status: &common.Status{
					State:  string(module.Status.State),
					Health: string(module.Status.Health),
				},
			},

			InterfaceType: string(module.InterfaceType),
		}

		// include additional firmware attributes from redfish firmware inventory
		c.firmwareAttributes(common.SlugTPM, "TPM", tpm.Firmware, softwareInventory)

		device.TPMs = append(device.TPMs, tpm)
	}

	return nil
}

// collectNICs collects network interface component information
func (c *Client) collectNICs(sys *schemas.ComputerSystem, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	if sys == nil || device == nil {
		return nil
	}

	// collect network interface information
	nics, err := sys.NetworkInterfaces()
	if err != nil {
		return err
	}

	// collect network ethernet interface information, these attributes are not available in NetworkAdapter, NetworkInterfaces
	ethernetInterfaces, err := sys.EthernetInterfaces()
	if err != nil {
		return err
	}

	// Fallback: add any EthernetInterfaces not discovered via NetworkInterfaces.
	// Some adapters (e.g. Supermicro AOC add-on cards) only appear in EthernetInterfaces.
	discoveredMACs := make(map[string]struct{})

	for _, nic := range nics {
		// collect network interface adaptor information
		adapter, err := nic.NetworkAdapter()
		if err != nil {
			return err
		}

		if adapter == nil {
			continue
		}

		n := &common.NIC{
			Common: common.Common{
				Vendor:      common.FormatVendorName(adapter.Manufacturer),
				Model:       adapter.Model,
				Serial:      adapter.SerialNumber,
				ProductName: adapter.PartNumber,
				Status: &common.Status{
					State:  string(nic.Status.State),
					Health: string(nic.Status.Health),
				},
			},

			ID: nic.ID, // "Id": "NIC.Slot.3",
		}

		ports, err := adapter.NetworkPorts()
		if err != nil {
			return err
		}

		portFirmwareVersion := getFirmwareVersionFromController(adapter.Controllers, len(ports))

		// NetworkPort.ID is only unique within its own adapter, so a card with
		// local port IDs "1"/"2" (e.g. an add-on AOC) collides with an unrelated
		// onboard NIC that also has ports "1"/"2". Resolve each port's MAC
		// authoritatively via NetworkDeviceFunctions, which links a specific
		// port to its function by odata.id rather than by the ambiguous local ID.
		portMACs := c.networkPortMACs(adapter)

		for _, networkPort := range ports {
			// populate network ports general data
			nicPort := &common.NICPort{}
			c.collectNetworkPortInfo(nicPort, adapter, networkPort, portFirmwareVersion, softwareInventory)

			if networkPort.ActiveLinkTechnology == schemas.EthernetLinkNetworkTechnology {
				// ethernet specific data
				c.collectEthernetInfo(nicPort, ethernetInterfaces, portMACs[networkPort.ODataID])
			}
			n.NICPorts = append(n.NICPorts, nicPort)

			if nicPort.MacAddress != "" {
				discoveredMACs[strings.ToLower(nicPort.MacAddress)] = struct{}{}
			}
		}

		// include additional firmware attributes from redfish firmware inventory
		c.firmwareAttributes(common.SlugNIC, n.ID, n.Firmware, softwareInventory)
		if portFirmwareVersion != "" {
			if n.Firmware == nil {
				n.Firmware = &common.Firmware{}
			}
			n.Firmware.Installed = portFirmwareVersion
		}

		device.NICs = append(device.NICs, n)
	}
	for _, eth := range ethernetInterfaces {
		mac := strings.ToLower(eth.MACAddress)
		if mac == "" || mac == "00:00:00:00:00:00" {
			continue
		}
		if _, seen := discoveredMACs[mac]; seen {
			continue
		}

		nicPort := &common.NICPort{
			Common: common.Common{
				Description: eth.Description,
				Status: &common.Status{
					Health: string(eth.Status.Health),
					State:  string(eth.Status.State),
				},
			},
			ID:         eth.ID,
			MacAddress: eth.MACAddress,
			AutoNeg:    eth.AutoNeg,
			MTUSize:    gofish.Deref(eth.MTUSize),
		}
		if eth.SpeedMbps != nil {
			nicPort.SpeedBits = int64(gofish.Deref(eth.SpeedMbps)) * int64(math.Pow10(6))
		}

		device.NICs = append(device.NICs, &common.NIC{
			Common: common.Common{
				Description: eth.Name,
				Status: &common.Status{
					State:  string(eth.Status.State),
					Health: string(eth.Status.Health),
				},
			},
			ID:       eth.ID,
			NICPorts: []*common.NICPort{nicPort},
		})
	}

	return nil
}

func (c *Client) collectNetworkPortInfo(
	nicPort *common.NICPort,
	adapter *schemas.NetworkAdapter,
	networkPort *schemas.NetworkPort,
	firmware string,
	softwareInventory []*schemas.SoftwareInventory,
) {
	if adapter != nil {
		nicPort.Vendor = adapter.Manufacturer
		nicPort.Model = adapter.Model
	}

	if networkPort != nil {
		nicPort.Description = networkPort.Description
		nicPort.PCIVendorID = networkPort.VendorID
		nicPort.Status = &common.Status{
			Health: string(networkPort.Status.Health),
			State:  string(networkPort.Status.State),
		}
		nicPort.ID = networkPort.ID
		nicPort.PhysicalID = networkPort.PhysicalPortNumber
		nicPort.LinkStatus = string(networkPort.LinkStatus)
		nicPort.ActiveLinkTechnology = string(networkPort.ActiveLinkTechnology)

		if networkPort.CurrentLinkSpeedMbps != nil {
			nicPort.SpeedBits = int64(gofish.Deref(networkPort.CurrentLinkSpeedMbps)) * int64(math.Pow10(6))
		}

		if len(networkPort.AssociatedNetworkAddresses) > 0 {
			for _, macAddress := range networkPort.AssociatedNetworkAddresses {
				if macAddress != "" && macAddress != "00:00:00:00:00:00" {
					nicPort.MacAddress = macAddress // first valid value only
					break
				}
			}
		}

		c.firmwareAttributes(common.SlugNIC, networkPort.ID, nicPort.Firmware, softwareInventory)
	}
	if firmware != "" {
		if nicPort.Firmware == nil {
			nicPort.Firmware = &common.Firmware{}
		}
		nicPort.Firmware.Installed = firmware
	}
}

// networkPortMACs returns a map of NetworkPort odata.id -> MAC address for the
// given adapter, resolved via its NetworkDeviceFunctions. This is the
// authoritative link between a physical port and its MAC address: unlike
// NetworkPort.ID (only unique within the adapter) or EthernetInterface.ID
// (system-wide, but on some Redfish implementations arbitrarily numbered),
// NetworkDeviceFunction.PhysicalPortAssignment resolves to the exact NetworkPort
// resource, keyed by its unique odata.id.
func (c *Client) networkPortMACs(adapter *schemas.NetworkAdapter) map[string]string {
	macs := make(map[string]string)
	if adapter == nil {
		return macs
	}

	functions, err := adapter.NetworkDeviceFunctions()
	if err != nil {
		return macs
	}

	for _, function := range functions {
		mac := function.Ethernet.MACAddress
		if mac == "" || mac == "00:00:00:00:00:00" {
			mac = function.Ethernet.PermanentMACAddress
		}
		if mac == "" || mac == "00:00:00:00:00:00" {
			continue
		}

		port, err := function.PhysicalPortAssignment()
		if err != nil || port == nil {
			continue
		}

		macs[port.ODataID] = mac
	}

	return macs
}

func (c *Client) collectEthernetInfo(nicPort *common.NICPort, ethernetInterfaces []*schemas.EthernetInterface, preferredMAC string) {
	if nicPort == nil {
		return
	}
	// populate mac address et al. from matching ethernet interface
	for _, ethInterface := range ethernetInterfaces {
		if preferredMAC != "" {
			// preferredMAC was resolved authoritatively via the adapter's
			// NetworkDeviceFunctions; match on it instead of the ambiguous
			// NetworkPort/EthernetInterface ID so that add-on cards whose local
			// port IDs collide with an unrelated adapter's IDs (e.g. both using
			// "1"/"2") don't inherit that adapter's ethernet attributes.
			if !strings.EqualFold(ethInterface.MACAddress, preferredMAC) {
				continue
			}
		} else if ethInterface.ID != nicPort.ID && !strings.HasPrefix(ethInterface.ID, nicPort.ID+"-") {
			// the ethernet interface includes the port, position number and function NIC.Slot.3-1-1;
			// require an exact match or a "-"-delimited prefix so that port "1" does not
			// incorrectly absorb EthernetInterface "10" (HasPrefix("10","1") is true).
			continue
		}

		// override values only if needed
		if ethInterface.Description != "" {
			nicPort.Description = ethInterface.Description
		}
		if len(ethInterface.Status.Health) > 0 {
			if nicPort.Status == nil {
				nicPort.Status = &common.Status{}
			}
			nicPort.Status.Health = string(ethInterface.Status.Health)
		}
		if len(ethInterface.Status.State) > 0 {
			if nicPort.Status == nil {
				nicPort.Status = &common.Status{}
			}
			nicPort.Status.State = string(ethInterface.Status.State)
		}
		nicPort.ID = ethInterface.ID // override ID
		if ethInterface.SpeedMbps != nil {
			nicPort.SpeedBits = int64(gofish.Deref(ethInterface.SpeedMbps)) * int64(math.Pow10(6))
		}
		nicPort.AutoNeg = ethInterface.AutoNeg
		nicPort.MTUSize = gofish.Deref(ethInterface.MTUSize)

		// always override mac address
		nicPort.MacAddress = ethInterface.MACAddress
		break // stop at first match
	}

	// preferredMAC is authoritative even if no EthernetInterface matched it
	// (e.g. the port isn't enumerated in the EthernetInterfaces collection at all).
	if preferredMAC != "" {
		nicPort.MacAddress = preferredMAC
	}
}

func getFirmwareVersionFromController(controllers []schemas.Controllers, portCount int) string {
	for i := range controllers {
		controller := &controllers[i]
		if gofish.Deref(controller.ControllerCapabilities.NetworkPortCount) == portCount {
			return controller.FirmwarePackageVersion
		}
	}
	return ""
}

func (c *Client) collectBIOS(sys *schemas.ComputerSystem, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	device.BIOS = &common.BIOS{
		Common: common.Common{
			Vendor: common.FormatVendorName(sys.Manufacturer),
			Firmware: &common.Firmware{
				Installed: sys.BiosVersion,
			},
		},
	}

	bios, err := sys.Bios()
	if err != nil {
		return err
	}

	if bios != nil {
		device.BIOS.Description = bios.Description
	}

	// include additional firmware attributes from redfish firmware inventory
	c.firmwareAttributes(common.SlugBIOS, "BIOS", device.BIOS.Firmware, softwareInventory)

	return nil
}

// collectDrives collects drive component information
func (c *Client) collectDrives(sys *schemas.ComputerSystem, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	storage, err := sys.Storage()
	if err != nil {
		return err
	}

	for _, member := range storage {
		// DrivesCount comes from the optional "Drives@odata.count" property.
		// Some implementations populate the "Drives" link array without ever
		// including a count property, which gofish reports as DrivesCount==0
		// even though real drives are linked. member.Drives() reads the
		// underlying link list directly and safely returns an empty slice
		// when there's genuinely nothing there, so skip the DrivesCount gate
		// entirely rather than short-circuiting on a field that may be unset.
		drives, err := member.Drives()
		if err != nil {
			return err
		}

		for _, drive := range drives {
			d := &common.Drive{
				Common: common.Common{
					ProductName: drive.Model,
					Description: drive.Description,
					Serial:      drive.SerialNumber,
					Vendor:      common.FormatVendorName(drive.Manufacturer),
					Model:       drive.Model,
					Firmware: &common.Firmware{
						Installed: drive.Revision,
					},
					Status: &common.Status{
						Health: string(drive.Status.Health),
						State:  string(drive.Status.State),
					},
				},

				ID:                  drive.ID,
				Type:                string(drive.MediaType),
				StorageController:   member.ID,
				Protocol:            string(drive.Protocol),
				CapacityBytes:       int64(gofish.Deref(drive.CapacityBytes)),
				CapableSpeedGbps:    int64(gofish.Deref(drive.CapableSpeedGbs)),
				NegotiatedSpeedGbps: int64(gofish.Deref(drive.NegotiatedSpeedGbs)),
				BlockSizeBytes:      int64(gofish.Deref(drive.BlockSizeBytes)),
			}

			// include additional firmware attributes from redfish firmware inventory
			c.firmwareAttributes("Disk", drive.ID, d.Firmware, softwareInventory)

			device.Drives = append(device.Drives, d)
		}
	}

	return nil
}

// collectStorageControllers populates the device with Storage controller component attributes
func (c *Client) collectStorageControllers(sys *schemas.ComputerSystem, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	storage, err := sys.Storage()
	if err != nil {
		return err
	}

	for _, member := range storage {
		controllers := member.StorageControllers //nolint:staticcheck // StorageControllers is deprecated but still read for backward compatibility
		for i := range controllers {
			controller := &controllers[i]

			cs := &common.StorageController{
				Common: common.Common{
					Description: controller.Name,
					Vendor:      common.FormatVendorName(controller.Manufacturer),
					Model:       controller.PartNumber,
					Serial:      controller.SerialNumber,
					Status: &common.Status{
						Health: string(controller.Status.Health),
						State:  string(controller.Status.State),
					},
					Firmware: &common.Firmware{
						Installed: controller.FirmwareVersion,
					},
				},

				ID:        controller.ID,
				SpeedGbps: int64(gofish.Deref(controller.SpeedGbps)),
			}

			// In some cases the storage controller model number is present in the Name field
			if strings.TrimSpace(cs.Model) == "" && strings.TrimSpace(controller.Name) != "" {
				cs.Model = controller.Name
			}

			// include additional firmware attributes from redfish firmware inventory
			c.firmwareAttributes(cs.Description, cs.ID, cs.Firmware, softwareInventory)

			device.StorageControllers = append(device.StorageControllers, cs)
		}
	}

	return nil
}

// collectCPUs populates the device with CPU component attributes
func (c *Client) collectCPUs(sys *schemas.ComputerSystem, device *common.Device, _ []*schemas.SoftwareInventory) (err error) {
	procs, err := sys.Processors()
	if err != nil {
		return err
	}

	for _, proc := range procs {
		if proc.ProcessorType != "CPU" {
			// TODO: handle this case
			continue
		}

		device.CPUs = append(device.CPUs, &common.CPU{
			Common: common.Common{
				Description: proc.Description,
				Vendor:      common.FormatVendorName(proc.Manufacturer),
				Model:       proc.Model,
				Serial:      "",
				Status: &common.Status{
					Health: string(proc.Status.Health),
					State:  string(proc.Status.State),
				},
				Firmware: &common.Firmware{
					Installed: proc.ProcessorID.MicrocodeInfo,
				},
			},
			ID:           proc.ID,
			Architecture: string(proc.ProcessorArchitecture),
			Slot:         proc.Socket,
			ClockSpeedHz: int64(gofish.Deref(proc.MaxSpeedMHz) * 1000 * 1000),
			Cores:        gofish.Deref(proc.TotalCores),
			Threads:      gofish.Deref(proc.TotalThreads),
		})
	}

	return nil
}

// collectDIMMs populates the device with memory component attributes
func (c *Client) collectDIMMs(sys *schemas.ComputerSystem, device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	dimms, err := sys.Memory()
	if err != nil {
		return err
	}

	for _, dimm := range dimms {
		// CapacityMiB is the overall module capacity and should be populated for
		// ordinary (volatile-only) DIMMs. VolatileSizeMiB/NonVolatileSizeMiB only
		// apply to memory with a volatile/non-volatile split (e.g. NVDIMM); fall
		// back to their sum when CapacityMiB is absent, rather than using
		// VolatileSizeMiB alone, which is 0 for plain DRAM on some platforms and
		// silently produced a zero SizeBytes.
		sizeMiB := gofish.Deref(dimm.CapacityMiB)
		if sizeMiB == 0 {
			sizeMiB = gofish.Deref(dimm.VolatileSizeMiB) + gofish.Deref(dimm.NonVolatileSizeMiB)
		}

		device.Memory = append(device.Memory, &common.Memory{
			Common: common.Common{
				Description: dimm.Description,
				Vendor:      common.FormatVendorName(dimm.Manufacturer),
				Model:       "",
				Serial:      dimm.SerialNumber,
				Status: &common.Status{
					Health: string(dimm.Status.Health),
					State:  string(dimm.Status.State),
				},
			},

			Slot:         dimm.ID,
			Type:         string(dimm.MemoryType),
			SizeBytes:    int64(sizeMiB) * 1024 * 1024,
			FormFactor:   "",
			PartNumber:   dimm.PartNumber,
			ClockSpeedHz: int64(gofish.Deref(dimm.OperatingSpeedMhz) * 1000 * 1000),
		})
	}

	return nil
}

// collecCPLDs populates the device with CPLD component attributes
func (c *Client) collectCPLDs(device *common.Device, softwareInventory []*schemas.SoftwareInventory) (err error) {
	cpld := &common.CPLD{
		Common: common.Common{
			Vendor:   common.FormatVendorName(device.Vendor),
			Model:    device.Model,
			Firmware: &common.Firmware{Metadata: make(map[string]string)},
		},
	}

	c.firmwareAttributes(common.SlugCPLD, "", cpld.Firmware, softwareInventory)
	name, exists := cpld.Firmware.Metadata["name"]
	if exists {
		cpld.Description = name
	}

	device.CPLDs = []*common.CPLD{cpld}

	return nil
}
