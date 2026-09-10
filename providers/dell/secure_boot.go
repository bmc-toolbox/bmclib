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
// iDRAC exposes both and keeps them in sync, but only the BIOS Setup attribute fits Dell's BIOS
// staging model. Multiple Bios/Settings attribute PATCHes stack into one pending job fine,
// regardless of order (confirmed live, repeatedly, on a PowerEdge R6715, iDRAC firmware
// 1.20.80.51). The one exception: a PATCH to the SecureBoot resource seals that pending job
// against any further Bios/Settings PATCH - confirmed on the same box with a same-attributes,
// order-only-swapped A/B: Bios/Settings then the resource merges cleanly; the resource then
// Bios/Settings fails with IDRAC.2.14.SYS011. (A single earlier observation, on a different
// R6715 at iDRAC firmware 1.5.3, first flagged the resource PATCH as the point of failure but
// wasn't a controlled A/B.)
//
// That makes SetSecureBoot the one call in this package that can break an otherwise-safe
// sequence of BIOS-affecting writes, purely because of which resource it targets. PATCHing the
// SecureBoot BIOS Setup attribute directly via SetBiosConfiguration instead removes that
// asymmetry: every Dell BIOS-backed feature here now goes through the same resource and stacks
// freely with the others, in either order.
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
