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
// every other Dell BIOS-attribute setter in this package.
//
// It does NOT remove the underlying constraint, and callers relying on this fix for anything
// more should know that: SetBiosConfiguration always requests "@Redfish.SettingsApplyTime":
// "OnReset" (required for iDRAC to ever actually apply a Bios/Settings write - confirmed
// independently by other Redfish client projects hitting this same iDRAC requirement), and
// asserting that on a PATCH is itself what creates the one job iDRAC allows at a time - a bare,
// ApplyTime-less Bios/Settings PATCH creates no job at all and merely soft-stages, confirmed via
// the same JobService/Jobs snapshots. So SetSecureBoot, SetSecureBootKeyManagement,
// SetNetworkBootEnabled, SetHTTPBootURI, and SetHTTPBootTLSMode still conflict with
// IDRAC.2.14.SYS011 pairwise, in either order, if called separately within the same boot cycle -
// this fix makes that symmetric and consistent across all of them, it doesn't eliminate it. A
// caller needing more than one of these changes in one maintenance window must still either
// batch them into a single SetBiosConfiguration call (confirmed live to merge cleanly even with
// @Redfish.SettingsApplyTime present) or handle/retry the conflict itself.
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
