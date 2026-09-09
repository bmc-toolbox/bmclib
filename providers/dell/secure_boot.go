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
// The currently *applied* value is read first, and a request matching it returns early without
// writing. This only helps when nothing has touched SecureBootPolicy since the last reboot:
// applied state doesn't change until then no matter how many times this is called in between, so
// back-to-back calls in the same boot cycle still write every time regardless of this check.
//
// Known limitation: because the check reads applied state, not pending state, it cannot tell a
// genuinely-already-satisfied request apart from one where a *different* value is already staged
// as pending from an earlier, unrelated call in the same boot cycle - in that case this returns
// early and leaves the stale pending value in place instead of correcting it. The write, when
// attempted, is staged into the Bios/Settings resource and only takes effect on the next POST, so
// a successful change always reports rebootRequired true.
//
// Implements bmc.SecureBootKeyManagementSetter.
func (c *Conn) SetSecureBootKeyManagement(ctx context.Context, enable bool) (rebootRequired bool, err error) {
	biosConfig, err := c.redfishwrapper.GetBiosConfiguration(ctx)
	if err != nil {
		return false, err
	}

	current, ok := biosConfig[secureBootPolicyAttribute]
	if !ok {
		return false, bmclibErrs.NewErrUnsupportedHardware(
			secureBootPolicyAttribute + " BIOS attribute not present: platform has no out-of-band Secure Boot key management control",
		)
	}

	want := secureBootPolicyStandard
	if enable {
		want = secureBootPolicyCustom
	}

	if current == want {
		return false, nil
	}

	if err := c.redfishwrapper.SetBiosConfiguration(ctx, map[string]string{secureBootPolicyAttribute: want}); err != nil {
		return false, errors.Wrapf(err, "failed to set %s", secureBootPolicyAttribute)
	}

	return true, nil
}
