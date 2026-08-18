package redfishwrapper

import (
	"context"

	"github.com/stmcginnis/gofish/schemas"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// GetSecureBoot returns whether UEFI Secure Boot is currently enabled for the system.
func (c *Client) GetSecureBoot(ctx context.Context) (enabled bool, err error) {
	sys, err := c.System()
	if err != nil {
		return false, err
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return false, bmclibErrs.ErrRedfishSystemOdataID
	}

	secureBoot, err := sys.SecureBoot()
	if err != nil {
		return false, err
	}

	return secureBoot.SecureBootEnable, nil
}

// SetSecureBoot enables or disables UEFI Secure Boot for the system. The system
// must be in UEFI boot mode for this to take effect.
func (c *Client) SetSecureBoot(ctx context.Context, enable bool) (err error) {
	sys, err := c.System()
	if err != nil {
		return err
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return bmclibErrs.ErrRedfishSystemOdataID
	}

	secureBoot, err := sys.SecureBoot()
	if err != nil {
		return err
	}

	secureBoot.SecureBootEnable = enable

	return secureBoot.Update()
}

// ResetSecureBootKeys resets the UEFI Secure Boot key databases. resetType is one
// of ResetAllKeysToDefault, DeleteAllKeys, or DeletePK.
func (c *Client) ResetSecureBootKeys(ctx context.Context, resetType string) (err error) {
	sys, err := c.System()
	if err != nil {
		return err
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return bmclibErrs.ErrRedfishSystemOdataID
	}

	secureBoot, err := sys.SecureBoot()
	if err != nil {
		return err
	}

	_, err = secureBoot.ResetKeys(schemas.ResetKeysType(resetType))

	return err
}
