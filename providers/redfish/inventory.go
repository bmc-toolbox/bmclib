package redfish

import (
	"context"
	"strings"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/pkg/errors"

	"github.com/bmc-toolbox/common"
	gofishrf "github.com/stmcginnis/gofish/redfish"
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
		// OpenBMC on ASRock
		"/redfish/v1/Chassis/ASRock_ROMED8HM3",
	}

	// Supported System Odata IDs
	systemsOdataIDs = []string{
		// Dells
		"/redfish/v1/Systems/System.Embedded.1",
		"/redfish/v1/Systems/System.Embedded.1/Bios",
		// Supermicros
		"/redfish/v1/Systems/1",
		// OpenBMC on ASRock
		"/redfish/v1/Systems/system",
	}

	// Supported Manager Odata IDs (BMCs)
	managerOdataIDs = []string{
		"/redfish/v1/Managers/iDRAC.Embedded.1",
		// Supermicros
		"/redfish/v1/Managers/1",
		// OpenBMC on ASRock
		"/redfish/v1/Managers/bmc",
	}
)

// inventory struct wraps redfish connection
type inventory struct {
	client            *redfishwrapper.Client
	failOnError       bool
	softwareInventory []*gofishrf.SoftwareInventory
}

func (c *Conn) Inventory(ctx context.Context) (device *common.Device, err error) {
	// initialize inventory object
	// the redfish client is assigned here to perform redfish Get/Delete requests
	inv := &inventory{client: c.redfishwrapper, failOnError: c.failInventoryOnError}

	updateService, err := c.redfishwrapper.UpdateService()
	if err != nil && inv.failOnError {
		return nil, errors.Wrap(bmclibErrs.ErrRedfishSoftwareInventory, err.Error())
	}

	if updateService != nil {
		inv.softwareInventory, err = updateService.FirmwareInventories()
		if err != nil && inv.failOnError {
			return nil, errors.Wrap(bmclibErrs.ErrRedfishSoftwareInventory, err.Error())
		}
	}

	// initialize device to be populated with inventory
	newDevice := common.NewDevice()
	device = &newDevice

	// populate device Chassis components attributes
	err = inv.chassisAttributes(ctx, device)
	if err != nil && inv.failOnError {
		return nil, err
	}

	// populate device System components attributes
	err = inv.systemAttributes(ctx, device)
	if err != nil && inv.failOnError {
		return nil, err
	}

	// populate device BMC component attributes
	err = inv.bmcAttributes(ctx, device)
	if err != nil && inv.failOnError {
		return nil, err
	}

	return device, nil
}

// DeviceVendorModel returns the device vendor and model attributes

// bmcAttributes collects BMC component attributes
func (i *inventory) bmcAttributes(ctx context.Context, device *common.Device) (err error) {
	managers, err := i.client.Managers(ctx)
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

		device.BMC = &common.BMC{
			Common: common.Common{
				Description: manager.Description,
				Vendor:      device.Vendor,
				Model:       device.Model,
				Status: &common.Status{
					Health: string(manager.Status.Health),
					State:  string(manager.Status.State),
				},
				Firmware: &common.Firmware{
					Installed: manager.FirmwareVersion,
				},
			},

			ID: manager.ID,
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
func (i *inventory) chassisAttributes(ctx context.Context, device *common.Device) (err error) {
	chassis, err := i.client.Chassis(ctx)
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
		if err != nil && i.failOnError {
			return err
		}

		err = i.collectPSUs(ch, device)
		if err != nil && i.failOnError {
			return err
		}

	}

	err = i.collectCPLDs(device)
	if err != nil && i.failOnError {
		return err
	}

	if compatible == 0 {
		return bmclibErrs.ErrRedfishChassisOdataID
	}

	return nil

}

func (i *inventory) systemAttributes(ctx context.Context, device *common.Device) (err error) {
	systems, err := i.client.Systems()
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
		funcs := []func(sys *gofishrf.ComputerSystem, device *common.Device) error{
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
			if err != nil && i.failOnError {
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
func (i *inventory) firmwareAttributes(slug, id string, firmwareObj *common.Firmware) {
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
					firmwareObj = &common.Firmware{}
				}

				if firmwareObj.Installed == inv.Version {
					continue
				}

				firmwareObj.Previous = append(firmwareObj.Previous, &common.Firmware{
					Installed:  inv.Version,
					SoftwareID: inv.SoftwareID,
				})
			}
		}

		// update firmwareObj with installed firmware attributes
		if strings.HasPrefix(inv.ID, "Installed") {
			if strings.Contains(inv.ID, id) || strings.EqualFold(slug, inv.Name) {

				if firmwareObj == nil {
					firmwareObj = &common.Firmware{}
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
