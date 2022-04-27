package devices

// Device type is composed of various components
type Device struct {
	Oem                bool                 `json:"oem"`
	HardwareType       string               `json:"hardware_type,omitempty"`
	Vendor             string               `json:"vendor,omitempty"`
	Model              string               `json:"model,omitempty"`
	Serial             string               `json:"serial,omitempty"`
	Chassis            string               `json:"chassis,omitempty"`
	BIOS               *BIOS                `json:"bios,omitempty"`
	BMC                *BMC                 `json:"bmc,omitempty"`
	Mainboard          *Mainboard           `json:"mainboard,omitempty"`
	CPLDs              []*CPLD              `json:"cplds"`
	TPMs               []*TPM               `json:"tpms,omitempty"`
	GPUs               []*GPU               `json:"gpus,omitempty"`
	CPUs               []*CPU               `json:"cpus,omitempty"`
	Memory             []*Memory            `json:"memory,omitempty"`
	NICs               []*NIC               `json:"nics,omitempty"`
	Drives             []*Drive             `json:"drives,omitempty"`
	StorageControllers []*StorageController `json:"storage_controller,omitempty"`
	PSUs               []*PSU               `json:"power_supplies,omitempty"`
	Enclosures         []*Enclosure         `json:"enclosures,omitempty"`
	Status             *Status              `json:"status,omitempty"`
	Metadata           map[string]string    `json:"metadata,omitempty"`
}

// NewDevice returns a pointer to an initialized Device type
func NewDevice() *Device {
	return &Device{
		BMC:                &BMC{NIC: &NIC{}},
		BIOS:               &BIOS{},
		Mainboard:          &Mainboard{},
		TPMs:               []*TPM{},
		CPLDs:              []*CPLD{},
		PSUs:               []*PSU{},
		NICs:               []*NIC{},
		GPUs:               []*GPU{},
		CPUs:               []*CPU{},
		Memory:             []*Memory{},
		Drives:             []*Drive{},
		StorageControllers: []*StorageController{},
		Enclosures:         []*Enclosure{},
		Status:             &Status{},
	}
}

// Firmware struct holds firmware attributes of a device component
type Firmware struct {
	Installed  string            `json:"installed,omitempty"`
	SoftwareID string            `json:"software_id,omitempty"`
	Previous   []*Firmware       `json:"previous,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// NewFirmwareObj returns a *Firmware object
func NewFirmwareObj() *Firmware {
	return &Firmware{Metadata: make(map[string]string)}
}

// Status is the health status of a component
type Status struct {
	Health         string
	State          string
	PostCode       int    `json:"post_code,omitempty"`
	PostCodeStatus string `json:"post_code_status,omitempty"`
}

// GPU component
type GPU struct {
}

// Enclosure component
type Enclosure struct {
	ID          string    `json:"id,omitempty"`
	Description string    `json:"description,omitempty"`
	ChassisType string    `json:"chassis_type,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	Model       string    `json:"model,omitempty"`
	Serial      string    `json:"serial,omitempty"`
	Firmware    *Firmware `json:"firmware,omitempty"`
	Status      *Status   `json:"status,omitempty"`
}

// TPM component
type TPM struct {
	InterfaceType string    `json:"interface_type,omitempty"`
	Firmware      *Firmware `json:"firmware,omitempty"`
	Status        *Status   `json:"status,omitempty"`
}

// CPLD component
type CPLD struct {
	Description string    `json:"description,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	Model       string    `json:"model,omitempty"`
	Serial      string    `json:"serial,omitempty"`
	Firmware    *Firmware `json:"firmware,omitempty"`
}

// PSU component
type PSU struct {
	ID                 string    `json:"id,omitempty"`
	Description        string    `json:"description,omitempty"`
	Vendor             string    `json:"vendor,omitempty"`
	Model              string    `json:"model,omitempty"`
	Serial             string    `json:"serial,omitempty"`
	PowerCapacityWatts int64     `json:"power_capacity_watts,omitempty"`
	Oem                bool      `json:"oem"`
	Status             *Status   `json:"status,omitempty"`
	Firmware           *Firmware `json:"firmware,omitempty"`
}

// BIOS component
type BIOS struct {
	Description   string    `json:"description,omitempty"`
	Vendor        string    `json:"vendor,omitempty"`
	Model         string    `json:"model,omitempty"`
	SizeBytes     int64     `json:"size_bytes,omitempty"`
	CapacityBytes int64     `json:"capacity_bytes,omitempty" diff:"immutable"`
	Firmware      *Firmware `json:"firmware,omitempty"`
}

// BMC component
type BMC struct {
	ID          string    `json:"id,omitempty"`
	Description string    `json:"description,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	Model       string    `json:"model,omitempty"`
	NIC         *NIC      `json:"nic,omitempty"`
	Status      *Status   `json:"status,omitempty"`
	Firmware    *Firmware `json:"firmware,omitempty"`
}

