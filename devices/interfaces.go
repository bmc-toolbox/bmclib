package devices

import (
	"crypto/x509"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// Bmc represents the requirement of items to be collected a server
type Bmc interface {
	ApplyCfg(*cfgresources.ResourcesConfig) error
	// embed Configure interface
	Configure
	BiosVersion() (string, error)
	BmcType() string
	BmcVersion() (string, error)
	CPU() (string, int, int, int, error)
	CheckCredentials() error
	Disks() ([]*Disk, error)
	IsBlade() (bool, error)
	License() (string, string, error)
	Close() error
	Memory() (int, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*Nic, error)
	PowerKw() (float64, error)
	PowerState() (string, error)
	IsOn() (bool, error)
	PowerOn() (status bool, err error)
	PowerOff() (status bool, err error)
	PxeOnce() (status bool, err error)
	PowerCycleBmc() (status bool, err error)
	PowerCycle() (status bool, err error)
	Serial() (string, error)
	Status() (string, error)
	TempC() (int, error)
	Vendor() string
	Screenshot() ([]byte, string, error)
	ServerSnapshot() (interface{}, error)
	UpdateCredentials(string, string)
	UpdateFirmware(string, string) (bool, error)
}

// BmcChassis represents the requirement of items to be collected from a chassis
type BmcChassis interface {
	ApplyCfg(*cfgresources.ResourcesConfig) error
	// embed Configure interface
	Configure
	// embed BmcChassisSetup interface
	BmcChassisSetup
	Blades() ([]*Blade, error)
	BmcType() string
	ChassisSnapshot() (*Chassis, error)
	CheckCredentials() error
	Close() error
	FindBladePosition(string) (int, error)
	FwVersion() (string, error)
	GetFirmwareVersion() (string, error)
	IsActive() bool
	IsOn() (bool, error)
	IsOnBlade(int) (bool, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*Nic, error)
	PassThru() (string, error)
	PowerCycle() (bool, error)
	PowerCycleBlade(int) (bool, error)
	PowerCycleBmcBlade(int) (bool, error)
	PowerKw() (float64, error)
	PowerOff() (bool, error)
	PowerOffBlade(int) (bool, error)
	PowerOn() (bool, error)
	PowerOnBlade(int) (bool, error)
	Psus() ([]*Psu, error)
	PxeOnceBlade(int) (bool, error)
	ReseatBlade(int) (bool, error)
	Serial() (string, error)
	Status() (string, error)
	StorageBlades() ([]*StorageBlade, error)
	TempC() (int, error)
	UpdateCredentials(string, string)
	UpdateFirmware(string, string) (bool, error)
	Vendor() string
}

// BmcChassisSetup interface declares methods
// that are used to apply one time configuration to a Chassis.
type BmcChassisSetup interface {
	ResourcesSetup() []string
	RemoveBladeBmcUser(string) error
	AddBladeBmcAdmin(string, string) error
	ModBladeBmcUser(string, string) error
	SetDynamicPower(bool) (bool, error)
	SetIpmiOverLan(int, bool) (bool, error)
	SetFlexAddressState(int, bool) (bool, error)
}

// Configure interface declares methods implemented
// to apply configuration to BMCs.
type Configure interface {
	Resources() []string
	User([]*cfgresources.User) error
	Syslog(*cfgresources.Syslog) error
	Ntp(*cfgresources.Ntp) error
	Ldap(*cfgresources.Ldap) error
	LdapGroup([]*cfgresources.LdapGroup, *cfgresources.Ldap) error
	Network(*cfgresources.Network) (bool, error)
	SetLicense(*cfgresources.License) error
	Bios(*cfgresources.Bios) error
	CurrentHTTPSCert() ([]*x509.Certificate, bool, error)
	GenerateCSR(*cfgresources.HTTPSCertAttributes) ([]byte, error)
	UploadHTTPSCert([]byte, string, []byte, string) (bool, error)
}
