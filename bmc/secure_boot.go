package bmc

import "context"

// SecureBootState describes the UEFI Secure Boot state of a system.
type SecureBootState struct {
	// Enabled reports whether UEFI Secure Boot takes effect on the next boot.
	Enabled bool
	// CurrentBoot reports the Secure Boot state during the current boot cycle
	// (e.g. "Enabled", "Disabled").
	CurrentBoot string
	// Mode reports the current UEFI Secure Boot mode (e.g. "UserMode",
	// "SetupMode", "AuditMode", "DeployedMode").
	Mode string
}

// SecureBootManager is implemented by providers that can read and manage UEFI
// Secure Boot on a system.
type SecureBootManager interface {
	// GetSecureBoot returns the current Secure Boot state.
	GetSecureBoot(ctx context.Context) (SecureBootState, error)
	// SetSecureBoot enables or disables Secure Boot (applied on next boot).
	SetSecureBoot(ctx context.Context, enabled bool) error
	// ResetSecureBootKeys resets the Secure Boot key databases. resetType is a
	// Redfish ResetKeysType, e.g. "ResetAllKeysToDefault", "DeleteAllKeys" or
	// "DeletePK". This is destructive and may require a subsequent system reset.
	ResetSecureBootKeys(ctx context.Context, resetType string) error
}
