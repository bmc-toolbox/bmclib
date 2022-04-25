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
