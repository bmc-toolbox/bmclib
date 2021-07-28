package ibmc

import (
	"crypto/x509"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// Resources returns a slice of supported resources and
// the order they are to be applied in.
func (i *Ibmc) Resources() []string {
	return []string{
		"user",
		"syslog",
		"ntp",
		"ldap",
		"ldap_group",
		"network",
	}
}

// Power implemented the Configure interface
func (i *Ibmc) Power(cfg *cfgresources.Power) (err error) {
	return err
}

// User method implements the Configure interface
func (i *Ibmc) User(cfg []*cfgresources.User) error {
	return nil
}

// Syslog method implements the Configure interface
func (i *Ibmc) Syslog(cfg *cfgresources.Syslog) error {
	return nil
}

// Ntp method implements the Configure interface
func (i *Ibmc) Ntp(cfg *cfgresources.Ntp) error {
	return nil
}

// Ldap method implements the Configure interface
func (i *Ibmc) Ldap(cfg *cfgresources.Ldap) error {
	return nil
}

// LdapGroups method implements the Configure interface
func (i *Ibmc) LdapGroups(cfgGroups []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) error {
	return nil
}

// Network method implements the Configure interface
func (i *Ibmc) Network(cfg *cfgresources.Network) (bool, error) {
	return false, nil
}

// SetLicense implements the Configure interface
func (i *Ibmc) SetLicense(*cfgresources.License) error {
	return nil
}

// Bios method implements the Configure interface
func (i *Ibmc) Bios(cfg *cfgresources.Bios) error {
	return nil
}

// GenerateCSR generates a CSR request on the BMC.
// GenerateCSR implements the Configure interface.
func (i *Ibmc) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {
	return []byte{}, nil
}

// UploadHTTPSCert uploads the given CRT cert,
// UploadHTTPSCert implements the Configure interface.
func (i *Ibmc) UploadHTTPSCert(cert []byte, certFileName string, key []byte, keyFileName string) (bool, error) {
	return false, nil
}

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// The bool value returned indicates if the BMC supports CSR generation.
// CurrentHTTPSCert implements the Configure interface.
func (i *Ibmc) CurrentHTTPSCert() ([]*x509.Certificate, bool, error) {
	return nil, false, nil
}
