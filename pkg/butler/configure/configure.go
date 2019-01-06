package configure

import (
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
)

// Configure struct declares attributes required to apply configuration.
type Configure struct {
	bmc       devices.Bmc
	configure devices.Configure
	config    *cfgresources.ResourcesConfig
	logger    *logrus.Logger
}

// New returns a new configure struct to apply configuration.
func New(bmc devices.Bmc, config *cfgresources.ResourcesConfig, logger *logrus.Logger) *Configure {
	return &Configure{
		// client is of type devices.Bmc
		bmc: bmc,
		// devices.Bmc is type asserted to apply configuration,
		// this is possible since devices.Bmc embeds the Configure interface.
		configure: bmc.(devices.Configure),
		config:    config,
		logger:    logger,
	}
}

// Apply applies configuration.
func (c *Configure) Apply() {

	// retrieve valid or known configuration resources for the bmc.
	resources := c.configure.Resources()

	vendor := c.bmc.Vendor()
	model, _ := c.bmc.Model()
	serial, _ := c.bmc.Serial()

	for _, resource := range resources {

		var err error

		switch resource {
		case "user":
			err = c.configure.User(c.config.User)
		case "syslog":
			err = c.configure.Syslog(c.config.Syslog)
		case "ntp":
			err = c.configure.Ntp(c.config.Ntp)
		case "ldap":
			err = c.configure.Ldap(c.config.Ldap)
		case "ldap_group":
			err = c.configure.LdapGroup(c.config.LdapGroup, c.config.Ldap)
		case "license":
			err = c.configure.SetLicense(c.config.License)
		case "network":
			err = c.configure.Network(c.config.Network)
		}

		if err != nil {
			c.logger.WithFields(logrus.Fields{
				"resource": resource,
				"IP":       "",
				"Vendor":   vendor,
				"Model":    model,
				"Serial":   serial,
				"Error":    err,
			}).Warn("Resource configuration returned errors.")
		}

		c.logger.WithFields(logrus.Fields{
			"resource": resource,
			"IP":       "",
			"Vendor":   vendor,
			"Model":    model,
			"Serial":   serial,
		}).Debug("Resource configuration applied.")

	}

	c.logger.WithFields(logrus.Fields{
		"IP":     "",
		"Vendor": vendor,
		"Model":  model,
		"Serial": serial,
	}).Debug("Configuration applied successfully.")
}
