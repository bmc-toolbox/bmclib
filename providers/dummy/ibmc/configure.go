package ibmc

import (
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

// LdapGroup method implements the Configure interface
func (i *Ibmc) LdapGroup(cfgGroup []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) error {
	return nil
}

// Network method implements the Configure interface
func (i *Ibmc) Network(cfg *cfgresources.Network) error {
	return nil
}

// SetLicense implements the Configure interface
func (i *Ibmc) SetLicense(*cfgresources.License) error {
	return nil
}

// Bios method implements the Configure interface
func (i *Ibmc) Bios(cfg *cfgresources.Bios) error {
	return nil
}
