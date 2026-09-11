package dell

import (
	"context"
)

// secureBootAttribute is Dell's vendor-specific BIOS Setup attribute name for the Secure Boot
// toggle, along with its two enumeration values. These MUST NOT leak into any exported bmclib
// identifier.
const (
	secureBootAttribute = "SecureBoot"
	secureBootEnabled   = "Enabled"
	secureBootDisabled  = "Disabled"
)

// SetSecureBoot enables or disables UEFI Secure Boot.
//
// Unlike the Redfish-generic provider, which PATCHes the standard ComputerSystem SecureBoot
// resource's SecureBootEnable property, this PATCHes Dell's own SecureBoot BIOS Setup attribute
// via SetBiosConfiguration.
//
// The ComputerSystem SecureBoot resource has an order-dependent bug on this box: PATCHing it
// unconditionally creates a real, exclusive BIOS Configuration Job immediately (confirmed live
// via JobService/Jobs snapshots - no @Redfish.SettingsApplyTime needed or even accepted on that
// resource). If that PATCH runs before another Bios/Settings write in the same maintenance
// window, the later write fails with IDRAC.2.14.SYS011, naming the attribute it was trying to
// set even though that attribute was never touched before. The reverse order merges cleanly,
// only because no job yet exists when the resource PATCH runs (confirmed with a same-box,
// same-attributes, order-only-swapped A/B, on a PowerEdge R6715, iDRAC firmware 1.20.80.51).
//
// PATCHing the SecureBoot BIOS Setup attribute via SetBiosConfiguration instead removes that
// specific order-dependent asymmetry: SetSecureBoot now fails the same way, in either order, as
// every other Dell BIOS-attribute setter in this package - and, like every one of them, that
// failure is transparently recovered from by recoveringRedfishClient (see
// bios_settings_recovery.go) rather than surfaced to the caller. SetBiosConfiguration always
// requests "@Redfish.SettingsApplyTime": "OnReset" (required for iDRAC to ever actually apply a
// Bios/Settings write - confirmed independently by other Redfish client projects hitting this
// same iDRAC requirement), and asserting that on a PATCH is itself what creates the one job
// iDRAC allows at a time, which is why two separate Dell BIOS-attribute calls in the same boot
// cycle would otherwise conflict on the second one regardless of which two they are. Routing
// SetSecureBoot through the same Bios/Settings mechanism as everything else means that recovery
// only ever has to reason about one job-creation trigger, not two.
//
// The write is staged into the Bios/Settings resource and only takes effect on the next POST.
//
// Implements bmc.SecureBootSetter.
func (c *Conn) SetSecureBoot(ctx context.Context, enable bool) (err error) {
	want := secureBootDisabled
	if enable {
		want = secureBootEnabled
	}

	return c.redfishwrapper.SetBiosConfiguration(ctx, map[string]string{secureBootAttribute: want})
}
