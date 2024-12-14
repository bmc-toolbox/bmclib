package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// SOLDeactivator for deactivating SOL sessions on a BMC.
type SOLDeactivator interface {
	DeactivateSOL(ctx context.Context) (err error)
}

// deactivatorProvider is an internal struct to correlate an implementation/provider and its name
type deactivatorProvider struct {
	name           string
	solDeactivator SOLDeactivator
}

// deactivateSOL tries all implementations for a successful SOL deactivation
func deactivateSOL(ctx context.Context, timeout time.Duration, b []deactivatorProvider) (metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range b {
		if elem.solDeactivator == nil {
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
			newErr := elem.solDeactivator.DeactivateSOL(ctx)
			if newErr != nil {
				err = multierror.Append(err, errors.WithMessagef(newErr, "provider: %v", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return metadataLocal, nil
		}
	}
	return metadataLocal, multierror.Append(err, errors.New("failed to deactivate SOL session"))
}

// DeactivateSOLFromInterfaces identifies implementations of the SOLDeactivator interface and passes them to the deactivateSOL() wrapper method.
func DeactivateSOLFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (metadata Metadata, err error) {
	deactivators := make([]deactivatorProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := deactivatorProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SOLDeactivator:
			temp.solDeactivator = p
			deactivators = append(deactivators, temp)
		default:
			e := fmt.Sprintf("not an SOLDeactivator implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(deactivators) == 0 {
		return metadata, multierror.Append(err, errors.New("no SOLDeactivator implementations found"))
	}
	return deactivateSOL(ctx, timeout, deactivators)
}
