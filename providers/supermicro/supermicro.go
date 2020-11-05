package supermicro

const (
	// VendorID represents the id of the vendor across all packages
	VendorID = "Supermicro"
)

// IPMI is the base structure that holds the information on queries to https://$ip/cgi/ipmi.cgi
type IPMI struct {
	*SmBiosInfo  `xml:",omitempty"`
	*Power       `xml:",omitempty"`
	ConfigInfo   *ConfigInfo    `xml:"CONFIG_INFO,omitempty"`
	FruInfo      *FruInfo       `xml:"FRU_INFO,omitempty"`
	GenericInfo  *GenericInfo   `xml:"GENERIC_INFO,omitempty"`
	PlatformInfo *PlatformInfo  `xml:"PLATFORM_INFO,omitempty"`
	Platform     *Platform      `xml:"Platform,omitempty"`
	PowerSupply  []*PowerSupply `xml:"PowerSupply,omitempty"`
	PowerInfo    *PowerInfo     `xml:"POWER_INFO"`
	NodeInfo     *NodeInfo      `xml:"NodeInfo,omitempty"`
	BiosLicense  *BiosLicense   `xml:"BIOS_LINCESNE,omitempty"`
	HealthInfo   *HealthInfo    `xml:"HEALTH_INFO,omitempty"`
	SensorInfo   *SensorInfo    `xml:"SENSOR_INFO,omitempty"`
}

// HealthInfo holds the health information
type HealthInfo struct {
	Health string `xml:"HEALTH,attr"`
}

type SmBiosInfo struct {
	Bios *Bios   `xml:"BIOS,omitempty"`
	Dimm []*Dimm `xml:"DIMM,omitempty"`
	CPU  []*CPU  `xml:"CPU,omitempty"`
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
	Generic       *Generic `xml:"GENERIC,omitempty"`
	BiosVersion   string   `xml:"BIOS_VERSION,attr"`
	BmcIP         string   `xml:"BMC_IP,attr"`
	BmcMac        string   `xml:"BMC_MAC,attr"`
	IpmiFwVersion string   `xml:"IPMIFW_VERSION,attr"`
}

// Generic holds the bmc information
type Generic struct {
	BiosVersion   string `xml:"BIOS_VERSION,attr"`
	BmcIP         string `xml:"BMC_IP,attr"`
	BmcMac        string `xml:"BMC_MAC,attr"`
	IpmiFwVersion string `xml:"IPMIFW_VERSION,attr"`
}

