package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/hashicorp/go-multierror"
)

// PowerSensorsGetter retrieves power consumption values
type PowerSensorsGetter interface {
	PowerSensors(ctx context.Context) ([]*devices.PowerSensor, error)
}

// TemperatureSensorGetter retrieves temperature values
type TemperatureSensorsGetter interface {
	TemperatureSensors(ctx context.Context) ([]*devices.TemperatureSensor, error)
}

// FanSensorsGetter retrieves fan speed data
type FanSensorsGetter interface {
	FanSensors(ctx context.Context) ([]*devices.FanSensor, error)
}

// ChassisHealthGetter retrieves chassis health data
type ChassisHealthGetter interface {
	ChassisHealth(ctx context.Context) ([]*devices.ChassisHealth, error)
}

// PowerGetter interface implementation identifier and passthrough methods

// GetPowerSensors returns power draw data, trying all interface implementations passed in
func GetPowerSensors(ctx context.Context, p []PowerSensorsGetter) (power []*devices.PowerSensor, err error) {
Loop:
	for _, elem := range p {
		if elem == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			power, vErr := elem.PowerSensors(ctx)
			if vErr != nil {
				err = multierror.Append(err, vErr)
				continue
			}
			return power, nil
		}
	}

	return power, multierror.Append(err, errors.New("failed to get power sensor data"))
}

// GetPowerSensorsFromInterfaces identfies implementations of the PowerSensorGetter interface and acts as a pass through method
func GetPowerSensorsFromInterfaces(ctx context.Context, generic []interface{}) (power []*devices.PowerSensor, err error) {
	powerDrawGetter := make([]PowerSensorsGetter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case PowerSensorsGetter:
			powerDrawGetter = append(powerDrawGetter, p)
		default:
			e := fmt.Sprintf("not a PowerSensorGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(powerDrawGetter) == 0 {
		return power, multierror.Append(err, errors.New("no PowerSensorGetter implementations found"))
	}

	return GetPowerSensors(ctx, powerDrawGetter)
}

// TemperatureSensorsGetter interface identifier and passthrough methods

// GetTemperatureSensors returns temperature data, trying all interface implementations passed in
func GetTemperatureSensors(ctx context.Context, p []TemperatureSensorsGetter) (temps []*devices.TemperatureSensor, err error) {
Loop:
	for _, elem := range p {
		if elem == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			temps, vErr := elem.TemperatureSensors(ctx)
			if vErr != nil {
				err = multierror.Append(err, vErr)
				continue
			}
			return temps, nil
		}
	}

	return temps, multierror.Append(err, errors.New("failed to get temperature sensor data"))
}

// GetTemperatureSensorsFromInterfaces identfies implementations of the TemperatureGetter interface and acts as a pass through method
func GetTemperatureSensorsFromInterfaces(ctx context.Context, generic []interface{}) (temps []*devices.TemperatureSensor, err error) {
	gets := make([]TemperatureSensorsGetter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case TemperatureSensorsGetter:
			gets = append(gets, p)
		default:
			e := fmt.Sprintf("not a TemperatureGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(gets) == 0 {
		return temps, multierror.Append(err, errors.New("no TemperatureGetter implementations found"))
	}

	return GetTemperatureSensors(ctx, gets)
}

// FanSensorsGetter interface identifier and passthrough methods

// GetFanSensors returns fan speed data, trying all interface implementations passed in
func GetFanSensors(ctx context.Context, p []FanSensorsGetter) (fanSensors []*devices.FanSensor, err error) {
Loop:
	for _, elem := range p {
		if elem == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			fanSensors, vErr := elem.FanSensors(ctx)
			if vErr != nil {
				err = multierror.Append(err, vErr)
				continue
			}
			return fanSensors, nil
		}
	}

	return fanSensors, multierror.Append(err, errors.New("failed to get fan sensor data"))
}

// GetFanSensorsFromInterfaces identfies implementations of the FanSpeedGetter interface and acts as a pass through method
func GetFanSensorsFromInterfaces(ctx context.Context, generic []interface{}) (fanSensors []*devices.FanSensor, err error) {
	gets := make([]FanSensorsGetter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case FanSensorsGetter:
			gets = append(gets, p)
		default:
			e := fmt.Sprintf("not a FanSensorGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(gets) == 0 {
		return fanSensors, multierror.Append(err, errors.New("no FanSensorGetter implementations found"))
	}

	return GetFanSensors(ctx, gets)
}

// ChassisHealthGetter interface identifier and passthrough methods

// GetChassisHealth gets all chassis health data, trying all interface implementations passed in
func GetChassisHealth(ctx context.Context, p []ChassisHealthGetter) (health []*devices.ChassisHealth, err error) {
Loop:
	for _, elem := range p {
		if elem == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			health, vErr := elem.ChassisHealth(ctx)
			if vErr != nil {
				err = multierror.Append(err, vErr)
				continue
			}
			return health, nil
		}
	}

	return health, multierror.Append(err, errors.New("failed to get chassis health"))
}

// GetChassisHealthFromInterfaces identfies implementations of the ChassisHealthGetter interface and acts as a pass through method
func GetChassisHealthFromInterfaces(ctx context.Context, generic []interface{}) (health []*devices.ChassisHealth, err error) {
	gets := make([]ChassisHealthGetter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case ChassisHealthGetter:
			gets = append(gets, p)
		default:
			e := fmt.Sprintf("not a ChassisHealthGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(gets) == 0 {
		return health, multierror.Append(err, errors.New("no FanSensorGetter implementations found"))
	}

	return GetChassisHealth(ctx, gets)
}
