package cfgresources

// Supermicro holds configuration for supermicro assets.
type Supermicro struct {
	NetworkCfg *SupermicroNetworkCfg `yaml:"network"`
}

// SupermicroNetworkCfg holds configuration for supermicro network.
type SupermicroNetworkCfg struct {
	Web              string `yaml:"web"`
	WebSsl           int    `yaml:"webSsl"`
	IkvmServerPort   int    `yaml:"ikvmServerPort"`
	VirtualMediaPort int    `yaml:"virtualMediaPort"`
	SSHPort          int    `yaml:"sshPort"`
}
