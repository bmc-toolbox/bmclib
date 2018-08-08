package idrac9

type userInfo map[int]User
type idracUsers map[string]userInfo

type User struct {
	UserName               string `json:"UserName",omitempty`
	Password               string `json:"Password",omitempty`
	Enable                 string `json:"Enable",omitempty`                 //Enabled, Disabled
	Privilege              string `json:"Privilege",omitempty`              //511, 499
	IpmiLanPrivilege       string `json:"IpmiLanPrivilege",omitempty`       //Administrator, Operator
	SolEnable              string `json:"SolEnable",omitempty`              //Disabled, Enabled
	ProtocolEnable         string `json:"ProtocolEnable",omitempty`         //Disabled, Enabled (SNMPv2)
	AuthenticationProtocol string `json:"AuthenticationProtocol",omitempty` //SHA, MD5, None
	PrivacyProtocol        string `json:"PrivacyProtocol",omitempty`        //AES, DES, None
}
