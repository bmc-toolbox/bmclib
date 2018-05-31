package cfgresources

type ResourcesConfig struct {
	Ldap      *Ldap        `yaml:"ldap"`
	LdapGroup []*LdapGroup `yaml:"ldapGroup"`
	Network   *Network     `yaml:"network"`
	Ntp       *Ntp         `yaml:"ntp"`
	Syslog    *Syslog      `yaml:"syslog"`
	User      []*User      `yaml:"user"`
	Ssl       *Ssl         `yaml:"ssl"`
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
}

type Ntp struct {
	Enable   bool   `yaml:"enable"`
	Server1  string `yaml:"server1"`
	Server2  string `yaml:"server2"`
	Server3  string `yaml:"server3"`
	Timezone string `yaml:"timezone"`
}
