package hp

import "github.com/ncode/bmc/devices"

const (
	// VendorID represents the id of the vendor across all packages
	VendorID = devices.HP
)

// Blade contains the unmarshalled data from the hp chassis
type Blade struct {
	Bay             *Bay   `xml:" BAY,omitempty"`
	Bsn             string `xml:" BSN,omitempty"`
	MgmtIPAddr      string `xml:" MGMTIPADDR,omitempty"`
	MgmtType        string `xml:" MGMTPN,omitempty"`
	MgmtVersion     string `xml:" MGMTFWVERSION,omitempty"`
	Name            string `xml:" NAME,omitempty"`
	Type            string `xml:" TYPE,omitempty"`
	Power           *Power `xml:" POWER,omitempty"`
	Status          string `xml:" STATUS,omitempty"`
	Spn             string `xml:" SPN,omitempty"`
	Temp            *Temp  `xml:" TEMPS>TEMP,omitempty"`
	BladeRomVer     string `xml:" BLADEROMVER,omitempty"`
	AssociatedBlade int    `xml:" ASSOCIATEDBLADE,omitempty"`
}

// Bay contains the position of the blade within the chassis
type Bay struct {
	Connection int `xml:" CONNECTION,omitempty"`
}

// Infra2 is the data retrieved from the chassis xml interface that contains all components
type Infra2 struct {
	Addr         string        `xml:" ADDR,omitempty"`
	Blades       []*Blade      `xml:" BLADES>BLADE,omitempty"`
	Switches     []*Switch     `xml:" SWITCHES>SWITCH,omitempty"`
	ChassisPower *ChassisPower `xml:" POWER,omitempty"`
	Status       string        `xml:" STATUS,omitempty"`
	Temp         *Temp         `xml:" TEMPS>TEMP,omitempty"`
	EnclSn       string        `xml:" ENCL_SN,omitempty"`
	Pn           string        `xml:" PN,omitempty"`
	Encl         string        `xml:" ENCL,omitempty"`
	Rack         string        `xml:" RACK,omitempty"`
	Managers     []*Manager    `xml:" MANAGERS>MANAGER,omitempty"`
}

// MP contains the firmware version and the model of the chassis or blade
type MP struct {
	Pn   string `xml:" PN,omitempty"`
	Sn   string `xml:" SN,omitempty"`
	Fwri string `xml:" FWRI,omitempty"`
}

// Switch contains the type of the switch
type Switch struct {
	Spn string `xml:" SPN,omitempty"`
}

// Power contains the power information of a blade
type Power struct {
	PowerConsumed float64 `xml:" POWER_CONSUMED,omitempty"`
	PowerState    string  `xml:" POWER_STATE,omitempty"`
}

// ChassisPower contains the power information of the chassis
type ChassisPower struct {
	PowerConsumed float64        `xml:" POWER_CONSUMED,omitempty"`
	Powersupply   []*Powersupply `xml:" POWERSUPPLY,omitempty"`
}

// Rimp is the entry data structure for the chassis
type Rimp struct {
	Infra2 *Infra2 `xml:" INFRA2,omitempty"`
	MP     *MP     `xml:" MP,omitempty"`
}

// Manager hold the information of the manager board of the chassis
type Manager struct {
	MgmtIPAddr string `xml:" MGMTIPADDR,omitempty"`
	Role       string `xml:" ROLE,omitempty"`
	MacAddr    string `xml:" MACADDR,omitempty"`
	Status     string `xml:" STATUS,omitempty"`
	Name       string `xml:" NAME,omitempty"`
}

// Powersupply contains the data of the power supply of the chassis
type Powersupply struct {
	Sn           string  `xml:" SN,omitempty"`
	Status       string  `xml:" STATUS,omitempty"`
	Capacity     float64 `xml:" CAPACITY,omitempty"`
	ActualOutput float64 `xml:" ACTUALOUTPUT,omitempty"`
}

// Temp contains the thermal data of a chassis or blade
type Temp struct {
	C    int    `xml:" C,omitempty" json:"C,omitempty"`
	Desc string `xml:" DESC,omitempty"`
}

// RimpBlade is the entry data structure for the blade when queries directly
type RimpBlade struct {
	MP          *MP          `xml:" MP,omitempty"`
	HSI         *HSI         `xml:" HSI,omitempty"`
	BladeSystem *BladeSystem `xml:" BLADESYSTEM,omitempty"`
}

// BladeSystem blade information from the hprimp of blades
type BladeSystem struct {
	Bay int `xml:" BAY,omitempty"`
}

// HSI contains the information about the components of the blade
type HSI struct {
	NICS []*NIC `xml:" NICS>NIC,omitempty"`
	Sbsn string `xml:" SBSN,omitempty" json:"SBSN,omitempty"`
	Spn  string `xml:" SPN,omitempty" json:"SPN,omitempty"`
}

// NIC contains the nic information of a blade
type NIC struct {
	Description string `xml:" DESCRIPTION,omitempty"`
	MacAddr     string `xml:" MACADDR,omitempty"`
	Status      string `xml:" STATUS,omitempty"`
}

