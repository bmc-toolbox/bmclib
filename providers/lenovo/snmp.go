package lenovo

import (
	"context"
	"net/url"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// snmpSubPath is the XCC OEM SNMP resource, relative to the Manager's
// NetworkProtocol resource.
const snmpSubPath = "NetworkProtocol/Oem/Lenovo/SNMP"

// compile-time assertion that the provider implements the interface.
var _ bmc.SNMPConfigurer = (*Conn)(nil)

// SNMP returns the XCC SNMP trap configuration.
//
// Implements bmc.SNMPConfigurer.
func (c *Conn) SNMP(ctx context.Context) (bmc.SNMPConfig, error) {
	path, err := c.snmpPath(ctx)
	if err != nil {
		return bmc.SNMPConfig{}, err
	}

	// On XCC the OEM SNMP resource carries CommunityNames at the top level (a
	// sibling of SNMPTraps), not inside SNMPTraps. The trap enable/port live
	// under SNMPTraps.
	var doc struct {
		CommunityNames []string `json:"CommunityNames"`
		SNMPTraps      struct {
			SNMPv1TrapEnabled bool `json:"SNMPv1TrapEnabled"`
			ProtocolEnabled   bool `json:"ProtocolEnabled"`
			Port              int  `json:"Port"`
		} `json:"SNMPTraps"`
	}
	if err := c.getJSON(path, &doc); err != nil {
		return bmc.SNMPConfig{}, err
	}

	return bmc.SNMPConfig{
		V1TrapEnabled:  doc.SNMPTraps.SNMPv1TrapEnabled,
		V3TrapEnabled:  doc.SNMPTraps.ProtocolEnabled,
		TrapPort:       doc.SNMPTraps.Port,
		CommunityNames: doc.CommunityNames,
	}, nil
}

// SetSNMPAlertFilter PATCHes the SNMP trap (SNMPTraps) properties.
//
// Implements bmc.SNMPConfigurer.
func (c *Conn) SetSNMPAlertFilter(ctx context.Context, attrs map[string]any) error {
	return c.patchSNMPTraps(ctx, attrs)
}

// EnableSNMPv1Trap enables or disables the SNMPv1 trap.
//
// Implements bmc.SNMPConfigurer.
func (c *Conn) EnableSNMPv1Trap(ctx context.Context, enabled bool) error {
	return c.patchSNMPTraps(ctx, map[string]any{"SNMPv1TrapEnabled": enabled})
}

// EnableSNMPv3Trap enables or disables the SNMPv3 trap.
//
// Implements bmc.SNMPConfigurer.
func (c *Conn) EnableSNMPv3Trap(ctx context.Context, enabled bool) error {
	return c.patchSNMPTraps(ctx, map[string]any{"ProtocolEnabled": enabled})
}

// patchSNMPTraps PATCHes the SNMP resource with the given SNMPTraps attributes.
func (c *Conn) patchSNMPTraps(ctx context.Context, traps map[string]any) error {
	path, err := c.snmpPath(ctx)
	if err != nil {
		return err
	}

	payload := map[string]any{"SNMPTraps": traps}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, path, payload, nil))
}

// snmpPath resolves the XCC OEM SNMP resource path from the managed Manager.
func (c *Conn) snmpPath(ctx context.Context) (string, error) {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return "", err
	}

	return url.JoinPath(manager.ODataID, snmpSubPath)
}
