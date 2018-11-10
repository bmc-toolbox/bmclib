package butler

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
	"github.com/bmc-toolbox/bmcbutler/pkg/inventory"
	"github.com/bmc-toolbox/bmcbutler/pkg/metrics"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
)

// SetupChassis struct holds attributes required to run chassis setup.
type SetupChassis struct {
	Asset          *asset.Asset
	Config         *config.Params //bmcbutler config, cli params
	Chassis        devices.BmcChassis
	AssetConfig    *cfgresources.SetupChassis
	MetricsEmitter *metrics.Emitter
	ID             int
	Log            *logrus.Logger
}

// SetupChassis method applies setup related configuration for chassis assets.
func (b *Butler) SetupChassis(assetConfig *cfgresources.SetupChassis, asset *asset.Asset, connection devices.BmcChassis) (err error) {

	log := b.log
	component := "setupChassis"
	metric := b.metricsEmitter

	defer metric.MeasureRuntime([]string{"butler", "setupChassis_runtime"}, time.Now())

	chassis := SetupChassis{
		Log:            log,
		Config:         b.config,
		MetricsEmitter: b.metricsEmitter,
		Asset:          asset,
		AssetConfig:    assetConfig,
		ID:             b.id,
		Chassis:        connection,
	}

	if b.config.DryRun {
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Dry run, asset setup will be skipped.")
		return nil
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": b.id,
		"Asset":     fmt.Sprintf("%+v", asset),
	}).Info("Chassis asset to be applied setup configuration.")

	//if chassis setup is done successfully invoke post action.
	if chassis.applyConfig() == true {
		chassis.Post(asset)
	}

	return
}

// Post method is when a chassis was setup successfully.
func (s *SetupChassis) Post(asset *asset.Asset) {

	log := s.Log
	enc := inventory.Enc{
		Config:         s.Config,
		Log:            log,
		MetricsEmitter: s.MetricsEmitter,
	}

	enc.SetChassisInstalled(asset.Serial)

	return
}

func (s *SetupChassis) applyConfig() (configured bool) {

	log := s.Log
	component := "setupChassis"
	configured = true

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.ID,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("Running setup actions on chassis..")

	cfg := reflect.ValueOf(s.AssetConfig).Elem()
	for r := 0; r < cfg.NumField(); r++ {
		if cfg.Field(r).Pointer() == 0 {
			continue
		}
		resourceName := cfg.Type().Field(r).Name
		switch resourceName {
		case "FlexAddress":
			err := s.setFlexAddressState(s.Chassis, s.AssetConfig.FlexAddress.Enable)
			if err != nil {
				configured = false
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.ID,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn("Failed to update FlexAddressState.")
			}
		case "DynamicPower":
			err := s.setDynamicPower(s.Chassis, s.AssetConfig.DynamicPower.Enable)
			if err != nil {
				configured = false
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.ID,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn("Failed to update Dynamic Power state.")
			}
		case "BladesPower":
			err := s.setBladesPower(s.Chassis, s.AssetConfig.BladesPower.Enable)
			if err != nil {
				configured = false
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.ID,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn("Failed to power up all blades in chassis.")
			}
		case "hostname":
		default:
		}
	}

	return configured
}

func (s *SetupChassis) setDynamicPower(chassis devices.BmcChassis, enable bool) (err error) {
	log := s.Log
	component := "setDynamicPower"
	_, err = chassis.SetDynamicPower(enable)
	if err != nil {
		msg := "Unable to update Dynamic Power status."
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": s.ID,
			"Asset":     fmt.Sprintf("%+v", s.Asset),
			"Error":     err,
		}).Warn(msg)
		return errors.New(msg)
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.ID,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("Dynamic Power config applied successfully.")
	return err

}

