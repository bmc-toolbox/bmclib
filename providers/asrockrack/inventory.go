package asrockrack

import (
	"context"

	"github.com/bmc-toolbox/bmclib/devices"
)

func (a *ASRockRack) Inventory(ctx context.Context) (device *devices.Device, err error) {
	// initialize device to be populated with inventory
	device = devices.NewDevice()
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
	err = a.systemHealth(ctx, device)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// systemHealth collects system health information based on the sensors data
func (a *ASRockRack) systemHealth(ctx context.Context, device *devices.Device) error {
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
func (a *ASRockRack) fruAttributes(ctx context.Context, device *devices.Device) error {
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
			device.Enclosures = append(device.Enclosures, &devices.Enclosure{
				Serial:      component.SerialNumber,
				Description: component.Type,
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
func (a *ASRockRack) systemAttributes(ctx context.Context, device *devices.Device) error {
	fwInfo, err := a.firmwareInfo(ctx)
	if err != nil {
		return err
	}

	device.BIOS = &devices.BIOS{
		Vendor:   device.Vendor,
		Model:    device.Model,
		Firmware: &devices.Firmware{Installed: fwInfo.BIOSVersion},
	}

	device.BMC = &devices.BMC{
		Vendor:   device.Vendor,
		Model:    device.Model,
		Firmware: &devices.Firmware{Installed: fwInfo.BMCVersion},
	}

	if fwInfo.CPLDVersion != "N/A" {
		device.CPLDs = append(device.CPLDs, &devices.CPLD{
			Vendor:   device.Vendor,
			Model:    device.Model,
			Firmware: &devices.Firmware{Installed: fwInfo.CPLDVersion},
		})
	}

	device.Metadata["node_id"] = fwInfo.NodeID

	components, err := a.inventoryInfo(ctx)
	if err != nil {
		return err
	}

	for _, component := range components {
		switch component.DeviceType {
		case "CPU":
			device.CPUs = append(device.CPUs,
				&devices.CPU{
					Vendor: component.ProductManufacturerName,
					Model:  component.ProductName,
					Firmware: &devices.Firmware{
						Installed: fwInfo.MicrocodeVersion,
						Metadata: map[string]string{
							"Intel_ME_version": fwInfo.MEVersion,
						},
					},
				},
			)
		case "Memory":
			device.Memory = append(device.Memory,
				&devices.Memory{
					Vendor:      component.ProductManufacturerName,
					Serial:      component.ProductSerialNumber,
					PartNumber:  component.ProductPartNumber,
					Type:        component.DeviceName,
					Description: component.ProductExtra,
				},
			)

		case "Storage device":
			var vendor string

			if component.ProductManufacturerName == "N/A" &&
				component.ProductPartNumber != "N/A" {
				vendor = devices.VendorFromProductName(component.ProductPartNumber)
			}

			device.Drives = append(device.Drives,
				&devices.Drive{
					Vendor:      vendor,
					Serial:      component.ProductSerialNumber,
					ProductName: component.ProductPartNumber,
				},
			)
		}

	}

	return nil
}
