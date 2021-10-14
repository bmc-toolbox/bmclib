package devices

import (
	"context"
	"crypto/x509"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// Bmc represents all the required bmc items
type Bmc interface {
	// Configure interface
	Configure

	// BmcCollection interface
	BmcCollection

	CheckCredentials() error
	Close(context.Context) error
	PowerOn() (bool, error)       // PowerSetter
	PowerOff() (bool, error)      // PowerSetter
	PxeOnce() (bool, error)       // BootDeviceSetter
	PowerCycleBmc() (bool, error) // BMCResetter
	PowerCycle() (bool, error)    // PowerSetter
	UpdateCredentials(string, string)
	UpdateFirmware(string, string) (bool, string, error)
	CheckFirmwareVersion() (string, error)
}

// BmcCollection represents the requirement of items to be collected a server
type BmcCollection interface {
	BiosVersion() (string, error)
	HardwareType() string // ilo4, ilo5, idrac8 or idrac9, etc
	Version() (string, error)
	CPU() (string, int, int, int, error)
	Disks() ([]*Disk, error)
	IsBlade() (bool, error)
	License() (string, string, error)
	Memory() (int, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*Nic, error)
	PowerKw() (float64, error)
	PowerState() (string, error) // PowerStateGetter
	IsOn() (bool, error)         // PowerStateGetter
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
	PowerCycle() (bool, error) // PowerSetter
	PowerCycleBlade(int) (bool, error)
	PowerCycleBmcBlade(int) (bool, error)
	PowerOff() (bool, error) // PowerSetter
	PowerOffBlade(int) (bool, error)
	PowerOn() (bool, error) // PowerSetter
	PowerOnBlade(int) (bool, error)
	PxeOnceBlade(int) (bool, error)
	ReseatBlade(int) (bool, error)
	UpdateCredentials(string, string)
	UpdateFirmware(string, string) (bool, string, error)
	CheckFirmwareVersion() (string, error)
}

// CmcCollection represents the requirement of items to be collected from a chassis
type CmcCollection interface {
	Blades() ([]*Blade, error)
	HardwareType() string
	FindBladePosition(string) (int, error)
	Version() (string, error)
	Fans() ([]*Fan, error)
	IsActive() bool
	IsOn() (bool, error) // PowerStateGetter
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
	User([]*cfgresources.User) error // UserCreator, UserUpdater, UserDeleter, UserReader
	Syslog(*cfgresources.Syslog) error
	Ntp(*cfgresources.Ntp) error
	Ldap(*cfgresources.Ldap) error
	LdapGroups([]*cfgresources.LdapGroup, *cfgresources.Ldap) error
	Network(*cfgresources.Network) (bool, error)
	SetLicense(*cfgresources.License) error
	Bios(*cfgresources.Bios) error
	Power(*cfgresources.Power) error
	CurrentHTTPSCert() ([]*x509.Certificate, bool, error)
	GenerateCSR(*cfgresources.HTTPSCertAttributes) ([]byte, error)
	UploadHTTPSCert([]byte, string, []byte, string) (bool, error)
}
