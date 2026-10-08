package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// PowerMetrics holds the chassis power readings and the configured power cap.
type PowerMetrics struct {
	// ConsumedWatts is the power currently drawn by the chassis.
	ConsumedWatts float64
	// CapacityWatts is the total power capacity available to the chassis.
	CapacityWatts float64
	// LimitInWatts is the configured power cap; nil when no cap is set.
	LimitInWatts *float64
}

// PowerMetricsGetter provides retrieval of the chassis power readings and power cap.
type PowerMetricsGetter interface {
	GetPowerMetrics(ctx context.Context) (metrics PowerMetrics, err error)
}

type powerMetricsGetterProvider struct {
	name string
	PowerMetricsGetter
}

// PowerCapSetter provides setting the chassis power cap. A nil limitWatts clears
// the cap.
type PowerCapSetter interface {
	SetPowerCap(ctx context.Context, limitWatts *float64) (err error)
}

type powerCapSetterProvider struct {
	name string
	PowerCapSetter
}

func getPowerMetrics(ctx context.Context, generic []powerMetricsGetterProvider) (metrics PowerMetrics, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.PowerMetricsGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			metrics, vErr := elem.GetPowerMetrics(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metrics, metadata, nil
		}
	}

	return metrics, metadata, multierror.Append(err, errors.New("failure to get power metrics"))
}

func setPowerCap(ctx context.Context, generic []powerCapSetterProvider, limitWatts *float64) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.PowerCapSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.SetPowerCap(ctx, limitWatts)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to set power cap"))
}

// GetPowerMetricsFromInterfaces identifies implementations of the PowerMetricsGetter
// interface and passes them to the getPowerMetrics() wrapper.
func GetPowerMetricsFromInterfaces(ctx context.Context, generic []interface{}) (metrics PowerMetrics, metadata Metadata, err error) {
	implementations := make([]powerMetricsGetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := powerMetricsGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case PowerMetricsGetter:
			temp.PowerMetricsGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a PowerMetricsGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metrics, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no PowerMetricsGetter implementations found",
			),
		)
	}

	return getPowerMetrics(ctx, implementations)
}

// SetPowerCapFromInterfaces identifies implementations of the PowerCapSetter
// interface and passes them to the setPowerCap() wrapper.
func SetPowerCapFromInterfaces(ctx context.Context, generic []interface{}, limitWatts *float64) (metadata Metadata, err error) {
	implementations := make([]powerCapSetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := powerCapSetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case PowerCapSetter:
			temp.PowerCapSetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a PowerCapSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no PowerCapSetter implementations found",
			),
		)
	}

	return setPowerCap(ctx, implementations, limitWatts)
}
