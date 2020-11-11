package devices

import "context"

// PowerCommand type
type PowerCommand string

// Setting type
type Setting string

// DataType represents a data point of a BMC
type DataType string

// BootDevice represents next boot type for a machine
type BootDevice string

// BootOptions is the boot type and options
type BootOptions struct {
	Device     BootDevice
	Persistent bool
	EfiBoot    bool
}

type UserSetting struct {
	Name     string
	Password string
	Role     string
	Enabled  bool
}

const (
	// PowerOn action
	PowerOn PowerCommand = "PowerOn"
	// PowerOff action
	PowerOff PowerCommand = "PowerOff"
	// PowerCycle action
	PowerCycle PowerCommand = "PowerCycle"

	// NoneBoot as next boot device
	NoneBoot BootDevice = "none"
	// BiosBoot as next boot device
	BiosBoot BootDevice = "bios"
	// CdromBoot as next boot device
	CdromBoot BootDevice = "cdrom"
	// DiskBoot as next boot device
	DiskBoot BootDevice = "disk"
	// PxeBoot as next boot device
	PxeBoot BootDevice = "pxe"

	// User settings
	User Setting = "User"
	// Syslog setting
	Syslog Setting = "Syslog"
	// NTP setting
	NTP Setting = "NTP"

	// CPU data
	CPU DataType = "CPU"
	// Memory data
	Memory DataType = "Memory"
	// SystemTemp data
	SystemTemp DataType = "SystemTemp"
	// SystemState data
	SystemState DataType = "SystemState"
)

// Configuration represents a before and after
// BMC setting
type Configuration struct {
	ConfigBefore Config
	ConfigAfter  Config
}

// Config represents the data of a
// single configuration setting
type Config struct {
	Name  string
	Value interface{}
	Type  Setting
}

// DataPoint represents a given data point from
// a resource in a BMC
type DataPoint struct {
	Name  string
	Value interface{}
	Type  DataType
}

// Connection opening and closing
type Connection interface {
	// Open a connection to a BMC
	// return error if unable to connect
	Open(context.Context) error
	// Close a connection to a BMC
	// return error if unable to close the connection
	Close(context.Context) error
}

// PowerBootRequester actions
type PowerBootRequester interface {
	// PowerRequest will do a power actions against a BMC
	PowerRequest(context.Context, PowerCommand) (bool, error)
	// BootDeviceRequest will set the next boot device for a machine
	BootDeviceRequest(context.Context, BootOptions) (bool, error)
}

// Configurer for configuring a BMC
type Configurer interface {
	// Configure will set a BMC property
	Configure(context.Context, Config) (*Configuration, error)
}

// DataRequester for retrieving information about a BMC
type DataRequester interface {
	// DataRequest retrieves info about a specific BMC data point
	DataRequest(context.Context, DataType) (*DataPoint, error)
}

// BmcWorker performs queries and actions against a BMC
type BmcWorker interface {
	Connection
	PowerBootRequester
	Configurer
	DataRequester
}
