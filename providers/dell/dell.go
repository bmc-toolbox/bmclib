package dell

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/devices"
)

const (
	// VendorID represents the id of the vendor across all packages
	VendorID = devices.Dell
)

// CMC is the entry of the json exposed by dell
// We don't need to use an maps[string] with Chassis, because we don't have clusters
type CMC struct {
	Chassis *Chassis `json:"0"`
}

// CMCTemp is the entry of the json exposed by dell when reading the temp metrics
type CMCTemp struct {
	ChassisTemp *ChassisTemp `json:"1"`
}

// ChassisTemp is where the chassis thermal data is kept
type ChassisTemp struct {
	TempHealth                 int    `json:"TempHealth"`
	TempUpperCriticalThreshold int    `json:"TempUpperCriticalThreshold"`
	TempSensorID               int    `json:"TempSensorID"`
	TempCurrentValue           int    `json:"TempCurrentValue"`
	TempLowerCriticalThreshold int    `json:"TempLowerCriticalThreshold"`
	TempPresence               int    `json:"TempPresence"`
	TempSensorName             string `json:"TempSensorName"`
}

// Chassis groups all the interresting stuff we will ready from the chassis
type Chassis struct {
	ChassisGroupMemberHealthBlob *ChassisGroupMemberHealthBlob `json:"ChassisGroupMemberHealthBlob"`
}

// ChassisGroupMemberHealthBlob has a collection of metrics from the chassis, psu and blades
type ChassisGroupMemberHealthBlob struct {
	Blades        map[string]*Blade `json:"blades_status"`
	PsuStatus     *PsuStatus        `json:"psu_status"`
	ChassisStatus *ChassisStatus    `json:"chassis_status"`
	CMCStatus     *CMCStatus        `json:"cmc_status"`
	// TODO: active_alerts
}

// ChassisStatus expose the basic information that identify the chassis
type ChassisStatus struct {
	ROCmcFwVersionString string `json:"RO_cmc_fw_version_string"`
	ROChassisServiceTag  string `json:"RO_chassis_service_tag"`
	ROChassisProductname string `json:"RO_chassis_productname"`
	CHASSISName          string `json:"CHASSIS_name"`
}

// CMCStatus brings the information about the cmc status itself we will use it to know if the chassis has errors
type CMCStatus struct {
	CMCActiveError string `json:"cmcActiveError"`
}

// Nic is the nic we have on a servers
type Nic struct {
	BladeNicName string `json:"bladeNicName"`
	BladeNicVer  string `json:"bladeNicVer"`
}

// Blade contains all the blade information
type Blade struct {
	BladeTemperature    string          `json:"bladeTemperature"`
	BladePresent        int             `json:"bladePresent"`
	IdracURL            string          `json:"idracURL"`
	BladeLogDescription string          `json:"bladeLogDescription"`
	StorageNumDrives    int             `json:"storageNumDrives"`
	BladeCPUInfo        string          `json:"bladeCpuInfo"`
	Nics                map[string]*Nic `json:"nic"`
	BladeMasterSlot     int             `json:"bladeMasterSlot"`
	BladeUSCVer         string          `json:"bladeUSCVer"`
	BladeSvcTag         string          `json:"bladeSvcTag"`
	BladeBIOSver        string          `json:"bladeBIOSver"`
	ActualPwrConsump    int             `json:"actualPwrConsump"`
	BladePowerState     int             `json:"bladePowerStatus"`
	IsStorageBlade      int             `json:"isStorageBlade"`
	BladeModel          string          `json:"bladeModel"`
	BladeName           string          `json:"bladeName"`
	BladeSerialNum      string          `json:"bladeSerialNum"`
}

// PsuStatus contains the information and power usage of the psus
type PsuStatus struct {
	AcPower  string    `json:"acPower"`
	PsuCount int64     `json:"psuCount"`
	Psus     []PsuData `json:"-"`
}

// PsuData contains information of the psu
type PsuData struct {
	PsuPosition    string `json:"-"`
	PsuCapacity    int    `json:"psuCapacity"`
	PsuPresent     int    `json:"psuPresent"`
	PsuActiveError string `json:"psuActiveError"`
	PsuHealth      int    `json:"psuHealth"`
	PsuAcCurrent   string `json:"psuAcCurrent"`
	PsuAcVolts     string `json:"psuAcVolts"`
}

// UnmarshalJSON custom unmarshalling for this "special" data structure
func (d *PsuStatus) UnmarshalJSON(data []byte) error {
	var jsonMapping map[string]json.RawMessage
	if err := json.Unmarshal(data, &jsonMapping); err != nil {
		return err
	}

	rfct := reflect.ValueOf(d).Elem()
	rfctType := rfct.Type()

	// TODO(jumartinez): Juliano of the future, if you know by the time a better way of
	//                   doing this. Please refactor it!!.
	for key, value := range jsonMapping {
		for i := 0; i < rfctType.NumField(); i++ {
			if strings.HasPrefix(key, "psu_") {
				p := PsuData{}
				err := json.Unmarshal(value, &p)
				if err != nil {
					return err
				}
				p.PsuPosition = key
				d.Psus = append(d.Psus, p)
				break
			} else if key == rfctType.Field(i).Tag.Get("json") {
				var data interface{}
				err := json.Unmarshal(value, &data)
				if err != nil {
					return err
				}

				name := rfctType.Field(i).Name
				f := reflect.Indirect(rfct).FieldByName(name)

				switch f.Kind() {
				case reflect.String:
					f.SetString(data.(string))
				case reflect.Int64:
					d := int64(data.(float64))
					if !f.OverflowInt(d) {
						f.SetInt(d)
					}
				}
				break
			}
		}
	}

	sort.Slice(d.Psus, func(i, j int) bool {
		return d.Psus[i].PsuPosition < d.Psus[j].PsuPosition
	})

	return nil
}

