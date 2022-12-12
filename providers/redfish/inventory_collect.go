package redfish

import (
	"strings"

	"github.com/bmc-toolbox/common"
	gofishrf "github.com/stmcginnis/gofish/redfish"
)

// defines various inventory collection helper methods

// collectEnclosure collects Enclosure information
func (i *inventory) collectEnclosure(ch *gofishrf.Chassis, device *common.Device) (err error) {
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
	i.firmwareAttributes(common.SlugEnclosure, e.ID, e.Firmware)

	device.Enclosures = append(device.Enclosures, e)

	return nil
}

// collectPSUs collects Power Supply Unit component information
func (i *inventory) collectPSUs(ch *gofishrf.Chassis, device *common.Device) (err error) {
	power, err := ch.Power()
	if err != nil {
		return err
	}

	if power == nil {
		return nil
	}

	for _, psu := range power.PowerSupplies {
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
			PowerCapacityWatts: int64(psu.PowerCapacityWatts),
		}

		// include additional firmware attributes from redfish firmware inventory
		i.firmwareAttributes(common.SlugPSU, psu.ID, p.Firmware)

		device.PSUs = append(device.PSUs, p)

	}
	return nil
}

// collectTPMs collects Trusted Platform Module component information
func (i *inventory) collectTPMs(sys *gofishrf.ComputerSystem, device *common.Device) (err error) {
	for _, module := range sys.TrustedModules {
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
		i.firmwareAttributes(common.SlugTPM, "TPM", tpm.Firmware)

		device.TPMs = append(device.TPMs, tpm)
	}

	return nil
}

// collectNICs collects network interface component information
func (i *inventory) collectNICs(sys *gofishrf.ComputerSystem, device *common.Device) (err error) {
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
				Vendor: common.FormatVendorName(adapter.Manufacturer),
				Model:  adapter.Model,
				Serial: adapter.SerialNumber,
				Status: &common.Status{
					State:  string(nic.Status.State),
					Health: string(nic.Status.Health),
				},
			},

			ID: nic.ID, // "Id": "NIC.Slot.3",
		}

		if len(adapter.Controllers) > 0 {
			n.Firmware = &common.Firmware{
				Installed: adapter.Controllers[0].FirmwarePackageVersion,
			}
		}

		// populate mac addresses from ethernet interfaces
		for _, ethInterface := range ethernetInterfaces {
			// the ethernet interface includes the port and position number NIC.Slot.3-1-1
			if !strings.HasPrefix(ethInterface.ID, adapter.ID) {
				continue
			}

			// The ethernet interface description is
			n.Description = ethInterface.Description
			n.MacAddress = ethInterface.MACAddress
			n.SpeedBits = int64(ethInterface.SpeedMbps*10 ^ 6)
		}

		// include additional firmware attributes from redfish firmware inventory
		i.firmwareAttributes(common.SlugNIC, n.ID, n.Firmware)

		device.NICs = append(device.NICs, n)
	}

	return nil
}

func (i *inventory) collectBIOS(sys *gofishrf.ComputerSystem, device *common.Device) (err error) {
	device.BIOS = &common.BIOS{
		Common: common.Common{
			Firmware: &common.Firmware{
				Installed: sys.BIOSVersion,
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
	i.firmwareAttributes(common.SlugBIOS, "BIOS", device.BIOS.Firmware)

	return nil
}

// collectDrives collects drive component information
func (i *inventory) collectDrives(sys *gofishrf.ComputerSystem, device *common.Device) (err error) {
	storage, err := sys.Storage()
	if err != nil {
		return err
	}

	for _, member := range storage {
		if member.DrivesCount == 0 {
			continue
		}

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
				CapacityBytes:       drive.CapacityBytes,
				CapableSpeedGbps:    int64(drive.CapableSpeedGbs),
				NegotiatedSpeedGbps: int64(drive.NegotiatedSpeedGbs),
				BlockSizeBytes:      int64(drive.BlockSizeBytes),
			}

			// include additional firmware attributes from redfish firmware inventory
			i.firmwareAttributes("Disk", drive.ID, d.Firmware)

			device.Drives = append(device.Drives, d)

		}

	}

	return nil
}

// collectStorageControllers populates the device with Storage controller component attributes
func (i *inventory) collectStorageControllers(sys *gofishrf.ComputerSystem, device *common.Device) (err error) {
	storage, err := sys.Storage()
	if err != nil {
		return err
	}

	for _, member := range storage {
		for _, controller := range member.StorageControllers {

			c := &common.StorageController{
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
				SpeedGbps: int64(controller.SpeedGbps),
			}

			// In some cases the storage controller model number is present in the Name field
			if strings.TrimSpace(c.Model) == "" && strings.TrimSpace(controller.Name) != "" {
				c.Model = controller.Name
			}

			// include additional firmware attributes from redfish firmware inventory
			i.firmwareAttributes(c.Description, c.ID, c.Firmware)

			device.StorageControllers = append(device.StorageControllers, c)
		}

	}

	return nil
}

// collectCPUs populates the device with CPU component attributes
func (i *inventory) collectCPUs(sys *gofishrf.ComputerSystem, device *common.Device) (err error) {
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
			ClockSpeedHz: int64(proc.MaxSpeedMHz * 1000 * 1000),
			Cores:        proc.TotalCores,
			Threads:      proc.TotalThreads,
		})
	}

	return nil
}

// collectDIMMs populates the device with memory component attributes
func (i *inventory) collectDIMMs(sys *gofishrf.ComputerSystem, device *common.Device) (err error) {
	dimms, err := sys.Memory()
	if err != nil {
		return err
	}

	for _, dimm := range dimms {
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
			SizeBytes:    int64(dimm.VolatileSizeMiB * 1024 * 1024),
			FormFactor:   "",
			PartNumber:   dimm.PartNumber,
			ClockSpeedHz: int64(dimm.OperatingSpeedMhz * 1000 * 1000),
		})
	}

	return nil
}

// collecCPLDs populates the device with CPLD component attributes
func (i *inventory) collectCPLDs(device *common.Device) (err error) {

	cpld := &common.CPLD{
		Common: common.Common{
			Vendor:   common.FormatVendorName(device.Vendor),
			Model:    device.Model,
			Firmware: &common.Firmware{Metadata: make(map[string]string)},
		},
	}

	i.firmwareAttributes(common.SlugCPLD, "", cpld.Firmware)
	name, exists := cpld.Firmware.Metadata["name"]
	if exists {
		cpld.Description = name
	}

	device.CPLDs = []*common.CPLD{cpld}

	return nil
}
