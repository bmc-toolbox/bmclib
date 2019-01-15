package configure

import (
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
)

// Bmc struct declares attributes required to apply configuration.
type Bmc struct {
	bmc       devices.Bmc
	resources []string
	configure devices.Configure
	config    *cfgresources.ResourcesConfig
	logger    *logrus.Logger
}

// NewBmcConfigurator returns a new configure struct to apply configuration.
func NewBmcConfigurator(bmc devices.Bmc, resources []string, config *cfgresources.ResourcesConfig, logger *logrus.Logger) *Bmc {
	return &Bmc{
		// client is of type devices.Bmc
		bmc: bmc,
		// devices.Bmc is type asserted to apply configuration,
		// this is possible since devices.Bmc embeds the Configure interface.
		configure: bmc.(devices.Configure),
		// if --resources was passed, only these resources will be applied
		resources: resources,
		config:    config,
		logger:    logger,
	}
}

// Apply applies configuration.
func (b *Bmc) Apply() {

	// slice of configuration resources to be applied.
	var resources []string

	// retrieve valid or known configuration resources for the bmc.
	if len(b.resources) > 0 {
		resources = b.resources
	} else {
		resources = b.configure.Resources()
	}

	vendor := b.bmc.Vendor()
	model, _ := b.bmc.Model()
	serial, _ := b.bmc.Serial()

	for _, resource := range resources {

		var err error

		switch resource {
		case "user":
			err = b.configure.User(b.config.User)
		case "syslog":
			err = b.configure.Syslog(b.config.Syslog)
		case "ntp":
			err = b.configure.Ntp(b.config.Ntp)
		case "ldap":
			err = b.configure.Ldap(b.config.Ldap)
		case "ldap_group":
			err = b.configure.LdapGroup(b.config.LdapGroup, b.config.Ldap)
		case "license":
			err = b.configure.SetLicense(b.config.License)
		case "network":
			err = b.configure.Network(b.config.Network)
		case "bios":
			err = b.configure.Bios(b.config.Bios)
		default:
			b.logger.WithFields(logrus.Fields{
				"resource": resource,
			}).Warn("Unknown resource.")
		}

		if err != nil {
			b.logger.WithFields(logrus.Fields{
				"resource": resource,
				"IP":       "",
				"Vendor":   vendor,
				"Model":    model,
				"Serial":   serial,
				"Error":    err,
			}).Warn("Resource configuration returned errors.")
		}

		b.logger.WithFields(logrus.Fields{
			"resource": resource,
			"IP":       "",
			"Vendor":   vendor,
			"Model":    model,
			"Serial":   serial,
		}).Debug("Resource configuration applied.")

	}

	b.logger.WithFields(logrus.Fields{
		"IP":     "",
		"Vendor": vendor,
		"Model":  model,
		"Serial": serial,
	}).Debug("Configuration applied successfully.")
}

// BmcChassis struct declares attributes required to apply configuration.
type BmcChassis struct {
	bmc       devices.BmcChassis
	resources []string
	configure devices.Configure
	config    *cfgresources.ResourcesConfig
	logger    *logrus.Logger
}

// NewBmcChassisConfigurator returns a new configure struct to apply configuration.
func NewBmcChassisConfigurator(bmc devices.BmcChassis, resources []string, config *cfgresources.ResourcesConfig, logger *logrus.Logger) *BmcChassis {
	return &BmcChassis{
		// client is of type devices.Bmc
		bmc: bmc,
		// if --resources was passed, only these resources will be applied
		resources: resources,
		// devices.BmcChassis is type asserted to apply configuration,
		// this is possible since devices.Bmc embeds the Configure interface.
		configure: bmc.(devices.Configure),
		config:    config,
		logger:    logger,
	}
}

// Apply applies configuration.
func (b *BmcChassis) Apply() {

	// slice of configuration resources to be applied.
	var resources []string

	// retrieve valid or known configuration resources for the bmc.
	if len(b.resources) > 0 {
		resources = b.resources
	} else {
		resources = b.configure.Resources()
	}

	vendor := b.bmc.Vendor()
	model, _ := b.bmc.Model()
	serial, _ := b.bmc.Serial()

	for _, resource := range resources {

		var err error

		switch resource {
		case "user":
			err = b.configure.User(b.config.User)
		case "syslog":
			err = b.configure.Syslog(b.config.Syslog)
		case "ntp":
			err = b.configure.Ntp(b.config.Ntp)
		case "ldap":
			err = b.configure.Ldap(b.config.Ldap)
		case "ldap_group":
			err = b.configure.LdapGroup(b.config.LdapGroup, b.config.Ldap)
		case "license":
			err = b.configure.SetLicense(b.config.License)
		case "network":
			err = b.configure.Network(b.config.Network)
		default:
			b.logger.WithFields(logrus.Fields{
				"resource": resource,
			}).Warn("Unknown resource.")
		}

		if err != nil {
			b.logger.WithFields(logrus.Fields{
				"resource": resource,
				"IP":       "",
				"Vendor":   vendor,
				"Model":    model,
				"Serial":   serial,
				"Error":    err,
			}).Warn("Resource configuration returned errors.")
		}

		b.logger.WithFields(logrus.Fields{
			"resource": resource,
			"IP":       "",
			"Vendor":   vendor,
			"Model":    model,
			"Serial":   serial,
		}).Debug("Resource configuration applied.")

	}

	b.logger.WithFields(logrus.Fields{
		"IP":     "",
		"Vendor": vendor,
		"Model":  model,
		"Serial": serial,
	}).Debug("Configuration applied successfully.")
}
