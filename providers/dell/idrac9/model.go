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

type Ldap struct {
	BaseDN               string `json"BaseDN"`                //dell
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
type LdapRoleGroups map[string]LdapRoleGroup

type LdapRoleGroup struct {
	DN        string `json:"DN"`        //cn=dell,cn=bmcAdmins
	Privilege string `json:"Privilege"` //511 (Administrator), 499 (Operator)
}