// Platform holds the information of the hardware type eg: fattwin or discrete
type Platform struct {
	MultiNode      string `xml:"EnMultiNode,attr"  json:",omitempty"`
	TwinNodeNumber string `xml:"TwinNodeNumber,attr"  json:",omitempty"`
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

// NodeInfo contains a lists of boards in the chassis
type NodeInfo struct {
	Nodes []*Node `xml:"Node,omitempty"`
}

// Node contains the power and thermal information of each board in the chassis
type Node struct {
	IP          string `xml:"IP,attr"`
	ID          int    `xml:"ID,attr"`
	Power       string `xml:"Power,attr"`
	PowerStatus string `xml:"PowerStatus,attr"`
	NodeSerial  string `xml:"NodeSerialNo,attr"`
	SystemTemp  string `xml:"SystemTemp,attr"`
}

// Power for x11 BMCs
type Power struct {
	POWER struct {
		HAVERAGE string `xml:"HAVERAGE,attr"`
		DAVERAGE string `xml:"DAVERAGE,attr"`
		WAVERAGE string `xml:"WAVERAGE,attr"`
		HMINIMUM string `xml:"HMINIMUM,attr"`
		DMINIMUM string `xml:"DMINIMUM,attr"`
		WMINIMUM string `xml:"WMINIMUM,attr"`
		HMINTIME string `xml:"HMINTIME,attr"`
		DMINTIME string `xml:"DMINTIME,attr"`
		WMINTIME string `xml:"WMINTIME,attr"`
		HMAXIMUM string `xml:"HMAXIMUM,attr"`
		DMAXIMUM string `xml:"DMAXIMUM,attr"`
		WMAXIMUM string `xml:"WMAXIMUM,attr"`
		HMAXTIME string `xml:"HMAXTIME,attr"`
		DMAXTIME string `xml:"DMAXTIME,attr"`
		WMAXTIME string `xml:"WMAXTIME,attr"`
	} `xml:"POWER"`
	HOUR struct {
		FMINS0 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS0"`
		FMINS1 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS1"`
		FMINS2 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS2"`
		FMINS3 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS3"`
		FMINS4 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS4"`
		FMINS5 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS5"`
		FMINS6 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS6"`
		FMINS7 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS7"`
		FMINS8 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS8"`
		FMINS9 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS9"`
		FMINS10 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS10"`
		FMINS11 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"FMINS11"`
	} `xml:"HOUR"`
	DAY struct {
		HOUR0 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR0"`
		HOUR1 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR1"`
		HOUR2 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR2"`
		HOUR3 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR3"`
		HOUR4 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR4"`
		HOUR5 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR5"`
		HOUR6 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR6"`
		HOUR7 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR7"`
		HOUR8 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR8"`
		HOUR9 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR9"`
		HOUR10 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR10"`
		HOUR11 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR11"`
		HOUR12 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR12"`
		HOUR13 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR13"`
		HOUR14 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR14"`
		HOUR15 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR15"`
		HOUR16 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR16"`
		HOUR17 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR17"`
		HOUR18 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR18"`
		HOUR19 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR19"`
		HOUR20 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR20"`
		HOUR21 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR21"`
		HOUR22 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR22"`
		HOUR23 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"HOUR23"`
	} `xml:"DAY"`
	WEEK struct {
		DAY0 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY0"`
		DAY1 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY1"`
		DAY2 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY2"`
		DAY3 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY3"`
		DAY4 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY4"`
		DAY5 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY5"`
		DAY6 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY6"`
		DAY7 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY7"`
		DAY8 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY8"`
		DAY9 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY9"`
		DAY10 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY10"`
		DAY11 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY11"`
		DAY12 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY12"`
		DAY13 struct {
			MAX string `xml:"MAX,attr"`
			AVR string `xml:"AVR,attr"`
			MIN string `xml:"MIN,attr"`
		} `xml:"DAY13"`
	} `xml:"WEEK"`
	NOW struct {
		MAX string `xml:"MAX,attr"`
		AVR string `xml:"AVR,attr"`
		MIN string `xml:"MIN,attr"`
	} `xml:"NOW"`
	PEAK struct {
		MAX      string `xml:"MAX,attr"`
		MIN      string `xml:"MIN,attr"`
		Current  string `xml:"Current,attr"`
		PMAXTIME string `xml:"PMAXTIME,attr"`
		PMINTIME string `xml:"PMINTIME,attr"`
	} `xml:"PEAK"`
	BBP struct {
		TIMEOUT    string `xml:"TIMEOUT,attr"`
		BBPSUPPORT string `xml:"BBPSUPPORT,attr"`
	} `xml:"BBP"`
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

// SensorInfo for x11 BMCs
type SensorInfo struct {
	SENSOR []struct {
		ID      string `xml:"ID,attr"`
		NUMBER  string `xml:"NUMBER,attr"`
		NAME    string `xml:"NAME,attr"`
		READING string `xml:"READING,attr"`
		OPTION  string `xml:"OPTION,attr"`
		UNR     string `xml:"UNR,attr"`
		UC      string `xml:"UC,attr"`
		UNC     string `xml:"UNC,attr"`
		LNC     string `xml:"LNC,attr"`
		LC      string `xml:"LC,attr"`
		LNR     string `xml:"LNR,attr"`
		STYPE   string `xml:"STYPE,attr"`
		RTYPE   string `xml:"RTYPE,attr"`
		ERTYPE  string `xml:"ERTYPE,attr"`
		UNIT1   string `xml:"UNIT1,attr"`
		UNIT    string `xml:"UNIT,attr"`
		L       string `xml:"L,attr"`
		M       string `xml:"M,attr"`
		B       string `xml:"B,attr"`
		RB      string `xml:"RB,attr"`
	} `xml:"SENSOR"`
}
