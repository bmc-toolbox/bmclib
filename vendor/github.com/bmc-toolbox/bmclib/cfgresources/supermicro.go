package cfgresources

// Supermicro holds configuration for supermicro assets.
type Supermicro struct {
	NetworkCfg *SupermicroNetworkCfg `yaml:"network"`
}

// SupermicroNetworkCfg holds configuration for supermicro network.
//this is an example for issue #34
type SupermicroNetworkCfg struct {
	Web              string `yaml:"web"`
	WebSsl           int    `yaml:"webSsl"`
	IkvmServerPort   int    `yaml:"ikvmServerPort"`
	VirtualMediaPort int    `yaml:"virtualMediaPort"`
	SSHPort          int    `yaml:"sshPort"`
}
