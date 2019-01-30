package configure

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
	"github.com/bmc-toolbox/bmcbutler/pkg/inventory"
	"github.com/bmc-toolbox/bmcbutler/pkg/metrics"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
)

// BmcChassisSetup struct holds various attributes for chassis setup methods.
type BmcChassisSetup struct {
	asset          *asset.Asset
	chassis        devices.BmcChassis
	setup          devices.BmcChassisSetup
	config         *cfgresources.SetupChassis
	resources      []string
	butlerConfig   *config.Params
	metricsEmitter *metrics.Emitter
	log            *logrus.Logger
	ip             string
	serial         string
	vendor         string
	model          string
	stopChan       <-chan struct{}
}

// NewBmcChassisSetup returns a new  struct to apply configuration.
func NewBmcChassisSetup(
	chassis devices.BmcChassis,
	asset *asset.Asset,
	resources []string,
	config *cfgresources.SetupChassis,
	butlerConfig *config.Params,
	metricsEmitter *metrics.Emitter,
	stopChan <-chan struct{},
	logger *logrus.Logger) *BmcChassisSetup {

	return &BmcChassisSetup{
		// asset to be setup
		asset: asset,
		// client is of type devices.Bmc
		chassis: chassis,
		// devices.Bmc is type asserted to apply one time setup configuration,
		// this is possible since devices.Bmc embeds the BmcChassisSetup interface.
		setup:        chassis.(devices.BmcChassisSetup),
		butlerConfig: butlerConfig,
		// if --resources was passed, only these resources will be applied
		resources:      resources,
		metricsEmitter: metricsEmitter,
		config:         config,
		log:            logger,
		stopChan:       stopChan,
	}
}

// Apply applies one time setup configuration.
func (b *BmcChassisSetup) Apply() { //nolint: gocyclo

	defer b.metricsEmitter.MeasureRuntime(
		[]string{"butler", "setupChassis_runtime"},
		time.Now(),
	)

	var interrupt bool
	go func() { <-b.stopChan; interrupt = true }()

	// slice of configuration resources to be applied.
	var resources []string

	// if any setup action fails, this is set to false
	// if this finally is true, the post actions are invoked.
	setupActionSuccess := true

	// retrieve valid or known setup configuration resources for the chassis.
	if len(b.resources) > 0 {
		resources = b.resources
	} else {
		resources = b.setup.ResourcesSetup()
	}

	b.vendor = b.chassis.Vendor()
	b.model, _ = b.chassis.Model()
	b.serial, _ = b.chassis.Serial()
	b.ip = b.asset.IPAddress

	var failed, success []string

	b.log.WithFields(logrus.Fields{
		"Vendor":    b.vendor,
		"Model":     b.model,
		"Serial":    b.serial,
		"IPAddress": b.ip,
		"To apply":  strings.Join(resources, ", "),
	}).Trace("Configuration resources to be applied.")

	for _, resource := range resources {

		var err error

		// check if an interrupt was received.
		if interrupt == true {
			b.log.WithFields(logrus.Fields{
				"Vendor":    b.vendor,
				"Model":     b.model,
				"Serial":    b.serial,
				"IPAddress": b.ip,
			}).Debug("Received interrupt.")
			break
		}

		err = b.ensurePoweredUp()
		if err != nil {
			b.log.WithFields(logrus.Fields{
				"resource":  resource,
				"Vendor":    b.vendor,
				"Model":     b.model,
				"Serial":    b.serial,
				"IPAddress": b.ip,
				"Error":     err,
			}).Warn("Chassis power status")
			return
		}

		b.log.WithFields(logrus.Fields{
			"resource":  resource,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
		}).Debug("Chassis is powered on, continuing setup.")

		switch resource {
		case "setipmioverlan":
			if b.config.IpmiOverLan != nil {
				err = b.setIpmiOverLan()
			}
		case "flexaddress":
			if b.config.FlexAddress != nil {
				err = b.setFlexAddressState()
			}
		case "dynamicpower":
			if b.config.DynamicPower != nil {
				err = b.setDynamicPower()
			}
		case "bladespower":
			if b.config.BladesPower != nil {
				err = b.setBladesPower()
			}
		case "add_blade_bmc_admins":
			if len(b.config.AddBladeBmcAdmins) > 0 {
				err = b.addBladeBmcAdmins()
			}
		case "remove_blade_bmc_users":
			if len(b.config.RemoveBladeBmcUsers) > 0 {
				err = b.removeBladeBmcUsers()
			}
		default:
			b.log.WithFields(logrus.Fields{
				"resource": resource,
			}).Warn("Unknown setup resource.")
		}

		if err != nil {
			setupActionSuccess = false
			failed = append(failed, resource)
			b.log.WithFields(logrus.Fields{
				"resource":  resource,
				"Vendor":    b.vendor,
				"Model":     b.model,
				"Serial":    b.serial,
				"IPAddress": b.ip,
				"Error":     err,
			}).Warn("Setup resource returned errors.")
		} else {
			success = append(success, resource)

		}

		b.log.WithFields(logrus.Fields{
			"resource":  resource,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
		}).Trace("Resource configuration applied.")

	}

	//if chassis setup is done successfully invoke post action.
	if setupActionSuccess {
		b.Post()
	}

	b.log.WithFields(logrus.Fields{
		"Vendor":       b.vendor,
		"Model":        b.model,
		"Serial":       b.serial,
		"IPAddress":    b.ip,
		"applied":      strings.Join(success, ", "),
		"unsuccessful": strings.Join(failed, ", "),
	}).Info("Chassis setup actions done.")

}

