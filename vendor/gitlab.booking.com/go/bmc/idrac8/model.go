package idrac8

//https://10.193.251.5/sysmgmt/2012/server/configgroup/iDRAC.SysLog
//{"iDRAC.SysLog":{"Port":"514","SysLogEnable":"Enabled","Server1":"example.com","Server2":"","Server3":""}}

type Syslog struct {
	Port    string `json:"Port"`
	Enable  string `json:"SysLogEnable"`
	Server1 string `json:"Server1"`
	Server2 string `json:"Server2"`
	Server3 string `json:"Server3"`
}

//https://10.193.251.5/sysmgmt/2012/server/configgroup/iDRAC.Users.3
//{"iDRAC.Users":{"UserName":"@066@06f@06f@062@061@072", "Password":"@066@06f@06f@062@061@072", "Enable":"Enabled", "Privilege":"511",
// "IpmiLanPrivilege":"Administrator", "SolEnable":"Enabled"}}
type User struct {
	UserName         string `json:"UserName"`
	Password         string `json:"Password"`
	Enable           string `json:"Enable"`
	Privilege        string `json:"Privilege"`
	IpmiLanPrivilege string `json:"IpmiLanPrivilege"`
	SolEnable        string `json:"SolEnable"`
}

//blink identifier led - with a 1 day timeout.
//https://10.193.251.10/data?set=IdentifyEnable:1,IdentifyTimeout:8640000

//GET - params as query string
//power cap
//https://10.193.251.10/data?set=pbtEnabled:0,

//GET - params as query string
//timezone
//https://10.193.251.10/data?set=tm_tz_str_zone:CET

//GET - params as query string
//ntp servers
//https://10.193.251.10/data?set=tm_ntp_int_opmode:1,tm_ntp_str_server1:ntp0.lhr4.example.com,tm_ntp_str_server2:ntp0.ams4.example.com,tm_ntp_str_server3:ntp0.fra4.example.com
type NtpServer struct {
	Enable  bool   `url:"tm_ntp_int_opmode,int"`
	Server1 string `url:"tm_ntp_str_server1"`
	Server2 string `url:"tm_ntp_str_server2"`
	Server3 string `url:"tm_ntp_str_server3"`
}

// LDAP settings
// Notes: - all non-numeric,alphabetic characters are escaped
//        - the idrac posts each payload twice?
//        - requests can be either POST or GET except for the final one - postset?ldapconf
//POST/GET
//Setup ldap server
//https://10.193.251.10/data?set=xGLServer:ldaps.example.com
//Set object class
// set=xGLSearchFilter:objectClass\=posixAccount
//Set ldap groups
// set=xGLGroup1Name:cn\=bmcAdmins\,ou\=Group\,dc\=example\,dc\=com
// set=xGLGroup2Name:cn\=bmcUsers\,ou\=Group\,dc\=example\,dc\=com

//POST
//Setup ldap group privileges and some other params
//Content-Type: application/x-www-form-urlencoded
//https://10.193.251.10/postset?ldapconf
//data=LDAPEnableMode:3,xGLNameSearchEnabled:0,xGLBaseDN:ou%5C%3DPeople%5C%2Cdc%5C%3Dexample%5C%2Cdc%5C%3Dcom,xGLUserLogin:uid,xGLGroupMem:memberUid,xGLBindDN:,xGLCertValidationEnabled:1,xGLGroup1Priv:511,xGLGroup2Priv:97,xGLGroup3Priv:0,xGLGroup4Priv:0,xGLGroup5Priv:0,xGLServerPort:636
