package redfishwrapper

import (
	"context"
	"strings"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/pkg/errors"

	common "github.com/metal-toolbox/bmc-common"
	redfish "github.com/stmcginnis/gofish/redfish"
)

var (
	// Supported Chassis Odata IDs
	KnownChassisOdataIDs = []string{
		// Dells
		"/redfish/v1/Chassis/Enclosure.Internal.0-1",
		"/redfish/v1/Chassis/System.Embedded.1",
		"/redfish/v1/Chassis/Enclosure.Internal.0-1:NonRAID.Integrated.1-1",
		// Supermicro
		"/redfish/v1/Chassis/1",
		// MegaRAC/ARockRack
		"/redfish/v1/Chassis/Self",
		// OpenBMC on ASRock
		"/redfish/v1/Chassis/ASRock_ROMED8HM3",
	}

	// Supported System Odata IDs
	knownSystemsOdataIDs = []string{
		// Dells
		"/redfish/v1/Systems/System.Embedded.1",
		"/redfish/v1/Systems/System.Embedded.1/Bios",
		// Supermicros
		"/redfish/v1/Systems/1",
		// MegaRAC/ARockRack
		"/redfish/v1/Systems/Self",
		// OpenBMC on ASRock
		"/redfish/v1/Systems/system",
	}

	// Supported Manager Odata IDs (BMCs)
	managerOdataIDs = []string{
		// Dells
		"/redfish/v1/Managers/iDRAC.Embedded.1",
		// Supermicros
		"/redfish/v1/Managers/1",
		// MegaRAC/ARockRack
		"/redfish/v1/Managers/Self",
		// OpenBMC on ASRock
		"/redfish/v1/Managers/bmc",
	}
)

// TODO: consider removing this
func (c *Client) compatibleOdataID(OdataID string, knownOdataIDs []string) bool {
	for _, url := range knownOdataIDs {
		if url == OdataID {
			return true
		}
	}

	return false
}

func (c *Client) Inventory(ctx context.Context, failOnError bool) (device *common.Device, err error) {
	updateService, err := c.UpdateService()
	if err != nil && failOnError {
		return nil, errors.Wrap(bmclibErrs.ErrRedfishSoftwareInventory, err.Error())
	}

	softwareInventory := []*redfish.SoftwareInventory{}

	if updateService != nil {
		// nolint
		softwareInventory, err = updateService.FirmwareInventories()
		if err != nil && failOnError {
			return nil, errors.Wrap(bmclibErrs.ErrRedfishSoftwareInventory, err.Error())
		}
	}

	// initialize device to be populated with inventory
	newDevice := common.NewDevice()
	device = &newDevice

	// populate device Chassis components attributes
	err = c.chassisAttributes(ctx, device, failOnError, softwareInventory)
	if err != nil && failOnError {
		return nil, err
	}

	// populate device System components attributes
	err = c.systemAttributes(device, failOnError, softwareInventory)
	if err != nil && failOnError {
		return nil, err
	}

	// populate device BMC component attributes
	err = c.bmcAttributes(ctx, device, softwareInventory)
	if err != nil && failOnError {
		return nil, err
	}

	return device, nil
}

// DeviceVendorModel returns the device vendor and model attributes

// bmcAttributes collects BMC component attributes
func (c *Client) bmcAttributes(ctx context.Context, device *common.Device, softwareInventory []*redfish.SoftwareInventory) (err error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return err
	}

	var compatible int
	for _, manager := range managers {
		if !c.compatibleOdataID(manager.ODataID, managerOdataIDs) {
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
		c.firmwareAttributes("", device.BMC.ID, device.BMC.Firmware, softwareInventory)
	}

	if compatible == 0 {
		return bmclibErrs.ErrRedfishManagerOdataID
	}

	return nil
}

// chassisAttributes populates the device chassis attributes
func (c *Client) chassisAttributes(ctx context.Context, device *common.Device, failOnError bool, softwareInventory []*redfish.SoftwareInventory) (err error) {
	chassis, err := c.Chassis(ctx)
	if err != nil {
		return err
	}

	compatible := 0
	for _, ch := range chassis {
		if !c.compatibleOdataID(ch.ODataID, KnownChassisOdataIDs) {
			continue
		}

		compatible++

		err = c.collectEnclosure(ch, device, softwareInventory)
		if err != nil && failOnError {
			return err
		}

		err = c.collectPSUs(ch, device, softwareInventory)
		if err != nil && failOnError {
			return err
		}

	}

	err = c.collectCPLDs(device, softwareInventory)
	if err != nil && failOnError {
		return err
	}

	if compatible == 0 {
		return bmclibErrs.ErrRedfishChassisOdataID
	}

	return nil

}

func (c *Client) systemAttributes(device *common.Device, failOnError bool, softwareInventory []*redfish.SoftwareInventory) (err error) {
	systems, err := c.Systems()
	if err != nil {
		return err
	}

	compatible := 0
	for _, sys := range systems {
		if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
			continue
		}

		compatible++

		if sys.Manufacturer != "" && sys.Model != "" && sys.SerialNumber != "" {
			device.Vendor = sys.Manufacturer
			device.Model = sys.Model
			device.Serial = sys.SerialNumber
		}

		type collectorFuncs []func(
			sys *redfish.ComputerSystem,
			device *common.Device,
			softwareInventory []*redfish.SoftwareInventory,
		) error

		// slice of collector methods
		funcs := collectorFuncs{
			c.collectCPUs,
			c.collectDIMMs,
			c.collectDrives,
			c.collectBIOS,
			c.collectNICs,
			c.collectTPMs,
			c.collectStorageControllers,
		}

		// execute collector methods
		for _, f := range funcs {
			err := f(sys, device, softwareInventory)
			if err != nil && failOnError {
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
func (c *Client) firmwareAttributes(slug, id string, firmwareObj *common.Firmware, softwareInventory []*redfish.SoftwareInventory) {
	if len(softwareInventory) == 0 {
		return
	}

	if id == "" {
		id = slug
	}

	for _, inv := range softwareInventory {
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
