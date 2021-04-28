package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmclib/providers/redfish"
	"github.com/hashicorp/go-multierror"
)

var (
	ErrVirtualMediaInsert = errors.New("failed to insert virtual media")
)

// VirtualMediaInserter inserts virtual media
type VirtualMediaInserter interface {
	InsertVirtualMedia(ctx context.Context, options *redfish.VirtualMediaOptions) (err error)
}

// VirtualMediaEjecter ejects inserted virtual media
type VirtualMediaEjecter interface {
	EjectVirtualMedia(ctx context.Context) (err error)
}

// InsertVirtualMedia inserts the virtual media declared in options
func InsertVirtualMedia(ctx context.Context, options *redfish.VirtualMediaOptions, p []VirtualMediaInserter) (err error) {
Loop:
	for _, elem := range p {
		if elem == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			mErr := elem.InsertVirtualMedia(ctx, options)
			if mErr != nil {
				err = multierror.Append(err, mErr)
				continue
			}
			return nil
		}
	}

	return multierror.Append(err, ErrVirtualMediaInsert)
}

//  VirtualMediaInserterFromInterfaces pass through to library function
func VirtualMediaInserterFromInterfaces(ctx context.Context, options *redfish.VirtualMediaOptions, generic []interface{}) (err error) {
	virtualMediaInserter := make([]VirtualMediaInserter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case VirtualMediaInserter:
			virtualMediaInserter = append(virtualMediaInserter, p)
		default:
			e := fmt.Sprintf("not a VirtualMediaInserter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(virtualMediaInserter) == 0 {
		return multierror.Append(err, errors.New("no VirtualMediaInserter implementations found"))
	}

	return InsertVirtualMedia(ctx, options, virtualMediaInserter)
}

// EjectVirtualMedia ejects virtual media
func EjectVirtualMedia(ctx context.Context, p []VirtualMediaEjecter) (err error) {
Loop:
	for _, elem := range p {
		if elem == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			mErr := elem.EjectVirtualMedia(ctx)
			if mErr != nil {
				err = multierror.Append(err, mErr)
				continue
			}
			return nil
		}
	}

	return multierror.Append(err, ErrVirtualMediaInsert)
}

//  VirtualMediaEjecterFromInterfaces pass through to library function
func VirtualMediaEjecterFromInterfaces(ctx context.Context, generic []interface{}) (err error) {
	virtualMediaEjecter := make([]VirtualMediaEjecter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case VirtualMediaEjecter:
			virtualMediaEjecter = append(virtualMediaEjecter, p)
		default:
			e := fmt.Sprintf("not a VirtualMediaEjecter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(virtualMediaEjecter) == 0 {
		return multierror.Append(err, errors.New("no VirtualMediaEjecter implementations found"))
	}

	return EjectVirtualMedia(ctx, virtualMediaEjecter)
}
