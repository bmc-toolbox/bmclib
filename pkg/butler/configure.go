package butler

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmclogin"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/butler/configure"
	"github.com/bmc-toolbox/bmcbutler/pkg/resource"
	"github.com/bmc-toolbox/bmclib/devices"
)

// applyConfig setups up the bmc connection
// gets any Asset config templated data rendered
// applies the asset configuration using bmclib
func (b *Butler) configureAsset(config []byte, asset *asset.Asset) (err error) {

	log := b.log
	component := "configureAsset"
	metric := b.metricsEmitter

	if b.config.DryRun {
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Dry run, asset configuration will be skipped.")
		return nil
	}

	defer metric.MeasureRuntime([]string{"butler", "configure_runtime"}, time.Now())

	b.log.WithFields(logrus.Fields{
		"component": component,
		"Serial":    asset.Serial,
		"IPAddress": asset.IPAddress,
	}).Debug("Connecting to asset.")

	bmcConn := bmclogin.Params{
		IpAddresses:     asset.IPAddresses,
		Credentials:     b.config.Credentials,
		CheckCredential: true,
		Retries:         1,
	}

	//connect to the bmc/chassis bmc
	client, loginInfo, err := bmcConn.Login()
	if err != nil {
		return err
	}

	asset.IPAddress = loginInfo.ActiveIpAddress

	switch client.(type) {
	case devices.Bmc:

		bmc := client.(devices.Bmc)

		asset.Type = "server"
		asset.Model = bmc.BmcType()
		asset.Vendor = bmc.Vendor()

		//Setup a resource instance
		//Get any templated values in the asset config rendered
		resourceInstance := resource.Resource{Log: log, Asset: asset}
		//rendered config is a *cfgresources.ResourcesConfig type
		renderedConfig := resourceInstance.LoadConfigResources(config)
		if renderedConfig == nil {
			return errors.New("No BMC configuration to be applied")
		}

		// Apply configuration
		c := configure.NewBmcConfigurator(bmc, asset, b.config.Resources, renderedConfig, b.stopChan, log)
		c.Apply()

		bmc.Close()
	case devices.BmcChassis:
		chassis := client.(devices.BmcChassis)

		asset.Type = "chassis"
		asset.Model = chassis.BmcType()
		asset.Vendor = chassis.Vendor()

		//Setup a resource instance
		//Get any templated values in the asset config rendered
		resourceInstance := resource.Resource{Log: log, Asset: asset}

		renderedConfig := resourceInstance.LoadConfigResources(config)
		if renderedConfig == nil {
			return errors.New("No BMC configuration to be applied")
		}

		if renderedConfig.SetupChassis != nil {
			s := configure.NewBmcChassisSetup(
				chassis,
				asset,
				b.config.Resources,
				renderedConfig.SetupChassis,
				b.config,
				b.metricsEmitter,
				b.stopChan,
				b.log,
			)
			s.Apply()
		}

		// Apply configuration
		c := configure.NewBmcChassisConfigurator(chassis, asset, b.config.Resources, renderedConfig, b.stopChan, log)
		c.Apply()

		chassis.Close()
	default:
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Warn("Unknown device type.")
		return errors.New("Unknown asset type")
	}

	return err
}
