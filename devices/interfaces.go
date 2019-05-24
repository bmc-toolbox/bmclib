package devices

import (
	"crypto/x509"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// Bmc represents all the required bmc items
type Bmc interface {
	// Configure interface
	Configure

	// BmcCollection interface
	BmcCollection

	ApplyCfg(*cfgresources.ResourcesConfig) error
	CheckCredentials() error
	Close() error
	PowerOn() (bool, error)
	PowerOff() (bool, error)
	PxeOnce() (bool, error)
	PowerCycleBmc() (bool, error)
	PowerCycle() (bool, error)
	UpdateCredentials(string, string)
	UpdateFirmware(string, string) (bool, error)
}

// BmcCollection represents the requirement of items to be collected a server
type BmcCollection interface {
	BiosVersion() (string, error)
	BmcType() string
	BmcVersion() (string, error)
	CPU() (string, int, int, int, error)
	Disks() ([]*Disk, error)
	IsBlade() (bool, error)
	License() (string, string, error)
	Memory() (int, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*Nic, error)
	PowerKw() (float64, error)
	PowerState() (string, error)
	IsOn() (bool, error)
	Serial() (string, error)
	Status() (string, error)
	TempC() (int, error)
	Vendor() string
	Slot() (int, error)
	Screenshot() ([]byte, string, error)
	ServerSnapshot() (interface{}, error)
	ChassisSerial() (string, error)
}

// Cmc represents all the required cmc items
type Cmc interface {
	//  Configure interface
	Configure

	// CmcSetup interface
	CmcSetup

	// CmcCollection Interface
	CmcCollection

	ApplyCfg(*cfgresources.ResourcesConfig) error
	ChassisSnapshot() (*Chassis, error)
	CheckCredentials() error
	Close() error
	PowerCycle() (bool, error)
	PowerCycleBlade(int) (bool, error)
	PowerCycleBmcBlade(int) (bool, error)
	PowerOff() (bool, error)
	PowerOffBlade(int) (bool, error)
	PowerOn() (bool, error)
	PowerOnBlade(int) (bool, error)
	PxeOnceBlade(int) (bool, error)
	ReseatBlade(int) (bool, error)
	UpdateCredentials(string, string)
	UpdateFirmware(string, string) (bool, error)
}

// CmcCollection represents the requirement of items to be collected from a chassis
type CmcCollection interface {
	Blades() ([]*Blade, error)
	BmcType() string
	FindBladePosition(string) (int, error)
	FwVersion() (string, error)
	GetFirmwareVersion() (string, error)
	Fans() ([]*Fan, error)
	IsActive() bool
	IsOn() (bool, error)
	IsOnBlade(int) (bool, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*Nic, error)
	PassThru() (string, error)
	PowerKw() (float64, error)
	Psus() ([]*Psu, error)
	Serial() (string, error)
	Status() (string, error)
	IsPsuRedundant() (bool, error)
	PsuRedundancyMode() (string, error)
	StorageBlades() ([]*StorageBlade, error)
	TempC() (int, error)
	Vendor() string
}

// CmcSetup interface declares methods
// that are used to apply one time configuration to a Chassis.
type CmcSetup interface {
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
