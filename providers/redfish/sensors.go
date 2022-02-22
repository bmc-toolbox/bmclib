package redfish

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bmc-toolbox/bmclib/devices"
)

func chassisCompatible(chassisOdataID string) bool {
	for _, url := range chassisOdataIdURLs {
		if url == chassisOdataID {
			return true
		}
	}

	return false
}

func (c *Conn) PowerSensors(ctx context.Context) ([]*devices.PowerSensor, error) {

	data := make([]*devices.PowerSensor, 0)

	service := c.conn.Service
	if service == nil {
		return nil, ErrRedfishServiceNil
	}

	chassis, err := service.Chassis()
	if err != nil {
		return nil, err
	}

	compatible := 0
	for _, ch := range chassis {
		fmt.Println(ch.ODataID)
		if !chassisCompatible(ch.ODataID) {
			continue
		}

		compatible++
		p, err := ch.Power()
		if err != nil {
			return nil, err
		}

		for idx, supply := range p.PowerSupplies {
			id := p.ID

			if p.ID == "" {
				id = strconv.Itoa(idx)
			}

			c := &devices.PowerSensor{
				Name:            supply.Name,
				ID:              id,
				InputWatts:      supply.PowerInputWatts,
				OutputWatts:     supply.PowerOutputWatts,
				LastOutputWatts: supply.LastPowerOutputWatts,
			}
			data = append(data, c)
		}

	}

	if compatible == 0 {
		return nil, ErrRedfishChassisOdataID
	}

	return data, nil
}

func (c *Conn) TemperatureSensors(ctx context.Context) ([]*devices.TemperatureSensor, error) {

	data := make([]*devices.TemperatureSensor, 0)

	service := c.conn.Service
	if service == nil {
		return nil, ErrRedfishServiceNil
	}

	chassis, err := service.Chassis()
	if err != nil {
		return nil, err
	}

	compatible := 0
	for _, ch := range chassis {
		if !chassisCompatible(ch.ODataID) {
			continue
		}

		compatible++
		p, err := ch.Thermal()
		if err != nil {
			return nil, err
		}

		for idx, s := range p.Temperatures {
			id := s.ID

			if s.ID == "" {
				id = strconv.Itoa(idx)
			}

			t := &devices.TemperatureSensor{
				ID:              id,
				Name:            s.Name,
				PhysicalContext: s.PhysicalContext,
				ReadingCelsius:  s.ReadingCelsius,
			}

			data = append(data, t)
		}
	}

	if compatible == 0 {
		return nil, ErrRedfishChassisOdataID
	}

	return data, nil
}

func (c *Conn) FanSensors(ctx context.Context) ([]*devices.FanSensor, error) {
	data := make([]*devices.FanSensor, 0)

	service := c.conn.Service
	if service == nil {
		return nil, ErrRedfishServiceNil
	}

	chassis, err := service.Chassis()
	if err != nil {
		return nil, err
	}

	compatible := 0
	for _, ch := range chassis {
		if !chassisCompatible(ch.ODataID) {
			continue
		}

		compatible++
		p, err := ch.Thermal()
		if err != nil {
			return nil, err
		}

		for idx, s := range p.Fans {
			id := s.ID

			if s.ID == "" {
				id = strconv.Itoa(idx)
			}

			t := &devices.FanSensor{
				ID:              id,
				Name:            s.Name,
				PhysicalContext: s.PhysicalContext,
				Reading:         s.Reading,
			}

			data = append(data, t)
		}
	}

	if compatible == 0 {
		return nil, ErrRedfishChassisOdataID
	}

	return data, nil
}

func (c *Conn) ChassisHealth(ctx context.Context) ([]*devices.ChassisHealth, error) {

	data := make([]*devices.ChassisHealth, 0)

	service := c.conn.Service
	if service == nil {
		return nil, ErrRedfishServiceNil
	}

	chassis, err := service.Chassis()
	if err != nil {
		return nil, err
	}

	compatible := 0
	for _, ch := range chassis {
		if !chassisCompatible(ch.ODataID) {
			continue
		}

		compatible++
		h := &devices.ChassisHealth{
			ID:     ch.ID,
			Name:   ch.Name,
			State:  string(ch.Status.State),
			Health: string(ch.Status.Health),
		}
		data = append(data, h)
	}

	if compatible == 0 {
		return nil, ErrRedfishChassisOdataID
	}

	return data, nil
}
