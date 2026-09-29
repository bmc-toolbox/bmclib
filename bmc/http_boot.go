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

// HTTPBootTLSMode identifies the TLS authentication mode UEFI HTTP Boot uses to connect to the
// HTTP boot server.
type HTTPBootTLSMode string

// HTTPBootTLSMode constants enumerate the supported UEFI HTTP Boot TLS authentication modes.
const (
	// HTTPBootTLSModeNone means neither the HTTP boot server nor the client authenticate each
	// other, allowing UEFI HTTP Boot to fetch a boot image over plain HTTP.
	HTTPBootTLSModeNone HTTPBootTLSMode = "None"
	// HTTPBootTLSModeOneWay means the HTTP boot server is authenticated by the client, requiring
	// UEFI HTTP Boot to fetch a boot image over HTTPS.
	HTTPBootTLSModeOneWay HTTPBootTLSMode = "OneWay"
)

// HTTPBootTLSModeSetter sets the TLS authentication mode UEFI HTTP Boot uses to connect to the
// HTTP boot server.
//
// Firmware that defaults to HTTPBootTLSModeOneWay refuses to fetch a boot image over plain HTTP,
// so a caller pointing HTTP Boot at an unencrypted endpoint (see HTTPBootURISetter) has to be
// able to relax this. Independent of both the HTTP Boot enable/disable toggle and the boot image
// URI: setting the mode does neither.
type HTTPBootTLSModeSetter interface {
	SetHTTPBootTLSMode(ctx context.Context, mode HTTPBootTLSMode) (ok bool, err error)
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

// httpBootTLSModeProviders is an internal struct to correlate an implementation/provider and its name
type httpBootTLSModeProviders struct {
	name                  string
	httpBootTLSModeSetter HTTPBootTLSModeSetter
}

// setHTTPBootTLSMode sets the HTTP Boot TLS authentication mode.
func setHTTPBootTLSMode(ctx context.Context, mode HTTPBootTLSMode, b []httpBootTLSModeProviders) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range b {
		if elem.httpBootTLSModeSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ok, setErr := elem.httpBootTLSModeSetter.SetHTTPBootTLSMode(ctx, mode)
			if setErr != nil {
				err = multierror.Append(err, errors.WithMessagef(setErr, "provider: %v", elem.name))
				continue
			}
			if !ok {
				err = multierror.Append(err, fmt.Errorf("provider: %v, failed to set http boot tls mode", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to set http boot tls mode"))
}

// SetHTTPBootTLSModeFromInterfaces identifies implementations of the HTTPBootTLSModeSetter interface and passes the found implementations to the setHTTPBootTLSMode() wrapper
func SetHTTPBootTLSModeFromInterfaces(ctx context.Context, mode HTTPBootTLSMode, generic []interface{}) (ok bool, metadata Metadata, err error) {
	setters := make([]httpBootTLSModeProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := httpBootTLSModeProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case HTTPBootTLSModeSetter:
			temp.httpBootTLSModeSetter = p
			setters = append(setters, temp)
		default:
			e := fmt.Sprintf("not a HTTPBootTLSModeSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(setters) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no HTTPBootTLSModeSetter implementations found"))
	}
	return setHTTPBootTLSMode(ctx, mode, setters)
}
