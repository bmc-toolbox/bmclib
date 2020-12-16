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

// OpenConnection opens a connection to a BMC, trying all interface implementations passed in
func OpenConnection(ctx context.Context, o []Opener) (err error) {
	var connOpen bool
Loop:
	for _, elem := range o {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem != nil {
				openErr := elem.Open(ctx)
				if openErr != nil {
					err = multierror.Append(err, openErr)
					continue
				}
				connOpen = true
			}
		}
	}
	if connOpen {
		return nil
	}
	return multierror.Append(err, errors.New("failed to open connection"))
}

// OpenConnectionFromInterfaces pass through to library function
func OpenConnectionFromInterfaces(ctx context.Context, generic []interface{}) (err error) {
	var openers []Opener
	for _, elem := range generic {
		switch p := elem.(type) {
		case Opener:
			openers = append(openers, p)
		default:
			e := fmt.Sprintf("not a Opener implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(openers) == 0 {
		return multierror.Append(err, errors.New("no Opener implementations found"))
	}
	return OpenConnection(ctx, openers)
}

// CloseConnection closes a connection to a BMC, trying all interface implementations passed in
func CloseConnection(ctx context.Context, c []Closer) (err error) {
	var connClosed bool
Loop:
	for _, elem := range c {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem != nil {
				openErr := elem.Close(ctx)
				if openErr != nil {
					err = multierror.Append(err, openErr)
					continue
				}
				connClosed = true
			}
		}
	}
	if connClosed {
		return nil
	}
	return multierror.Append(err, errors.New("failed to close connection"))
}

// CloseConnectionFromInterfaces pass through to library function
func CloseConnectionFromInterfaces(ctx context.Context, generic []interface{}) (err error) {
	var closers []Closer
	for _, elem := range generic {
		switch p := elem.(type) {
		case Closer:
			closers = append(closers, p)
		default:
			e := fmt.Sprintf("not a Closer implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(closers) == 0 {
		return multierror.Append(err, errors.New("no Closer implementations found"))
	}
	return CloseConnection(ctx, closers)
}
