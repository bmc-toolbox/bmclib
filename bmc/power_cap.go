package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// ChassisPowerMetrics holds the chassis power readings and the configured power cap.
type ChassisPowerMetrics struct {
	// ConsumedWatts is the power currently drawn by the chassis.
	ConsumedWatts float64
	// CapacityWatts is the total power capacity available to the chassis.
	CapacityWatts float64
	// LimitInWatts is the configured power cap; nil when no cap is set.
	LimitInWatts *float64
}

// ChassisPowerMetricsGetter provides retrieval of the chassis power readings and power cap.
type ChassisPowerMetricsGetter interface {
	GetChassisPowerMetrics(ctx context.Context) (metrics ChassisPowerMetrics, err error)
}

type chassisPowerMetricsGetterProvider struct {
	name string
	ChassisPowerMetricsGetter
}

// ChassisPowerCapSetter provides setting the chassis power cap. A nil limitWatts clears
// the cap.
type ChassisPowerCapSetter interface {
	SetChassisPowerCap(ctx context.Context, limitWatts *float64) (err error)
}

type chassisPowerCapSetterProvider struct {
	name string
	ChassisPowerCapSetter
}

func getChassisPowerMetrics(ctx context.Context, generic []chassisPowerMetricsGetterProvider) (metrics ChassisPowerMetrics, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ChassisPowerMetricsGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			metrics, vErr := elem.GetChassisPowerMetrics(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metrics, metadata, nil
		}
	}

	return metrics, metadata, multierror.Append(err, errors.New("failure to get chassis power metrics"))
}

func setChassisPowerCap(ctx context.Context, generic []chassisPowerCapSetterProvider, limitWatts *float64) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ChassisPowerCapSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.SetChassisPowerCap(ctx, limitWatts)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to set chassis power cap"))
}

// GetChassisPowerMetricsFromInterfaces identifies implementations of the ChassisPowerMetricsGetter
// interface and passes them to the getChassisPowerMetrics() wrapper.
func GetChassisPowerMetricsFromInterfaces(ctx context.Context, generic []interface{}) (metrics ChassisPowerMetrics, metadata Metadata, err error) {
	implementations := make([]chassisPowerMetricsGetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := chassisPowerMetricsGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ChassisPowerMetricsGetter:
			temp.ChassisPowerMetricsGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ChassisPowerMetricsGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metrics, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ChassisPowerMetricsGetter implementations found",
			),
		)
	}

	return getChassisPowerMetrics(ctx, implementations)
}

// SetChassisPowerCapFromInterfaces identifies implementations of the ChassisPowerCapSetter
// interface and passes them to the setChassisPowerCap() wrapper.
func SetChassisPowerCapFromInterfaces(ctx context.Context, generic []interface{}, limitWatts *float64) (metadata Metadata, err error) {
	implementations := make([]chassisPowerCapSetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := chassisPowerCapSetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ChassisPowerCapSetter:
			temp.ChassisPowerCapSetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ChassisPowerCapSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ChassisPowerCapSetter implementations found",
			),
		)
	}

	return setChassisPowerCap(ctx, implementations, limitWatts)
}
