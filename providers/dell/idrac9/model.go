package idrac9

type userInfo map[int]User
type idracUsers map[string]userInfo

// User struct declares user configuration payload.
type User struct {
	UserName               string `json:"UserName,omitempty"`
	Password               string `json:"Password,omitempty"`
	Enable                 string `json:"Enable,omitempty"`                 //Enabled, Disabled
	Privilege              string `json:"Privilege,omitempty"`              //511, 499
	IpmiLanPrivilege       string `json:"IpmiLanPrivilege,omitempty"`       //Administrator, Operator
	SolEnable              string `json:"SolEnable,omitempty"`              //Disabled, Enabled
	ProtocolEnable         string `json:"ProtocolEnable,omitempty"`         //Disabled, Enabled (SNMPv2)
	AuthenticationProtocol string `json:"AuthenticationProtocol,omitempty"` //SHA, MD5, None
	PrivacyProtocol        string `json:"PrivacyProtocol,omitempty"`        //AES, DES, None
}

// Ldap struct declares Ldap configuration payload.
type Ldap struct {
	BaseDN               string `json:"BaseDN"`               //dell
	BindDN               string `json:"BindDN"`               //cn=dell
	CertValidationEnable string `json:"CertValidationEnable"` //Disabled
	Enable               string `json:"Enable"`               //Enabled
	GroupAttribute       string `json:"GroupAttribute"`       //memberUid
	GroupAttributeIsDN   string `json:"GroupAttributeIsDN"`   //Enabled
	Port                 string `json:"Port"`                 //636
	SearchFilter         string `json:"SearchFilter"`         //objectClass=posixAccount
	Server               string `json:"Server"`               //ldap.example.com"
	UserAttribute        string `json:"UserAttribute"`        //uid
}

type idracLdapRoleGroups map[string]LdapRoleGroups

// LdapRoleGroups declares the format in which ldap role groups are un/marshalled.
type LdapRoleGroups map[string]LdapRoleGroup

// LdapRoleGroup declares Ldap role group configuration payload.
type LdapRoleGroup struct {
	DN        string `json:"DN"`        //cn=dell,cn=bmcAdmins
	Privilege string `json:"Privilege"` //511 (Administrator), 499 (Operator)
}

// Syslog declares syslog configuration payload.
type Syslog struct {
	Enable  string `json:"SysLogEnable"`
	Server1 string `json:"Server1"`
	Server2 string `json:"Server2"`
	Server3 string `json:"Server3"`
	Port    string `json:"Port"`
}

// NtpConfig declares NTP configuration payload.
type NtpConfig struct {
	Enable string `json:"NTPEnable"` //Enabled
	NTP1   string `json:"NTP1"`      //example0.ntp.com
	NTP2   string `json:"NTP2"`      //example1.ntp.com
	NTP3   string `json:"NTP3"`      //example2.ntp.com
}

// Ipv4 declares IPv4 configuration payload.
type Ipv4 struct {
	Enable      string `json:"Enable"`      //Enabled
	DHCPEnable  string `json:"DHCPEnable"`  //Enabled
	DNSFromDHCP string `json:"DNSFromDHCP"` //Enabled
}

// IpmiOverLan declares IpmiOverLan configuration payload.
type IpmiOverLan struct {
	Enable        string `json:"Enable"`        //Enabled
	PrivLimit     string `json:"PrivLimit"`     //Administrator
	EncryptionKey string `json:"EncryptionKey"` //0000000000000000000000000000000000000000
}

// SerialRedirection declares serial console configuration payload.
type SerialRedirection struct {
	Enable  string `json:"Enable"`  //Enabled
	QuitKey string `json:"QuitKey"` //^\\
}

// SerialOverLan declares serial over lan configuration payload.
type SerialOverLan struct {
	Enable       string `json:"Enable"`       //Enabled
	BaudRate     string `json:"BaudRate"`     //115200
	MinPrivilege string `json:"MinPrivilege"` //Administrator
}

// Timezone declares timezone configuration payload.
type Timezone struct {
	Timezone string `json:"Timezone"` //CET
}

// CSRInfo declares SSL/TLS CSR request payloads.
type CSRInfo struct {
	CommonName       string `json:"CsrCommonName"`
	CountryCode      string `json:"CsrCountryCode"`
	LocalityName     string `json:"CsrLocalityName"`
	OrganizationName string `json:"CsrOrganizationName"`
	OrganizationUnit string `json:"CsrOrganizationUnit"`
	StateName        string `json:"CsrStateName"`
	EmailAddr        string `json:"CsrEmailAddr"`
	SubjectAltName   string `json:"CsrSubjectAltName"`
}

// certStore is the response received when uploading a multipart form,
// that includes the certificate, this cert is stored in a transient store.
// {"File":{"ResourceURI":"/var/volatile/tmp/upload/idrac9.crt"}}
type certStore struct {
	File struct {
		ResourceURI string `json:"ResourceURI"`
	} `json:"File"`
}

// AlertEnable is the payload to enable/disable Alerts
type AlertEnable struct {
	Enabled string `json:"AlertEnable"`
}