// Post method is when a chassis was setup successfully.
func (b *BmcChassisSetup) Post() {

	enc := inventory.Enc{
		Config:         b.butlerConfig,
		Log:            b.log,
		MetricsEmitter: b.metricsEmitter,
	}

	enc.SetChassisInstalled(b.asset.Serial)

	return
}

// ensurePoweredUp method checks if a chassis is powered off
// and powers it back on.
func (b *BmcChassisSetup) ensurePoweredUp() (err error) {

	status, _ := b.chassis.IsOn()
	if status == false {
		_, err := b.chassis.PowerOn()
		if err != nil {
			return err
		}

		return errors.New("Chassis power status was off, powering on.. retry in a few minutes")
	}

	return nil
}

func (b *BmcChassisSetup) addBladeBmcAdmins() (err error) {

	component := "addBladeBmcAdmins"
	cfg := b.config.AddBladeBmcAdmins

	for _, user := range cfg {
		if user.Name == "" {
			return fmt.Errorf("AddbladeBmcAdmins resource expects parameter: Name")
		}

		if user.Password == "" {
			return fmt.Errorf("AddbladeBmcAdmins resource expects parameter: Password")
		}

		err = b.setup.AddBladeBmcAdmin(user.Name, user.Password)
		if err != nil {
			return err
		}

		// in cases where the user may already exist, we modify the credentials
		err = b.setup.ModBladeBmcUser(user.Name, user.Password)
		if err != nil {
			return err
		}

		b.log.WithFields(logrus.Fields{
			"component": component,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
			"User":      user.Name,
		}).Debug("Blade BMC admin account added.")
	}

	return err
}

func (b *BmcChassisSetup) removeBladeBmcUsers() (err error) {

	component := "removeBladeBmcUsers"

	cfg := b.config.RemoveBladeBmcUsers
	for _, user := range cfg {
		if user.Name == "" {
			return fmt.Errorf("RemoveBladeBmcUsers resource expects parameter: Name")
		}

		err = b.setup.RemoveBladeBmcUser(user.Name)
		if err != nil {
			return err
		}

		b.log.WithFields(logrus.Fields{
			"component": component,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
			"User":      user.Name,
		}).Debug("Blade BMC user account removed.")
	}

	return err
}

func (b *BmcChassisSetup) setDynamicPower() (err error) {

	log := b.log
	component := "setDynamicPower"

	_, err = b.setup.SetDynamicPower(b.config.DynamicPower.Enable)
	if err != nil {
		msg := "Unable to update Dynamic Power status."
		log.WithFields(logrus.Fields{
			"component": component,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
			"Error":     err,
		}).Warn(msg)
		return errors.New(msg)
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"Vendor":    b.vendor,
		"Model":     b.model,
		"Serial":    b.serial,
		"IPAddress": b.ip,
	}).Debug("Dynamic Power config applied successfully.")
	return err

}

