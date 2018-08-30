package butler

import (
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
)

// applyConfig setups up the bmc connection
// gets any config templated data rendered
// applies the configuration using bmclib
func (b *Butler) executeCommand(command string, asset *asset.Asset) (err error) {

	component := "executeCommand"
	log := b.log

	if b.config.DryRun {
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Dry run, won't execute cmd on asset.")
		return nil
	}

	//connect to the bmc/chassis bmc
	client, err := b.setupConnection(asset, true)
	if err != nil {
		return err
	}

	switch client.(type) {
	case devices.Bmc:
		bmc := client.(devices.Bmc)
		success, err := b.executeCommandBmc(bmc, command)
		if err != nil || success != true {
			log.WithFields(logrus.Fields{
				"component":          component,
				"butler-id":          b.id,
				"Serial":             asset.Serial,
				"AssetType":          asset.Type,
				"Vendor":             asset.Vendor, //at this point the vendor may or may not be known.
				"Location":           asset.Location,
				"Command":            command,
				"Command successful": success,
				"Error":              err,
			}).Warn("Command execute returned error.")
		} else {
			log.WithFields(logrus.Fields{
				"component":          component,
				"butler-id":          b.id,
				"Serial":             asset.Serial,
				"AssetType":          asset.Type,
				"Vendor":             asset.Vendor,
				"Location":           asset.Location,
				"Command":            command,
				"Command successful": success,
			}).Debug("Command executed.")

		}
		bmc.Close()
	case devices.BmcChassis:
		chassis := client.(devices.BmcChassis)
		//b.executeCommandChassis(chassis, command)
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Command executed.")
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

func (b *Butler) executeCommandBmc(bmc devices.Bmc, command string) (success bool, err error) {

	switch command {
	case "bmc-reset":
		success, err := bmc.PowerCycleBmc()
		return success, err
	case "powercycle":
		success, err := bmc.PowerCycle()
		return success, err
	default:
		return success, errors.New(fmt.Sprintf("Unknown command: %s", command))
	}

	return success, err
}

//func (b *Butler) executeCommandChassis(chassis devices.BmcChassis, command []byte) (err error) {
//
//	switch string(command) {
//	case "Chassis reset":
//		chassis.PowerCycleBmc()
//	default:
//		return errors.New(fmt.Sprintf("Unknown command: %s", command))
//	}
//
//	return err
//}
