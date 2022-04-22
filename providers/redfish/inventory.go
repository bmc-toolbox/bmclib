package redfish

import (
	"context"
	"strings"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/redfish"

	bmclibErrs "github.com/bmc-toolbox/bmclib/errors"
)

var (

	// Supported Chassis Odata IDs
	chassisOdataIDs = []string{
		// Dells
		"/redfish/v1/Chassis/Enclosure.Internal.0-1",
		"/redfish/v1/Chassis/System.Embedded.1",
		"/redfish/v1/Chassis/Enclosure.Internal.0-1:NonRAID.Integrated.1-1",
		// Supermicro
		"/redfish/v1/Chassis/1",
		// MegaRAC
		"/redfish/v1/Chassis/Self",
	}

	// Supported System Odata IDs
	systemsOdataIDs = []string{
		// Dells
		"/redfish/v1/Systems/System.Embedded.1",
	}

	// Supported Manager Odata IDs (BMCs)
	managerOdataIDs = []string{
		"/redfish/v1/Managers/iDRAC.Embedded.1",
	}
)

// inventory struct wraps redfish connection
type inventory struct {
	conn              *gofish.APIClient
	softwareInventory []*redfish.SoftwareInventory
}

// DeviceVendorModel returns the device vendor and model attributes
func (c *Conn) DeviceVendorModel(ctx context.Context) (vendor, model string, err error) {
	systems, err := c.conn.Service.Systems()
	if err != nil {
		return vendor, model, err
	}

	for _, sys := range systems {
		if !compatibleOdataID(sys.ODataID, systemsOdataIDs) {
			continue
		}

		return sys.Manufacturer, sys.Model, nil
	}

	return vendor, model, bmclibErrs.ErrRedfishSystemOdataID
}

