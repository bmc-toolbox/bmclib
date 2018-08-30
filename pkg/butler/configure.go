package butler

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/resource"
	"github.com/bmc-toolbox/bmclib/devices"
)

// applyConfig setups up the bmc connection
// gets any Asset config templated data rendered
// applies the asset configuration using bmclib
func (b *Butler) configureAsset(config []byte, asset *asset.Asset) (err error) {

	log := b.log
	component := "configureAsset"

	if b.config.DryRun {
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Dry run, asset configuration will be skipped.")
		return nil
	}

	//connect to the bmc/chassis bmc
	client, err := b.setupConnection(asset, false)
	if err != nil {
		return err
	}

	switch client.(type) {
	case devices.Bmc:

		bmc := client.(devices.Bmc)

		asset.Model = bmc.BmcType()

		//Setup a resource instance
		//Get any templated values in the asset config rendered
		resourceInstance := resource.Resource{Log: log, Vendor: asset.Vendor}
		//rendered config is a *cfgresources.ResourcesConfig type
		renderedConfig := resourceInstance.LoadConfigResources(config)

		//time how long it takes to run configure
		metricPrefix := fmt.Sprintf("%s.%s.%s", asset.Location, asset.Vendor, asset.Type)
		defer b.metricsEmitter.MeasureRunTime(
			time.Now().Unix(), fmt.Sprintf("%s.%s", metricPrefix, component))

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

		asset.Model = chassis.BmcType()
		//Setup a resource instance
		//Get any templated values in the asset config rendered
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
