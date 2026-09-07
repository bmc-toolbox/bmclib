package redfishwrapper

import (
	"context"
	"fmt"
)

const (
	attrEnabled  = "Enabled"
	attrDisabled = "Disabled"
	attrUEFI     = "UEFI"
)

// networkBootFingerprintTables resolves the BIOS attribute PATCH body needed to enable or
// disable UEFI HTTP Boot and/or legacy PXE boot capability, based on a fingerprint match against
// a machine's current BIOS configuration.
//
// fingerprint keys are attribute names that, if present in a GetBiosConfiguration() result,
// identify this table as applicable. Extend with one entry per additional vendor/board as
// fleets need them — do not attempt to support unknown BIOS/vendor combinations silently.
//
// httpEnabled/httpDisabled and pxeEnabled/pxeDisabled are independent attribute sets: HTTP Boot
// and legacy PXE are separate capabilities that can be on or off in any combination. NetworkStack
// and BootModeSelect are prerequisites shared by both protocols, so they're only asserted on the
// enabled side of each set — disabling one protocol must not turn off the network stack the
// other protocol may still depend on.
var networkBootFingerprintTables = []struct {
	fingerprint  string // attribute name unique enough to identify this BIOS/vendor
	httpEnabled  map[string]string
	httpDisabled map[string]string
	pxeEnabled   map[string]string
	pxeDisabled  map[string]string
}{
	{
		fingerprint: "IPv4HTTPSupport", // Supermicro H12SSW-NTR / AMI Aptio, confirmed live
		httpEnabled: map[string]string{
			"NetworkStack":    attrEnabled,
			"BootModeSelect":  attrUEFI,
			"IPv4HTTPSupport": attrEnabled,
			"IPv6HTTPSupport": attrEnabled,
		},
		httpDisabled: map[string]string{
			"IPv4HTTPSupport": attrDisabled,
			"IPv6HTTPSupport": attrDisabled,
		},
		pxeEnabled: map[string]string{
			"NetworkStack":   attrEnabled,
			"BootModeSelect": attrUEFI,
			"IPv4PXESupport": attrEnabled,
		},
		pxeDisabled: map[string]string{
			"IPv4PXESupport": attrDisabled,
		},
	},
}

// SetNetworkBootEnabled enables/disables UEFI HTTP Boot and/or legacy PXE boot capability by
// PATCHing the BIOS attributes resolved from the machine's current BIOS configuration via a
// fingerprint match (see networkBootFingerprintTables). httpEnabled and pxeEnabled are
// independent: a nil pointer leaves that protocol's attributes untouched, so either, both, or
// neither may be set.
func (c *Client) SetNetworkBootEnabled(ctx context.Context, httpEnabled, pxeEnabled *bool) (ok bool, err error) {
	if httpEnabled == nil && pxeEnabled == nil {
		return false, fmt.Errorf("at least one of httpEnabled or pxeEnabled must be set")
	}

	current, err := c.GetBiosConfiguration(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to read current BIOS configuration: %w", err)
	}

	attrs, err := networkBootAttributes(httpEnabled, pxeEnabled, current)
	if err != nil {
		return false, err
	}

	if err := c.SetBiosConfiguration(ctx, attrs); err != nil {
		return false, err
	}

	return true, nil
}

// networkBootAttributes returns the BIOS attribute PATCH body for the requested HTTP Boot / PXE
// Boot enabled states, selecting the table whose fingerprint attribute is present in current
// (the machine's live GetBiosConfiguration result). Returns an error if no known table matches —
// silently no-op'ing on an unrecognized BIOS is worse than a clear, actionable failure.
func networkBootAttributes(httpEnabled, pxeEnabled *bool, current map[string]string) (map[string]string, error) {
	for _, t := range networkBootFingerprintTables {
		if _, ok := current[t.fingerprint]; !ok {
			continue
		}

		attrs := map[string]string{}
		if httpEnabled != nil {
			mergeNetworkBootAttrs(attrs, pickNetworkBootAttrs(*httpEnabled, t.httpEnabled, t.httpDisabled))
		}
		if pxeEnabled != nil {
			mergeNetworkBootAttrs(attrs, pickNetworkBootAttrs(*pxeEnabled, t.pxeEnabled, t.pxeDisabled))
		}
		return attrs, nil
	}
	return nil, fmt.Errorf("no known BIOS attribute mapping for this machine (fingerprint attributes not found in GetBiosConfiguration result)")
}

func pickNetworkBootAttrs(enabled bool, onTrue, onFalse map[string]string) map[string]string {
	if enabled {
		return onTrue
	}
	return onFalse
}

func mergeNetworkBootAttrs(dst, src map[string]string) {
	for k, v := range src {
		dst[k] = v
	}
}