func (c *Conn) Inventory(ctx context.Context) (device *devices.Device, err error) {
	// initialize inventory object
	inv := &inventory{conn: c.conn}
	// TODO: this can soft fail
	inv.softwareInventory, err = inv.collectSoftwareInventory()
	if err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrRedfishSoftwareInventory, err.Error())
	}

	// initialize device to be populated with inventory
	device = devices.NewDevice()

	// populate device Chassis components attributes
	err = inv.chassisAttributes(device)
	if err != nil {
		return nil, err
	}

	// populate device System components attributes
	err = inv.systemAttributes(device)
	if err != nil {
		return nil, err
	}

	// populate device BMC component attributes
	err = inv.bmcAttributes(device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// collectSoftwareInventory returns redfish software inventory
func (i *inventory) collectSoftwareInventory() ([]*redfish.SoftwareInventory, error) {
	service := i.conn.Service
	if service == nil {
		return nil, bmclibErrs.ErrRedfishServiceNil
	}

	updateService, err := service.UpdateService()
	if err != nil {
		return nil, err
	}

	return updateService.FirmwareInventories()
}

// bmcAttributes collects BMC component attributes
func (i *inventory) bmcAttributes(device *devices.Device) (err error) {
	service := i.conn.Service
	if service == nil {
		return bmclibErrs.ErrRedfishServiceNil
	}

	managers, err := service.Managers()
	if err != nil {
		return err
	}

	var compatible int
	for _, manager := range managers {
		if !compatibleOdataID(manager.ODataID, managerOdataIDs) {
			continue
		}

		compatible++

		if manager.ManagerType != "BMC" {
			continue
		}

		device.BMC = &devices.BMC{
			ID:          manager.ID,
			Description: manager.Description,
			Vendor:      device.Vendor,
			Model:       device.Model,
			Status: &devices.Status{
				Health: string(manager.Status.Health),
				State:  string(manager.Status.State),
			},
			Firmware: &devices.Firmware{
				Installed: manager.FirmwareVersion,
			},
		}

		// include additional firmware attributes from redfish firmware inventory
		i.firmwareAttributes("", device.BMC.ID, device.BMC.Firmware)
	}

	if compatible == 0 {
		return bmclibErrs.ErrRedfishManagerOdataID
	}

	return nil
}

// chassisAttributes populates the device chassis attributes
func (i *inventory) chassisAttributes(device *devices.Device) (err error) {
	service := i.conn.Service
	if service == nil {
		return bmclibErrs.ErrRedfishServiceNil
	}

	chassis, err := service.Chassis()
	if err != nil {
		return err
	}

	compatible := 0
	for _, ch := range chassis {
		if !compatibleOdataID(ch.ODataID, chassisOdataIDs) {
			continue
		}

		compatible++

		err = i.collectEnclosure(ch, device)
		if err != nil {
			return err
		}

		err = i.collectPSUs(ch, device)
		if err != nil {
			return err
		}

	}

	err = i.collectCPLDs(device)
	if err != nil {
		return err
	}

	if compatible == 0 {
		return bmclibErrs.ErrRedfishChassisOdataID
	}

	return nil

}

func (i *inventory) systemAttributes(device *devices.Device) (err error) {
	service := i.conn.Service
	if service == nil {
		return bmclibErrs.ErrRedfishServiceNil
	}

	systems, err := service.Systems()
	if err != nil {
		return err
	}

	compatible := 0
	for _, sys := range systems {
		if !compatibleOdataID(sys.ODataID, systemsOdataIDs) {
			continue
		}

		compatible++

		if sys.Manufacturer != "" && sys.Model != "" && sys.SerialNumber != "" {
			device.Vendor = sys.Manufacturer
			device.Model = sys.Model
			device.Serial = sys.SerialNumber
		}

		// slice of collector methods
		funcs := []func(sys *redfish.ComputerSystem, device *devices.Device) error{
			i.collectCPUs,
			i.collectDIMMs,
			i.collectDrives,
			i.collectBIOS,
			i.collectNICs,
			i.collectTPMs,
			i.collectStorageControllers,
		}

		// execute collector methods
		for _, f := range funcs {
			err := f(sys, device)
			if err != nil {
				return err
			}
		}
	}

	if compatible == 0 {
		return bmclibErrs.ErrRedfishSystemOdataID
	}

	return nil
}

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

// firmwareInventory looks up the redfish inventory for objects that
// match - 1. slug, 2. id
// and returns the intalled or previous firmware for objects that matched
//
// slug - the component slug constant
// id - the component ID
// previous - when true returns previously installed firmware, else returns the current
func (i *inventory) firmwareAttributes(slug, id string, firmwareObj *devices.Firmware) {
	if len(i.softwareInventory) == 0 {
		return
	}

	if id == "" {
		id = slug
	}

	for _, inv := range i.softwareInventory {
		// include previously installed firmware attributes
		if strings.HasPrefix(inv.ID, "Previous") {
			if strings.Contains(inv.ID, id) || strings.EqualFold(slug, inv.Name) {

				if firmwareObj == nil {
					firmwareObj = &devices.Firmware{}
				}

				if firmwareObj.Installed == inv.Version {
					continue
				}

				firmwareObj.Previous = append(firmwareObj.Previous, &devices.Firmware{
					Installed:  inv.Version,
					SoftwareID: inv.SoftwareID,
				})
			}
		}

		// update firmwareObj with installed firmware attributes
		if strings.HasPrefix(inv.ID, "Installed") {
			if strings.Contains(inv.ID, id) || strings.EqualFold(slug, inv.Name) {

				if firmwareObj == nil {
					firmwareObj = &devices.Firmware{}
				}

				if firmwareObj.Installed == "" || firmwareObj.Installed != inv.Version {
					firmwareObj.Installed = inv.Version
				}

				firmwareObj.Metadata = map[string]string{"name": inv.Name}
				firmwareObj.SoftwareID = inv.SoftwareID
			}
		}
	}
}

func compatibleOdataID(OdataID string, knownOdataIDs []string) bool {
	for _, url := range knownOdataIDs {
		if url == OdataID {
			return true
		}
	}

	return false
}

func stringInSlice(s string, sl []string) bool {
	for _, elem := range sl {
		if elem == s {
			return true
		}
	}

	return false
}
