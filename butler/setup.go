package butler

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmcbutler/resource"
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	bmcerros "github.com/bmc-toolbox/bmclib/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type SetupAction struct {
	Asset *asset.Asset
	Id    int
	Log   *logrus.Logger
}

func (b *Butler) setupAsset(id int, config []byte, asset *asset.Asset) (err error) {
	log := b.Log
	component := "setupAsset"

	bmcUser := viper.GetString("bmcUser")
	bmcPassword := viper.GetString("bmcPassword")

	client, err := discover.ScanAndConnect(asset.IpAddress, bmcUser, bmcPassword)
	if err != nil {
		return fmt.Errorf("unable to connect to bmc")
	}

	setup := SetupAction{Log: log, Asset: asset, Id: id}

	switch deviceType := client.(type) {
	case devices.Bmc:
		log.Error("Setup not implemented for BMCs ", deviceType)
	case devices.BmcChassis:
		chassis := client.(devices.BmcChassis)
		defer chassis.Close()

		asset.Model = chassis.BmcType()
		err := chassis.CheckCredentials()
		if err == bmcerros.ErrLoginFailed {
			log.WithFields(logrus.Fields{
				"component": component,
				"Asset":     fmt.Sprintf("%+v", asset),
				"Error":     err,
			}).Warn("unable to login to bmc, trying default credentials")

			DefaultbmcUser := viper.GetString(fmt.Sprintf("bmcDefaults.%s.user", asset.Model))
			DefaultbmcPassword := viper.GetString(fmt.Sprintf("bmcDefaults.%s.password", asset.Model))
			chassis.UpdateCredentials(DefaultbmcUser, DefaultbmcPassword)
			err := chassis.CheckCredentials()
			if err != nil {
				return fmt.Errorf("unable to login to bmc with default credentialsc")
			}
		}

		//login successful
		//At this point bmc lib can tell us the vendor.
		asset.Vendor = chassis.Vendor()

		//Setup a resource instance
		//Get any templated values in the config rendered
		resourceInstance := resource.Resource{Log: log, Vendor: asset.Vendor}

		//rendered config is a *cfgresources.ResourcesSetup type
		renderedConfig := resourceInstance.LoadSetupResources(config)

		//if a chassis was setup successfully,
		//call some post setup actions.
		if setup.Chassis(chassis, renderedConfig) == true {
			setup.Post(asset)
		}

		chassis.Close()
	default:
		log.WithFields(logrus.Fields{
			"component":   component,
			"butler-id":   b.id,
			"Device type": fmt.Sprintf("%T", client),
			"Asset":       fmt.Sprintf("%+v", asset),
		}).Error("Unknown device type.")
		return
	}

	return
}

func (s *SetupAction) Chassis(chassis devices.BmcChassis, config *cfgresources.ResourcesSetup) (configured bool) {

	log := s.Log
	component := "setupChassis"
	configured = true

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.Id,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("Running setup actions on chassis..")

	cfg := reflect.ValueOf(config).Elem()
	for r := 0; r < cfg.NumField(); r++ {
		if cfg.Field(r).Pointer() == 0 {
			continue
		}
		resourceName := cfg.Type().Field(r).Name
		switch resourceName {
		case "FlexAddress":
			err := s.setFlexAddressState(chassis, config.FlexAddress.Enable)
			if err != nil {
				configured = false
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn("Failed to update FlexAddressState.")
			}
		case "DynamicPower":
			err := s.setDynamicPower(chassis, config.DynamicPower.Enable)
			if err != nil {
				configured = false
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn("Failed to update Dynamic Power state.")
			}
		case "IpmiOverLan":
			err := s.setIpmiOverLan(chassis, config.IpmiOverLan.Enable)
			if err != nil {
				configured = false
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn("Failed to update IpmiOverLan state.")
			}
		case "hostname":
		default:
		}
	}

	return configured
}

func (s *SetupAction) setDynamicPower(chassis devices.BmcChassis, enable bool) (err error) {
	log := s.Log
	component := "setDynamicPower"
	_, err = chassis.SetDynamicPower(enable)
	if err != nil {
		msg := "Unable to update Dynamic Power status."
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": s.Id,
			"Asset":     fmt.Sprintf("%+v", s.Asset),
			"Error":     err,
		}).Warn(msg)
		return errors.New(msg)
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.Id,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("Dynamic Power config applied successfully.")
	return err

}

