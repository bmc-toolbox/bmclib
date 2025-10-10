package redfishwrapper

import (
	"context"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

func (c *Client) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	sys, err := c.System()
	if err != nil {
		return nil, err
	}

	biosConfig = make(map[string]string)
	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return biosConfig, nil
	}

	bios, err := sys.Bios()
	if err != nil {
		return nil, err
	}

	if bios == nil {
		return nil, bmclibErrs.ErrNoBiosAttributes
	}

	for attr := range bios.Attributes {
		biosConfig[attr] = bios.Attributes.String(attr)
	}

	return biosConfig, nil
}

func (c *Client) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	sys, err := c.System()
	if err != nil {
		return err
	}

	settingsAttributes := make(redfish.SettingsAttributes)

	for attr, value := range biosConfig {
		settingsAttributes[attr] = value
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return nil
	}

	bios, err := sys.Bios()
	if err != nil {
		return err
	}

	// TODO(jwb) We should handle passing different apply times here
	return bios.UpdateBiosAttributesApplyAt(settingsAttributes, common.OnResetApplyTime)
}

func (c *Client) ResetBiosConfiguration(ctx context.Context) (err error) {
	sys, err := c.System()
	if err != nil {
		return err
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return nil
	}

	bios, err := sys.Bios()
	if err != nil {
		return err
	}

	return bios.ResetBios()
}
