package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.BMCResetter = (*Conn)(nil)

// BmcReset restarts the BMC via the Manager.Reset action.
//
// resetType is a Redfish ResetType, typically "GracefulRestart" or
// "ForceRestart". Implements bmc.BMCResetter.
func (c *Conn) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	return c.redfishwrapper.BMCReset(ctx, resetType)
}

// ResetToFactoryDefaults resets the BMC to its factory defaults via the
// Manager.ResetToDefaults action.
//
// resetType is a Redfish ResetToDefaultsType, e.g. "ResetAll",
// "PreserveNetworkAndUsers" or "PreserveNetwork". This is destructive and
// disconnects the session. It is an XCC-specific provider method (the
// bmc.BMCResetter interface only covers Manager.Reset).
func (c *Conn) ResetToFactoryDefaults(ctx context.Context, resetType string) error {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return err
	}

	if resetType == "" {
		resetType = "ResetAll"
	}

	target := singleTrailingSlashJoin(manager.ODataID, "Actions/Manager.ResetToDefaults")
	payload := map[string]any{"ResetType": resetType}

	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, target, payload, nil))
}

// UpdateManager PATCHes Manager properties (e.g. the OEM time zone and other
// OEM fields).
//
// This is an XCC-specific provider method.
func (c *Conn) UpdateManager(ctx context.Context, properties map[string]any) error {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return err
	}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, manager.ODataID, properties, nil))
}
