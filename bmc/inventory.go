package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	common "github.com/metal-toolbox/bmc-common"
	"github.com/pkg/errors"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
)

// InventoryGetter defines methods to retrieve device hardware and firmware inventory
type InventoryGetter interface {
	Inventory(ctx context.Context) (device *common.Device, err error)
}

type inventoryGetterProvider struct {
	name string
	InventoryGetter
}

// inventory returns hardware and firmware inventory
func inventory(ctx context.Context, generic []inventoryGetterProvider) (device *common.Device, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.InventoryGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return device, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			device, vErr := elem.Inventory(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return device, metadataLocal, nil
		}
	}

	return device, metadataLocal, multierror.Append(err, errors.New("failure to get device inventory"))
}

// GetInventoryFromInterfaces identifies implementations of the InventoryGetter interface and passes the found implementations to the inventory() wrapper method
func GetInventoryFromInterfaces(ctx context.Context, generic []interface{}) (device *common.Device, metadata Metadata, err error) {
	metadata = newMetadata()

	implementations := make([]inventoryGetterProvider, 0)
	for _, elem := range generic {
		temp := inventoryGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case InventoryGetter:
			temp.InventoryGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a InventoryGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return device, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no InventoryGetter implementations found"),
			),
		)
	}

	return inventory(ctx, implementations)
}
