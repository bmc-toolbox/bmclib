package redfishwrapper

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/schemas"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// SetHTTPBootURI sets the URI UEFI HTTP Boot fetches its boot image from, via the standard
// Redfish ComputerSystem.Boot.HttpBootUri property (Redfish schema v1.9.0+). It only sets that
// one property, leaving BootSourceOverrideTarget/Enabled/Mode untouched — setting the URI does
// not select UEFI HTTP Boot as the boot target; use SetBootDevice(..., "uefi_http", ...) for that.
func (c *Client) SetHTTPBootURI(_ context.Context, uri string) (ok bool, err error) {
	if err := c.SessionActive(); err != nil {
		return false, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	system, err := c.System()
	if err != nil {
		return false, err
	}

	boot := schemas.Boot{HTTPBootURI: uri}
	if err := system.SetBoot(&boot); err != nil {
		return false, err
	}

	return true, nil
}