// Firmware is the struct used to render the data from https://$ip/json/fw_info, it contains firmware data of the blade
type Firmware struct {
	Firmware []struct {
		FwName    string `json:"fw_name"`
		FwVersion string `json:"fw_version"`
	} `json:"firmware"`
}

// Procs is the struct used to render the data from https://$ip/json/proc_info, it contains the processor data
type Procs struct {
	Processors []struct {
		ProcName       string `json:"proc_name"`
		ProcNumCores   int    `json:"proc_num_cores"`
		ProcNumThreads int    `json:"proc_num_threads"`
	} `json:"processors"`
}

// Mem is the struct used to render the data from https://$ip/json/mem_info, it contains the ram data
type Mem struct {
	MemTotalMemSize int        `json:"mem_total_mem_size"`
	Memory          []*MemSlot `json:"memory"`
}

// MemSlot is part of the payload returned from https://$ip/json/mem_info
type MemSlot struct {
	MemDevLoc string `json:"mem_dev_loc"`
	MemSize   int    `json:"mem_size"`
	MemSpeed  int    `json:"mem_speed"`
}

// Overview is the struct used to render the data from https://$ip/json/overview, it contains information about bios version, ilo license and a bit more
type Overview struct {
	ServerName    string `json:"server_name"`
	ProductName   string `json:"product_name"`
	SerialNum     string `json:"serial_num"`
	SystemRom     string `json:"system_rom"`
	SystemRomDate string `json:"system_rom_date"`
	BackupRomDate string `json:"backup_rom_date"`
	License       string `json:"license"`
	IloFwVersion  string `json:"ilo_fw_version"`
	IPAddress     string `json:"ip_address"`
	SystemHealth  string `json:"system_health"`
	Power         string `json:"power"`
}

// PowerSummary is the struct used to render the data from https://$ip/json/power_summary, it contains the basic information about the power usage of the machine
type PowerSummary struct {
	HostpwrState          string `json:"hostpwr_state"`
	PowerSupplyInputPower int    `json:"power_supply_input_power"`
}

// HealthTemperature is the struct used to render the data from https://$ip/json/health_temperature, it contains the information about the thermal status of the machine
type HealthTemperature struct {
	HostpwrState string         `json:"hostpwr_state"`
	InPost       int            `json:"in_post"`
	Temperature  []*Temperature `json:"temperature"`
}

// Temperature is part of the data rendered from https://$ip/json/health_temperature, it contains the names of each component and their current temp
type Temperature struct {
	Label          string `json:"label"`
	Location       string `json:"location"`
	Status         string `json:"status"`
	Currentreading int    `json:"currentreading"`
	TempUnit       string `json:"temp_unit"`
}

// IloLicense is the struct used to render the data from https://$ip/json/license, it contains the license information of the ilo
type IloLicense struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// IloPowerSupply holds the information of power supplies exposed via ilo
type IloPowerSupply struct {
	Supplies []struct {
		Unhealthy        int    `json:"unhealthy"`
		Enabled          int    `json:"enabled"`
		Mismatch         int    `json:"mismatch"`
		PsBay            int    `json:"ps_bay"`
		PsPresent        string `json:"ps_present"`
		PsCondition      string `json:"ps_condition"`
		PsErrorCode      string `json:"ps_error_code"`
		PsIpduCapable    string `json:"ps_ipdu_capable"`
		PsHotplugCapable string `json:"ps_hotplug_capable"`
		PsModel          string `json:"ps_model"`
		PsSpare          string `json:"ps_spare"`
		PsSerialNum      string `json:"ps_serial_num"`
		PsMaxCapWatts    int    `json:"ps_max_cap_watts"`
		PsFwVer          string `json:"ps_fw_ver"`
		PsInputVolts     int    `json:"ps_input_volts"`
		PsOutputWatts    int    `json:"ps_output_watts"`
		Avg              int    `json:"avg"`
		Max              int    `json:"max"`
		Supply           bool   `json:"supply"`
		Bbu              bool   `json:"bbu"`
		Charge           int    `json:"charge"`
		Age              int    `json:"age"`
		BatteryHealth    int    `json:"battery_health"`
	} `json:"supplies"`
	PresentPowerReading int `json:"present_power_reading"`
}

// IloDisks is the struct used to render the data from https://$ip/json/health_phy_drives, it contains the list of disks and their current health state
type IloDisks struct {
	PhyDriveArrays []struct {
		PhysicalDrives []struct {
			Name           string `json:"name"`
			Status         string `json:"status"`
			SerialNo       string `json:"serial_no"`
			Model          string `json:"model"`
			Capacity       string `json:"capacity"`
			Location       string `json:"location"`
			FwVersion      string `json:"fw_version"`
			PhysStatus     string `json:"phys_status"`
			DriveType      string `json:"drive_type"`
			EncrStat       string `json:"encr_stat"`
			PhysIdx        int    `json:"phys_idx"`
			DriveMediatype string `json:"drive_mediatype"`
		} `json:"physical_drives"`
	} `json:"phy_drive_arrays"`
}
