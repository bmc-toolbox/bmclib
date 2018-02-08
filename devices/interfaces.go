package devices

// Bmc represents the requirement of items to be collected a server
type Bmc interface {
	BiosVersion() (string, error)
	BmcType() (string, error)
	BmcVersion() (string, error)
	CPU() (string, int, int, int, error)
	IsBlade() (bool, error)
	License() (string, string, error)
	Login() error
	Logout() error
	Memory() (int, error)
	Model() (string, error)
	Name() (string, error)
	Nics() ([]*Nic, error)
	PowerKw() (float64, error)
	Serial() (string, error)
	Status() (string, error)
	TempC() (int, error)
}

// BmcChassis represents the requirement of items to be collected from a chassis
type BmcChassis interface {
	Blades() ([]*Blade, error)
	FwVersion() (string, error)
	Login() error
	Logout() error
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
}
