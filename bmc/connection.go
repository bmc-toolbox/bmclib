package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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
func OpenConnectionFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (opened []interface{}, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		// return immediately if the context is done/terminated/etc
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			return nil, metadata, err
		default:
		}
		// get the provider name for use in metadata
		providerName := getProviderName(elem)
		// now, try to open connections
		switch p := elem.(type) {
		case Opener:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, providerName)
			tmpctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			er := p.Open(tmpctx)
			if er != nil {
				err = multierror.Append(err, errors.WithMessagef(er, "provider: %v", providerName))
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
		return nil, metadataLocal, multierror.Append(err, errors.New("no Opener implementations found"))
	}
	return opened, metadataLocal, nil
}

// closeConnection closes a connection to a BMC, trying all interface implementations passed in
func closeConnection(ctx context.Context, c []connectionProviders) (metadata Metadata, err error) {
	var metadataLocal Metadata
	var connClosed bool

	for _, elem := range c {
		if elem.closer == nil {
			continue
		}
		metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
		closeErr := elem.closer.Close(ctx)
		if closeErr != nil {
			err = multierror.Append(err, errors.WithMessagef(closeErr, "provider: %v", elem.name))
			continue
		}
		connClosed = true
		metadataLocal.SuccessfulCloseConns = append(metadataLocal.SuccessfulCloseConns, elem.name)
	}
	if connClosed {
		return metadataLocal, nil
	}
	return metadataLocal, multierror.Append(err, errors.New("failed to close connection"))
}

// CloseConnectionFromInterfaces identifies implementations of the Closer() interface and and passes the found implementations to the closeConnection() wrapper
func CloseConnectionFromInterfaces(ctx context.Context, generic []interface{}) (metadata Metadata, err error) {
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
		return metadata, multierror.Append(err, errors.New("no Closer implementations found"))
	}
	return closeConnection(ctx, closers)
}
