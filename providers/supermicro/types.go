package supermicro

type IPMI struct {
	FruInfo *FruInfo `xml:"FRU_INFO,omitempty"`
}

// FruInfo contains the FRU information
type FruInfo struct {
	Board *Board `xml:"BOARD,omitempty"`
}

// Board contains the product baseboard information
type Board struct {
	MfcName   string `xml:"MFC_NAME,attr"`
	PartNum   string `xml:"PART_NUM,attr"`
	ProdName  string `xml:"PROD_NAME,attr"`
	SerialNum string `xml:"SERIAL_NUM,attr"`
}

type Supermicro struct {
	BIOS map[string]bool `json:"BIOS,omitempty"`
	BMC  map[string]bool `json:"BMC,omitempty"`
}

type OEM struct {
	Supermicro `json:"Supermicro"`
}
