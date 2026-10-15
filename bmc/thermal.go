package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// ChassisThermal holds the chassis temperature sensor and fan readings.
type ChassisThermal struct {
	Temperatures []ChassisTemperature
	Fans         []ChassisFan
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
	Health string
	// State is the sensor state, e.g. "Enabled" or "Absent".
	State string
}

// ChassisFan is a single chassis fan reading.
type ChassisFan struct {
	// Name is the fan name, e.g. "Fan 1 Tach".
	Name string
	// Reading is the current fan reading in ReadingUnits; nil when the fan reports no reading.
	Reading *int
	// ReadingUnits is the unit of Reading, "RPM" or "Percent".
	ReadingUnits string
	// Health is the fan health, e.g. "OK", "Warning" or "Critical".
	Health string
	// State is the fan state, e.g. "Enabled" or "Absent".
	State string
}

// ChassisThermalGetter provides retrieval of the chassis temperature sensor and fan readings.
type ChassisThermalGetter interface {
	GetChassisThermal(ctx context.Context) (thermal ChassisThermal, err error)
}

type chassisThermalGetterProvider struct {
	name string
	ChassisThermalGetter
}

func getChassisThermal(ctx context.Context, generic []chassisThermalGetterProvider) (thermal ChassisThermal, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ChassisThermalGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			thermal, vErr := elem.GetChassisThermal(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return thermal, metadata, nil
		}
	}

	return thermal, metadata, multierror.Append(err, errors.New("failure to get chassis thermal"))
}

// GetChassisThermalFromInterfaces identifies implementations of the ChassisThermalGetter
// interface and passes them to the getChassisThermal() wrapper.
func GetChassisThermalFromInterfaces(ctx context.Context, generic []interface{}) (thermal ChassisThermal, metadata Metadata, err error) {
	implementations := make([]chassisThermalGetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := chassisThermalGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ChassisThermalGetter:
			temp.ChassisThermalGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ChassisThermalGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return thermal, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ChassisThermalGetter implementations found",
			),
		)
	}

	return getChassisThermal(ctx, implementations)
}
