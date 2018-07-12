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
	"sync"

	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/discover"
	bmclibLogger "github.com/bmc-toolbox/bmclib/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ButlerMsg struct {
	Asset  asset.Asset
	Config *cfgresources.ResourcesConfig
	Setup  *cfgresources.ResourcesSetup
}

type Butler struct {
	Log            *logrus.Logger
	SpawnCount     int
	SyncWG         sync.WaitGroup
	Channel        <-chan ButlerMsg
	IgnoreLocation bool
}

// spawn a pool of butlers
func (b *Butler) Spawn() {

	log := b.Log
	component := "butler-spawn"

	for i := 1; i <= b.SpawnCount; i++ {
		b.SyncWG.Add(1)
		go b.butler(i)
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"count":     b.SpawnCount,
	}).Info("Spawned butlers.")

	//runtime.Goexit()

}

func (b *Butler) Wait() {
	b.SyncWG.Wait()
}

func myLocation(location string) bool {
	myLocations := viper.GetStringSlice("locations")
	for _, l := range myLocations {
		if l == location {
			return true
		}
	}

	return false
}

// butler recieves config, assets over channel
// iterate over assets and apply config
func (b *Butler) butler(id int) {

	log := b.Log
	component := "butler-worker"
	defer b.SyncWG.Done()

	//set bmclib logger params
	bmclibLogger.SetFormatter(&logrus.TextFormatter{})
	if log.Level == logrus.DebugLevel {
		bmclibLogger.SetLevel(logrus.DebugLevel)
	}

	for {
		msg, ok := <-b.Channel
		if !ok {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
			}).Debug("butler msg channel was closed, goodbye.")
			return
		}

		//if asset has no IPAddress, we can't do anything about it
		if msg.Asset.IpAddress == "" {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
				"Serial":    msg.Asset.Serial,
				"AssetType": msg.Asset.Type,
			}).Warn("Asset was retrieved without any IP address info, skipped.")
			continue
		}

		//if asset has a location defined, we may want to filter it
		if msg.Asset.Location != "" {
			if !myLocation(msg.Asset.Location) && !b.IgnoreLocation {
				log.WithFields(logrus.Fields{
					"Asset": msg.Asset,
				}).Info("Ignored asset since location did not match.")
				continue
			}
		}

		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": id,
			"IP":        msg.Asset.IpAddress,
			"Serial":    msg.Asset.Serial,
			"AssetType": msg.Asset.Type,
			"Vendor":    msg.Asset.Vendor,
			"Location":  msg.Asset.Location,
		}).Info("Configuring asset..")

		//this asset needs to be setup
		if msg.Asset.Setup == true {
			err = b.setupAsset(id, msg.Setup, &msg.Asset)
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": id,
					"Serial":    msg.Asset.Serial,
					"AssetType": msg.Asset.Type,
					"Error":     err,
				}).Warn("Unable to setup asset.")
			}
			continue
		}

		b.applyConfig(id, msg.Config, &msg.Asset)

	}
}

// connects to the asset and returns the bmc connection
func (b *Butler) connectAsset(asset *asset.Asset, useDefaultLogin bool) (bmcConnection interface{}, err error) {

	var bmcUser, bmcPassword string
	log := b.Log
	component := "butler-connect-asset"

	if useDefaultLogin {
		if asset.Model == "" {
			log.WithFields(logrus.Fields{
				"component":     component,
				"default-login": useDefaultLogin,
				"Asset":         fmt.Sprintf("%+v", asset),
				"Error":         err,
			}).Warn("Unable to use default credentials to connect since asset.Model is unknown.")
			return
		}

		bmcUser = viper.GetString(fmt.Sprintf("bmcDefaults.%s.user", asset.Model))
		bmcPassword = viper.GetString(fmt.Sprintf("bmcDefaults.%s.password", asset.Model))
	} else {
		bmcUser = viper.GetString("bmcUser")
		bmcPassword = viper.GetString("bmcPassword")
	}

	bmcConnection, err = discover.ScanAndConnect(asset.IpAddress, bmcUser, bmcPassword)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component":     component,
			"default-login": useDefaultLogin,
			"Asset":         fmt.Sprintf("%+v", asset),
			"Error":         err,
		}).Warn("Unable to connect to bmc.")
		return bmcConnection, err
	}

	return bmcConnection, err

}
