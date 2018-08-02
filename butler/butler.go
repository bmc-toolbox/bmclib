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
	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	bmcerros "github.com/bmc-toolbox/bmclib/errors"
	bmclibLogger "github.com/bmc-toolbox/bmclib/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ButlerMsg struct {
	Asset   asset.Asset //Asset to be configured
	Config  []byte      //The BMC configuration read in from configuration.yml
	Setup   []byte      //The One time setup configuration read from setup.yml
	Execute string      //Commands to be executed on the BMC
}

type ButlerManager struct {
	Log            *logrus.Logger
	SpawnCount     int
	SyncWG         sync.WaitGroup
	Channel        <-chan ButlerMsg
	IgnoreLocation bool
}

// spawn a pool of butlers
func (bm *ButlerManager) SpawnButlers() {

	log := bm.Log
	component := "SpawnButlers"

	for i := 1; i <= bm.SpawnCount; i++ {
		bm.SyncWG.Add(1)
		butlerInstance := Butler{id: i, log: bm.Log, syncWG: &bm.SyncWG, channel: bm.Channel, ignoreLocation: bm.IgnoreLocation}
		go butlerInstance.Run()
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"count":     bm.SpawnCount,
	}).Info("Spawned butlers.")

	//runtime.Goexit()
}

func (bm *ButlerManager) Wait() {
	bm.SyncWG.Wait()
}

type Butler struct {
	id             int
	log            *logrus.Logger
	syncWG         *sync.WaitGroup
	channel        <-chan ButlerMsg
	ignoreLocation bool
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

	var err error
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
					"component":     component,
					"butler-id":     id,
					"Serial":        msg.Asset.Serial,
					"AssetType":     msg.Asset.Type,
					"AssetLocation": msg.Asset.Location,
				}).Info("Butler wont manage asset based on its current location.")
				continue
			}
		}

		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": id,
			"IP":        msg.Asset.IpAddress,
			"Serial":    msg.Asset.Serial,
			"AssetType": msg.Asset.Type,
			"Vendor":    msg.Asset.Vendor, //at this point the vendor may or may not be known.
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

		err = b.applyConfig(id, msg.Config, &msg.Asset)
		if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": id,
				"Serial":    msg.Asset.Serial,
				"AssetType": msg.Asset.Type,
				"Error":     err,
			}).Warn("Unable to configure asset.")
		}

	}
}
