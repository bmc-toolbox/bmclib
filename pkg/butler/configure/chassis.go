package configure

import (
	"strings"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
)

// BmcChassis struct declares attributes required to apply configuration.
type BmcChassis struct {
	asset     *asset.Asset
	bmc       devices.BmcChassis
	resources []string
	configure devices.Configure
	config    *cfgresources.ResourcesConfig
	logger    *logrus.Logger
	ip        string
	serial    string
	vendor    string
	model     string
}

// NewBmcChassisConfigurator returns a new configure struct to apply configuration.
func NewBmcChassisConfigurator(bmc devices.BmcChassis,
	asset *asset.Asset,
	resources []string,
	config *cfgresources.ResourcesConfig,
	logger *logrus.Logger) *BmcChassis {

	return &BmcChassis{
		// asset to be setup
		asset: asset,
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
func (b *BmcChassis) Apply() { //nolint: gocyclo

	// slice of configuration resources to be applied.
	var resources []string

	// retrieve valid or known configuration resources for the bmc.
	if len(b.resources) > 0 {
		resources = b.resources
	} else {
		resources = b.configure.Resources()
	}

	b.vendor = b.bmc.Vendor()
	b.model, _ = b.bmc.Model()
	b.serial, _ = b.bmc.Serial()
	b.ip = b.asset.IPAddress

	var failed, success []string

	b.logger.WithFields(logrus.Fields{
		"Vendor":    b.vendor,
		"Model":     b.model,
		"Serial":    b.serial,
		"IPAddress": b.ip,
		"To apply":  strings.Join(resources, ", "),
	}).Trace("Configuration resources to be applied.")

	for _, resource := range resources {

		var err error

		switch resource {
		case "user":
			if b.config.User != nil {
				err = b.configure.User(b.config.User)
			}
		case "syslog":
			if b.config.Syslog != nil {
				err = b.configure.Syslog(b.config.Syslog)
			}
		case "ntp":
			if b.config.Ntp != nil {
				err = b.configure.Ntp(b.config.Ntp)
			}
		case "ldap":
			if b.config.Ldap != nil {
				err = b.configure.Ldap(b.config.Ldap)
			}
		case "ldap_group":
			if b.config.LdapGroup != nil && b.config.Ldap != nil {
				err = b.configure.LdapGroup(b.config.LdapGroup, b.config.Ldap)
			}
		case "license":
			if b.config.License != nil {
				err = b.configure.SetLicense(b.config.License)
			}
		case "network":
			if b.config.Network != nil {
				err = b.configure.Network(b.config.Network)
			}
		default:
			b.logger.WithFields(logrus.Fields{
				"resource": resource,
			}).Warn("Unknown resource.")
		}

		if err != nil {
			failed = append(failed, resource)
			b.logger.WithFields(logrus.Fields{
				"resource":  resource,
				"Vendor":    b.vendor,
				"Model":     b.model,
				"Serial":    b.serial,
				"IPAddress": b.ip,
				"Error":     err,
			}).Warn("Resource configuration returned errors.")
		} else {
			success = append(success, resource)
		}

		b.logger.WithFields(logrus.Fields{
			"resource":  resource,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
		}).Trace("Resource configuration applied.")

	}

	b.logger.WithFields(logrus.Fields{
		"Vendor":       b.vendor,
		"Model":        b.model,
		"Serial":       b.serial,
		"applied":      strings.Join(success, ", "),
		"unsuccessful": strings.Join(failed, ", "),
	}).Debug("Chassis BMC configuration actions done.")
}
