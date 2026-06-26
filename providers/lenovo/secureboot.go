package lenovo

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/stmcginnis/gofish/schemas"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.SecureBootManager = (*Conn)(nil)

// GetSecureBoot returns the current UEFI Secure Boot state from the
// ComputerSystem SecureBoot resource.
//
// Implements bmc.SecureBootManager.
func (c *Conn) GetSecureBoot(ctx context.Context) (bmc.SecureBootState, error) {
	sb, err := c.secureBoot()
	if err != nil {
		return bmc.SecureBootState{}, err
	}

	return bmc.SecureBootState{
		Enabled:     sb.SecureBootEnable,
		CurrentBoot: string(sb.SecureBootCurrentBoot),
		Mode:        string(sb.SecureBootMode),
	}, nil
}

// SetSecureBoot enables or disables Secure Boot by PATCHing
// SecureBoot.SecureBootEnable. The change takes effect on the next boot and only
// in UEFI boot mode. Implements bmc.SecureBootManager.
func (c *Conn) SetSecureBoot(ctx context.Context, enabled bool) error {
	sb, err := c.secureBoot()
	if err != nil {
		return err
	}

	payload := map[string]any{"SecureBootEnable": enabled}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, sb.ODataID, payload, nil))
}

// ResetSecureBootKeys resets the Secure Boot key databases via the
// SecureBoot.ResetKeys action. resetType is a Redfish ResetKeysType such as
// "ResetAllKeysToDefault", "DeleteAllKeys" or "DeletePK". This is destructive.
// Implements bmc.SecureBootManager.
func (c *Conn) ResetSecureBootKeys(ctx context.Context, resetType string) error {
	sb, err := c.secureBoot()
	if err != nil {
		return err
	}

	if _, err := sb.ResetKeys(schemas.ResetKeysType(resetType)); err != nil {
		return fmt.Errorf("resetting secure boot keys: %w", err)
	}

	return nil
}

// secureBoot resolves the SecureBoot resource of the managed system.
func (c *Conn) secureBoot() (*schemas.SecureBoot, error) {
	system, err := c.redfishwrapper.System()
	if err != nil {
		return nil, err
	}

	sb, err := system.SecureBoot()
	if err != nil {
		return nil, fmt.Errorf("reading secure boot resource: %w", err)
	}
	if sb == nil {
		return nil, fmt.Errorf("device does not expose a SecureBoot resource")
	}

	return sb, nil
}
