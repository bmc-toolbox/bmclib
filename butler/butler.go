// Copyright © 2018 Joel Rebello <joel.rebello@booking.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package butler

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ncode/bmc/cfgresources"
	"github.com/ncode/bmc/devices"
	"github.com/ncode/bmc/discover"
	"github.com/ncode/bmcbutler/asset"
	"sync"
)

type ButlerMsg struct {
	Assets []asset.Asset
	Config *cfgresources.ResourcesConfig
}

type Butler struct {
	Log        *logrus.Logger
	SpawnCount int
	SyncWG     sync.WaitGroup
	Channel    <-chan ButlerMsg
}

// spawn a pool of butlers
func (b *Butler) Spawn() {

	log := b.Log
	component := "butler-spawn"

	for i := 0; i <= b.SpawnCount; i++ {
		b.SyncWG.Add(1)
		go b.butler(i)
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"count":     b.SpawnCount,
	}).Info("Spawned butlers.")

	b.SyncWG.Wait()
	runtime.Goexit()

}

// butler recieves config, assets over channel
// iterate over assets and apply config
func (b *Butler) butler(id int) {

	log := b.Log
	component := "butler"

	defer b.SyncWG.Done()

	for {
		msg, ok := <-b.Channel
		if !ok {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
			}).Info("butler msg channel was closed, goodbye.")
			return
		}

		for _, asset := range msg.Assets {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
				"AssetType": asset.Type,
				"IP":        asset.IpAddress,
				"Vendor":    asset.Vendor,
				"Serial":    asset.Serial,
				"Location":  asset.Location,
			}).Info("Applying config.")

			b.applyConfig(id, msg.Config, asset)
		}
	}
}

// applyConfig setups up the bmc connection,
// and iterates over the config to be applied.
func (b *Butler) applyConfig(id int, config *cfgresources.ResourcesConfig, asset asset.Asset) {

	log := b.Log
	component := "butler-apply-config"
	bmcUser := viper.GetString("bmcUser")
	bmcPassword := viper.GetString("bmcPassword")

	client, err := discover.ScanAndConnect(asset.IpAddress, bmcUser, bmcPassword)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": id,
			"Asset":     fmt.Sprintf("%+v", asset),
			"Error":     err,
		}).Warn("Unable to connect to bmc.")
		return
	}

	switch deviceType := client.(type) {
	case devices.Bmc:
		bmc := client.(devices.Bmc)
		err := bmc.Login()
		if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
				"Asset":     fmt.Sprintf("%+v", asset),
				"Error":     err,
			}).Warn("Unable to login to bmc.")
			return
		} else {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
				"Asset":     fmt.Sprintf("%+v", asset),
			}).Info("Successfully logged into asset.")
		}

		bmc.ApplyCfg(config)
		fmt.Printf("%+v\n", deviceType)
		fmt.Println("Device is a blade")
	case devices.BmcChassis:
		chassis := client.(devices.BmcChassis)

		err := chassis.Login()
		if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
				"Asset":     fmt.Sprintf("%+v", asset),
				"Error":     err,
			}).Warn("Unable to login to bmc.")
			return
		} else {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
				"Asset":     fmt.Sprintf("%+v", asset),
			}).Debug("Logged into asset.")
		}

		chassis.ApplyCfg(config)
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Info("Config applied.")

		chassis.Logout()
	default:
		fmt.Println("--> Unknown device")
		fmt.Printf("%+v\n", client)
		return
	}

	return

}
