package bmc

import (
	"context"
	"fmt"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sanity-io/litter"
)

type BiosConfigurationGetter interface {
	GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error)
}

type biosConfigurationGetterProvider struct {
	name string
	BiosConfigurationGetter
}

func biosConfiguration(ctx context.Context, generic []biosConfigurationGetterProvider) (biosConfig map[string]string, metadata Metadata, err error) {
	var metadataLocal Metadata
Loop:
	for _, elem := range generic {
		if elem.BiosConfigurationGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			biosConfig, vErr := elem.GetBiosConfiguration(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return biosConfig, metadataLocal, nil
		}
	}

	return biosConfig, metadataLocal, multierror.Append(err, errors.New("failure to get bios configuration"))
}

func GetBiosConfigurationInterfaces(ctx context.Context, generic []interface{}) (biosConfig map[string]string, metadata Metadata, err error) {
	implementations := make([]biosConfigurationGetterProvider, 0)
	litter.Dump(generic)
	for _, elem := range generic {
		litter.Dump(elem)
		temp := biosConfigurationGetterProvider{name: getProviderName(elem)}
		litter.Dump(getProviderName(elem))
		litter.Dump(temp)
		switch p := elem.(type) {
		case BiosConfigurationGetter:
			temp.BiosConfigurationGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a BiosConfigurationGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return biosConfig, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no BiosConfigurationGetter implementations found"),
			),
		)
	}

	return biosConfiguration(ctx, implementations)
}
