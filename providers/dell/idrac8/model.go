package idrac8

import (
	"encoding/xml"
)

type UserInfo map[int]User

//https://10.193.251.5/sysmgmt/2012/server/configgroup/iDRAC.SysLog
type Syslog struct {
	Port    string `json:"Port"`
	Enable  string `json:"SysLogEnable"`
	Server1 string `json:"Server1"`
	Server2 string `json:"Server2"`
	Server3 string `json:"Server3"`
}

//https://10.193.251.5/sysmgmt/2012/server/configgroup/iDRAC.Users.3
type User struct {
	UserName         string `json:"UserName"`
	Password         string `json:"Password"`
	Enable           string `json:"Enable"`
	Privilege        string `json:"Privilege"`
	IpmiLanPrivilege string `json:"IpmiLanPrivilege"`
	SolEnable        string `json:"SolEnable"`
}

//GET - params as query string
//https://10.193.251.10/data?set=tm_ntp_int_opmode:1,tm_ntp_str_server1:ntp0.lhr4.example.com,tm_ntp_str_server2:ntp0.ams4.example.com,tm_ntp_str_server3:ntp0.fra4.example.com
type NtpServer struct {
	Enable  bool   `url:"tm_ntp_int_opmode,int"`
	Server1 string `url:"tm_ntp_str_server1"`
	Server2 string `url:"tm_ntp_str_server2"`
	Server3 string `url:"tm_ntp_str_server3"`
}

type XmlRoot struct {
	XMLName        xml.Name         `xml:"root"`
	Text           string           `xml:",chardata"`
	XmlUserAccount []XmlUserAccount `xml:"user"`
	Status         string           `xml:"status"`
}

type XmlUserAccount struct {
	Name          string `xml:"name"`
	Id            int    `xml:"id"`
	Privileges    int    `xml:"privileges"` //511 = Administrator, 0 = None,
	Enabled       int    `xml:"enabled"`
	LanPriv       int    `xml:"lanPriv"`    //4 = Administrator, 3 = Operator, 2 = User, 15 = None
	SerialPriv    int    `xml:"serialPriv"` //4 = Administrator, 3 = Operator, 2 = User, 15 = None
	SolEnabled    int    `xml:"solEnabled"`
	SnmpV3Enabled int    `xml:"SNMPV3Enabled"`
	SnmpPrivType  int    `xml:"snmpPrivType"`
}
