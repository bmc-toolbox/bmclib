package redfishwrapper

import (
	"context"
	"errors"
	"strings"

	"github.com/stmcginnis/gofish/schemas"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// rejectsSettingsApplyTime reports whether err is a Redfish error indicating
// the BMC doesn't recognize the @Redfish.SettingsApplyTime property at all.
// Some BMCs don't declare @Redfish.Settings.SupportedApplyTimes on their Bios
// resource and reject the property outright rather than ignoring it.
func rejectsSettingsApplyTime(err error) bool {
	var redfishErr *schemas.Error
	if !errors.As(err, &redfishErr) {
		return false
	}

	for i := range redfishErr.ExtendedInfos {
		info := &redfishErr.ExtendedInfos[i]
		if info.MessageID != "Base.1.10.PropertyUnknown" {
			continue
		}
		for _, prop := range info.RelatedProperties {
			if strings.Contains(prop, "SettingsApplyTime") {
				return true
			}
		}
	}

	return false
}

// GetBiosConfiguration returns the current BIOS configuration attributes for the system.
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

// SetBiosConfiguration applies the given BIOS configuration attributes, to take effect on the next reset.
func (c *Client) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	sys, err := c.System()
	if err != nil {
		return err
	}

	settingsAttributes := make(schemas.SettingsAttributes)

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
	err = bios.UpdateBiosAttributesApplyAt(settingsAttributes, schemas.OnResetSettingsApplyTime)
	if err != nil && rejectsSettingsApplyTime(err) {
		// This BMC's Bios resource doesn't declare @Redfish.Settings.SupportedApplyTimes
		// at all and rejects the @Redfish.SettingsApplyTime property outright, rather than
		// ignoring it. Retry without an apply-time hint - the settings still go through the
		// resource's separate Settings URI (@Redfish.Settings.SettingsObject), which by
		// Redfish convention means they're staged rather than applied immediately.
		return bios.UpdateBiosAttributes(settingsAttributes)
	}
	return err
}

// ResetBiosConfiguration resets the BIOS configuration to its default values.
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

	_, err = bios.ResetBios()
	return err
}