// CPU component
type CPU struct {
	ID           string    `json:"id,omitempty"`
	Description  string    `json:"description,omitempty"`
	Vendor       string    `json:"vendor,omitempty"`
	Model        string    `json:"model,omitempty"`
	Serial       string    `json:"serial,omitempty"`
	Slot         string    `json:"slot,omitempty"`
	Architecture string    `json:"architecture,omitempty"`
	ClockSpeedHz int64     `json:"clock_speeed_hz,omitempty"`
	Cores        int       `json:"cores,omitempty"`
	Threads      int       `json:"threads,omitempty"`
	Status       *Status   `json:"status,omitempty"`
	Firmware     *Firmware `json:"firmware,omitempty"`
}

// Memory component
type Memory struct {
	ID           string    `json:"id,omitempty"`
	Description  string    `json:"description,omitempty"`
	Slot         string    `json:"slot,omitempty"`
	Type         string    `json:"type,omitempty"`
	Vendor       string    `json:"vendor,omitempty"`
	Model        string    `json:"model,omitempty"`
	Serial       string    `json:"serial,omitempty"`
	SizeBytes    int64     `json:"size_bytes,omitempty"`
	FormFactor   string    `json:"form_factor,omitempty"`
	PartNumber   string    `json:"part_number,omitempty"`
	ClockSpeedHz int64     `json:"clock_speed_hz,omitempty"`
	Status       *Status   `json:"status,omitempty"`
	Firmware     *Firmware `json:"firmware,omitempty"`
}

// NIC component
type NIC struct {
	ID          string            `json:"id,omitempty"`
	Description string            `json:"description,omitempty"`
	Vendor      string            `json:"vendor,omitempty"`
	Model       string            `json:"model,omitempty"`
	Serial      string            `json:"serial,omitempty" diff:"identifier"`
	SpeedBits   int64             `json:"speed_bits,omitempty"`
	PhysicalID  string            `json:"physid,omitempty"`
	MacAddress  string            `json:"macaddress,omitempty"`
	Oem         bool              `json:"oem"`
	Metadata    map[string]string `json:"metadata"`
	Status      *Status           `json:"status,omitempty"`
	Firmware    *Firmware         `json:"firmware,omitempty"`
}

// StorageController component
type StorageController struct {
	ID                           string            `json:"id,omitempty"`
	Description                  string            `json:"description,omitempty"`
	Vendor                       string            `json:"vendor,omitempty"`
	Model                        string            `json:"model,omitempty"`
	Serial                       string            `json:"serial,omitempty"`
	SupportedControllerProtocols string            `json:"supported_controller_protocol,omitempty"` // PCIe
	SupportedDeviceProtocols     string            `json:"supported_device_protocol,omitempty"`     // Attached device protocols - SAS, SATA
	SupportedRAIDTypes           string            `json:"supported_raid_types,omitempty"`
	PhysicalID                   string            `json:"physid,omitempty"`
	SpeedGbps                    int64             `json:"speed_gbps,omitempty"`
	Oem                          bool              `json:"oem"`
	Status                       *Status           `json:"status,omitempty"`
	Metadata                     map[string]string `json:"metadata"`
	Firmware                     *Firmware         `json:"firmware,omitempty"`
}

// Mainboard component
type Mainboard struct {
	ProductName string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Vendor      string    `json:"vendor,omitempty"`
	Model       string    `json:"model,omitempty"`
	Serial      string    `json:"serial,omitempty"`
	PhysicalID  string    `json:"physid,omitempty"`
	Firmware    *Firmware `json:"firmware,omitempty"`
}

// Drive component
type Drive struct {
	ID                  string            `json:"id,omitempty"`
	ProductName         string            `json:"name,omitempty"`
	Type                string            `json:"drive_type,omitempty"`
	Description         string            `json:"description,omitempty"`
	Serial              string            `json:"serial,omitempty" diff:"identifier"`
	StorageController   string            `json:"storage_controller,omitempty"`
	Vendor              string            `json:"vendor,omitempty"`
	Model               string            `json:"model,omitempty"`
	WWN                 string            `json:"wwn,omitempty"`
	Protocol            string            `json:"protocol,omitempty"`
	CapacityBytes       int64             `json:"capacity_bytes,omitempty"`
	BlockSizeBytes      int64             `json:"block_size_bytes,omitempty"`
	CapableSpeedGbps    int64             `json:"capable_speed_gbps,omitempty"`
	NegotiatedSpeedGbps int64             `json:"negotiated_speed_gbps,omitempty"`
	Metadata            map[string]string `json:"metadata,omitempty"` // Additional metadata if any
	Oem                 bool              `json:"oem,omitempty"`      // Component is an OEM component
	Firmware            *Firmware         `json:"firmware,omitempty"`
	Status              *Status           `json:"status,omitempty"`
}
