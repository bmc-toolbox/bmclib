package butler

import (
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmcbutler/resource"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	bmcerros "github.com/bmc-toolbox/bmclib/errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// applyConfig setups up the bmc connection
// gets any config templated data rendered
// applies the configuration using bmclib
func (b *Butler) applyConfig(id int, config []byte, asset *asset.Asset) (err error) {

	log := b.Log
	component := "butler-apply-config"

	bmcUser := viper.GetString("bmcUser")
	bmcPassword := viper.GetString("bmcPassword")

	client, err := discover.ScanAndConnect(asset.IpAddress, bmcUser, bmcPassword)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"Asset":     fmt.Sprintf("%+v", asset),
			"Error":     err,
		}).Warn("Unable to connect to bmc.")
		return err
	}

	switch client.(type) {
	case devices.Bmc:

		bmc := client.(devices.Bmc)
		asset.Model = bmc.BmcType()

		//attempt to login with credentials
		err := bmc.CheckCredentials()
		if err == bmcerros.ErrLoginFailed {
			log.WithFields(logrus.Fields{
				"component": component,
				"Asset":     fmt.Sprintf("%+v", asset),
				"Error":     err,
			}).Warn("Unable to login to bmc, trying default credentials")

			DefaultbmcUser := viper.GetString(fmt.Sprintf("bmcDefaults.%s.user", asset.Model))
			DefaultbmcPassword := viper.GetString(fmt.Sprintf("bmcDefaults.%s.password", asset.Model))
			bmc.UpdateCredentials(DefaultbmcUser, DefaultbmcPassword)
			err := bmc.CheckCredentials()
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"Asset":     fmt.Sprintf("%+v", asset),
					"Error":     err,
				}).Warn("Unable to login to bmc with default credentials.")
				return err
			}
		} else if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"Asset":     fmt.Sprintf("%+v", asset),
				"Error":     err,
			}).Warn("Something went wrong")
			return err
		}

		//login successful
		//At this point bmc lib can tell us the vendor.
		asset.Vendor = bmc.Vendor()

		//Setup a resource instance
		//Get any templated values in the config rendered
		resourceInstance := resource.Resource{Log: log, Vendor: asset.Vendor}
		//rendered config is a *cfgresources.ResourcesConfig type
		renderedConfig := resourceInstance.LoadConfigResources(config)

		// Apply configuration
		bmc.ApplyCfg(renderedConfig)
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Config applied.")

		bmc.Close()

	case devices.BmcChassis:
		chassis := client.(devices.BmcChassis)
		asset.Model = chassis.BmcType()

		err := chassis.CheckCredentials()
		if err == bmcerros.ErrLoginFailed {
			log.WithFields(logrus.Fields{
				"component": component,
				"Asset":     fmt.Sprintf("%+v", asset),
				"Error":     err,
			}).Warn("Unable to login to bmc, trying default credentials")

			DefaultbmcUser := viper.GetString(fmt.Sprintf("bmcDefaults.%s.user", asset.Model))
			DefaultbmcPassword := viper.GetString(fmt.Sprintf("bmcDefaults.%s.password", asset.Model))
			chassis.UpdateCredentials(DefaultbmcUser, DefaultbmcPassword)
			err := chassis.CheckCredentials()
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"Asset":     fmt.Sprintf("%+v", asset),
					"Error":     err,
				}).Warn("Unable to login to bmc with default credentials.")
				return err
			}
		} else if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"Asset":     fmt.Sprintf("%+v", asset),
				"Error":     err,
			}).Warn("Something went wrong")
			return err
		}

		//login successful
		//At this point we know the vendor.
		asset.Vendor = chassis.Vendor()

		//Setup a resource instance
		//Get any templated values in the config rendered
		resourceInstance := resource.Resource{Log: log, Vendor: asset.Vendor}
		renderedConfig := resourceInstance.LoadConfigResources(config)

		chassis.ApplyCfg(renderedConfig)
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Config applied.")

		chassis.Close()
	default:
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Warn("Unkown device type.")
		return errors.New("Unknown asset type.")
	}

	return err
}
