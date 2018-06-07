package butler

import (
	"errors"
	"fmt"
	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
	"time"
)

type SetupAction struct {
	Asset       *asset.Asset
	Id          int
	Log         *logrus.Logger
	SetupConfig *cfgresources.ResourcesSetup
}

func (b *Butler) setupAsset(id int, config *cfgresources.ResourcesSetup, asset *asset.Asset) {

	log := b.Log
	component := "setupAsset"

	useDefaultLogin := false
	client, err := b.connectAsset(asset, useDefaultLogin)
	if err != nil {
		log.WithFields(logrus.Fields{
			"butler-id": id,
			"Asset":     asset,
		}).Error("Unable to connect to asset.")
		return
	}

	setup := SetupAction{Log: log, SetupConfig: config, Asset: asset, Id: id}

	switch deviceType := client.(type) {
	case devices.Bmc:
		log.Error("Setup not implemented for BMCs ", deviceType)
	case devices.BmcChassis:

		chassis := client.(devices.BmcChassis)
		defer chassis.Close()

		asset.Model = chassis.BmcType()
		setup.Chassis(chassis)
	default:
		log.WithFields(logrus.Fields{
			"component":   component,
			"butler-id":   id,
			"Device type": fmt.Sprintf("%T", client),
			"Asset":       fmt.Sprintf("%+v", asset),
		}).Error("Unknown device type.")
		return
	}

	return
}

func (s *SetupAction) Chassis(chassis devices.BmcChassis) {

	log := s.Log
	component := "setupChassis"
	config := s.SetupConfig

	cfg := reflect.ValueOf(config.Chassis).Elem()

	for r := 0; r < cfg.NumField(); r++ {
		resourceName := cfg.Type().Field(r).Name
		switch resourceName {
		case "FlexAddressState":
			err := s.applyFlexAddressState(chassis, config.Chassis.FlexAddressState)
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": s.Id,
					"Asset":     fmt.Sprintf("%+v", s.Asset),
					"Error":     err,
				}).Warn("Failed to update FlexAddressState.")
			}
		case "IpmiOverLanState":
			//err := s.applyIpmiOverLanState(chassis, config.Chassis.IpmiOverLanState)
			//if err != nil {
			//	log.WithFields(logrus.Fields{
			//		"component": component,
			//		"butler-id": s.Id,
			//		"Asset":     fmt.Sprintf("%+v", s.Asset),
			//		"Error":     err,
			//	}).Warn("Failed to update IpmiOverLanState.")
			//}
		case "hostname":
		default:
		}
	}

}

//func (s *SetupAction) s.applyIpmiOverLanState(chassis devices.BmcChassis, status string) (err error) {
//	log := s.Log
//	component := "applyFlexAddressConfig"
//
//	status = strings.ToLower(status)
//
//	var enable bool
//	if status == "enable" {
//		enable = true
//	} else if status == "disable" {
//		enable = false
//	} else {
//		msg := "Invalid FlexAddressStatus parameter, expected enable/disable"
//		log.WithFields(logrus.Fields{
//			"component":   component,
//			"butler-id":   s.Id,
//			"Invalid arg": status,
//		}).Error(msg)
//		return errors.New(msg)
//	}
//
//
//}

// Enables/ Disables FlexAddress status for each blade in a chassis.
// Each blade is powered down, flex state updated, powered up
func (s *SetupAction) applyFlexAddressState(chassis devices.BmcChassis, status string) (err error) {

	log := s.Log
	component := "applyFlexAddressConfig"

	status = strings.ToLower(status)

	var enable bool
	if status == "enable" {
		enable = true
	} else if status == "disable" {
		enable = false
	} else {
		msg := "Invalid FlexAddressStatus parameter, expected enable/disable"
		log.WithFields(logrus.Fields{
			"component":   component,
			"butler-id":   s.Id,
			"Invalid arg": status,
		}).Error(msg)
		return errors.New(msg)
	}

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
