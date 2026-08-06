package bmc

import "context"

// LicenseInfo describes an installed BMC license.
type LicenseInfo struct {
	// ID is the Redfish License resource Id.
	ID string
	// EntitlementID is the license entitlement identifier.
	EntitlementID string
	// LicenseType is the license type (e.g. "Production", "Trial").
	LicenseType string
	// State is the license status state (e.g. "Enabled").
	State string
	// Removable reports whether the license can be deleted.
	Removable bool
}

// LicenseManager is implemented by providers that can read, install and delete
// BMC licenses.
type LicenseManager interface {
	// Licenses returns the installed licenses.
	Licenses(ctx context.Context) ([]LicenseInfo, error)
	// LicenseInstall installs a license from its (typically base64-encoded)
	// license string.
	LicenseInstall(ctx context.Context, license string) error
	// LicenseDelete removes an installed license by its Id.
	LicenseDelete(ctx context.Context, licenseID string) error
}
