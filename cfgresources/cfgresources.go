package cfgresources

type ResourcesSetup struct {
	FlexAddress  *flexAddress  `yaml:"flexAddress"`
	IpmiOverLan  *ipmiOverLan  `yaml:"ipmiOverLan"`
	DynamicPower *dynamicPower `yaml:"dynamicPower"`
}

type ResourcesConfig struct {
	Ldap       *Ldap        `yaml:"ldap"`
	LdapGroup  []*LdapGroup `yaml:"ldapGroup"`
	Network    *Network     `yaml:"network"`
	Ntp        *Ntp         `yaml:"ntp"`
	Syslog     *Syslog      `yaml:"syslog"`
	User       []*User      `yaml:"user"`
	Ssl        *Ssl         `yaml:"ssl"`
	Supermicro *Supermicro  `yaml:"supermicro"` //supermicro specific config, example of issue #34
}

//Enable/Disable Virtual Mac addresses for blades in a chassis.
//FlexAddresses in M1000e jargon.
//Virtual connect in HP C7000 jargon.
type flexAddress struct {
	Enable bool `yaml:"enable"`
}

//Enable/Disable ipmi over lan
type ipmiOverLan struct {
	Enable bool `yaml:"enable"`
}

//'Dynamic Power' in HP C7000 Jargon.
//'DPSE' (dynamic PSU engagement) in M1000e Dell jargon.
type dynamicPower struct {
	Enable bool `yaml:"enable"`
}

type User struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	Role     string `yaml:"role"`
	Enable   bool   `yaml:"enable,omitempty"`
}

type Syslog struct {
	Server string `yaml:"server"`
	Port   int    `yaml:"port",omitempty`
	Enable bool   `yaml:"enable",omitempty`
}

type Ldap struct {
	Server         string `yaml:"server"`
	Port           int    `yaml:"port"`
	Enable         bool   `yaml:"enable"`
	Role           string `yaml:"role"`
	BaseDn         string `yaml:"baseDn"`
	Group          string `yaml:"group"`
	GroupBaseDn    string `yaml:"groupBaseDn"`
	UserAttribute  string `yaml:"userAttribute"`
	GroupAttribute string `yaml:"groupAttribute"`
	SearchFilter   string `yaml:"searchFilter"`
}

type LdapGroup struct {
	Role        string `yaml:"role"`
	Group       string `yaml:"group"`
	GroupBaseDn string `yaml:"groupBaseDn"`
	Enable      bool   `yaml:"enable"`
}

type Ssl struct {
	CertFile string `yaml:"certfile"`
	KeyFile  string `yaml:"keyfile"`
}

type Network struct {
	Hostname    string `yaml:"hostname"`
	DNSFromDHCP bool   `yaml:"dnsfromdhcp"`
	SshEnable   bool   `yaml:"sshEnable"`
	SshPort     int    `yaml:"sshPort"`
	IpmiEnable  bool   `yaml:"ipmiEnable"`
	DhcpEnable  bool   `yaml:"dhcpEnable"`
	IpmiPort    int    `yaml:"ipmiPort"`
}

type Ntp struct {
	Enable   bool   `yaml:"enable"`
	Server1  string `yaml:"server1"`
	Server2  string `yaml:"server2"`
	Server3  string `yaml:"server3"`
	Timezone string `yaml:"timezone"`
}
