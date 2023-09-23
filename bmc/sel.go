package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// System Event Log Services for related services
type SystemEventLog interface {
	ClearSystemEventLog(ctx context.Context) (err error)
}

type systemEventLogProviders struct {
	name                   string
	systemEventLogProvider SystemEventLog
}

func clearSystemEventLog(ctx context.Context, timeout time.Duration, s []systemEventLogProviders) (metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range s {
		if elem.systemEventLogProvider == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			selErr := elem.systemEventLogProvider.ClearSystemEventLog(ctx)
			if selErr != nil {
				err = multierror.Append(err, errors.WithMessagef(selErr, "provider: %v", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return metadataLocal, nil
		}

	}

	return metadataLocal, multierror.Append(err, errors.New("failed to reset System Event Log"))
}

func ClearSystemEventLogFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (metadata Metadata, err error) {
	selServices := make([]systemEventLogProviders, 0)
	for _, elem := range generic {
		temp := systemEventLogProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SystemEventLog:
			temp.systemEventLogProvider = p
			selServices = append(selServices, temp)
		default:
			e := fmt.Sprintf("not a SystemEventLog service implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(selServices) == 0 {
		return metadata, multierror.Append(err, errors.New("no SystemEventLog implementations found"))
	}
	return clearSystemEventLog(ctx, timeout, selServices)
}
