package ibmc

import (
	"github.com/bmc-toolbox/bmclib/cfgresources"
)

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

// Timezone method implements the Configure interface
func (i *Ibmc) Timezone(timezone string) error {
	return nil
}

// Network method implements the Configure interface
func (i *Ibmc) Network(cfg *cfgresources.Network) error {
	return nil
}
