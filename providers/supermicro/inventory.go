package supermicro

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/common"
)

// Inventory returns hardware and firmware inventory
func (s *Supermicro) Inventory(ctx context.Context) (device *common.Device, err error) {
	// initialize device to be populated with inventory
	newDevice := common.NewDevice()
	device = &newDevice
	device.Status = &common.Status{}

	device.Metadata = map[string]string{}

	// populate device BMC, BIOS component attributes
	err = s.fruAttributes(ctx, device)
	if err != nil {
		return nil, err
	}

	// populate device System components attributes
	err = s.systemAttributes(ctx, device)
	if err != nil {
		return nil, err
	}

	//	// populate device health based on sensor readings
	//	err = s.systemHealth(ctx, device)
	//	if err != nil {
	//		return nil, err
	//	}

	return device, nil
}

func (s *Supermicro) systemAttributes(ctx context.Context, device *common.Device) (err error) {
	err = s.platformInfo(ctx, device)
	if err != nil {
		return err
	}

	err = s.smbiosInfo(ctx, device)
	if err != nil {
		return err
	}

	return nil
}

func (s *Supermicro) smbiosInfo(ctx context.Context, device *common.Device) (err error) {
	data, err := s.queryCGI(ctx, "op=SMBIOS_INFO.XML&r=(0,0)")
	if err != nil {
		return err
	}

	if data == nil || data.SmBiosInfo == nil {
		return fmt.Errorf("got empty SMBIOS info attributes")
	}

	if data.Dimm != nil {
		for _, dimm := range data.Dimm {
			device.Memory = append(device.Memory,
				&common.Memory{
					Common: common.Common{
						Vendor: dimm.Manufacturer,
						Serial: dimm.Serial,
					},
					PartNumber:   dimm.ProductNumber,
					SizeBytes:    sizeBytes(dimm.SizeMb),
					ClockSpeedHz: speedHertz(dimm.SpeedMhz),
					Slot:         dimm.Location,
				},
			)
		}
	}

	if data.CPU != nil {
		for _, cpu := range data.CPU {
			device.CPUs = append(device.CPUs,
				&common.CPU{
					Common: common.Common{
						Vendor:      cpu.Manufacturer,
						Description: cpu.Version,
					},
					ClockSpeedHz: speedHertz(cpu.SpeedMhz),
					Cores:        stringToInt(cpu.Core),
				},
			)
		}
	}

	if data.PowerSupply != nil {
		for _, psu := range data.PowerSupply {
			device.PSUs = append(device.PSUs,
				&common.PSU{
					Common: common.Common{
						Vendor:      psu.Manufacturer,
						Serial:      psu.Serial,
						Description: psu.Name,
						Status: &common.Status{
							State: psu.Status,
						},
						Metadata: map[string]string{
							"Present":   psu.Present,
							"IVRS":      psu.Ivrs,
							"Group":     psu.Group,
							"Rev":       psu.Rev,
							"Unplugged": psu.Unplugged,
							"Type":      psu.Type,
						},
					},
					PowerCapacityWatts: capacityWatts(psu.MaxPower),
					ID:                 psu.Location,
				},
			)
		}
	}

	return nil
}

func capacityWatts(c string) int64 {
	wattParts := strings.Split(c, " Watts")
	if len(wattParts) == 0 {
		return 0
	}

	watts, err := strconv.Atoi(wattParts[0])
	if err != nil {
		return 0
	}

	return int64(watts)
}

func stringToInt(s string) int {
	i, _ := strconv.Atoi(s)

	return i
}

func sizeBytes(size string) int64 {
	parts := strings.Split(size, " MiB")
	if len(parts) == 0 {
		return 0
	}

	sizeMB, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	return int64(sizeMB * 1024)
}

func speedHertz(speed string) int64 {
	parts := strings.Split(speed, " Mhz")
	if len(parts) == 0 {
		return 0
	}

	speedHz, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0
	}

	return int64(speedHz * 1000)
}

func (s *Supermicro) platformInfo(ctx context.Context, device *common.Device) (err error) {
	data, err := s.queryCGI(ctx, "op=Get_PlatformInfo.XML&r=(0,0)")
	if err != nil {
		return err
	}

	if data == nil || data.PlatformInfo == nil {
		return fmt.Errorf("got empty Platform info attributes")
	}

	if data.PlatformInfo.BiosVersion != "" {
		device.BIOS = &common.BIOS{
			Common: common.Common{
				Firmware: &common.Firmware{
					Installed: data.PlatformInfo.BiosVersion,
				},
			},
		}
	}

	if data.PlatformInfo.CpldRev != "" {
		device.CPLDs = append(device.CPLDs,
			&common.CPLD{
				Common: common.Common{
					Firmware: &common.Firmware{
						Installed: data.PlatformInfo.CpldRev,
					},
				},
			},
		)
	}

	return nil
}

// fruAttributes looks up the service FRU attributes
func (s *Supermicro) fruAttributes(ctx context.Context, device *common.Device) (err error) {
	data, err := s.queryCGI(ctx, "op=FRU_INFO.XML&r=(0,0)")
	if err != nil {
		return err
	}

	if data == nil || data.FruInfo == nil {
		return fmt.Errorf("got empty FRU attributes")
	}

	if data.FruInfo.Board != nil {
		device.Mainboard = &common.Mainboard{
			Common: common.Common{
				Vendor:      data.FruInfo.Board.MfcName,
				Model:       data.FruInfo.Board.PartNum,
				ProductName: data.FruInfo.Board.ProdName,
				Serial:      data.FruInfo.Board.SerialNum,
			},
		}
	}

	if data.FruInfo.Chassis != nil {
		device.Enclosures = append(device.Enclosures,
			&common.Enclosure{
				Common: common.Common{
					Serial: data.FruInfo.Chassis.SerialNum,
				},
			},
		)

		device.Metadata["chassis.part_number"] = data.FruInfo.Chassis.PartNum
	}

	if data.FruInfo.Product != nil {
		device.Metadata["product.serial_number"] = data.FruInfo.Product.SerialNum
	}

	return nil
}
