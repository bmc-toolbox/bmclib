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

// ResetBMC tries all implementations for a success BMC reset
func ResetBMC(ctx context.Context, resetType string, b []BMCResetter) (ok bool, err error) {
Loop:
	for _, elem := range b {
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			if elem != nil {
				ok, setErr := elem.BmcReset(ctx, resetType)
				if setErr != nil {
					err = multierror.Append(err, setErr)
					continue
				}
				if !ok {
					err = multierror.Append(err, errors.New("failed to reset BMC"))
					continue
				}
				return ok, nil
			}
		}
	}
	return ok, multierror.Append(err, errors.New("failed to reset BMC"))
}

// ResetBMCFromInterfaces pass through to library function
func ResetBMCFromInterfaces(ctx context.Context, resetType string, generic []interface{}) (ok bool, err error) {
	bmcSetters := make([]BMCResetter, 0)
	for _, elem := range generic {
		switch p := elem.(type) {
		case BMCResetter:
			bmcSetters = append(bmcSetters, p)
		default:
			e := fmt.Sprintf("not a BMCResetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(bmcSetters) == 0 {
		return ok, multierror.Append(err, errors.New("no BMCResetter implementations found"))
	}
	return ResetBMC(ctx, resetType, bmcSetters)
}