//  alertConfigPayload basically sets up the alert configuration,
// 1. disable informational messages and enables warning/critical messages
// 2. send all of these messages to remote syslog
// If we are to enable events to be sent over Email or SNMP or Redfish, this would need updating
// endpoint: /sysmgmt/2012/server/eventpolicy
// method: PUT
var alertConfigPayload = []byte(`{
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#AMP_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "AMP"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#BAT_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "BAT"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#BAT_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "BAT"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#BAT_2_2": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "BAT"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#BAT_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "BAT"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#CPU_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "CPU"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#CTL_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "CTL"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#CTL_2_2": {
	  "filter_actions": 259,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "CTL"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#CTL_2_3": {
	  "filter_actions": 259,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "CTL"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#CTL_5_3": {
	  "filter_actions": 256,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "CTL"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#DIS_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "DIS"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#ENC_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "ENC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#ENC_2_2": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "ENC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#ENC_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "ENC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#FAN_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "FAN"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#FAN_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "FAN"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#FAN_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "FAN"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#FC_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "FC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#GMGR_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "GMGR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#HWC_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "HWC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#IOID_5_2": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 2,
	  "subcategory": "IOID"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#IOID_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "IOID"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#IOV_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "IOV"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#IOV_5_3": {
	  "filter_actions": 256,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "IOV"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#IPA_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "IPA"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#JCP_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "JCP"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#LNK_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "LNK"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#MEM_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "MEM"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#NIC_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "NIC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PCI_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "PCI"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PCI_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "PCI"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PDR_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "PDR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PDR_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "PDR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PDR_2_2": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "PDR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PDR_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "PDR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PSU_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "PSU"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PSU_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "PSU"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PSU_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "PSU"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PWR_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "PWR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#PWR_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "PWR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RAC_3_3": {
	  "filter_actions": 257,
	  "category": 3,
	  "severity": 3,
	  "subcategory": "RAC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RAC_5_2": {
	  "filter_actions": 256,
	  "category": 5,
	  "severity": 2,
	  "subcategory": "RAC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RAC_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "RAC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RDU_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "RDU"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RED_3_1": {
	  "filter_actions": 257,
	  "category": 3,
	  "severity": 1,
	  "subcategory": "RED"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RED_3_2": {
	  "filter_actions": 257,
	  "category": 3,
	  "severity": 2,
	  "subcategory": "RED"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RED_3_3": {
	  "filter_actions": 257,
	  "category": 3,
	  "severity": 3,
	  "subcategory": "RED"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RFL_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "RFL"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#RRDU_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "RRDU"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SEC_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "SEC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SEC_2_2": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "SEC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SEC_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "SEC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SEC_5_2": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 2,
	  "subcategory": "SEC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SSD_2_2": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "SSD"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#STOR_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "STOR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#STOR_2_2": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "STOR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#STOR_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "STOR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SWC_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "SWC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SWC_5_1": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 1,
	  "subcategory": "SWC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SWC_5_2": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 2,
	  "subcategory": "SWC"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#SWU_3_2": {
	  "filter_actions": 257,
	  "category": 3,
	  "severity": 2,
	  "subcategory": "SWU"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#TMP_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "TMP"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#TMP_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "TMP"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#TMP_2_2": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "TMP"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#TMP_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "TMP"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#UEFI_3_2": {
	  "filter_actions": 257,
	  "category": 3,
	  "severity": 2,
	  "subcategory": "UEFI"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#UEFI_3_3": {
	  "filter_actions": 257,
	  "category": 3,
	  "severity": 3,
	  "subcategory": "UEFI"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#UEFI_5_1": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 1,
	  "subcategory": "UEFI"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#UEFI_5_3": {
	  "filter_actions": 257,
	  "category": 5,
	  "severity": 3,
	  "subcategory": "UEFI"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#VDR_2_1": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 1,
	  "subcategory": "VDR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#VDR_2_2": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 2,
	  "subcategory": "VDR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#VDR_2_3": {
	  "filter_actions": 257,
	  "category": 2,
	  "severity": 3,
	  "subcategory": "VDR"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#VFLA_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "VFLA"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#VFL_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "VFL"
	},
	"iDRAC.Embedded.1#RACEvtFilterCfgRoot#VLT_1_3": {
	  "filter_actions": 1,
	  "category": 1,
	  "severity": 3,
	  "subcategory": "VLT"
	}
  }`)

