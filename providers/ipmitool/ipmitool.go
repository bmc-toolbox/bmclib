package ipmitool

import (
	"github.com/bmc-toolbox/bmclib/registry"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "ipmitool"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "ipmi"
)

// Conn for Ipmitool connection details
type Conn struct {
	Host string
	Port string
	User string
	Pass string
}

func init() {
	registry.Register(ProviderName, ProviderProtocol, func(host, user, pass string) (interface{}, error) {
		return &Conn{Host: host, User: user, Pass: pass, Port: "623"}, nil
	}, []registry.Feature{
		registry.FeaturePowerSet,
		registry.FeaturePowerState,
		registry.FeatureUserRead,
		registry.FeatureBmcReset,
		registry.FeatureBootDeviceSet,
	})
}