func (b *BmcChassisSetup) setIpmiOverLan() (err error) {

	log := b.log
	component := "setIpmiOverLan"

	enable := b.config.IpmiOverLan.Enable
	chassis := b.chassis
	setup := b.setup

	//retrive list of blades in chassis
	blades, err := chassis.Blades()
	if err != nil {
		msg := "Unable to list blades for chassis."
		log.WithFields(logrus.Fields{
			"component": component,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
			"Error":     err,
		}).Error(msg)
		return errors.New(msg)
	}

	for _, blade := range blades {
		log.WithFields(logrus.Fields{
			"component":      component,
			"Vendor":         b.vendor,
			"Model":          b.model,
			"Serial":         b.serial,
			"IPAddress":      b.ip,
			"Blade Serial":   blade.Serial,
			"Blade Position": blade.BladePosition,
			"Enable":         enable,
		}).Debug("Updating IpmiOverLan config.")

		//blade needs to be powered on to set this parameter
		isPoweredOn, _ := chassis.IsOnBlade(blade.BladePosition)
		if isPoweredOn == false {
			_, err = chassis.PowerOnBlade(blade.BladePosition)
			if err != nil {
				msg := "Unable to power up blade to enable IpmiOverLan."
				log.WithFields(logrus.Fields{
					"component": component,
					"Vendor":    b.vendor,
					"Model":     b.model,
					"Serial":    b.serial,
					"IPAddress": b.ip,
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

			//give it a few seconds to power on
			time.Sleep(20 * time.Second)
		}

		_, err = setup.SetIpmiOverLan(blade.BladePosition, enable)
		if err != nil {
			msg := "Unable to update IpmiOverLan status."
			log.WithFields(logrus.Fields{
				"component":      component,
				"Vendor":         b.vendor,
				"Model":          b.model,
				"Serial":         b.serial,
				"IPAddress":      b.ip,
				"Blade Serial":   blade.Serial,
				"Blade Position": blade.BladePosition,
				"Error":          err,
			}).Warn(msg)
			return errors.New(msg)
		}
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"Vendor":    b.vendor,
		"Model":     b.model,
		"Serial":    b.serial,
		"IPAddress": b.ip,
	}).Debug("IpmiOverLan config applied successfully.")

	return err

}

// Enables/ Disables FlexAddress status for each blade in a chassis.
// Each blade is powered down, flex state updated, powered up
func (b *BmcChassisSetup) setFlexAddressState() (err error) { // nolint: gocyclo

	component := "setFlexAddressState"

	chassis := b.chassis
	setup := b.setup
	log := b.log

	enable := b.config.FlexAddress.Enable

	//retrive list of blades in chassis
	blades, err := chassis.Blades()
	if err != nil {
		msg := "Unable to list blades for chassis."
		log.WithFields(logrus.Fields{
			"component": component,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
			"Error":     err,
		}).Error(msg)
		return errors.New(msg)
	}

	for _, blade := range blades {
		//Flex addresses are enabled, disable them.
		if blade.FlexAddressEnabled == true && enable == false {

			log.WithFields(logrus.Fields{
				"component":      component,
				"Vendor":         b.vendor,
				"Model":          b.model,
				"Serial":         b.serial,
				"IPAddress":      b.ip,
				"Blade Serial":   blade.Serial,
				"Blade Position": blade.BladePosition,
				"Current state":  blade.FlexAddressEnabled,
				"Expected state": enable,
			}).Debug("Disabling FlexAddress on blade.")

			isPoweredOn, _ := chassis.IsOnBlade(blade.BladePosition)
			if isPoweredOn {
				_, err = chassis.PowerOffBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable to disable FlexAddress - blade power off failed."
					log.WithFields(logrus.Fields{
						"component": component,
						"Vendor":    b.vendor,
						"Model":     b.model,
						"Serial":    b.serial,
						"IPAddress": b.ip,
						"Error":     err,
					}).Warn(msg)
					return errors.New(msg)
				}

				//generally 10 seconds is enough for the blade to power off
				time.Sleep(10 * time.Second)

			}

			_, err = setup.SetFlexAddressState(blade.BladePosition, false)
			if err != nil {
				msg := "Unable to disable FlexAddress - action failed."
				log.WithFields(logrus.Fields{
					"component": component,
					"Vendor":    b.vendor,
					"Model":     b.model,
					"Serial":    b.serial,
					"IPAddress": b.ip,
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

			//give it a few seconds to change the flex state
			time.Sleep(10 * time.Second)

			chassis.PowerOnBlade(blade.BladePosition)
			if err != nil {
				msg := "Unable to disable FlexAddress - blade power on failed."
				log.WithFields(logrus.Fields{
					"component": component,
					"Vendor":    b.vendor,
					"Model":     b.model,
					"Serial":    b.serial,
					"IPAddress": b.ip,
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

		}
		//flex addresses are disabled, enable them
		if blade.FlexAddressEnabled == false && enable == true {

			log.WithFields(logrus.Fields{
				"component":      component,
				"Vendor":         b.vendor,
				"Model":          b.model,
				"Serial":         b.serial,
				"IPAddress":      b.ip,
				"Blade Serial":   blade.Serial,
				"Blade Position": blade.BladePosition,
				"Current state":  blade.FlexAddressEnabled,
				"Expected state": enable,
			}).Info("Enabling FlexAddress on blade.")

			isPoweredOn, _ := chassis.IsOnBlade(blade.BladePosition)
			if isPoweredOn {

				log.WithFields(logrus.Fields{
					"component":      component,
					"Vendor":         b.vendor,
					"Model":          b.model,
					"Serial":         b.serial,
					"IPAddress":      b.ip,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
				}).Info("Powering off blade, this takes a few seconds..")

				_, err = chassis.PowerOffBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable to enable FlexAddress - blade power off failed."
					log.WithFields(logrus.Fields{
						"component":      component,
						"Vendor":         b.vendor,
						"Model":          b.model,
						"Serial":         b.serial,
						"IPAddress":      b.ip,
						"Blade Serial":   blade.Serial,
						"Blade Position": blade.BladePosition,
						"Error":          err,
					}).Warn(msg)
					return errors.New(msg)
				}

				//generally 10 seconds is enough for the blade to power off
				time.Sleep(10 * time.Second)
			}

			_, err = setup.SetFlexAddressState(blade.BladePosition, true)
			if err != nil {
				msg := "Unable to enable FlexAddress - action failed."
				log.WithFields(logrus.Fields{
					"component":      component,
					"Vendor":         b.vendor,
					"Model":          b.model,
					"Serial":         b.serial,
					"IPAddress":      b.ip,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
					"Error":          err,
				}).Error(msg)
				return errors.New(msg)
			}

			//give it a few seconds to change the flex state
			time.Sleep(10 * time.Second)

			_, err = chassis.PowerOnBlade(blade.BladePosition)
			if err != nil {
				msg := "Unable to enable FlexAddress - blade power on failed."
				log.WithFields(logrus.Fields{
					"component": component,
					"Vendor":    b.vendor,
					"Model":     b.model,
					"Serial":    b.serial,
					"IPAddress": b.ip,
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

		}

	}

	log.WithFields(logrus.Fields{
		"component": component,
		"Vendor":    b.vendor,
		"Model":     b.model,
		"Serial":    b.serial,
		"IPAddress": b.ip,
	}).Debug("FlexAddress config applied successfully.")

	return err
}

// Powers up/down blades as defined in config.
func (b *BmcChassisSetup) setBladesPower() (err error) {

	log := b.log
	component := "setBladesPower"

	chassis := b.chassis
	powerEnable := b.config.BladesPower.Enable

	blades, err := chassis.Blades()
	if err != nil {
		msg := "Unable to list blades for chassis."
		log.WithFields(logrus.Fields{
			"component": component,
			"Vendor":    b.vendor,
			"Model":     b.model,
			"Serial":    b.serial,
			"IPAddress": b.ip,
			"Error":     err,
		}).Error(msg)
		return errors.New(msg)
	}

	for _, blade := range blades {

		bladeIsPoweredOn, _ := chassis.IsOnBlade(blade.BladePosition)

		if bladeIsPoweredOn != powerEnable {
			if powerEnable == true {
				_, err = chassis.PowerOnBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable power up blade."
					log.WithFields(logrus.Fields{
						"component":      component,
						"Vendor":         b.vendor,
						"Model":          b.model,
						"Serial":         b.serial,
						"IPAddress":      b.ip,
						"Blade Serial":   blade.Serial,
						"Blade Position": blade.BladePosition,
						"Error":          err,
					}).Warn(msg)
					return errors.New(msg)
				}

				log.WithFields(logrus.Fields{
					"component":      component,
					"Vendor":         b.vendor,
					"Model":          b.model,
					"Serial":         b.serial,
					"IPAddress":      b.ip,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
				}).Debug("Set blade power state on.")
			}

			if powerEnable == false {
				chassis.PowerOffBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable power down blade."
					log.WithFields(logrus.Fields{
						"component":      component,
						"Vendor":         b.vendor,
						"Model":          b.model,
						"Serial":         b.serial,
						"IPAddress":      b.ip,
						"Blade Serial":   blade.Serial,
						"Blade Position": blade.BladePosition,
						"Error":          err,
					}).Warn(msg)
					return errors.New(msg)
				}

				log.WithFields(logrus.Fields{
					"component":      component,
					"Vendor":         b.vendor,
					"Model":          b.model,
					"Serial":         b.serial,
					"IPAddress":      b.ip,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
				}).Info("Set blade power state off.")
			}
		}
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"Vendor":    b.vendor,
		"Model":     b.model,
		"Serial":    b.serial,
		"IPAddress": b.ip,
	}).Debug("BladesPower config applied successfully.")

	return err
}