func (s *SetupChassis) setIpmiOverLan(chassis devices.BmcChassis, enable bool) (err error) {
	log := s.Log
	component := "setIpmiOverLan"

	//retrive list of blades in chassis
	blades, err := chassis.Blades()
	if err != nil {
		msg := "Unable to list blades for chassis."
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": s.ID,
			"Asset":     fmt.Sprintf("%+v", s.Asset),
			"Error":     err,
		}).Error(msg)
		return errors.New(msg)
	}

	for _, blade := range blades {
		log.WithFields(logrus.Fields{
			"component":      component,
			"butler-id":      s.ID,
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
					"butler-id": s.ID,
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
				"butler-id":      s.ID,
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
		"butler-id": s.ID,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("IpmiOverLan config applied successfully.")

	return err

}

// Enables/ Disables FlexAddress status for each blade in a chassis.
// Each blade is powered down, flex state updated, powered up
func (s *SetupChassis) setFlexAddressState(chassis devices.BmcChassis, enable bool) (err error) {

	log := s.Log
	component := "setFlexAddressState"

	//retrive list of blades in chassis
	blades, err := chassis.Blades()
	if err != nil {
		msg := "Unable to list blades for chassis."
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": s.ID,
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
				"butler-id":      s.ID,
				"Blade Serial":   blade.Serial,
				"Blade Position": blade.BladePosition,
				"Current state":  blade.FlexAddressEnabled,
				"Expected state": enable,
			}).Info("Disabling FlexAddress on blade.")

			isPoweredOn, _ := chassis.IsOnBlade(blade.BladePosition)
			if isPoweredOn {
				_, err = chassis.PowerOffBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable to disable FlexAddress - blade power off failed."
					log.WithFields(logrus.Fields{
						"component": component,
						"butler-id": s.ID,
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
					"butler-id": s.ID,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
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
					"butler-id": s.ID,
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
				"butler-id":      s.ID,
				"Blade Serial":   blade.Serial,
				"Blade Position": blade.BladePosition,
				"Current state":  blade.FlexAddressEnabled,
				"Expected state": enable,
			}).Info("Enabling FlexAddress on blade.")

			isPoweredOn, _ := chassis.IsOnBlade(blade.BladePosition)
			if isPoweredOn {

				log.WithFields(logrus.Fields{
					"component":      component,
					"butler-id":      s.ID,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
				}).Info("Powering off blade, this takes a few seconds..")

				_, err = chassis.PowerOffBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable to enable FlexAddress - blade power off failed."
					log.WithFields(logrus.Fields{
						"component":      component,
						"butler-id":      s.ID,
						"Blade Serial":   blade.Serial,
						"Blade Position": blade.BladePosition,
						"Asset":          fmt.Sprintf("%+v", s.Asset),
						"Error":          err,
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
					"component":      component,
					"butler-id":      s.ID,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
					"Asset":          fmt.Sprintf("%+v", s.Asset),
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
					"butler-id": s.ID,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn(msg)
				return errors.New(msg)
			}

		}

	}

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.ID,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("FlexAddress config applied successfully.")

	return err
}

// Powers up/down blades as defined in config.
func (s *SetupChassis) setBladesPower(chassis devices.BmcChassis, powerEnable bool) (err error) {

	log := s.Log
	component := "setBladesPower"

	//retrieve list of blades in chassis
	blades, err := chassis.Blades()
	if err != nil {
		msg := "Unable to list blades for chassis."
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": s.ID,
			"Asset":     fmt.Sprintf("%+v", s.Asset),
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
						"butler-id":      s.ID,
						"Blade Serial":   blade.Serial,
						"Blade Position": blade.BladePosition,
						"Error":          err,
						"Asset":          fmt.Sprintf("%+v", s.Asset),
					}).Warn(msg)
					return errors.New(msg)
				}

				log.WithFields(logrus.Fields{
					"component":      component,
					"butler-id":      s.ID,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
				}).Info("Set blade power state on.")
			}

			if powerEnable == false {
				chassis.PowerOffBlade(blade.BladePosition)
				if err != nil {
					msg := "Unable power down blade."
					log.WithFields(logrus.Fields{
						"component":      component,
						"butler-id":      s.ID,
						"Blade Serial":   blade.Serial,
						"Blade Position": blade.BladePosition,
						"Error":          err,
						"Asset":          fmt.Sprintf("%+v", s.Asset),
					}).Warn(msg)
					return errors.New(msg)
				}

				log.WithFields(logrus.Fields{
					"component":      component,
					"butler-id":      s.ID,
					"Blade Serial":   blade.Serial,
					"Blade Position": blade.BladePosition,
				}).Info("Set blade power state off.")
			}
		}
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.ID,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("BladesPower config applied successfully.")

	return err
}