// Timezones declares all known timezones, taken from the idrac web interface.
var Timezones = map[string]string{
	"Africa/Abidjan":                   "Africa/Abidjan",
	"Africa/Accra":                     "Africa/Accra",
	"Africa/Addis_Ababa":               "Africa/Addis_Ababa",
	"Africa/Algiers":                   "Africa/Algiers",
	"Africa/Asmara":                    "Africa/Asmara",
	"Africa/Asmera":                    "Africa/Asmera",
	"Africa/Bamako":                    "Africa/Bamako",
	"Africa/Bangui":                    "Africa/Bangui",
	"Africa/Banjul":                    "Africa/Banjul",
	"Africa/Bissau":                    "Africa/Bissau",
	"Africa/Blantyre":                  "Africa/Blantyre",
	"Africa/Brazzaville":               "Africa/Brazzaville",
	"Africa/Bujumbura":                 "Africa/Bujumbura",
	"Africa/Cairo":                     "Africa/Cairo",
	"Africa/Casablanca":                "Africa/Casablanca",
	"Africa/Ceuta":                     "Africa/Ceuta",
	"Africa/Conakry":                   "Africa/Conakry",
	"Africa/Dakar":                     "Africa/Dakar",
	"Africa/Dar_es_Salaam":             "Africa/Dar_es_Salaam",
	"Africa/Djibouti":                  "Africa/Djibouti",
	"Africa/Douala":                    "Africa/Douala",
	"Africa/El_Aaiun":                  "Africa/El_Aaiun",
	"Africa/Freetown":                  "Africa/Freetown",
	"Africa/Gaborone":                  "Africa/Gaborone",
	"Africa/Harare":                    "Africa/Harare",
	"Africa/Johannesburg":              "Africa/Johannesburg",
	"Africa/Juba":                      "Africa/Juba",
	"Africa/Kampala":                   "Africa/Kampala",
	"Africa/Khartoum":                  "Africa/Khartoum",
	"Africa/Kigali":                    "Africa/Kigali",
	"Africa/Kinshasa":                  "Africa/Kinshasa",
	"Africa/Lagos":                     "Africa/Lagos",
	"Africa/Libreville":                "Africa/Libreville",
	"Africa/Lome":                      "Africa/Lome",
	"Africa/Luanda":                    "Africa/Luanda",
	"Africa/Lubumbashi":                "Africa/Lubumbashi",
	"Africa/Lusaka":                    "Africa/Lusaka",
	"Africa/Malabo":                    "Africa/Malabo",
	"Africa/Maputo":                    "Africa/Maputo",
	"Africa/Maseru":                    "Africa/Maseru",
	"Africa/Mbabane":                   "Africa/Mbabane",
	"Africa/Mogadishu":                 "Africa/Mogadishu",
	"Africa/Monrovia":                  "Africa/Monrovia",
	"Africa/Nairobi":                   "Africa/Nairobi",
	"Africa/Ndjamena":                  "Africa/Ndjamena",
	"Africa/Niamey":                    "Africa/Niamey",
	"Africa/Nouakchott":                "Africa/Nouakchott",
	"Africa/Ouagadougou":               "Africa/Ouagadougou",
	"Africa/Porto-Novo":                "Africa/Porto-Novo",
	"Africa/Sao_Tome":                  "Africa/Sao_Tome",
	"Africa/Timbuktu":                  "Africa/Timbuktu",
	"Africa/Tripoli":                   "Africa/Tripoli",
	"Africa/Tunis":                     "Africa/Tunis",
	"Africa/Windhoek":                  "Africa/Windhoek",
	"America/Adak":                     "America/Adak",
	"America/Anchorage":                "America/Anchorage",
	"America/Anguilla":                 "America/Anguilla",
	"America/Antigua":                  "America/Antigua",
	"America/Araguaina":                "America/Araguaina",
	"America/Argentina/Buenos_Aires":   "America/Argentina/Buenos_Aires",
	"America/Argentina/Catamarca":      "America/Argentina/Catamarca",
	"America/Argentina/ComodRivadavia": "America/Argentina/ComodRivadavia",
	"America/Argentina/Cordoba":        "America/Argentina/Cordoba",
	"America/Argentina/Jujuy":          "America/Argentina/Jujuy",
	"America/Argentina/La_Rioja":       "America/Argentina/La_Rioja",
	"America/Argentina/Mendoza":        "America/Argentina/Mendoza",
	"America/Argentina/Rio_Gallegos":   "America/Argentina/Rio_Gallegos",
	"America/Argentina/Salta":          "America/Argentina/Salta",
	"America/Argentina/San_Juan":       "America/Argentina/San_Juan",
	"America/Argentina/San_Luis":       "America/Argentina/San_Luis",
	"America/Argentina/Tucuman":        "America/Argentina/Tucuman",
	"America/Argentina/Ushuaia":        "America/Argentina/Ushuaia",
	"America/Aruba":                    "America/Aruba",
	"America/Asuncion":                 "America/Asuncion",
	"America/Atikokan":                 "America/Atikokan",
	"America/Atka":                     "America/Atka",
	"America/Bahia":                    "America/Bahia",
	"America/Bahia_Banderas":           "America/Bahia_Banderas",
	"America/Barbados":                 "America/Barbados",
	"America/Belem":                    "America/Belem",
	"America/Belize":                   "America/Belize",
	"America/Blanc-Sablon":             "America/Blanc-Sablon",
	"America/Boa_Vista":                "America/Boa_Vista",
	"America/Bogota":                   "America/Bogota",
	"America/Boise":                    "America/Boise",
	"America/Buenos_Aires":             "America/Buenos_Aires",
	"America/Cambridge_Bay":            "America/Cambridge_Bay",
	"America/Campo_Grande":             "America/Campo_Grande",
	"America/Cancun":                   "America/Cancun",
	"America/Caracas":                  "America/Caracas",
	"America/Catamarca":                "America/Catamarca",
	"America/Cayenne":                  "America/Cayenne",
	"America/Cayman":                   "America/Cayman",
	"America/Chicago":                  "America/Chicago",
	"America/Chihuahua":                "America/Chihuahua",
	"America/Coral_Harbour":            "America/Coral_Harbour",
	"America/Cordoba":                  "America/Cordoba",
	"America/Costa_Rica":               "America/Costa_Rica",
	"America/Cuiaba":                   "America/Cuiaba",
	"America/Curacao":                  "America/Curacao",
	"America/Danmarkshavn":             "America/Danmarkshavn",
	"America/Dawson":                   "America/Dawson",
	"America/Dawson_Creek":             "America/Dawson_Creek",
	"America/Denver":                   "America/Denver",
	"America/Detroit":                  "America/Detroit",
	"America/Dominica":                 "America/Dominica",
	"America/Edmonton":                 "America/Edmonton",
	"America/Eirunepe":                 "America/Eirunepe",
	"America/El_Salvador":              "America/El_Salvador",
	"America/Ensenada":                 "America/Ensenada",
	"America/Fort_Wayne":               "America/Fort_Wayne",
	"America/Fortaleza":                "America/Fortaleza",
	"America/Glace_Bay":                "America/Glace_Bay",
	"America/Godthab":                  "America/Godthab",
	"America/Goose_Bay":                "America/Goose_Bay",
	"America/Grand_Turk":               "America/Grand_Turk",
	"America/Grenada":                  "America/Grenada",
	"America/Guadeloupe":               "America/Guadeloupe",
	"America/Guatemala":                "America/Guatemala",
	"America/Guayaquil":                "America/Guayaquil",
	"America/Guyana":                   "America/Guyana",
	"America/Halifax":                  "America/Halifax",
	"America/Havana":                   "America/Havana",
	"America/Hermosillo":               "America/Hermosillo",
	"America/Indiana/Indianapolis":     "America/Indiana/Indianapolis",
	"America/Indiana/Knox":             "America/Indiana/Knox",
	"America/Indiana/Marengo":          "America/Indiana/Marengo",
	"America/Indiana/Petersburg":       "America/Indiana/Petersburg",
	"America/Indiana/Tell_City":        "America/Indiana/Tell_City",
	"America/Indiana/Vevay":            "America/Indiana/Vevay",
	"America/Indiana/Vincennes":        "America/Indiana/Vincennes",
	"America/Indiana/Winamac":          "America/Indiana/Winamac",
	"America/Indianapolis":             "America/Indianapolis",
	"America/Inuvik":                   "America/Inuvik",
	"America/Iqaluit":                  "America/Iqaluit",
	"America/Jamaica":                  "America/Jamaica",
	"America/Jujuy":                    "America/Jujuy",
	"America/Juneau":                   "America/Juneau",
	"America/Kentucky/Louisville":      "America/Kentucky/Louisville",
	"America/Kentucky/Monticello":      "America/Kentucky/Monticello",
	"America/Knox_IN":                  "America/Knox_IN",
	"America/Kralendijk":               "America/Kralendijk",
	"America/La_Paz":                   "America/La_Paz",
	"America/Lima":                     "America/Lima",
	"America/Los_Angeles":              "America/Los_Angeles",
	"America/Louisville":               "America/Louisville",
	"America/Lower_Princes":            "America/Lower_Princes",
	"America/Maceio":                   "America/Maceio",
	"America/Managua":                  "America/Managua",
	"America/Manaus":                   "America/Manaus",
	"America/Marigot":                  "America/Marigot",
	"America/Martinique":               "America/Martinique",
	"America/Matamoros":                "America/Matamoros",
	"America/Mazatlan":                 "America/Mazatlan",
	"America/Mendoza":                  "America/Mendoza",
	"America/Menominee":                "America/Menominee",
	"America/Merida":                   "America/Merida",
	"America/Metlakatla":               "America/Metlakatla",
	"America/Mexico_City":              "America/Mexico_City",
	"America/Miquelon":                 "America/Miquelon",
	"America/Moncton":                  "America/Moncton",
	"America/Monterrey":                "America/Monterrey",
	"America/Montevideo":               "America/Montevideo",
	"America/Montreal":                 "America/Montreal",
	"America/Montserrat":               "America/Montserrat",
	"America/Nassau":                   "America/Nassau",
	"America/New_York":                 "America/New_York",
	"America/Nipigon":                  "America/Nipigon",
	"America/Nome":                     "America/Nome",
	"America/Noronha":                  "America/Noronha",
	"America/North_Dakota/Beulah":      "America/North_Dakota/Beulah",
	"America/North_Dakota/Center":      "America/North_Dakota/Center",
	"America/North_Dakota/New_Salem":   "America/North_Dakota/New_Salem",
	"America/Ojinaga":                  "America/Ojinaga",
	"America/Panama":                   "America/Panama",
	"America/Pangnirtung":              "America/Pangnirtung",
	"America/Paramaribo":               "America/Paramaribo",
	"America/Phoenix":                  "America/Phoenix",
	"America/Port-au-Prince":           "America/Port-au-Prince",
	"America/Port_of_Spain":            "America/Port_of_Spain",
	"America/Porto_Acre":               "America/Porto_Acre",
	"America/Porto_Velho":              "America/Porto_Velho",
	"America/Puerto_Rico":              "America/Puerto_Rico",
	"America/Rainy_River":              "America/Rainy_River",
	"America/Rankin_Inlet":             "America/Rankin_Inlet",
	"America/Recife":                   "America/Recife",
	"America/Regina":                   "America/Regina",
	"America/Resolute":                 "America/Resolute",
	"America/Rio_Branco":               "America/Rio_Branco",
	"America/Rosario":                  "America/Rosario",
	"America/Santa_Isabel":             "America/Santa_Isabel",
	"America/Santarem":                 "America/Santarem",
	"America/Santiago":                 "America/Santiago",
	"America/Santo_Domingo":            "America/Santo_Domingo",
	"America/Sao_Paulo":                "America/Sao_Paulo",
	"America/Scoresbysund":             "America/Scoresbysund",
	"America/Shiprock":                 "America/Shiprock",
	"America/Sitka":                    "America/Sitka",
	"America/St_Barthelemy":            "America/St_Barthelemy",
	"America/St_Johns":                 "America/St_Johns",
	"America/St_Kitts":                 "America/St_Kitts",
	"America/St_Lucia":                 "America/St_Lucia",
	"America/St_Thomas":                "America/St_Thomas",
	"America/St_Vincent":               "America/St_Vincent",
	"America/Swift_Current":            "America/Swift_Current",
	"America/Tegucigalpa":              "America/Tegucigalpa",
	"America/Thule":                    "America/Thule",
	"America/Thunder_Bay":              "America/Thunder_Bay",
	"America/Tijuana":                  "America/Tijuana",
	"America/Toronto":                  "America/Toronto",
	"America/Tortola":                  "America/Tortola",
	"America/Vancouver":                "America/Vancouver",
	"America/Virgin":                   "America/Virgin",
	"America/Whitehorse":               "America/Whitehorse",
	"America/Winnipeg":                 "America/Winnipeg",
	"America/Yakutat":                  "America/Yakutat",
	"America/Yellowknife":              "America/Yellowknife",
	"Antarctica/Casey":                 "Antarctica/Casey",
	"Antarctica/Davis":                 "Antarctica/Davis",
	"Antarctica/DumontDUrville":        "Antarctica/DumontDUrville",
	"Antarctica/Macquarie":             "Antarctica/Macquarie",
	"Antarctica/Mawson":                "Antarctica/Mawson",
	"Antarctica/McMurdo":               "Antarctica/McMurdo",
	"Antarctica/Palmer":                "Antarctica/Palmer",
	"Antarctica/Rothera":               "Antarctica/Rothera",
	"Antarctica/South_Pole":            "Antarctica/South_Pole",
	"Antarctica/Syowa":                 "Antarctica/Syowa",
	"Antarctica/Vostok":                "Antarctica/Vostok",
	"Asia/Aden":                        "Asia/Aden",
	"Asia/Almaty":                      "Asia/Almaty",
	"Asia/Amman":                       "Asia/Amman",
	"Asia/Anadyr":                      "Asia/Anadyr",
	"Asia/Aqtau":                       "Asia/Aqtau",
	"Asia/Aqtobe":                      "Asia/Aqtobe",
	"Asia/Ashgabat":                    "Asia/Ashgabat",
	"Asia/Ashkhabad":                   "Asia/Ashkhabad",
	"Asia/Baghdad":                     "Asia/Baghdad",
	"Asia/Bahrain":                     "Asia/Bahrain",
	"Asia/Baku":                        "Asia/Baku",
	"Asia/Bangkok":                     "Asia/Bangkok",
	"Asia/Beirut":                      "Asia/Beirut",
	"Asia/Bishkek":                     "Asia/Bishkek",
	"Asia/Brunei":                      "Asia/Brunei",
	"Asia/Calcutta":                    "Asia/Calcutta",
	"Asia/Choibalsan":                  "Asia/Choibalsan",
	"Asia/Chongqing":                   "Asia/Chongqing",
	"Asia/Chungking":                   "Asia/Chungking",
	"Asia/Colombo":                     "Asia/Colombo",
	"Asia/Dacca":                       "Asia/Dacca",
	"Asia/Damascus":                    "Asia/Damascus",
	"Asia/Dhaka":                       "Asia/Dhaka",
	"Asia/Dili":                        "Asia/Dili",
	"Asia/Dubai":                       "Asia/Dubai",
	"Asia/Dushanbe":                    "Asia/Dushanbe",
	"Asia/Gaza":                        "Asia/Gaza",
	"Asia/Harbin":                      "Asia/Harbin",
	"Asia/Ho_Chi_Minh":                 "Asia/Ho_Chi_Minh",
	"Asia/Hong_Kong":                   "Asia/Hong_Kong",
	"Asia/Hovd":                        "Asia/Hovd",
	"Asia/Irkutsk":                     "Asia/Irkutsk",
	"Asia/Istanbul":                    "Asia/Istanbul",
	"Asia/Jakarta":                     "Asia/Jakarta",
	"Asia/Jayapura":                    "Asia/Jayapura",
	"Asia/Jerusalem":                   "Asia/Jerusalem",
	"Asia/Kabul":                       "Asia/Kabul",
	"Asia/Kamchatka":                   "Asia/Kamchatka",
	"Asia/Karachi":                     "Asia/Karachi",
	"Asia/Kashgar":                     "Asia/Kashgar",
	"Asia/Kathmandu":                   "Asia/Kathmandu",
	"Asia/Katmandu":                    "Asia/Katmandu",
	"Asia/Kolkata":                     "Asia/Kolkata",
	"Asia/Krasnoyarsk":                 "Asia/Krasnoyarsk",
	"Asia/Kuala_Lumpur":                "Asia/Kuala_Lumpur",
	"Asia/Kuching":                     "Asia/Kuching",
	"Asia/Kuwait":                      "Asia/Kuwait",
	"Asia/Macao":                       "Asia/Macao",
	"Asia/Macau":                       "Asia/Macau",
	"Asia/Magadan":                     "Asia/Magadan",
	"Asia/Makassar":                    "Asia/Makassar",
	"Asia/Manila":                      "Asia/Manila",
	"Asia/Muscat":                      "Asia/Muscat",
	"Asia/Nicosia":                     "Asia/Nicosia",
	"Asia/Novokuznetsk":                "Asia/Novokuznetsk",
	"Asia/Novosibirsk":                 "Asia/Novosibirsk",
	"Asia/Omsk":                        "Asia/Omsk",
	"Asia/Oral":                        "Asia/Oral",
	"Asia/Phnom_Penh":                  "Asia/Phnom_Penh",
	"Asia/Pontianak":                   "Asia/Pontianak",
	"Asia/Pyongyang":                   "Asia/Pyongyang",
	"Asia/Qatar":                       "Asia/Qatar",
	"Asia/Qyzylorda":                   "Asia/Qyzylorda",
	"Asia/Rangoon":                     "Asia/Rangoon",
	"Asia/Riyadh":                      "Asia/Riyadh",
	"Asia/Saigon":                      "Asia/Saigon",
	"Asia/Sakhalin":                    "Asia/Sakhalin",
	"Asia/Samarkand":                   "Asia/Samarkand",
	"Asia/Seoul":                       "Asia/Seoul",
	"Asia/Shanghai":                    "Asia/Shanghai",
	"Asia/Singapore":                   "Asia/Singapore",
	"Asia/Taipei":                      "Asia/Taipei",
	"Asia/Tashkent":                    "Asia/Tashkent",
	"Asia/Tbilisi":                     "Asia/Tbilisi",
	"Asia/Tehran":                      "Asia/Tehran",
	"Asia/Tel_Aviv":                    "Asia/Tel_Aviv",
	"Asia/Thimbu":                      "Asia/Thimbu",
	"Asia/Thimphu":                     "Asia/Thimphu",
	"Asia/Tokyo":                       "Asia/Tokyo",
	"Asia/Ujung_Pandang":               "Asia/Ujung_Pandang",
	"Asia/Ulaanbaatar":                 "Asia/Ulaanbaatar",
	"Asia/Ulan_Bator":                  "Asia/Ulan_Bator",
	"Asia/Urumqi":                      "Asia/Urumqi",
	"Asia/Vientiane":                   "Asia/Vientiane",
	"Asia/Vladivostok":                 "Asia/Vladivostok",
	"Asia/Yakutsk":                     "Asia/Yakutsk",
	"Asia/Yekaterinburg":               "Asia/Yekaterinburg",
	"Asia/Yerevan":                     "Asia/Yerevan",
	"Atlantic/Azores":                  "Atlantic/Azores",
	"Atlantic/Bermuda":                 "Atlantic/Bermuda",
	"Atlantic/Canary":                  "Atlantic/Canary",
	"Atlantic/Cape_Verde":              "Atlantic/Cape_Verde",
	"Atlantic/Faeroe":                  "Atlantic/Faeroe",
	"Atlantic/Faroe":                   "Atlantic/Faroe",
	"Atlantic/Jan_Mayen":               "Atlantic/Jan_Mayen",
	"Atlantic/Madeira":                 "Atlantic/Madeira",
	"Atlantic/Reykjavik":               "Atlantic/Reykjavik",
	"Atlantic/South_Georgia":           "Atlantic/South_Georgia",
	"Atlantic/St_Helena":               "Atlantic/St_Helena",
	"Atlantic/Stanley":                 "Atlantic/Stanley",
	"Australia/ACT":                    "Australia/ACT",
	"Australia/Adelaide":               "Australia/Adelaide",
	"Australia/Brisbane":               "Australia/Brisbane",
	"Australia/Broken_Hill":            "Australia/Broken_Hill",
	"Australia/Canberra":               "Australia/Canberra",
	"Australia/Currie":                 "Australia/Currie",
	"Australia/Darwin":                 "Australia/Darwin",
	"Australia/Eucla":                  "Australia/Eucla",
	"Australia/Hobart":                 "Australia/Hobart",
	"Australia/LHI":                    "Australia/LHI",
	"Australia/Lindeman":               "Australia/Lindeman",
	"Australia/Lord_Howe":              "Australia/Lord_Howe",
	"Australia/Melbourne":              "Australia/Melbourne",
	"Australia/NSW":                    "Australia/NSW",
	"Australia/North":                  "Australia/North",
	"Australia/Perth":                  "Australia/Perth",
	"Australia/Queensland":             "Australia/Queensland",
	"Australia/South":                  "Australia/South",
	"Australia/Sydney":                 "Australia/Sydney",
	"Australia/Tasmania":               "Australia/Tasmania",
	"Australia/Victoria":               "Australia/Victoria",
	"Australia/West":                   "Australia/West",
	"Australia/Yancowinna":             "Australia/Yancowinna",
	"Brazil/Acre":                      "Brazil/Acre",
	"Brazil/DeNoronha":                 "Brazil/DeNoronha",
	"Brazil/East":                      "Brazil/East",
	"Brazil/West":                      "Brazil/West",
	"CET":                              "CET",
	"CST6CDT":                          "CST6CDT",
	"Canada/Atlantic":                  "Canada/Atlantic",
	"Canada/Central":                   "Canada/Central",
	"Canada/East-Saskatchewan":         "Canada/East-Saskatchewan",
	"Canada/Eastern":                   "Canada/Eastern",
	"Canada/Mountain":                  "Canada/Mountain",
	"Canada/Newfoundland":              "Canada/Newfoundland",
	"Canada/Pacific":                   "Canada/Pacific",
	"Canada/Saskatchewan":              "Canada/Saskatchewan",
	"Canada/Yukon":                     "Canada/Yukon",
	"Chile/Continental":                "Chile/Continental",
	"Chile/EasterIsland":               "Chile/EasterIsland",
	"Cuba":                             "Cuba",
	"EET":                              "EET",
	"EST":                              "EST",
	"EST5EDT":                          "EST5EDT",
	"Egypt":                            "Egypt",
	"Eire":                             "Eire",
	"Etc/GMT":                          "Etc/GMT",
	"Etc/GMT+0":                        "Etc/GMT+0",
	"Etc/GMT+1":                        "Etc/GMT+1",
	"Etc/GMT+10":                       "Etc/GMT+10",
	"Etc/GMT+11":                       "Etc/GMT+11",
	"Etc/GMT+12":                       "Etc/GMT+12",
	"Etc/GMT+2":                        "Etc/GMT+2",
	"Etc/GMT+3":                        "Etc/GMT+3",
	"Etc/GMT+4":                        "Etc/GMT+4",
	"Etc/GMT+5":                        "Etc/GMT+5",
	"Etc/GMT+6":                        "Etc/GMT+6",
	"Etc/GMT+7":                        "Etc/GMT+7",
	"Etc/GMT+8":                        "Etc/GMT+8",
	"Etc/GMT+9":                        "Etc/GMT+9",
	"Etc/GMT-0":                        "Etc/GMT-0",
	"Etc/GMT-1":                        "Etc/GMT-1",
	"Etc/GMT-10":                       "Etc/GMT-10",
	"Etc/GMT-11":                       "Etc/GMT-11",
	"Etc/GMT-12":                       "Etc/GMT-12",
	"Etc/GMT-13":                       "Etc/GMT-13",
	"Etc/GMT-14":                       "Etc/GMT-14",
	"Etc/GMT-2":                        "Etc/GMT-2",
	"Etc/GMT-3":                        "Etc/GMT-3",
	"Etc/GMT-4":                        "Etc/GMT-4",
	"Etc/GMT-5":                        "Etc/GMT-5",
	"Etc/GMT-6":                        "Etc/GMT-6",
	"Etc/GMT-7":                        "Etc/GMT-7",
	"Etc/GMT-8":                        "Etc/GMT-8",
	"Etc/GMT-9":                        "Etc/GMT-9",
	"Etc/GMT0":                         "Etc/GMT0",
	"Etc/Greenwich":                    "Etc/Greenwich",
	"Etc/UCT":                          "Etc/UCT",
	"Etc/UTC":                          "Etc/UTC",
	"Etc/Universal":                    "Etc/Universal",
	"Etc/Zulu":                         "Etc/Zulu",
	"Europe/Amsterdam":                 "Europe/Amsterdam",
	"Europe/Andorra":                   "Europe/Andorra",
	"Europe/Athens":                    "Europe/Athens",
	"Europe/Belfast":                   "Europe/Belfast",
	"Europe/Belgrade":                  "Europe/Belgrade",
	"Europe/Berlin":                    "Europe/Berlin",
	"Europe/Bratislava":                "Europe/Bratislava",
	"Europe/Brussels":                  "Europe/Brussels",
	"Europe/Bucharest":                 "Europe/Bucharest",
	"Europe/Budapest":                  "Europe/Budapest",
	"Europe/Chisinau":                  "Europe/Chisinau",
	"Europe/Copenhagen":                "Europe/Copenhagen",
	"Europe/Dublin":                    "Europe/Dublin",
	"Europe/Gibraltar":                 "Europe/Gibraltar",
	"Europe/Guernsey":                  "Europe/Guernsey",
	"Europe/Helsinki":                  "Europe/Helsinki",
	"Europe/Isle_of_Man":               "Europe/Isle_of_Man",
	"Europe/Istanbul":                  "Europe/Istanbul",
	"Europe/Jersey":                    "Europe/Jersey",
	"Europe/Kaliningrad":               "Europe/Kaliningrad",
	"Europe/Kiev":                      "Europe/Kiev",
	"Europe/Lisbon":                    "Europe/Lisbon",
	"Europe/Ljubljana":                 "Europe/Ljubljana",
	"Europe/London":                    "Europe/London",
	"Europe/Luxembourg":                "Europe/Luxembourg",
	"Europe/Madrid":                    "Europe/Madrid",
	"Europe/Malta":                     "Europe/Malta",
	"Europe/Mariehamn":                 "Europe/Mariehamn",
	"Europe/Minsk":                     "Europe/Minsk",
	"Europe/Monaco":                    "Europe/Monaco",
	"Europe/Moscow":                    "Europe/Moscow",
	"Europe/Nicosia":                   "Europe/Nicosia",
	"Europe/Oslo":                      "Europe/Oslo",
	"Europe/Paris":                     "Europe/Paris",
	"Europe/Podgorica":                 "Europe/Podgorica",
	"Europe/Prague":                    "Europe/Prague",
	"Europe/Riga":                      "Europe/Riga",
	"Europe/Rome":                      "Europe/Rome",
	"Europe/Samara":                    "Europe/Samara",
	"Europe/San_Marino":                "Europe/San_Marino",
	"Europe/Sarajevo":                  "Europe/Sarajevo",
	"Europe/Simferopol":                "Europe/Simferopol",
	"Europe/Skopje":                    "Europe/Skopje",
	"Europe/Sofia":                     "Europe/Sofia",
	"Europe/Stockholm":                 "Europe/Stockholm",
	"Europe/Tallinn":                   "Europe/Tallinn",
	"Europe/Tirane":                    "Europe/Tirane",
	"Europe/Tiraspol":                  "Europe/Tiraspol",
	"Europe/Uzhgorod":                  "Europe/Uzhgorod",
	"Europe/Vaduz":                     "Europe/Vaduz",
	"Europe/Vatican":                   "Europe/Vatican",
	"Europe/Vienna":                    "Europe/Vienna",
	"Europe/Vilnius":                   "Europe/Vilnius",
	"Europe/Volgograd":                 "Europe/Volgograd",
	"Europe/Warsaw":                    "Europe/Warsaw",
	"Europe/Zagreb":                    "Europe/Zagreb",
	"Europe/Zaporozhye":                "Europe/Zaporozhye",
	"Europe/Zurich":                    "Europe/Zurich",
	"GB":                               "GB",
	"GB-Eire":                          "GB-Eire",
	"GMT":                              "GMT",
	"GMT+0":                            "GMT+0",
	"GMT-0":                            "GMT-0",
	"GMT0":                             "GMT0",
	"Greenwich":                        "Greenwich",
	"HST":                              "HST",
	"Hongkong":                         "Hongkong",
	"Iceland":                          "Iceland",
	"Indian/Antananarivo":              "Indian/Antananarivo",
	"Indian/Chagos":                    "Indian/Chagos",
	"Indian/Christmas":                 "Indian/Christmas",
	"Indian/Cocos":                     "Indian/Cocos",
	"Indian/Comoro":                    "Indian/Comoro",
	"Indian/Kerguelen":                 "Indian/Kerguelen",
	"Indian/Mahe":                      "Indian/Mahe",
	"Indian/Maldives":                  "Indian/Maldives",
	"Indian/Mauritius":                 "Indian/Mauritius",
	"Indian/Mayotte":                   "Indian/Mayotte",
	"Indian/Reunion":                   "Indian/Reunion",
	"Iran":                             "Iran",
	"Israel":                           "Israel",
	"Jamaica":                          "Jamaica",
	"Japan":                            "Japan",
	"Kwajalein":                        "Kwajalein",
	"Libya":                            "Libya",
	"MET":                              "MET",
	"MST":                              "MST",
	"MST7MDT":                          "MST7MDT",
	"Mexico/BajaNorte":                 "Mexico/BajaNorte",
	"Mexico/BajaSur":                   "Mexico/BajaSur",
	"Mexico/General":                   "Mexico/General",
	"NZ":                               "NZ",
	"NZ-CHAT":                          "NZ-CHAT",
	"Navajo":                           "Navajo",
	"PRC":                              "PRC",
	"PST8PDT":                          "PST8PDT",
	"Pacific/Apia":                     "Pacific/Apia",
	"Pacific/Auckland":                 "Pacific/Auckland",
	"Pacific/Chatham":                  "Pacific/Chatham",
	"Pacific/Chuuk":                    "Pacific/Chuuk",
	"Pacific/Easter":                   "Pacific/Easter",
	"Pacific/Efate":                    "Pacific/Efate",
	"Pacific/Enderbury":                "Pacific/Enderbury",
	"Pacific/Fakaofo":                  "Pacific/Fakaofo",
	"Pacific/Fiji":                     "Pacific/Fiji",
	"Pacific/Funafuti":                 "Pacific/Funafuti",
	"Pacific/Galapagos":                "Pacific/Galapagos",
	"Pacific/Gambier":                  "Pacific/Gambier",
	"Pacific/Guadalcanal":              "Pacific/Guadalcanal",
	"Pacific/Guam":                     "Pacific/Guam",
	"Pacific/Honolulu":                 "Pacific/Honolulu",
	"Pacific/Johnston":                 "Pacific/Johnston",
	"Pacific/Kiritimati":               "Pacific/Kiritimati",
	"Pacific/Kosrae":                   "Pacific/Kosrae",
	"Pacific/Kwajalein":                "Pacific/Kwajalein",
	"Pacific/Majuro":                   "Pacific/Majuro",
	"Pacific/Marquesas":                "Pacific/Marquesas",
	"Pacific/Midway":                   "Pacific/Midway",
	"Pacific/Nauru":                    "Pacific/Nauru",
	"Pacific/Niue":                     "Pacific/Niue",
	"Pacific/Norfolk":                  "Pacific/Norfolk",
	"Pacific/Noumea":                   "Pacific/Noumea",
	"Pacific/Pago_Pago":                "Pacific/Pago_Pago",
	"Pacific/Palau":                    "Pacific/Palau",
	"Pacific/Pitcairn":                 "Pacific/Pitcairn",
	"Pacific/Pohnpei":                  "Pacific/Pohnpei",
	"Pacific/Ponape":                   "Pacific/Ponape",
	"Pacific/Port_Moresby":             "Pacific/Port_Moresby",
	"Pacific/Rarotonga":                "Pacific/Rarotonga",
	"Pacific/Saipan":                   "Pacific/Saipan",
	"Pacific/Samoa":                    "Pacific/Samoa",
	"Pacific/Tahiti":                   "Pacific/Tahiti",
	"Pacific/Tarawa":                   "Pacific/Tarawa",
	"Pacific/Tongatapu":                "Pacific/Tongatapu",
	"Pacific/Truk":                     "Pacific/Truk",
	"Pacific/Wake":                     "Pacific/Wake",
	"Pacific/Wallis":                   "Pacific/Wallis",
	"Pacific/Yap":                      "Pacific/Yap",
	"Poland":                           "Poland",
	"Portugal":                         "Portugal",
	"ROC":                              "ROC",
	"ROK":                              "ROK",
	"Singapore":                        "Singapore",
	"Turkey":                           "Turkey",
	"UCT":                              "UCT",
	"US/Alaska":                        "US/Alaska",
	"US/Aleutian":                      "US/Aleutian",
	"US/Arizona":                       "US/Arizona",
	"US/Central":                       "US/Central",
	"US/East-Indiana":                  "US/East-Indiana",
	"US/Eastern":                       "US/Eastern",
	"US/Hawaii":                        "US/Hawaii",
	"US/Indiana-Starke":                "US/Indiana-Starke",
	"US/Michigan":                      "US/Michigan",
	"US/Mountain":                      "US/Mountain",
	"US/Pacific":                       "US/Pacific",
	"US/Samoa":                         "US/Samoa",
	"UTC":                              "UTC",
	"Universal":                        "Universal",
	"W-SU":                             "W-SU",
	"Zulu":                             "Zulu",
	"WET":                              "WET",
}
