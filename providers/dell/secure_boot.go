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

// SetSecureBootKeyManagement sets the SecureBootPolicy BIOS attribute to
// Custom or Standard.
//
// The attribute is read first only to reject platforms that don't expose it at all - not to
// skip the write when the currently *applied* value already matches what's requested.
// Currently-applied state can match while a different value is genuinely *pending* from an
// earlier call in the same boot cycle (e.g. a stale pending PATCH left staged by an interrupted,
// unrelated caller); confirmed live, skipping the write in that case silently left
// SecureBootPolicy staged as Standard despite a request to set it to Custom, on the same class
// of bug UpdateBiosAttributesExact was written to fix in gofish. The write is staged into the
// Bios/Settings resource and only takes effect on the next POST, so a successful call always
// reports rebootRequired true.
//
// Implements bmc.SecureBootKeyManagementSetter.
func (c *Conn) SetSecureBootKeyManagement(ctx context.Context, enable bool) (rebootRequired bool, err error) {
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
