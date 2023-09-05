package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// SELServices for various SEL related services
type SELService interface {
	ClearSEL(ctx context.Context) (err error)
}

type selProviders struct {
	name        string
	selProvider SELService
}

func clearSEL(ctx context.Context, timeout time.Duration, s []selProviders) (metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range s {
		if elem.selProvider == nil {
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
			selErr := elem.selProvider.ClearSEL(ctx)
			if selErr != nil {
				err = multierror.Append(err, errors.WithMessagef(selErr, "provider: %v", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return metadataLocal, nil
		}

	}

	return metadataLocal, multierror.Append(err, errors.New("failed to reset SEL"))
}

func ClearSELFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (metadata Metadata, err error) {
	selServices := make([]selProviders, 0)
	for _, elem := range generic {
		temp := selProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SELService:
			temp.selProvider = p
			selServices = append(selServices, temp)
		default:
			e := fmt.Sprintf("not a SELService implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(selServices) == 0 {
		return metadata, multierror.Append(err, errors.New("no SelService implementations found"))
	}
	return clearSEL(ctx, timeout, selServices)
}
