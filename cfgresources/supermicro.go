package cfgresources

type Supermicro struct {
	NetworkCfg *SupermicroNetworkCfg `yaml:"network"`
}

//this is an example for issue #34
type SupermicroNetworkCfg struct {
	Web              string `yaml:"web"`
	WebSsl           int    `yaml:"webSsl"`
	IkvmServerPort   int    `yaml:"ikvmServerPort"`
	VirtualMediaPort int    `yaml:"virtualMediaPort"`
	SshPort          int    `yaml:"sshPort"`
}
