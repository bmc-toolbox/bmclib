package devices

import (
	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// Bmc represents the requirement of items to be collected a server
type Bmc interface {
	ApplyCfg(*cfgresources.ResourcesConfig) error
	BiosVersion() (string, error)
	BmcType() string
	BmcVersion() (string, error)
	CPU() (string, int, int, int, error)
	Disks() ([]*Disk, error)
	IsBlade() (bool, error)
	License() (string, string, error)
	Login() error
	Logout() error
	Memory() (int, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*Nic, error)
	PowerKw() (float64, error)
	PowerState() (string, error)
	Serial() (string, error)
	Status() (string, error)
	TempC() (int, error)
	Vendor() string
	ServerSnapshot() (interface{}, error)
	UpdateCredentials(string, string)
}

// BmcChassis represents the requirement of items to be collected from a chassis
type BmcChassis interface {
	ApplyCfg(*cfgresources.ResourcesConfig) error
	Blades() ([]*Blade, error)
	BmcType() string
	FwVersion() (string, error)
	Close() error
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*Nic, error)
	PassThru() (string, error)
	PowerKw() (float64, error)
	Psus() ([]*Psu, error)
	Serial() (string, error)
	Status() (string, error)
	IsActive() bool
	StorageBlades() ([]*StorageBlade, error)
	TempC() (int, error)
	Vendor() string
	ChassisSnapshot() (*Chassis, error)
	UpdateCredentials(string, string)
}
