package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// ChassisFan is a single chassis fan reading.
type ChassisFan struct {
	// Name is the fan name, e.g. "Fan 1 Tach".
	Name string
	// PhysicalContext is the area the fan cools, e.g. "SystemBoard" or "CPU".
	PhysicalContext string
	// SpeedRPM is the fan speed in revolutions per minute; nil when not reported.
	SpeedRPM *float64
	// SpeedPercent is the fan speed as a percentage of its maximum; nil when not reported.
	SpeedPercent *float64
	// Health is the fan health, e.g. "OK", "Warning" or "Critical".
	Health string
	// State is the fan state, e.g. "Enabled" or "Absent".
	State string
}

// ChassisTemperature is a single chassis temperature sensor reading.
type ChassisTemperature struct {
	// Name is the sensor name, e.g. "Ambient Temp".
	Name string
	// PhysicalContext is the area the sensor measures, e.g. "Intake" or "CPU".
	PhysicalContext string
	// ReadingCelsius is the current temperature; nil when the sensor reports no reading.
	ReadingCelsius *float64
	// Health is the sensor health, e.g. "OK", "Warning" or "Critical".
	// Empty when the BMC reports temperatures through ThermalMetrics, which carries readings only.
	Health string
	// State is the sensor state, e.g. "Enabled" or "Absent".
	// Empty when the BMC reports temperatures through ThermalMetrics, which carries readings only.
	State string
}

// ChassisFanSpeedsGetter provides retrieval of the chassis fan readings.
type ChassisFanSpeedsGetter interface {
	GetChassisFanSpeeds(ctx context.Context) (fans []ChassisFan, err error)
}

// ChassisTemperaturesGetter provides retrieval of the chassis temperature sensor readings.
type ChassisTemperaturesGetter interface {
	GetChassisTemperatures(ctx context.Context) (temperatures []ChassisTemperature, err error)
}

type chassisFanSpeedsGetterProvider struct {
	name string
	ChassisFanSpeedsGetter
}

type chassisTemperaturesGetterProvider struct {
	name string
	ChassisTemperaturesGetter
}

func getChassisFanSpeeds(ctx context.Context, generic []chassisFanSpeedsGetterProvider) (fans []ChassisFan, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ChassisFanSpeedsGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			fans, vErr := elem.GetChassisFanSpeeds(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return fans, metadata, nil
		}
	}

	return nil, metadata, multierror.Append(err, errors.New("failure to get chassis fan speeds"))
}

func getChassisTemperatures(ctx context.Context, generic []chassisTemperaturesGetterProvider) (temperatures []ChassisTemperature, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ChassisTemperaturesGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			temperatures, vErr := elem.GetChassisTemperatures(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return temperatures, metadata, nil
		}
	}

	return nil, metadata, multierror.Append(err, errors.New("failure to get chassis temperatures"))
}

// GetChassisFanSpeedsFromInterfaces identifies implementations of the ChassisFanSpeedsGetter
// interface and passes them to the getChassisFanSpeeds() wrapper.
func GetChassisFanSpeedsFromInterfaces(ctx context.Context, generic []interface{}) (fans []ChassisFan, metadata Metadata, err error) {
	implementations := make([]chassisFanSpeedsGetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := chassisFanSpeedsGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ChassisFanSpeedsGetter:
			temp.ChassisFanSpeedsGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ChassisFanSpeedsGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return nil, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ChassisFanSpeedsGetter implementations found",
			),
		)
	}

	return getChassisFanSpeeds(ctx, implementations)
}

// GetChassisTemperaturesFromInterfaces identifies implementations of the ChassisTemperaturesGetter
// interface and passes them to the getChassisTemperatures() wrapper.
func GetChassisTemperaturesFromInterfaces(ctx context.Context, generic []interface{}) (temperatures []ChassisTemperature, metadata Metadata, err error) {
	implementations := make([]chassisTemperaturesGetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := chassisTemperaturesGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ChassisTemperaturesGetter:
			temp.ChassisTemperaturesGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ChassisTemperaturesGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return nil, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ChassisTemperaturesGetter implementations found",
			),
		)
	}

	return getChassisTemperatures(ctx, implementations)
}
