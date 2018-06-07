package butler

import (
	"fmt"
	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
)

// applyConfig setups up the bmc connection,
//
// and iterates over the config to be applied.
func (b *Butler) applyConfig(id int, config *cfgresources.ResourcesConfig, asset *asset.Asset) {

	var useDefaultLogin bool
	var err error
	log := b.Log
	component := "butler-apply-config"

	//this bit is ugly, but I need a way to retry connecting and login,
	//without having to pass around the specific bmc/chassis types (*m1000.M1000e etc..),
	//maybe this can be done in bmclib instead.
	client, err := b.connectAsset(asset, useDefaultLogin)
	if err != nil {
		return
	}

	switch client.(type) {
	case devices.Bmc:

		bmc := client.(devices.Bmc)
		asset.Model = bmc.BmcType()
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
		return
	}

	return

}
