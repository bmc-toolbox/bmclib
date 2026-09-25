package dell

import (
	"context"

	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// secureBootPolicyAttribute is Dell's vendor-specific BIOS attribute name and
// MUST NOT leak into any exported bmclib identifier.
const (
	secureBootPolicyAttribute = "SecureBootPolicy"
	secureBootPolicyCustom    = "Custom"
	secureBootPolicyStandard  = "Standard"
)

// AllowCustomSecureBootKeys sets the SecureBootPolicy BIOS attribute to
// Custom or Standard.
//
// The attribute is read first only to reject platforms that don't expose it at all - not to
// skip the write when the currently *applied* value already matches what's requested.
// Currently-applied state can match while a different value is genuinely *pending* from an
// earlier call in the same boot cycle (e.g. a stale pending PATCH left staged by an interrupted,
// unrelated caller); confirmed live, skipping the write in that case silently left
// SecureBootPolicy staged as Standard despite a request to set it to Custom - the same class of
// bug stmcginnis/gofish#571 fixes one layer down, in SetBiosConfiguration's own diff baseline.
// This early return happens before SetBiosConfiguration is ever called, so #571 can't reach it.
// The write is staged into the
// Bios/Settings resource and only takes effect on the next POST, so a successful call always
// reports rebootRequired true.
//
// Implements bmc.CustomSecureBootKeysAllower.
func (c *Conn) AllowCustomSecureBootKeys(ctx context.Context, enable bool) (rebootRequired bool, err error) {
	biosConfig, err := c.redfishwrapper.GetBiosConfiguration(ctx)
	if err != nil {
		return false, err
	}

	if _, ok := biosConfig[secureBootPolicyAttribute]; !ok {
		return false, bmclibErrs.NewErrUnsupportedHardware(
			secureBootPolicyAttribute + " BIOS attribute not present: platform has no out-of-band Secure Boot key management control",
		)
	}

	want := secureBootPolicyStandard
	if enable {
		want = secureBootPolicyCustom
	}

	if err := c.redfishwrapper.SetBiosConfiguration(ctx, map[string]string{secureBootPolicyAttribute: want}); err != nil {
		return false, errors.Wrapf(err, "failed to set %s", secureBootPolicyAttribute)
	}

	return true, nil
}

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
// window, the later write fails with the same pending-settings conflict, naming the attribute it
// was trying to set even though that attribute was never touched before. The reverse order
// merges cleanly, only because no job yet exists when the resource PATCH runs (confirmed with a
// same-box, same-attributes, order-only-swapped A/B, on a PowerEdge R6715,
// iDRAC firmware 1.20.80.51).
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
