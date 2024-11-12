package asrockrack

import (
	"context"

	common "github.com/metal-toolbox/bmc-common"
)

// Inventory returns hardware and firmware inventory
func (a *ASRockRack) Inventory(ctx context.Context) (device *common.Device, err error) {
	// initialize device to be populated with inventory
	newDevice := common.NewDevice()
	device = &newDevice
	device.Status = &common.Status{}

	device.Metadata = map[string]string{}

	// populate device BMC, BIOS component attributes
	err = a.fruAttributes(ctx, device)
	if err != nil {
		return nil, err
	}

	// populate device System components attributes
	err = a.systemAttributes(ctx, device)
	if err != nil {
		return nil, err
	}

	// populate device health based on sensor readings
	//
	// sensor data collection can fail for a myriad of reasons
	// we log the error and keep going
	err = a.systemHealth(ctx, device)
	if err != nil {
		a.log.V(2).Error(err, "sensor data collection error", "deviceModel", a.deviceModel)
	}

	return device, nil
}

// systemHealth collects system health information based on the sensors data
func (a *ASRockRack) systemHealth(ctx context.Context, device *common.Device) error {
	sensors, err := a.sensors(ctx)
	if err != nil {
		return err
	}

	ok := true
	device.Status.Health = "OK"
	for _, sensor := range sensors {
		switch sensor.Name {
		case "CPU_CATERR", "CPU_THERMTRIP", "CPU_PROCHOT":
			if sensor.SensorState != 0 {
				ok = false
				device.Status.State = sensor.Name
				break
			}
		default:
			if sensor.SensorState != 1 {
				ok = false
				device.Status.State = sensor.Name
				break
			}
		}
	}

	if !ok {
		device.Status.Health = "CRITICAL"
	}

	// we don't want to fail inventory collection hence ignore POST code collection error
	device.Status.PostCodeStatus, device.Status.PostCode, _ = a.PostCode(ctx)

	return nil
}

// fruAttributes collects chassis information
func (a *ASRockRack) fruAttributes(ctx context.Context, device *common.Device) error {
	components, err := a.fruInfo(ctx)
	if err != nil {
		return err
	}

	for _, component := range components {
		switch component.Component {
		case "board":
			device.Vendor = component.Manufacturer
			device.Model = component.ProductName
			device.Serial = component.SerialNumber
		case "chassis":
			device.Enclosures = append(device.Enclosures, &common.Enclosure{
				Common: common.Common{
					Serial:      component.SerialNumber,
					Description: component.Type,
				},
			})
		case "product":
			device.Metadata["product.manufacturer"] = component.Manufacturer
			device.Metadata["product.name"] = component.ProductName
			device.Metadata["product.part_number"] = component.PartNumber
			device.Metadata["product.version"] = component.ProductVersion
			device.Metadata["product.serialnumber"] = component.SerialNumber
		}
	}

	return nil
}

// systemAttributes collects system component attributes
func (a *ASRockRack) systemAttributes(ctx context.Context, device *common.Device) error {
	fwInfo, err := a.firmwareInfo(ctx)
	if err != nil {
		return err
	}

	device.BIOS = &common.BIOS{
		Common: common.Common{
			Vendor:   device.Vendor,
			Model:    device.Model,
			Firmware: &common.Firmware{Installed: fwInfo.BIOSVersion},
		},
	}

	device.BMC = &common.BMC{
		Common: common.Common{
			Vendor:   device.Vendor,
			Model:    device.Model,
			Firmware: &common.Firmware{Installed: fwInfo.BMCVersion},
		},
	}

	if fwInfo.CPLDVersion != "N/A" {
		device.CPLDs = append(device.CPLDs, &common.CPLD{
			Common: common.Common{
				Vendor:   device.Vendor,
				Model:    device.Model,
				Firmware: &common.Firmware{Installed: fwInfo.CPLDVersion},
			},
		})
	}

	device.Metadata["node_id"] = fwInfo.NodeID

	switch device.Model {
	case E3C246D4ID_NL, E3C246D4I_NL:
		return a.componentAttributesE3C246(ctx, fwInfo, device)
	default:
		return nil
	}
}

func (a *ASRockRack) componentAttributesE3C246(ctx context.Context, fwInfo *firmwareInfo, device *common.Device) error {
	// TODO: implement newer device inventory
	components, err := a.inventoryInfoE3C246D41D(ctx)
	if err != nil {
		return err
	}

	for _, component := range components {
		switch component.DeviceType {
		case "CPU":
			device.CPUs = append(device.CPUs,
				&common.CPU{
					Common: common.Common{
						Vendor: component.ProductManufacturerName,
						Model:  component.ProductName,
						Firmware: &common.Firmware{
							Installed: fwInfo.MicrocodeVersion,
							Metadata: map[string]string{
								"Intel_ME_version": fwInfo.MEVersion,
							},
						},
					},
				},
			)
		case "Memory":
			device.Memory = append(device.Memory,
				&common.Memory{
					Common: common.Common{
						Vendor:      component.ProductManufacturerName,
						Serial:      component.ProductSerialNumber,
						Description: component.ProductExtra,
					},

					PartNumber: component.ProductPartNumber,
					Type:       component.DeviceName,
				},
			)

		case "Storage device":
			var vendor string

			if component.ProductManufacturerName == "N/A" &&
				component.ProductPartNumber != "N/A" {
				vendor = common.FormatVendorName(component.ProductPartNumber)
			}

			device.Drives = append(device.Drives,
				&common.Drive{
					Common: common.Common{
						Vendor:      vendor,
						Serial:      component.ProductSerialNumber,
						ProductName: component.ProductPartNumber,
					},
				},
			)
		}

	}

	return nil
}
