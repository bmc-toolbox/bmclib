package redfish

// defines various inventory collection helper methods

// collectEnclosure collects Enclosure information
func (i *inventory) collectEnclosure(ch *redfish.Chassis, device *devices.Device) (err error) {

	e := &devices.Enclosure{
		ID:          ch.ID,
		Description: ch.Description,
		Vendor:      ch.Manufacturer,
		Model:       ch.Model,
		ChassisType: string(ch.ChassisType),
		Status: &devices.Status{
			Health: string(ch.Status.Health),
			State:  string(ch.Status.State),
		},
		Firmware: &devices.Firmware{},
	}

	// include additional firmware attributes from redfish firmware inventory
	i.firmwareAttributes(devices.SlugEnclosure, e.ID, e.Firmware)

	device.Enclosures = append(device.Enclosures, e)

	return nil
}

// collectPSUs collects Power Supply Unit component information
func (i *inventory) collectPSUs(ch *redfish.Chassis, device *devices.Device) (err error) {
	power, err := ch.Power()
	if err != nil {
		return err
	}

	if power == nil {
		return nil
	}

	for _, psu := range power.PowerSupplies {
		p := &devices.PSU{
			ID:                 psu.ID,
			Description:        psu.Name,
			Vendor:             psu.Manufacturer,
			Model:              psu.Model,
			Serial:             psu.SerialNumber,
			PowerCapacityWatts: int64(psu.PowerCapacityWatts),
			Status: &devices.Status{
				Health: string(psu.Status.Health),
				State:  string(psu.Status.State),
			},
			Firmware: &devices.Firmware{
				Installed: psu.FirmwareVersion,
			},
		}

		// include additional firmware attributes from redfish firmware inventory
		i.firmwareAttributes(devices.SlugPSU, psu.ID, p.Firmware)

		device.PSUs = append(device.PSUs, p)

	}
	return nil
}

// collectTPMs collects Trusted Platform Module component information
func (i *inventory) collectTPMs(sys *redfish.ComputerSystem, device *devices.Device) (err error) {
	for _, module := range sys.TrustedModules {

		tpm := &devices.TPM{
			InterfaceType: string(module.InterfaceType),
			Firmware: &devices.Firmware{
				Installed: module.FirmwareVersion,
			},
			Status: &devices.Status{
				State:  string(module.Status.State),
				Health: string(module.Status.Health),
			},
		}

		// include additional firmware attributes from redfish firmware inventory
		i.firmwareAttributes(devices.SlugTPM, "TPM", tpm.Firmware)

		device.TPMs = append(device.TPMs, tpm)
	}

	return nil
}

// collectNICs collects network interface component information
func (i *inventory) collectNICs(sys *redfish.ComputerSystem, device *devices.Device) (err error) {
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

		n := &devices.NIC{
			ID:     nic.ID, // "Id": "NIC.Slot.3",
			Vendor: adapter.Manufacturer,
			Model:  adapter.Model,
			Serial: adapter.SerialNumber,
			Status: &devices.Status{
				State:  string(nic.Status.State),
				Health: string(nic.Status.Health),
			},
		}

		if len(adapter.Controllers) > 0 {
			n.Firmware = &devices.Firmware{
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
		i.firmwareAttributes(devices.SlugNIC, n.ID, n.Firmware)

		device.NICs = append(device.NICs, n)
	}

	return nil
}

func (i *inventory) collectBIOS(sys *redfish.ComputerSystem, device *devices.Device) (err error) {
	bios, err := sys.Bios()
	if err != nil {
		return err
	}

	device.BIOS = &devices.BIOS{
		Description: bios.Description,
		Firmware: &devices.Firmware{
			Installed: sys.BIOSVersion,
		},
	}

	// include additional firmware attributes from redfish firmware inventory
	i.firmwareAttributes(devices.SlugBIOS, "BIOS", device.BIOS.Firmware)

	return nil
}

// collectDrives collects drive component information
func (i *inventory) collectDrives(sys *redfish.ComputerSystem, device *devices.Device) (err error) {
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
			d := &devices.Drive{
				ID:                  drive.ID,
				ProductName:         drive.Model,
				Type:                string(drive.MediaType),
				Description:         drive.Description,
				Serial:              drive.SerialNumber,
				StorageController:   member.ID,
				Vendor:              drive.Manufacturer,
				Model:               drive.Model,
				Protocol:            string(drive.Protocol),
				CapacityBytes:       drive.CapacityBytes,
				CapableSpeedGbps:    int64(drive.CapableSpeedGbs),
				NegotiatedSpeedGbps: int64(drive.NegotiatedSpeedGbs),
				BlockSizeBytes:      int64(drive.BlockSizeBytes),
				Firmware: &devices.Firmware{
					Installed: drive.Revision,
				},
				Status: &devices.Status{
					Health: string(drive.Status.Health),
					State:  string(drive.Status.State),
				},
			}

			// include additional firmware attributes from redfish firmware inventory
			i.firmwareAttributes("Disk", drive.ID, d.Firmware)

			device.Drives = append(device.Drives, d)

		}

	}

	return nil
}