func (s *SetupAction) setIpmiOverLan(chassis devices.BmcChassis, enable bool) (err error) {
	log := s.Log
	component := "setIpmiOverLan"

	//retrive list of blades in chassis
	blades, err := chassis.Blades()
	if err != nil {
		msg := "Unable to list blades for chassis."
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": s.Id,
			"Asset":     fmt.Sprintf("%+v", s.Asset),
			"Error":     err,
		}).Error(msg)
		return errors.New(msg)
	}

	for _, blade := range blades {
		log.WithFields(logrus.Fields{
			"component":      component,
			"butler-id":      s.Id,
			"Blade Serial":   blade.Serial,
			"Blade Position": blade.BladePosition,
			"Enable":         enable,
		}).Debug("Updating IpmiOverLan config.")

		//blade needs to be powered on to set this parameter
		isPoweredOn, err := chassis.IsOnBlade(blade.BladePosition)
		if isPoweredOn == false {
			_, err = chassis.PowerOnBlade(blade.BladePosition)
			if err != nil {
				msg := "Unable to power up blade to enable IpmiOverLan."
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

			//give it a few seconds to power on
			time.Sleep(20 * time.Second)
		}

		_, err = chassis.SetIpmiOverLan(blade.BladePosition, enable)
		if err != nil {
			msg := "Unable to update IpmiOverLan status."
			log.WithFields(logrus.Fields{
				"component":      component,
				"butler-id":      s.Id,
				"Blade Serial":   blade.Serial,
				"Blade Position": blade.BladePosition,
				"Asset":          fmt.Sprintf("%+v", s.Asset),
				"Error":          err,
			}).Warn(msg)
			return errors.New(msg)
		}
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.Id,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("IpmiOverLan config applied successfully.")

	return err

}

// Enables/ Disables FlexAddress status for each blade in a chassis.
// Each blade is powered down, flex state updated, powered up
func (s *SetupAction) setFlexAddressState(chassis devices.BmcChassis, enable bool) (err error) {

	log := s.Log
	component := "setFlexAddressState"

	//retrive list of blades in chassis
	blades, err := chassis.Blades()
	if err != nil {
		msg := "Unable to list blades for chassis."
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": s.Id,
			"Asset":     fmt.Sprintf("%+v", s.Asset),
			"Error":     err,
		}).Error(msg)
		return errors.New(msg)
	}

	for _, blade := range blades {
		//Flex addresses are enabled, disable them.
		if blade.FlexAddressEnabled == true && enable == false {

			log.WithFields(logrus.Fields{
				"component":      component,
				"butler-id":      s.Id,
				"Blade Serial":   blade.Serial,
				"Blade Position": blade.BladePosition,
			}).Info("Disabling FlexAddress on blade.")

			isPoweredOn, err := chassis.IsOnBlade(blade.BladePosition)
			if isPoweredOn {
				_, err = chassis.PowerOffBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable to disable FlexAddress - blade power off failed."
					log.WithFields(logrus.Fields{
						"component": component,
						"butler-id": s.Id,
						"Asset":     fmt.Sprintf("%+v", s.Asset),
						"Error":     err,
					}).Warn(msg)
					return errors.New(msg)
				}

				//generally 10 seconds is enough for the blade to power off
				time.Sleep(10 * time.Second)

			}

			_, err = chassis.SetFlexAddressState(blade.BladePosition, false)
			if err != nil {
				msg := "Unable to disable FlexAddress - action failed."
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

			//give it a few seconds to change the flex state
			time.Sleep(5 * time.Second)

			chassis.PowerOnBlade(blade.BladePosition)
			if err != nil {
				msg := "Unable to disable FlexAddress - blade power on failed."
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

		}
		//flex addresses are disabled, enable them
		if blade.FlexAddressEnabled == false && enable == true {

			log.WithFields(logrus.Fields{
				"component":      component,
				"butler-id":      s.Id,
				"Blade Serial":   blade.Serial,
				"Blade Position": blade.BladePosition,
			}).Info("Enabling FlexAddress on blade.")

			isPoweredOn, err := chassis.IsOnBlade(blade.BladePosition)
			if isPoweredOn {

				log.WithFields(logrus.Fields{
					"component":      component,
					"butler-id":      s.Id,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
				}).Info("Powering off blade, this takes a few seconds..")

				_, err = chassis.PowerOffBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable to disable FlexAddress - blade power off failed."
					log.WithFields(logrus.Fields{
						"component": component,
						"butler-id": s.Id,
						"Asset":     fmt.Sprintf("%+v", s.Asset),
						"Error":     err,
					}).Warn(msg)
					return errors.New(msg)
				}

				//generally 10 seconds is enough for the blade to power off
				time.Sleep(10 * time.Second)
			}

			_, err = chassis.SetFlexAddressState(blade.BladePosition, true)
			if err != nil {
				msg := "Unable to enable FlexAddress - action failed."
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Error(msg)
				return errors.New(msg)
			}

			//give it a few seconds to change the flex state
			time.Sleep(5 * time.Second)

			chassis.PowerOnBlade(blade.BladePosition)
			if err != nil {
				msg := "Unable to disable FlexAddress - blade power on failed."
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

		}

	}

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.Id,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("FlexAddress config applied successfully.")

	return err
}
