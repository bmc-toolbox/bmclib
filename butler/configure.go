package butler

import (
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmcbutler/resource"
	"github.com/bmc-toolbox/bmclib/devices"

	"github.com/sirupsen/logrus"
)

// applyConfig setups up the bmc connection
// gets any config templated data rendered
// applies the configuration using bmclib
func (b *Butler) configureAsset(config []byte, asset *asset.Asset) (err error) {

	log := b.log
	component := "configureAsset"

	//connect to the bmc/chassis bmc
	client, err := b.setupConnection(asset, false)
	if err != nil {
		return err
	}

	switch client.(type) {
	case devices.Bmc:

		bmc := client.(devices.Bmc)

		//Setup a resource instance
		//Get any templated values in the config rendered
		resourceInstance := resource.Resource{Log: log, Vendor: asset.Vendor}
		//rendered config is a *cfgresources.ResourcesConfig type
		renderedConfig := resourceInstance.LoadConfigResources(config)

		// Apply configuration
		bmc.ApplyCfg(renderedConfig)
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Config applied.")

		bmc.Close()
	case devices.BmcChassis:
		chassis := client.(devices.BmcChassis)

		//Setup a resource instance
		//Get any templated values in the config rendered
		resourceInstance := resource.Resource{Log: log, Vendor: asset.Vendor}
		renderedConfig := resourceInstance.LoadConfigResources(config)

		chassis.ApplyCfg(renderedConfig)
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Config applied.")

		chassis.Close()
	default:
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Warn("Unkown device type.")
		return errors.New("Unknown asset type.")
	}

	return err
}
