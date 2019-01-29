package supermicro

import (
	"net/http"
)

const (
	// VendorID represents the id of the vendor across all packages
	VendorID = "Supermicro"
)

// IPMI is the base structure that holds the information on queries to https://$ip/cgi/ipmi.cgi
type IPMI struct {
	Bios         *Bios          `xml:"BIOS,omitempty"`
	CPU          []*CPU         `xml:"CPU,omitempty"`
	ConfigInfo   *ConfigInfo    `xml:"CONFIG_INFO,omitempty"`
	Dimm         []*Dimm        `xml:"DIMM,omitempty"`
	FruInfo      *FruInfo       `xml:"FRU_INFO,omitempty"`
	GenericInfo  *GenericInfo   `xml:"GENERIC_INFO,omitempty"`
	PlatformInfo *PlatformInfo  `xml:"PLATFORM_INFO,omitempty"`
	PowerSupply  []*PowerSupply `xml:"PowerSupply,omitempty"`
	PowerInfo    *PowerInfo     `xml:"POWER_INFO"`
	NodeInfo     *NodeInfo      `xml:"NodeInfo,omitempty"`
	BiosLicense  *BiosLicense   `xml:"BIOS_LINCESNE,omitempty" json:"BIOS_LINCESNE,omitempty"`
}

// Bios holds the bios information
type Bios struct {
	Date    string `xml:"REL_DATE,attr"`
	Vendor  string `xml:"VENDOR,attr"`
	Version string `xml:"VER,attr"`
}

// CPU holds the cpu information
type CPU struct {
	Core    string `xml:"CORE,attr"`
	Version string `xml:"VER,attr"`
}

// ConfigInfo holds the bmc configuration
type ConfigInfo struct {
	Hostname     *Hostname       `xml:"HOSTNAME,omitempty"`
	UserAccounts []*UserAccounts `xml:"USER,omitempty"`
}

// Hostname is the bmc hostname
type Hostname struct {
	Name string `xml:"NAME,attr"`
}

// UserAccounts contains the user account information
type UserAccounts struct {
	Name string `xml:"NAME,attr"`
}

// Dimm holds the ram information
type Dimm struct {
	Size string `xml:"SIZE,attr"`
}

// FruInfo holds the fru ipmi information (serial numbers and so on)
type FruInfo struct {
	Board   *Board   `xml:"BOARD,omitempty"`
	Chassis *Chassis `xml:"CHASSIS,omitempty"`
	Product *Product `xml:"PRODUCT,omitempty" json:"PRODUCT,omitempty"`
}

// Chassis holds the chassis information
type Chassis struct {
	PartNum   string `xml:"PART_NUM,attr"`
	SerialNum string `xml:"SERIAL_NUM,attr"`
}

// Board holds the mother board information
type Board struct {
	MfcName   string `xml:"MFC_NAME,attr"`
	PartNum   string `xml:"PART_NUM,attr"`
	ProdName  string `xml:"PROD_NAME,attr"`
	SerialNum string `xml:"SERIAL_NUM,attr"`
}

// Product hold the product information
type Product struct {
	SerialNum string `xml:"SERIAL_NUM,attr"  json:",omitempty"`
}

// GenericInfo holds the bmc information
type GenericInfo struct {
	Generic *Generic `xml:"GENERIC,omitempty"`
}

// Generic holds the bmc information
type Generic struct {
	BiosVersion   string `xml:"BIOS_VERSION,attr"`
	BmcIP         string `xml:"BMC_IP,attr"`
	BmcMac        string `xml:"BMC_MAC,attr"`
	IpmiFwVersion string `xml:"IPMIFW_VERSION,attr"`
}

// PlatformInfo holds the hardware related information
type PlatformInfo struct {
	BiosVersion string `xml:"BIOS_VERSION,attr"`
	MbMacAddr1  string `xml:"MB_MAC_ADDR1,attr"`
	MbMacAddr2  string `xml:"MB_MAC_ADDR2,attr"`
	MbMacAddr3  string `xml:"MB_MAC_ADDR3,attr"`
	MbMacAddr4  string `xml:"MB_MAC_ADDR4,attr"`
}

// PowerSupply holds the power supply information
type PowerSupply struct {
	Location  string `xml:"LOCATION,attr"`
	Status    string `xml:"STATUS,attr"`
	Unplugged string `xml:"UNPLUGGED,attr"`
}

// Reader holds the status and properties of a connection to a supermicro bmc
type Reader struct {
	ip       *string
	username *string
	password *string
	client   *http.Client
}

// NodeInfo contains a lists of boards in the chassis
type NodeInfo struct {
	Nodes []*Node `xml:"Node,omitempty"`
}

// Node contains the power and thermal information of each board in the chassis
type Node struct {
	IP          string `xml:"IP,attr"`
	Power       string `xml:"Power,attr"`
	PowerStatus string `xml:"PowerStatus,attr"`
	SystemTemp  string `xml:"SystemTemp,attr"`
}

// BiosLicense contains the license of bmc
type BiosLicense struct {
	Check string `xml:"CHECK,attr"`
}

// PowerInfo renders the current power state of the machine
type PowerInfo struct {
	Power struct {
		Status string `xml:"STATUS,attr"`
	} `xml:"POWER,omitempty"`
}
