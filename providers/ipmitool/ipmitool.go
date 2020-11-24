package ipmitool

import (
	"github.com/bmc-toolbox/bmclib/registry"
	"github.com/go-logr/logr"
)

// ProviderName for the implementation
const ProviderName = "ipmitool"

// Conn for Ipmitool connection details
type Conn struct {
	Host string
	Port string
	User string
	Pass string
	Log  logr.Logger
}

func init() {
	registry.Register(ProviderName, "ipmi", func(host, user, pass string) (interface{}, error) {
		port := "623"
		return &Conn{Host: host, User: user, Pass: pass, Port: port}, nil
	}, []registry.Feature{registry.FeaturePowerSetting, registry.FeaturePowerState, registry.FeatureUserRead})
}
