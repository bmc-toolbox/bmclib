package lenovo

import (
	"context"
)

// GetSecureBoot returns whether UEFI Secure Boot is currently enabled.
//
// Implements bmc.SecureBootStateGetter.
func (c *Conn) GetSecureBoot(ctx context.Context) (enabled bool, err error) {
	return c.redfishwrapper.GetSecureBoot(ctx)
}

// SetSecureBoot enables or disables UEFI Secure Boot.
//
// Implements bmc.SecureBootSetter.
func (c *Conn) SetSecureBoot(ctx context.Context, enable bool) (err error) {
	return c.redfishwrapper.SetSecureBoot(ctx, enable)
}

// ResetSecureBootKeys resets the UEFI Secure Boot key databases.
//
// Implements bmc.SecureBootKeysResetter.
func (c *Conn) ResetSecureBootKeys(ctx context.Context, resetType string) (err error) {
	return c.redfishwrapper.ResetSecureBootKeys(ctx, resetType)
}
