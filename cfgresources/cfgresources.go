package cfgresources

//if a resource is added/updated here it needs to be updated in bmc butler as well,
//bmc butlers resources need to be the same as bmclib resources,
//this is until we figure how we want these two packages to depend on each other.
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
	GroupDn        string `yaml:"groupDn"`
	UserAttribute  string `yaml:"userAttribute"`
	GroupAttribute string `yaml:"groupAttribute"`
	SearchFilter   string `yaml:"searchFilter"`
}

type Network struct {
	Hostname    string `yaml:"hostname"`
	DNSFromDHCP bool   `yaml:"dnsfromdhcp"`
}

type Ntp struct {
	Server1  string `yaml:"server1"`
	Server2  string `yaml:"server2"`
	Server3  string `yaml:"server3"`
	Timezone string `yaml:"timezone"`
}

type ResourcesConfig struct {
	Ldap    *Ldap    `yaml:"ldap"`
	Network *Network `yaml:"network"`
	Ntp     *Ntp     `yaml:"ntp"`
	Syslog  *Syslog  `yaml:"syslog"`
	User    []*User  `yaml:"user'`
}