// BladeMemoryEndpoint is the struct used to collect data from "https://$ip/sysmgmt/2012/server/memory" when passing the header X_SYSMGMT_OPTIMIZE:true
type BladeMemoryEndpoint struct {
	Memory *BladeMemory `json:"Memory"`
}

// BladeMemory is part of the payload returned by "https://$ip/sysmgmt/2012/server/memory"
type BladeMemory struct {
	Capacity       int `json:"capacity"`
	ErrCorrection  int `json:"err_correction"`
	MaxCapacity    int `json:"max_capacity"`
	SlotsAvailable int `json:"slots_available"`
	SlotsUsed      int `json:"slots_used"`
}

// BladeProcessorEndpoint is the struct used to collect data from "https://$ip/sysmgmt/2012/server/processor" when passing the header X_SYSMGMT_OPTIMIZE:true
type BladeProcessorEndpoint struct {
	Proccessors map[string]*BladeProcessor `json:"Processor"`
}

// BladeProcessor contains the processor data information
type BladeProcessor struct {
	Brand             string                 `json:"brand"`
	CoreCount         int                    `json:"core_count"`
	CurrentSpeed      int                    `json:"current_speed"`
	DeviceDescription string                 `json:"device_description"`
	HyperThreading    []*BladeHyperThreading `json:"hyperThreading"`
}

// BladeHyperThreading contains the hyperthread information
type BladeHyperThreading struct {
	Capable int `json:"capable"`
	Enabled int `json:"enabled"`
}

// IDracAuth is the struct used to verify the iDrac authentication
type IDracAuth struct {
	Status     string `xml:"status"`
	AuthResult int    `xml:"authResult" json:"authResult"`
	ForwardURL string `xml:"forwardUrl"`
	ErrorMsg   string `xml:"errorMsg"`
}

// IDracLicense is the struct used to collect data from "https://$ip/sysmgmt/2012/server/license" and it contains the license information for the bmc
type IDracLicense struct {
	License struct {
		VConsole int `json:"VCONSOLE"`
	} `json:"License"`
}

// IDracRoot is the structure used to render the data when querying -> https://$ip/data?get
type IDracRoot struct {
	BiosVer          string                 `xml:"biosVer"`
	FwVersion        string                 `xml:"fwVersion"`
	SysDesc          string                 `xml:"sysDesc"`
	Powermonitordata *IDracPowermonitordata `xml:"powermonitordata,omitempty"`
	PsSensorList     []*IDracPsSensor       `xml:"psSensorList>sensor,omitempty"`
}

// IDracPsSensorList contains a list of psu sensors
type IDracPsSensorList struct {
}

// IDracPsSensor contains the information regarding the psu devices
type IDracPsSensor struct {
	FwVersion    string `xml:"fwVersion,omitempty"`
	InputWattage int    `xml:"inputWattage,omitempty"`
	MaxWattage   int    `xml:"maxWattage,omitempty"`
	Name         string `xml:"name,omitempty"`
	SensorHealth int    `xml:"sensorHealth,omitempty"`
	SensorStatus int    `xml:"sensorStatus,omitempty"`
}

// IDracPowermonitordata contains the power consumption data for the iDrac
type IDracPowermonitordata struct {
	PresentReading *IDracPresentReading `xml:"presentReading,omitempty"`
}

// IDracPresentReading contains the present reading data
type IDracPresentReading struct {
	Reading *IDracReading `xml:"reading,omitempty"`
}

// IDracReading is used to express the power data
type IDracReading struct {
	ProbeName string `xml:"probeName,omitempty"`
	Reading   string `xml:"reading"`
}

// SVMInventory is the struct used to collect data from "https://$ip/sysmgmt/2012/server/inventory/software"
type SVMInventory struct {
	Device []*IDracDevice `xml:"Device"`
}

// IDracDevice contains the list of devices and their information
type IDracDevice struct {
	Display     string            `xml:"display,attr"`
	Application *IDracApplication `xml:"Application"`
}

// IDracApplication contains the name of the device and it's version
type IDracApplication struct {
	Display string `xml:"display,attr"`
	Version string `xml:"version,attr"`
}

// SystemServerOS contains the hostname, os name and os version
type SystemServerOS struct {
	SystemServerOS struct {
		HostName  string `json:"HostName"`
		OSName    string `json:"OSName"`
		OSVersion string `json:"OSVersion"`
	} `json:"system.ServerOS"`
}

