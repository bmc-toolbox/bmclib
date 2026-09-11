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

// SetBiosConfiguration applies the given BIOS configuration attributes, to take effect on the
// next reset.
//
// Every requested attribute is included in the PATCH, whether or not its value already matches
// what the resource currently reports. Omitting the ones that match - which is what gofish's
// Bios.UpdateBiosAttributesApplyAt does, and what this used to inherit - silently drops exactly
// the writes a caller most needs to land: BIOS attribute writes are staged and only take effect
// on the next reset, so any caller that stages a change and then revises it before that reset is
// asking to write a value equal to the still-applied one. Diffing against applied state turns
// that into a no-op that reports success, leaving the superseded value staged to commit.
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
	err = bios.UpdateBiosAttributesExactApplyAt(settingsAttributes, schemas.OnResetSettingsApplyTime)
	if err != nil && rejectsSettingsApplyTime(err) {
		// This BMC's Bios resource doesn't declare @Redfish.Settings.SupportedApplyTimes
		// at all and rejects the @Redfish.SettingsApplyTime property outright, rather than
		// ignoring it. Retry without an apply-time hint - the settings still go through the
		// resource's separate Settings URI (@Redfish.Settings.SettingsObject), which by
		// Redfish convention means they're staged rather than applied immediately.
		return bios.UpdateBiosAttributesExact(settingsAttributes)
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
