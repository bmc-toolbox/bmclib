package cfgresources

import (
	"encoding/json"
	"os/exec"
	"time"
)

// SetupChassis struct holds attributes for one time chassis setup.
type SetupChassis struct {
	FlexAddress         *flexAddress       `yaml:"flexAddress"`
	IpmiOverLan         *ipmiOverLan       `yaml:"ipmiOverLan"`
	DynamicPower        *dynamicPower      `yaml:"dynamicPower"`
	BladesPower         *bladesPower       `yaml:"bladesPower"`
	AddBladeBmcAdmins   []*BladeBmcAccount `yaml:"addBladeBmcAdmins"`
	RemoveBladeBmcUsers []*BladeBmcAccount `yaml:"removeBladeBmcUsers"`
}

// ResourcesConfig struct holds all the configuration to be applied.
type ResourcesConfig struct {
	Ldap         *Ldap         `yaml:"ldap"`
	LdapGroups   *LdapGroups   `yaml:"ldapGroups"`
	License      *License      `yaml:"license"`
	Network      *Network      `yaml:"network"`
	Syslog       *Syslog       `yaml:"syslog"`
	User         []*User       `yaml:"user"`
	HTTPSCert    *HTTPSCert    `yaml:"httpsCert"`
	Ntp          *Ntp          `yaml:"ntp"`
	Bios         *Bios         `yaml:"bios"`
	Power        *Power        `yaml:"power"`
	Supermicro   *Supermicro   `yaml:"supermicro"` // supermicro specific config, example of issue #34
	SetupChassis *SetupChassis `yaml:"setupChassis"`
}

// Power struct holds Power settings configuration for each vendor.
type Power struct {
	HPE *HPE `yaml:"hpe"`
}

// Bios struct holds bios configuration for each vendor.
type Bios struct {
	Dell *Dell `yaml:"dell"`
}

// BladeBmcAccount declares attributes for a Blade BMC user to be managed through the chassis.
type BladeBmcAccount struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

// Enable/Disable Virtual Mac addresses for blades in a chassis.
// FlexAddresses in M1000e jargon.
// Virtual connect in HP C7000 jargon.
type flexAddress struct {
	Enable bool `yaml:"enable"`
}

// Enable/Disable ipmi over lan
type ipmiOverLan struct {
	Enable bool `yaml:"enable"`
}

// 'Dynamic Power' in HP C7000 Jargon.
// 'DPSE' (dynamic PSU engagement) in M1000e Dell jargon.
type dynamicPower struct {
	Enable bool `yaml:"enable"`
}

// Ensure power state on all blades in chassis.
type bladesPower struct {
	Enable bool `yaml:"enable"`
}

// User struct holds a BMC user account configuration.
type User struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	Role     string `yaml:"role"`
	Enable   bool   `yaml:"enable,omitempty"`
}

// Syslog struct holds BMC syslog configuration.
type Syslog struct {
	Server string `yaml:"server"`
	Port   int    `yaml:"port,omitempty"`
	Enable bool   `yaml:"enable,omitempty"`
}

// Ldap struct holds BMC LDAP configuration.
type Ldap struct {
	Server         string `yaml:"server"`
	Port           int    `yaml:"port"`
	Enable         bool   `yaml:"enable"`
	Role           string `yaml:"role"`
	BaseDn         string `yaml:"baseDn"` // BaseDN is the starting point of the LDAP tree search.
	BindDn         string `yaml:"bindDn"` // BindDN is used to gain access to the LDAP tree.
	Group          string `yaml:"group"`
	GroupBaseDn    string `yaml:"groupBaseDn"`
	UserAttribute  string `yaml:"userAttribute"`
	GroupAttribute string `yaml:"groupAttribute"`
	SearchFilter   string `yaml:"searchFilter"`
}

// License struct holds BMC licensing configuration.
type License struct {
	Key string `yaml:"key"`
}

type LdapBin struct {
	Executor string `yaml:"executor"`
	Path     string `yaml:"path"`
}

// LdapGroups holds all group-related configuration parameters.
// ExtraGroups is used in combination with Bin to add more groups at runtime.
type LdapGroups struct {
	Bin              *LdapBin     `yaml:"bin"`
	Groups           []*LdapGroup `yaml:"groups"`
	ExtraAdminGroups []*LdapGroup `json:"admins"`
	ExtraUserGroups  []*LdapGroup `json:"users"`
}

// If you want to add extra groups at runtime using a script, you have
//   the option of specifying
//   * Bin.Executor: Usually /bin/sh or /bin/bash and the like
//   * Bin.Path: Path your actual script
// You get the serial of the asset and its vendor as two arguments
// If you want more, create a GitHub issue and we will take a look
func (l *LdapGroups) GetExtraGroups(serial, vendor string) (string, error) {
	if l.Bin.Path == "" {
		return "nothing", nil
	}

	cmd := exec.Command(l.Bin.Executor, l.Bin.Path, serial, vendor)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return string(stdout), err
	}

	err = json.Unmarshal(stdout, &l)
	if err != nil {
		return string(stdout), err
	}

	l.Groups = append(l.Groups, l.ExtraAdminGroups...)
	l.Groups = append(l.Groups, l.ExtraUserGroups...)

	return "success", nil
}

// LdapGroup struct holds a single BMC LDAP role group configuration.
type LdapGroup struct {
	Role        string `yaml:"role"`
	Group       string `yaml:"group"`
	GroupBaseDn string `yaml:"groupBaseDn"`
	Enable      bool   `yaml:"enable"`
}

// HTTPSCert struct holds BMC HTTPs cert configuration.
type HTTPSCert struct {
	// Renew cert if it will expire in this time period.
	RenewBeforeExpiry time.Duration `yaml:"renewBeforeExpiry"`
	// Validate these attributes when renewing certs
	ValidateAttributes []string             `yaml:"validateAttributes"`
	Attributes         *HTTPSCertAttributes `yaml:"attributes"`
}

// HTTPSCertAttributes declares attributes that are part of a cert.
type HTTPSCertAttributes struct {
	CommonName       string `yaml:"commonName"`
	OrganizationName string `yaml:"organizationName"`
	OrganizationUnit string `yaml:"organizationUnit"`
	Locality         string `yaml:"locality"`
	StateName        string `yaml:"stateName"`
	CountryCode      string `yaml:"countryCode"`
	Email            string `yaml:"email"`
	SubjectAltName   string `yaml:"subjectAltName"`
}

// Network struct holds BMC network configuration.
type Network struct {
	Hostname       string `yaml:"hostname"`
	DNSFromDHCP    bool   `yaml:"dnsFromDhcp"`
	SSHEnable      bool   `yaml:"sshEnable"`
	SSHPort        int    `yaml:"sshPort"`
	SolEnable      bool   `yaml:"solEnable"` // Serial over lan
	IpmiEnable     bool   `yaml:"ipmiEnable"`
	DhcpEnable     bool   `yaml:"dhcpEnable"`
	IpmiPort       int    `yaml:"ipmiPort"`
	KVMMediaPort   int    `yaml:"kvmMediaPort"`
	KVMConsolePort int    `yaml:"kvmConsolePort"`
	DDNSEnable     bool   `yaml:"ddnsEnable"`
}

// Ntp struct holds BMC NTP configuration.
type Ntp struct {
	Enable   bool   `yaml:"enable"`
	Server1  string `yaml:"server1"`
	Server2  string `yaml:"server2"`
	Server3  string `yaml:"server3"`
	Timezone string `yaml:"timezone"`
}
