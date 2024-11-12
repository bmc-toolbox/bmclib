package redfishwrapper

import (
	"context"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
)

func (c *Client) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	systems, err := c.Systems()
	if err != nil {
		return nil, err
	}

	biosConfig = make(map[string]string)
	for _, sys := range systems {
		if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
			continue
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
	}

	return biosConfig, nil
}

func (c *Client) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	systems, err := c.Systems()
	if err != nil {
		return err
	}

	settingsAttributes := make(redfish.SettingsAttributes)

	for attr, value := range biosConfig {
		settingsAttributes[attr] = value
	}

	for _, sys := range systems {
		if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
			continue
		}

		bios, err := sys.Bios()
		if err != nil {
			return err
		}

		// TODO(jwb) We should handle passing different apply times here
		err = bios.UpdateBiosAttributesApplyAt(settingsAttributes, common.OnResetApplyTime)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) ResetBiosConfiguration(ctx context.Context) (err error) {
	systems, err := c.Systems()
	if err != nil {
		return err
	}

	for _, sys := range systems {
		if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
			continue
		}

		bios, err := sys.Bios()
		if err != nil {
			return err
		}

		err = bios.ResetBios()

		if err != nil {
			return err
		}
	}

	return nil
}