// collectStorageControllers populates the device with Storage controller component attributes
func (i *inventory) collectStorageControllers(sys *redfish.ComputerSystem, device *devices.Device) (err error) {
	storage, err := sys.Storage()
	if err != nil {
		return err
	}

	for _, member := range storage {
		for _, controller := range member.StorageControllers {

			c := &devices.StorageController{
				ID:          controller.ID,
				Description: controller.Name,
				Vendor:      controller.Manufacturer,
				Model:       controller.PartNumber,
				Serial:      controller.SerialNumber,
				SpeedGbps:   int64(controller.SpeedGbps),
				Status: &devices.Status{
					Health: string(controller.Status.Health),
					State:  string(controller.Status.State),
				},
				Firmware: &devices.Firmware{
					Installed: controller.FirmwareVersion,
				},
			}

			// include additional firmware attributes from redfish firmware inventory
			i.firmwareAttributes(c.Description, c.ID, c.Firmware)

			device.StorageControllers = append(device.StorageControllers, c)
		}

	}

	return nil
}

// collectCPUs populates the device with CPU component attributes
func (i *inventory) collectCPUs(sys *redfish.ComputerSystem, device *devices.Device) (err error) {
	procs, err := sys.Processors()
	if err != nil {
		return err
	}

	for _, proc := range procs {
		if proc.ProcessorType != "CPU" {
			// TODO: handle this case
			continue
		}

		device.CPUs = append(device.CPUs, &devices.CPU{
			ID:           proc.ID,
			Description:  proc.Description,
			Vendor:       proc.Manufacturer,
			Model:        proc.Model,
			Architecture: string(proc.ProcessorArchitecture),
			Serial:       "",
			Slot:         proc.Socket,
			ClockSpeedHz: int64(proc.MaxSpeedMHz),
			Cores:        proc.TotalCores,
			Threads:      proc.TotalThreads,
			Status: &devices.Status{
				Health: string(proc.Status.Health),
				State:  string(proc.Status.State),
			},
			Firmware: &devices.Firmware{
				Installed: proc.ProcessorID.MicrocodeInfo,
			},
		})
	}

	return nil
}

// collectDIMMs populates the device with memory component attributes
func (i *inventory) collectDIMMs(sys *redfish.ComputerSystem, device *devices.Device) (err error) {
	dimms, err := sys.Memory()
	if err != nil {
		return err
	}

	for _, dimm := range dimms {
		device.Memory = append(device.Memory, &devices.Memory{
			Description:  dimm.Description,
			Slot:         dimm.ID,
			Type:         string(dimm.MemoryType),
			Vendor:       dimm.Manufacturer,
			Model:        "",
			Serial:       dimm.SerialNumber,
			SizeBytes:    int64(dimm.VolatileSizeMiB),
			FormFactor:   "",
			PartNumber:   dimm.PartNumber,
			ClockSpeedHz: int64(dimm.OperatingSpeedMhz),
			Status: &devices.Status{
				Health: string(dimm.Status.Health),
				State:  string(dimm.Status.State),
			},
		})
	}

	return nil
}

// collecCPLDs populates the device with CPLD component attributes
func (i *inventory) collectCPLDs(device *devices.Device) (err error) {

	cpld := &devices.CPLD{
		Vendor:   device.Vendor,
		Model:    device.Model,
		Firmware: &devices.Firmware{Metadata: make(map[string]string)},
	}

	i.firmwareAttributes(devices.SlugCPLD, "", cpld.Firmware)
	name, exists := cpld.Firmware.Metadata["name"]
	if exists {
		cpld.Description = name
	}

	device.CPLDs = []*devices.CPLD{cpld}

	return nil
}
