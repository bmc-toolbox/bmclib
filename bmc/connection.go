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
	opener Opener
	closer Closer
}

// OpenConnection opens a connection to a BMC, trying all interface implementations passed in
func OpenConnection(ctx context.Context, o []connectionProviders, metadata ...*Metadata) (err error) {
	if len(metadata) == 0 || metadata[0] == nil {
		metadata = []*Metadata{&Metadata{}}
	}
	var connOpen bool
Loop:
	for _, elem := range o {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem.opener != nil {
				*metadata[0] = Metadata{ProvidersAttempted: append(metadata[0].ProvidersAttempted, elem.name)}
				openErr := elem.opener.Open(ctx)
				if openErr != nil {
					err = multierror.Append(err, openErr)
					continue
				}
				connOpen = true
				*metadata[0] = Metadata{SuccessfulOpenConns: append(metadata[0].SuccessfulOpenConns, elem.name), ProvidersAttempted: metadata[0].ProvidersAttempted}
			}
		}
	}
	if connOpen {
		return nil
	}
	return multierror.Append(err, errors.New("failed to open connection"))
}

// OpenConnectionFromInterfaces pass through to library function
func OpenConnectionFromInterfaces(ctx context.Context, generic []interface{}, metadata ...*Metadata) (err error) {
	openers := make([]connectionProviders, 0)
	for _, elem := range generic {
		var temp connectionProviders
		switch p := elem.(type) {
		case Provider:
			temp.name = p.Name()
		}
		switch p := elem.(type) {
		case Opener:
			temp.opener = p
			openers = append(openers, temp)
		default:
			e := fmt.Sprintf("not a Opener implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(openers) == 0 {
		return multierror.Append(err, errors.New("no Opener implementations found"))
	}
	return OpenConnection(ctx, openers, metadata...)
}

// CloseConnection closes a connection to a BMC, trying all interface implementations passed in
func CloseConnection(ctx context.Context, c []connectionProviders, metadata ...*Metadata) (err error) {
	if len(metadata) == 0 || metadata[0] == nil {
		metadata = []*Metadata{&Metadata{}}
	}
	var connClosed bool
Loop:
	for _, elem := range c {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem.closer != nil {
				*metadata[0] = Metadata{ProvidersAttempted: append(metadata[0].ProvidersAttempted, elem.name)}
				openErr := elem.closer.Close(ctx)
				if openErr != nil {
					err = multierror.Append(err, openErr)
					continue
				}
				connClosed = true
				*metadata[0] = Metadata{SuccessfulCloseConns: append(metadata[0].SuccessfulCloseConns, elem.name), ProvidersAttempted: metadata[0].ProvidersAttempted}
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
		var temp connectionProviders
		switch p := elem.(type) {
		case Provider:
			temp.name = p.Name()
		}
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
