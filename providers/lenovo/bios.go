package lenovo

import (
	"context"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/stmcginnis/gofish/schemas"
)

// GetBiosConfiguration returns the current BIOS attributes as a key/value map,
// read from the ComputerSystem Bios resource Attributes.
//
// Implements bmc.BiosConfigurationGetter.
func (c *Conn) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	return c.redfishwrapper.GetBiosConfiguration(ctx)
}

// SetBiosConfiguration writes pending BIOS attributes, staged for the next host
// reset.
//
// This is an XCC-specific override of the shared
// redfishwrapper.SetBiosConfiguration. The wrapper PATCHes the settings target
// with an "@Redfish.SettingsApplyTime" annotation (gofish
// UpdateBiosAttributesApplyAt with OnReset), but the XCC BIOS settings resource
// rejects that property — it returns Base.1.11.PropertyUnknown for both
// "@Redfish.SettingsApplyTime" and "ApplyTime" and fails the whole request.
// XCC stages BIOS changes for the next reset implicitly, so this override
// PATCHes "Attributes" only (gofish UpdateBiosAttributes, no apply-time), while
// still GETting the settings target first so gofish supplies the If-Match etag.
//
// Implements bmc.BiosConfigurationSetter.
func (c *Conn) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	sys, err := c.redfishwrapper.System()
	if err != nil {
		return err
	}

	bios, err := sys.Bios()
	if err != nil {
		return err
	}
	if bios == nil {
		return bmclibErrs.ErrNoBiosAttributes
	}

	settingsAttributes := make(schemas.SettingsAttributes, len(biosConfig))
	for attr, value := range biosConfig {
		settingsAttributes[attr] = value
	}

	return bios.UpdateBiosAttributes(settingsAttributes)
}

// ResetBiosConfiguration restores BIOS settings to their default values via the
// Bios.ResetBios action.
//
// Implements bmc.BiosConfigurationResetter.
func (c *Conn) ResetBiosConfiguration(ctx context.Context) (err error) {
	return c.redfishwrapper.ResetBiosConfiguration(ctx)
}
