package idrac8

import (
	"encoding/xml"
)

// UserInfo type is used to unmarshal user accounts payload.
type UsersInfo map[int]UserInfo

// Syslog struct holds syslog configuration payload
// https://10.193.251.5/sysmgmt/2012/server/configgroup/iDRAC.SysLog
type Syslog struct {
	Port    string `json:"Port"`
	Enable  string `json:"SysLogEnable"`
	Server1 string `json:"Server1"`
	Server2 string `json:"Server2"`
	Server3 string `json:"Server3"`
}

// User struct holds user account configuration payload
// https://10.193.251.5/sysmgmt/2012/server/configgroup/iDRAC.Users.3
type UserInfo struct {
	UserName         string `json:"UserName"`
	Password         string `json:"Password"`
	Enable           string `json:"Enable"`
	Privilege        string `json:"Privilege"`
	IpmiLanPrivilege string `json:"IpmiLanPrivilege"`
	SolEnable        string `json:"SolEnable"`
	SNMPv3Enable     string `json:"SNMPv3Enable"`
}

// certStore is the response received when uploading a multipart form,
// that includes the certificate, this cert is stored in a transient store.
// {"File":{"ResourceURI":"/sysmgmt/2012/server/transient/filestore/721k12.bmc.dummy.com.crt"}}
type certStore struct {
	File struct {
		ResourceURI string `json:"ResourceURI"`
	} `json:"File"`
}

// NtpServer struct holds NTP configuration payload
// GET - params as query string
// https://10.193.251.10/data?set=tm_ntp_int_opmode:1,tm_ntp_str_server1:ntp0.lhr4.example.com,tm_ntp_str_server2:ntp0.ams4.example.com,tm_ntp_str_server3:ntp0.fra4.example.com
type NtpServer struct {
	Enable  bool   `url:"tm_ntp_int_opmode,int"`
	Server1 string `url:"tm_ntp_str_server1"`
	Server2 string `url:"tm_ntp_str_server2"`
	Server3 string `url:"tm_ntp_str_server3"`
}

// XMLRoot is used to unmarshal XML response payloads.
type XMLRoot struct {
	XMLName        xml.Name         `xml:"root"`
	Text           string           `xml:",chardata"`
	XMLUserAccount []XMLUserAccount `xml:"user"`
	Status         string           `xml:"status"`
}

// XMLUserAccount is used to unmarshal XML user account response payloads.
type XMLUserAccount struct {
	Name          string `xml:"name"`
	ID            int    `xml:"id"`
	Privileges    int    `xml:"privileges"` // 511 = Administrator, 0 = None,
	Enabled       int    `xml:"enabled"`
	LanPriv       int    `xml:"lanPriv"`    // 4 = Administrator, 3 = Operator, 2 = User, 15 = None
	SerialPriv    int    `xml:"serialPriv"` // 4 = Administrator, 3 = Operator, 2 = User, 15 = None
	SolEnabled    int    `xml:"solEnabled"`
	SnmpV3Enabled int    `xml:"SNMPV3Enabled"`
	SnmpPrivType  int    `xml:"snmpPrivType"`
}

// setAlertFilterPayload enables all warning, critical alerts to be sent over syslog
// yes its ugly, if there are other parameters to be enabled, we'd need to get a fresh dump of these parameters from the POST request on the BMC.
var setAlertFilterPayload = `setAlertFilter(1\:2\:AMP\:261\,1\:1\:AMP\:261\,1\:1\:ASR\:261\,1\:2\:BAT\:261\,1\:1\:BAT\:261\,1\:1\:CBL\:257\,1\:2\:CMC\:1817\,1\:1\:CMC\:1817\,1\:2\:CPU\:261\,1\:1\:CPU\:261\,1\:1\:CPUA\:261\,1\:2\:FAN\:1285\,1\:1\:FAN\:261\,1\:2\:FC\:281\,1\:2\:HWC\:1817\,1\:1\:HWC\:1817\,1\:2\:IOV\:256\,1\:1\:IOV\:1817\,1\:2\:LNK\:1817\,1\:1\:LNK\:257\,1\:2\:MEM\:257\,1\:1\:MEM\:257\,1\:2\:NIC\:281\,1\:2\:PCI\:257\,1\:1\:PCI\:1817\,1\:2\:PDR\:257\,1\:1\:PDR\:257\,1\:2\:PFM\:261\,1\:1\:PST\:257\,1\:2\:PSU\:261\,1\:1\:PSU\:261\,1\:1\:PSUA\:261\,1\:2\:PWR\:1817\,1\:1\:PWR\:1817\,1\:2\:RDU\:261\,1\:1\:RDU\:261\,1\:2\:RFL\:261\,1\:1\:RFL\:261\,1\:1\:RFLA\:261\,1\:2\:RRDU\:261\,1\:1\:RRDU\:261\,1\:1\:SEC\:773\,1\:2\:SEL\:257\,1\:1\:SEL\:1309\,1\:2\:SWC\:257\,1\:1\:SWC\:257\,1\:1\:SYS\:1823\,1\:2\:TMP\:261\,1\:1\:TMP\:261\,1\:2\:TMPS\:257\,1\:1\:TMPS\:257\,1\:2\:VFL\:261\,1\:1\:VFL\:261\,1\:2\:VLT\:261\,1\:1\:VLT\:261\,2\:2\:BAT\:257\,2\:1\:BAT\:257\,2\:2\:CTL\:257\,2\:1\:CTL\:257\,2\:2\:ENC\:257\,2\:1\:ENC\:257\,2\:1\:FAN\:257\,2\:2\:PDR\:257\,2\:1\:PDR\:257\,2\:2\:PSU\:257\,2\:1\:PSU\:257\,2\:2\:SEC\:1305\,2\:1\:SEC\:1305\,2\:2\:STOR\:1817\,2\:1\:STOR\:1817\,2\:2\:TMP\:257\,2\:1\:TMP\:257\,2\:2\:VDR\:257\,2\:1\:VDR\:257\,5\:2\:IOID\:281\,5\:2\:RAC\:256\,5\:2\:SEC\:257\,5\:2\:SWC\:1817\,5\:1\:SWC\:1817\,4\:2\:CMC\:1817\,4\:1\:CMC\:1817\,4\:2\:FSD\:257\,4\:2\:LIC\:1305\,4\:1\:LIC\:1817\,4\:2\:PCI\:1817\,4\:2\:PSU\:1817\,4\:1\:PSU\:1817\,4\:2\:PWR\:1817\,4\:1\:PWR\:1817\,4\:3\:SYS\:769\,4\:3\:USR\:769\,4\:2\:USR\:1817\,3\:2\:RED\:257\,3\:2\:SWU\:1817)`
