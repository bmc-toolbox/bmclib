package bmc

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// BMCResetter for resetting a BMC.
// resetType: "warm" resets the management console without rebooting the BMC
// resetType: "cold" reboots the BMC
type BMCResetter interface {
	BmcReset(ctx context.Context, resetType string) (ok bool, err error)
}

// bmcProviders is an internal struct to correlate an implementation/provider and its name
type bmcProviders struct {
	name        string
	bmcResetter BMCResetter
}

// ResetBMC tries all implementations for a success BMC reset
// if a successfulProviderName is passed in, it will be updated to be the name of the provider that successfully executed
func ResetBMC(ctx context.Context, resetType string, b []bmcProviders, metadata ...*Metadata) (ok bool, err error) {
	if len(metadata) == 0 || metadata[0] == nil {
		metadata = []*Metadata{&Metadata{}}
	}
Loop:
	for _, elem := range b {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem.bmcResetter != nil {
				*metadata[0] = Metadata{ProvidersAttempted: append(metadata[0].ProvidersAttempted, elem.name)}
				ok, setErr := elem.bmcResetter.BmcReset(ctx, resetType)
				if setErr != nil {
					err = multierror.Append(err, setErr)
					continue
				}
				if !ok {
					err = multierror.Append(err, errors.New("failed to reset BMC"))
					continue
				}
				*metadata[0] = Metadata{SuccessfulProvider: elem.name, ProvidersAttempted: append(metadata[0].ProvidersAttempted, elem.name)}
				return ok, nil
			}
		}
	}
	return ok, multierror.Append(err, errors.New("failed to reset BMC"))
}

// ResetBMCFromInterfaces pass through to library function
// if a successfulProviderName is passed in, it will be updated to be the name of the provider that successfully executed
func ResetBMCFromInterfaces(ctx context.Context, resetType string, generic []interface{}, metadata ...*Metadata) (ok bool, err error) {
	bmcSetters := make([]bmcProviders, 0)
	for _, elem := range generic {
		var temp bmcProviders
		switch p := elem.(type) {
		case Provider:
			temp.name = p.Name()
		}
		switch p := elem.(type) {
		case BMCResetter:
			temp.bmcResetter = p
			bmcSetters = append(bmcSetters, temp)
		default:
			e := fmt.Sprintf("not a BMCResetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(bmcSetters) == 0 {
		return ok, multierror.Append(err, errors.New("no BMCResetter implementations found"))
	}
	return ResetBMC(ctx, resetType, bmcSetters, metadata...)
}
