package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// SystemEventLog provides access to a host's System Event Log (SEL) services.
type SystemEventLog interface {
	ClearSystemEventLog(ctx context.Context) (err error)
	GetSystemEventLog(ctx context.Context) (entries [][]string, err error)
	GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error)
}

type systemEventLogProviders struct {
	name                   string
	systemEventLogProvider SystemEventLog
}

// SystemEventLogEntries holds System Event Log entries as rows of string columns.
type SystemEventLogEntries [][]string

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
			selErr := elem.systemEventLogProvider.ClearSystemEventLog(ctx)
			cancel()
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

// ClearSystemEventLogFromInterfaces identifies implementations of the SystemEventLog interface and clears the System Event Log using the first successful provider.
func ClearSystemEventLogFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (metadata Metadata, err error) {
	selServices := make([]systemEventLogProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
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

func getSystemEventLog(ctx context.Context, timeout time.Duration, s []systemEventLogProviders) (sel SystemEventLogEntries, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range s {
		if elem.systemEventLogProvider == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return sel, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)

			sel, selErr := elem.systemEventLogProvider.GetSystemEventLog(ctx)
			cancel()
			if selErr != nil {
				err = multierror.Append(err, errors.WithMessagef(selErr, "provider: %v", elem.name))
				continue
			}

			metadataLocal.SuccessfulProvider = elem.name
			return sel, metadataLocal, nil
		}
	}

	return nil, metadataLocal, multierror.Append(err, errors.New("failed to get System Event Log"))
}

// GetSystemEventLogFromInterfaces identifies implementations of the SystemEventLog interface and returns the System Event Log entries from the first successful provider.
func GetSystemEventLogFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (sel SystemEventLogEntries, metadata Metadata, err error) {
	selServices := make([]systemEventLogProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
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
		return sel, metadata, multierror.Append(err, errors.New("no SystemEventLog implementations found"))
	}
	return getSystemEventLog(ctx, timeout, selServices)
}

func getSystemEventLogRaw(ctx context.Context, timeout time.Duration, s []systemEventLogProviders) (eventlog string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range s {
		if elem.systemEventLogProvider == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return eventlog, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)

			eventlog, selErr := elem.systemEventLogProvider.GetSystemEventLogRaw(ctx)
			cancel()
			if selErr != nil {
				err = multierror.Append(err, errors.WithMessagef(selErr, "provider: %v", elem.name))
				continue
			}

			metadataLocal.SuccessfulProvider = elem.name
			return eventlog, metadataLocal, nil
		}
	}

	return eventlog, metadataLocal, multierror.Append(err, errors.New("failed to get System Event Log"))
}

// GetSystemEventLogRawFromInterfaces identifies implementations of the SystemEventLog interface and returns the raw System Event Log from the first successful provider.
func GetSystemEventLogRawFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (eventlog string, metadata Metadata, err error) {
	selServices := make([]systemEventLogProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
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
		return eventlog, metadata, multierror.Append(err, errors.New("no SystemEventLog implementations found"))
	}
	return getSystemEventLogRaw(ctx, timeout, selServices)
}
