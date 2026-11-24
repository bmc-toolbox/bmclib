package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// VirtualMediaSetter controls the virtual media attached to a machine
type VirtualMediaSetter interface {
	SetVirtualMedia(ctx context.Context, kind, mediaURL string) (ok bool, err error)
}

// VirtualMediaEjecter ejects the virtual media attached to a machine
type VirtualMediaEjecter interface {
	EjectVirtualMedia(ctx context.Context, kind string) (ok bool, err error)
}

// VirtualMediaProviders is an internal struct to correlate an implementation/provider and its name
type virtualMediaProviders struct {
	name               string
	virtualMediaSetter VirtualMediaSetter
}

// setVirtualMedia sets the virtual media.
func setVirtualMedia(ctx context.Context, kind, mediaURL string, b []virtualMediaProviders) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range b {
		if elem.virtualMediaSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ok, setErr := elem.virtualMediaSetter.SetVirtualMedia(ctx, kind, mediaURL)
			if setErr != nil {
				err = multierror.Append(err, errors.WithMessagef(setErr, "provider: %v", elem.name))
				continue
			}
			if !ok {
				err = multierror.Append(err, fmt.Errorf("provider: %v, failed to set virtual media", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to set virtual media"))
}

// SetVirtualMediaFromInterfaces identifies implementations of the virtualMediaSetter interface and passes the found implementations to the setVirtualMedia() wrapper
func SetVirtualMediaFromInterfaces(ctx context.Context, kind, mediaURL string, generic []interface{}) (ok bool, metadata Metadata, err error) {
	bdSetters := make([]virtualMediaProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := virtualMediaProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case VirtualMediaSetter:
			temp.virtualMediaSetter = p
			bdSetters = append(bdSetters, temp)
		default:
			e := fmt.Sprintf("not a VirtualMediaSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(bdSetters) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no VirtualMediaSetter implementations found"))
	}
	return setVirtualMedia(ctx, kind, mediaURL, bdSetters)
}

// virtualMediaEjecterProvider is an internal struct to correlate an implementation/provider and its name
type virtualMediaEjecterProvider struct {
	name                string
	virtualMediaEjecter VirtualMediaEjecter
}

// ejectVirtualMedia ejects the virtual media.
func ejectVirtualMedia(ctx context.Context, kind string, b []virtualMediaEjecterProvider) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range b {
		if elem.virtualMediaEjecter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ok, ejectErr := elem.virtualMediaEjecter.EjectVirtualMedia(ctx, kind)
			if ejectErr != nil {
				err = multierror.Append(err, errors.WithMessagef(ejectErr, "provider: %v", elem.name))
				continue
			}
			if !ok {
				err = multierror.Append(err, fmt.Errorf("provider: %v, failed to eject virtual media", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to eject virtual media"))
}

// EjectVirtualMediaFromInterfaces identifies implementations of the VirtualMediaEjecter interface and passes the found implementations to the ejectVirtualMedia() wrapper
func EjectVirtualMediaFromInterfaces(ctx context.Context, kind string, generic []interface{}) (ok bool, metadata Metadata, err error) {
	ejecters := make([]virtualMediaEjecterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := virtualMediaEjecterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case VirtualMediaEjecter:
			temp.virtualMediaEjecter = p
			ejecters = append(ejecters, temp)
		default:
			e := fmt.Sprintf("not a VirtualMediaEjecter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(ejecters) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no VirtualMediaEjecter implementations found"))
	}
	return ejectVirtualMedia(ctx, kind, ejecters)
}
