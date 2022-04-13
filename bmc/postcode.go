package bmc

import (
	"context"
	"fmt"

	bmclibErrs "github.com/bmc-toolbox/bmclib/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// PostCodeGetter defines methods to retrieve device BIOS/UEFI POST code
type PostCodeGetter interface {
	// PostCode retrieves the BIOS/UEFI POST code from a device
	//
	// returns 'status' which is a (bmclib specific) string identifier for the POST code
	// and 'code' with the actual POST code returned to bmclib by the device
	PostCode(ctx context.Context) (status string, code int, err error)
}

type postCodeGetterProvider struct {
	name string
	PostCodeGetter
}

// PostCode returns the device BIOS/UEFI POST code
func PostCode(ctx context.Context, generic []postCodeGetterProvider) (status string, code int, metadata Metadata, err error) {
	var metadataLocal Metadata
Loop:
	for _, elem := range generic {
		if elem.PostCodeGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			status, code, vErr := elem.PostCode(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return status, code, metadataLocal, nil
		}
	}

	return status, code, metadataLocal, multierror.Append(err, errors.New("failure to get device POST code"))
}

// GetPostCodeFromInterfaces is a pass through to library function
func GetPostCodeInterfaces(ctx context.Context, generic []interface{}) (status string, code int, metadata Metadata, err error) {
	implementations := make([]postCodeGetterProvider, 0)
	for _, elem := range generic {
		temp := postCodeGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case PostCodeGetter:
			temp.PostCodeGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a PostCodeGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return status, code, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no PostCodeGetter implementations found"),
			),
		)
	}

	return PostCode(ctx, implementations)
}