// IDracInventory contains the whole hardware inventory exposed thru https://$ip/sysmgmt/2012/server/inventory/hardware
type IDracInventory struct {
	Version   string            `xml:"version,attr"`
	Component []*IDracComponent `xml:"Component,omitempty"`
}

// IDracComponent holds the information from each component detected by the iDrac
type IDracComponent struct {
	Classname  string           `xml:"Classname,attr"`
	Key        string           `xml:"Key,attr"`
	Properties []*IDracProperty `xml:"PROPERTY,omitempty"`
}

// IDracProperty is the property of each component exposed to iDrac
type IDracProperty struct {
	Name         string `xml:"NAME,attr"`
	Type         string `xml:"TYPE,attr"`
	DisplayValue string `xml:"DisplayValue,omitempty"`
	Value        string `xml:"VALUE,omitempty"`
}

// IDracTemp contains the data structure to render the thermal data from iDrac http://$ip/sysmgmt/2012/server/temperature
type IDracTemp struct {
	Statistics   string `json:"Statistics"`
	Temperatures struct {
		IDRACEmbedded1SystemBoardInletTemp struct {
			MaxFailure         int    `json:"max_failure"`
			MaxWarning         int    `json:"max_warning"`
			MaxWarningSettable int    `json:"max_warning_settable"`
			MinFailure         int    `json:"min_failure"`
			MinWarning         int    `json:"min_warning"`
			MinWarningSettable int    `json:"min_warning_settable"`
			Name               string `json:"name"`
			Reading            int    `json:"reading"`
			SensorStatus       int    `json:"sensor_status"`
		} `json:"iDRAC.Embedded.1#SystemBoardInletTemp"`
	} `json:"Temperatures"`
	IsFreshAirCompliant int `json:"is_fresh_air_compliant"`
}

// IDracHealthStatus contains the list of component status rendered from iDrac http://$ip/sysmgmt/2016/server/extended_health
type IDracHealthStatus struct {
	HealthStatus []int `json:"healthStatus"`
}

// IDracPowerData contains the power usage data from iDrac http://$ip/sysmgmt/2015/server/sensor/power
type IDracPowerData struct {
	Root struct {
		Powermonitordata struct {
			PresentReading struct {
				Reading struct {
					Reading float64 `json:"reading,string"`
				} `json:"reading"`
			} `json:"presentReading"`
		} `json:"powermonitordata"`
	} `json:"root"`
}

// CMCWWN is the structure used to render the data when querying /json?method=blades-wwn-info
type CMCWWN struct {
	SlotMacWwn CMCSlotMacWwn `json:"slot_mac_wwn"`
}

// CMCWWNBlade contains the blade structure used by CMCWWN
type CMCWWNBlade struct {
	BladeSlotName     string `json:"bladeSlotName"`
	IsNotDoubleHeight struct {
		IsInstalled string `json:"isInstalled"`
		PortPMAC    string `json:"portPMAC"`
		PortFMAC    string `json:"portFMAC"`
		IsSelected  int    `json:"isSelected"` //flexaddress enabled/disabled
	} `json:"is_not_double_height"`
}

// CMCSlotMacWwn contains index of blade by position inside of the chassis
type CMCSlotMacWwn struct {
	SlotMacWwnList map[int]CMCWWNBlade `json:"-"`
}

// UnmarshalJSON custom unmarshalling for this "special" data structure
func (d *CMCSlotMacWwn) UnmarshalJSON(data []byte) error {
	d.SlotMacWwnList = make(map[int]CMCWWNBlade, 0)
	var slotMacWwn map[string]json.RawMessage
	if err := json.Unmarshal(data, &slotMacWwn); err != nil {
		return err
	}

	if data, ok := slotMacWwn["slot_mac_wwn_list"]; ok {
		var slotMacWwnList map[string]json.RawMessage
		if err := json.Unmarshal(data, &slotMacWwnList); err != nil {
			return err
		}

		for slot, slotData := range slotMacWwnList {
			if pos, err := strconv.Atoi(slot); err == nil {
				var blade map[string]json.RawMessage
				if err := json.Unmarshal(slotData, &blade); err != nil {
					return err
				}

				b := CMCWWNBlade{}
				for key, value := range blade {
					switch key {
					case "bladeSlotName":
						if err := json.Unmarshal(value, &b.BladeSlotName); err != nil {
							return err
						}
					case "is_not_double_height":
						if err := json.Unmarshal(value, &b.IsNotDoubleHeight); err != nil {
							return err
						}
					}
				}
				d.SlotMacWwnList[pos] = b
			}
		}

	}

	return nil
}

// HwDetection used to render the content of session?aimGetProp=hostname,gui_str_title_bar,OEMHostName,fwVersion,sysDesc
type HwDetection struct {
	AimGetProp struct {
		Hostname       string `json:"hostname"`
		GuiStrTitleBar string `json:"gui_str_title_bar"`
		OEMHostName    string `json:"OEMHostName"`
		FwVersion      string `json:"fwVersion"`
		SysDesc        string `json:"sysDesc"`
		Status         string `json:"status"`
	} `json:"aimGetProp"`
}
