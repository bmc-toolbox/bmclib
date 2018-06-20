package butler

import (
	"errors"
	"fmt"
	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// applyConfig setups up the bmc connection,
//
// and iterates over the config to be applied.
func (b *Butler) applyConfig(id int, config *cfgresources.ResourcesConfig, asset *asset.Asset) (err error) {

	log := b.Log
	component := "butler-apply-config"

	//if asset.Model == "" {
	//	log.WithFields(logrus.Fields{
	//		"component": component,
	//		"Asset":     fmt.Sprintf("%+v", asset),
	//		"Error":     err,
	//	}).Warn("Unable to use default credentials to connect since asset.Model is unknown.")
	//	return errors.New("asset.Model is unknown.")
	//}

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
		if err != nil {
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
		}
		bmc.ApplyCfg(config)
	case devices.BmcChassis:
		chassis := client.(devices.BmcChassis)
		asset.Model = chassis.BmcType()
		chassis.ApplyCfg(config)
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Config applied.")
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
