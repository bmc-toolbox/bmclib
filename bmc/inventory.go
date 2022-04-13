package bmc

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/errors"
)

// InventoryGetter defines methods to retrieve device hardware and firmware inventory
type InventoryGetter interface {
	Inventory(ctx context.Context) (device *devices.Device, err error)
}

type inventoryGetterProvider struct {
	name string
	InventoryGetter
}

// Inventory returns hardware and firmware inventory
func Inventory(ctx context.Context, generic []inventoryGetterProvider) (device *devices.Device, metadata Metadata, err error) {
	var metadataLocal Metadata
Loop:
	for _, elem := range generic {
		if elem.InventoryGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
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

// GetInventoryFromInterfaces is a pass through to library function
func GetInventoryFromInterfaces(ctx context.Context, generic []interface{}) (device *devices.Device, metadata Metadata, err error) {
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

	return Inventory(ctx, implementations)
}
