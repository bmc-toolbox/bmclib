package ilo

import (
	"fmt"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// compares the current Network IPv4 config with the given Network configuration
func (i Ilo) cmpNetworkIPv4Settings(cfg *cfgresources.Network) (NetworkIPv4, bool, error) {

	// setup some params as int for comparison
	var dnsFromDHCP, dhcpEnable, ddnsEnable int

	if cfg.DhcpEnable {
		dhcpEnable = 1
	}

	if cfg.DNSFromDHCP {
		dnsFromDHCP = 1
	}

	if cfg.DDNSEnable {
		ddnsEnable = 1
	}

	// get current config
	currentConfig, err := i.queryNetworkIPv4()
	if err != nil {
		return NetworkIPv4{}, false, fmt.Errorf("Unable to query existing IPv4 network config")

	}

	settingsMatch := func() bool {
		if currentConfig.DhcpEnabled != dhcpEnable {
			return false
		}

		if currentConfig.RegDdnsServer != ddnsEnable {
			return false
		}

		if currentConfig.UseDhcpSuppliedDomainName != dnsFromDHCP {
			return false
		}

		return true
	}

	if settingsMatch() {
		return NetworkIPv4{}, false, nil
	}

	currentConfig.DhcpEnabled = dhcpEnable
	currentConfig.RegDdnsServer = ddnsEnable
	currentConfig.UseDhcpSuppliedDomainName = dnsFromDHCP
	currentConfig.SessionKey = i.sessionKey
	currentConfig.Method = "set_ipv4"

	// configuration update required.
	return currentConfig, true, nil
}

// compares the current AccessSettings struct field values
// with the given Network configuration resource,
// returning an updated AccessSettings struct if an update is required.
// nolint: gocyclo
func (i *Ilo) cmpAccessSettings(cfg *cfgresources.Network) (AccessSettings, bool, error) {

	// setup some params as int for comparison
	var sshEnable, ipmiEnable, serialEnable int

	if cfg.SSHEnable == true {
		sshEnable = 1
	}

	if cfg.IpmiEnable == true {
		ipmiEnable = 1
	}

	if cfg.SolEnable == true {
		// enable with Auth
		serialEnable = 2
	}

	currentConfig, err := i.queryAccessSettings()
	if err != nil {
		return AccessSettings{}, false, err
	}

	// compare current configuration with configuration declared.
	settingsMatch := func() bool {

		// compare currentConfig cofiguration with declared.
		if currentConfig.SSHStatus != sshEnable {
			return false
		}

		if currentConfig.IpmiLanStatus != ipmiEnable {
			return false
		}

		if currentConfig.SerialCliStatus != serialEnable {
			return false
		}

		if currentConfig.SSHPort != cfg.SSHPort {
			return false
		}

		if currentConfig.IpmiPort != cfg.IpmiPort {
			return false
		}

		if currentConfig.RemoteConsolePort != cfg.KVMConsolePort {
			return false
		}

		if currentConfig.VirtualMediaPort != cfg.KVMMediaPort {
			return false
		}

		return true
	}

	if settingsMatch() {
		return AccessSettings{}, false, nil
	}

	currentConfig.IpmiPort = cfg.IpmiPort
	currentConfig.SSHStatus = sshEnable
	currentConfig.SSHPort = cfg.SSHPort
	currentConfig.RemoteConsolePort = cfg.KVMConsolePort
	currentConfig.VirtualMediaPort = cfg.KVMMediaPort
	currentConfig.IpmiLanStatus = ipmiEnable
	currentConfig.SerialCliStatus = serialEnable
	currentConfig.SessionKey = i.sessionKey
	currentConfig.Method = "set_services"

	// configuration update required.
	return currentConfig, true, nil

}
