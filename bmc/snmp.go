package bmc

import "context"

// SNMPConfig describes the BMC SNMP trap configuration.
type SNMPConfig struct {
	// V1TrapEnabled reports whether the SNMPv1 trap is enabled.
	V1TrapEnabled bool
	// V3TrapEnabled reports whether the SNMPv3 trap is enabled.
	V3TrapEnabled bool
	// TrapPort is the trap receiver port.
	TrapPort int
	// CommunityNames are the configured SNMPv1 community names.
	CommunityNames []string
}

// SNMPConfigurer is implemented by providers that can read and configure SNMP
// traps.
type SNMPConfigurer interface {
	// SNMP returns the SNMP trap configuration.
	SNMP(ctx context.Context) (SNMPConfig, error)
	// SetSNMPAlertFilter PATCHes the SNMP trap (SNMPTraps) properties with the
	// given attributes (alert recipients, community names, addresses, ...).
	SetSNMPAlertFilter(ctx context.Context, attrs map[string]any) error
	// EnableSNMPv1Trap enables or disables the SNMPv1 trap.
	EnableSNMPv1Trap(ctx context.Context, enabled bool) error
	// EnableSNMPv3Trap enables or disables the SNMPv3 trap.
	EnableSNMPv3Trap(ctx context.Context, enabled bool) error
}
