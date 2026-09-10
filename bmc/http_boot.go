package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// HTTPBootURISetter sets the URI UEFI HTTP Boot fetches its boot image from.
type HTTPBootURISetter interface {
	SetHTTPBootURI(ctx context.Context, uri string) (ok bool, err error)
}

// httpBootURIProviders is an internal struct to correlate an implementation/provider and its name
type httpBootURIProviders struct {
	name              string
	httpBootURISetter HTTPBootURISetter
}

// setHTTPBootURI sets the HTTP Boot URI.
func setHTTPBootURI(ctx context.Context, uri string, b []httpBootURIProviders) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range b {
		if elem.httpBootURISetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ok, setErr := elem.httpBootURISetter.SetHTTPBootURI(ctx, uri)
			if setErr != nil {
				err = multierror.Append(err, errors.WithMessagef(setErr, "provider: %v", elem.name))
				continue
			}
			if !ok {
				err = multierror.Append(err, fmt.Errorf("provider: %v, failed to set http boot uri", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to set http boot uri"))
}

// SetHTTPBootURIFromInterfaces identifies implementations of the HTTPBootURISetter interface and passes the found implementations to the setHTTPBootURI() wrapper
func SetHTTPBootURIFromInterfaces(ctx context.Context, uri string, generic []interface{}) (ok bool, metadata Metadata, err error) {
	setters := make([]httpBootURIProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := httpBootURIProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case HTTPBootURISetter:
			temp.httpBootURISetter = p
			setters = append(setters, temp)
		default:
			e := fmt.Sprintf("not a HTTPBootURISetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(setters) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no HTTPBootURISetter implementations found"))
	}
	return setHTTPBootURI(ctx, uri, setters)
}
