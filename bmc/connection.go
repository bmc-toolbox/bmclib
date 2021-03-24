package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// Opener interface for opening a connection to a BMC
type Opener interface {
	Open(ctx context.Context) error
}

// Closer interface for closing a connection to a BMC
type Closer interface {
	Close(ctx context.Context) error
}

// connectionProviders is an internal struct to correlate an implementation/provider and its name
type connectionProviders struct {
	name   string
	closer Closer
}

// OpenConnectionFromInterfaces will try all opener interfaces and remove failed ones.
// The reason failed ones need to be removed is so that when other methods are called (like powerstate)
// implementations that have connections wont nil pointer error when their connection fails.
func OpenConnectionFromInterfaces(ctx context.Context, generic []interface{}, metadata ...*Metadata) (opened []interface{}, err error) {
	var metadataLocal Metadata
	defer func() {
		if len(metadata) > 0 && metadata[0] != nil {
			*metadata[0] = metadataLocal
		}
	}()
Loop:
	for _, elem := range generic {
		// return immediately if the context is done/terminated/etc
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
		}
		// get the provider name for use in metadata
		providerName := getProviderName(elem)
		// now, try to open connections
		switch p := elem.(type) {
		case Opener:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, providerName)
			er := p.Open(ctx)
			if er != nil {
				err = multierror.Append(err, er)
				continue
			}
			opened = append(opened, elem)
			metadataLocal.SuccessfulOpenConns = append(metadataLocal.SuccessfulOpenConns, providerName)
		default:
			e := fmt.Sprintf("not a Opener implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(opened) == 0 {
		return nil, multierror.Append(err, errors.New("no Opener implementations found"))
	}
	return opened, nil
}

// CloseConnection closes a connection to a BMC, trying all interface implementations passed in
func CloseConnection(ctx context.Context, c []connectionProviders, metadata ...*Metadata) (err error) {
	var metadataLocal Metadata
	defer func() {
		if len(metadata) > 0 && metadata[0] != nil {
			*metadata[0] = metadataLocal
		}
	}()
	var connClosed bool
Loop:
	for _, elem := range c {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem.closer != nil {
				metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
				openErr := elem.closer.Close(ctx)
				if openErr != nil {
					err = multierror.Append(err, openErr)
					continue
				}
				connClosed = true
				metadataLocal.SuccessfulCloseConns = append(metadataLocal.SuccessfulCloseConns, elem.name)
			}
		}
	}
	if connClosed {
		return nil
	}
	return multierror.Append(err, errors.New("failed to close connection"))
}

// CloseConnectionFromInterfaces pass through to library function
func CloseConnectionFromInterfaces(ctx context.Context, generic []interface{}, metadata ...*Metadata) (err error) {
	closers := make([]connectionProviders, 0)
	for _, elem := range generic {
		temp := connectionProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case Closer:
			temp.closer = p
			closers = append(closers, temp)
		default:
			e := fmt.Sprintf("not a Closer implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(closers) == 0 {
		return multierror.Append(err, errors.New("no Closer implementations found"))
	}
	return CloseConnection(ctx, closers, metadata...)
}
