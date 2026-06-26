package lenovo

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// licensesURI is the XCC License collection.
const licensesURI = "/redfish/v1/LicenseService/Licenses"

// compile-time assertion that the provider implements the interface.
var _ bmc.LicenseManager = (*Conn)(nil)

// Licenses returns the installed XCC licenses.
//
// Implements bmc.LicenseManager.
func (c *Conn) Licenses(ctx context.Context) ([]bmc.LicenseInfo, error) {
	var collection struct {
		Members []odataID `json:"Members"`
	}
	if err := c.getJSON(licensesURI, &collection); err != nil {
		// Some XCC firmware levels do not expose the LicenseService collection
		// and return 404 here. Treat an absent service as "no licenses" rather
		// than failing the read. TODO: resolve the per-firmware LicenseService
		// path variance (see examples/lenovo-smoketest/HARDWARE-TEST-PLAN.md
		// TC-BMC-4); other errors still propagate.
		if errors.Is(err, errResourceNotFound) {
			return []bmc.LicenseInfo{}, nil
		}
		return nil, err
	}

	out := make([]bmc.LicenseInfo, 0, len(collection.Members))
	for _, m := range collection.Members {
		var lic struct {
			ID            string `json:"Id"`
			EntitlementID string `json:"EntitlementId"`
			LicenseType   string `json:"LicenseType"`
			Removable     bool   `json:"Removable"`
			Status        struct {
				State string `json:"State"`
			} `json:"Status"`
		}
		if err := c.getJSON(m.ODataID, &lic); err != nil {
			return nil, err
		}

		out = append(out, bmc.LicenseInfo{
			ID:            lic.ID,
			EntitlementID: lic.EntitlementID,
			LicenseType:   lic.LicenseType,
			State:         lic.Status.State,
			Removable:     lic.Removable,
		})
	}

	return out, nil
}

// LicenseInstall installs a license by POSTing its license string to the
// License collection.
//
// Implements bmc.LicenseManager.
func (c *Conn) LicenseInstall(ctx context.Context, license string) error {
	payload := map[string]any{"LicenseString": license}
	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, licensesURI, payload, nil))
}

// LicenseDelete removes an installed license by Id.
//
// Implements bmc.LicenseManager.
func (c *Conn) LicenseDelete(ctx context.Context, licenseID string) error {
	if licenseID == "" {
		return fmt.Errorf("license id is required")
	}
	target, err := url.JoinPath(licensesURI, licenseID)
	if err != nil {
		return err
	}
	return checkResponse(c.redfishwrapper.Delete(target))
}
