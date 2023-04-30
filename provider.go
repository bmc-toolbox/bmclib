package bmclib

import "net/http"

// providerConfig contains per provider specific configuration.
type providerConfig struct {
	ipmitool ipmitoolConfig
	asrock   asrockrackConfig
	gofish   gofishConfig
	intelamt intelAMTConfig
}

type ipmitoolConfig struct {
	cipherSuite int
	port        string
}

type asrockrackConfig struct {
	httpClient *http.Client
	port       string
}

type gofishConfig struct {
	httpClient *http.Client
	port       string
	// versionsNotCompatible	is the list of incompatible redfish versions.
	//
	// With this option set, The bmclib.Registry.FilterForCompatible(ctx) method will not proceed on
	// devices with the given redfish version(s).
	versionsNotCompatible []string
}

type intelAMTConfig struct {
	// hostScheme should be either "http" or "https".
	hostScheme string
	port       string
}
