package ilo

import (
	"fmt"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// cmdPowerSettings
func (i *Ilo) cmpPowerSettings(regulatorMode string) (PowerRegulator, bool, error) {
	// get current config
	currentConfig, err := i.queryPowerRegulator()
	if err != nil {
		return PowerRegulator{}, false, fmt.Errorf("Unable to query existing Power regulator config")
	}

	settingsMatch := func() bool {
		return currentConfig.PowerMode == regulatorMode
	}

	if settingsMatch() {
		return currentConfig, false, nil
	}

	// configuration update required.
	return currentConfig, true, nil
}

// compares the current Network IPv4 config with the given Network configuration
func (i *Ilo) cmpNetworkIPv4Settings(cfg *cfgresources.Network) (NetworkIPv4, bool, error) {
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

func setSnmpSettings(hardwareType string, snmpEnable bool) *SNMPSettings {
	snmpSettings := new(SNMPSettings)
	snmpSettings.SnmpPort = 161 // TODO: Change this to something user-configurable
	snmpSettings.TrapPort = 162 // TODO: Change this to something user-configurable

	switch hardwareType {
	case Ilo4:
		snmpDisable := new(int)
		if snmpEnable {
			*snmpDisable = 0
		} else {
			*snmpDisable = 1
		}
		snmpSettings.SnmpExternalDisableIlo4 = snmpDisable
	case Ilo5:
		snmpEnabled := new(int)
		if snmpEnable {
			*snmpEnabled = 1
		} else {
			*snmpEnabled = 0
		}
		snmpSettings.SnmpExternalEnabledIlo5 = snmpEnabled
	}

	return snmpSettings

}

func isUpdateRequiredSNMPIlo4(snmpEnable bool, bmcSettings SNMPSettings) bool {
	if (snmpEnable && *bmcSettings.SnmpExternalDisableIlo4 == 1) ||
		(!snmpEnable && *bmcSettings.SnmpExternalDisableIlo4 == 0) {
		return true // means the current settings applied on the BMC and the declared configuration are different
	}
	return false // settings are equal, update not required
}

func isUpdateRequiredSNMPIlo5(snmpEnable bool, bmcSettings SNMPSettings) bool {
	if (snmpEnable && *bmcSettings.SnmpExternalEnabledIlo5 == 0) ||
		(!snmpEnable && *bmcSettings.SnmpExternalEnabledIlo5 == 1) {
		return true // means the current settings applied on the BMC and the declared configuration are different
	}
	return false //settings are equal, update not required
}

// compares the current AccessSettings struct field values
// with the given Network configuration resource,
// returning an updated AccessSettings struct if an update is required.
// nolint: gocyclo
func (i *Ilo) cmpAccessSettings(cfg *cfgresources.Network) (AccessSettings, bool, error) {
	// setup some params as int for comparison
	var sshEnable, ipmiEnable, serialEnable int

	if cfg.SSHEnable {
		sshEnable = 1
	}

	if cfg.IpmiEnable {
		ipmiEnable = 1
	}

	if cfg.SolEnable {
		// enable with Auth
		serialEnable = 2
	}

	// SNMP status is in cfg.SNMPEnable as a boolean

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
		// Comparing SNMP settings for iLO 4 and iLO5
		switch i.HardwareType() {
		case Ilo4:
			if isUpdateRequiredSNMPIlo4(cfg.SNMPEnable, currentConfig.SNMPSettings) {
				return false
			}
		case Ilo5:
			if isUpdateRequiredSNMPIlo5(cfg.SNMPEnable, currentConfig.SNMPSettings) {
				return false
			}
		}

		return true
	}

	if settingsMatch() {
		return AccessSettings{}, false, nil
	}

	currentConfig.IpmiPort = cfg.IpmiPort
	currentConfig.SSHStatus = sshEnable
	currentConfig.SSHPort = cfg.SSHPort
	currentConfig.SNMPSettings = *setSnmpSettings(i.HardwareType(), cfg.SNMPEnable)
	currentConfig.RemoteConsolePort = cfg.KVMConsolePort
	currentConfig.VirtualMediaPort = cfg.KVMMediaPort
	currentConfig.IpmiLanStatus = ipmiEnable
	currentConfig.SerialCliStatus = serialEnable
	currentConfig.SessionKey = i.sessionKey
	currentConfig.Method = "set_services"

	// configuration update required.
	return currentConfig, true, nil
}
