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
	"github.com/bmc-toolbox/bmcbutler/metrics"

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
	ButlerChan     <-chan ButlerMsg
	MetricsChan    chan []metrics.MetricsMsg
	IgnoreLocation bool
}

// spawn a pool of butlers
func (bm *ButlerManager) SpawnButlers() {

	log := bm.Log
	component := "SpawnButlers"

	for i := 1; i <= bm.SpawnCount; i++ {
		bm.SyncWG.Add(1)
		butlerInstance := Butler{
			id:             i,
			log:            bm.Log,
			syncWG:         &bm.SyncWG,
			butlerChan:     bm.ButlerChan,
			metricsChan:    bm.MetricsChan,
			ignoreLocation: bm.IgnoreLocation,
		}
		go butlerInstance.Run()
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"count":     bm.SpawnCount,
	}).Info("Spawned butlers.")

}

func (bm *ButlerManager) Wait() {
	bm.SyncWG.Wait()
}

type Butler struct {
	id             int
	log            *logrus.Logger
	syncWG         *sync.WaitGroup
	butlerChan     <-chan ButlerMsg
	metricsChan    chan<- []metrics.MetricsMsg
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
func (b *Butler) Run() {

	var err error
	log := b.log
	component := "Run"
	defer b.syncWG.Done()

	var exitFlag bool

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	//set bmclib logger params
	bmclibLogger.SetFormatter(&logrus.TextFormatter{})
	if log.Level == logrus.DebugLevel {
		bmclibLogger.SetLevel(logrus.DebugLevel)
	}

	go func() {
		_ = <-sigChan
		exitFlag = true

		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
		}).Warn("Interrupt SIGINT/SIGTERM recieved, butlers will exit gracefully.")
	}()

	for {
		msg, ok := <-b.butlerChan
		if !ok {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
			}).Debug("Butler message channel closed, goodbye.")
			return
		}

		if exitFlag {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
			}).Debug("Butler exited.")
			return
		}

		//if asset has no IPAddress, we can't do anything about it
		if msg.Asset.IpAddress == "" {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
				"Serial":    msg.Asset.Serial,
				"AssetType": msg.Asset.Type,
			}).Debug("Asset was retrieved without any IP address info, skipped.")
			continue
		}

		//if asset has a location defined, we may want to filter it
		if msg.Asset.Location != "" {
			if !myLocation(msg.Asset.Location) && !b.ignoreLocation {
				log.WithFields(logrus.Fields{
					"component":     component,
					"butler-id":     b.id,
					"Serial":        msg.Asset.Serial,
					"AssetType":     msg.Asset.Type,
					"AssetLocation": msg.Asset.Location,
				}).Debug("Butler wont manage asset based on its current location.")
				continue
			}
		}

		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"IP":        msg.Asset.IpAddress,
			"Serial":    msg.Asset.Serial,
			"AssetType": msg.Asset.Type,
			"Vendor":    msg.Asset.Vendor, //at this point the vendor may or may not be known.
			"Location":  msg.Asset.Location,
		}).Info("Connecting to asset..")

		switch {
		case msg.Asset.Setup == true:
			err = b.setupAsset(msg.Setup, &msg.Asset)
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
					"Serial":    msg.Asset.Serial,
					"AssetType": msg.Asset.Type,
					"Vendor":    msg.Asset.Vendor, //at this point the vendor may or may not be known.
					"Location":  msg.Asset.Location,
					"Error":     err,
				}).Warn("Unable to setup asset.")
			}
			continue
		case msg.Asset.Execute == true:
			err = b.executeCommand(msg.Execute, &msg.Asset)
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
					"Serial":    msg.Asset.Serial,
					"AssetType": msg.Asset.Type,
					"Vendor":    msg.Asset.Vendor, //at this point the vendor may or may not be known.
					"Location":  msg.Asset.Location,
					"Error":     err,
				}).Warn("Unable Execute command(s) on asset.")
			}
			continue
		case msg.Asset.Configure == true:
			err = b.configureAsset(msg.Config, &msg.Asset)
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
					"Serial":    msg.Asset.Serial,
					"AssetType": msg.Asset.Type,
					"Vendor":    msg.Asset.Vendor, //at this point the vendor may or may not be known.
					"Location":  msg.Asset.Location,
					"Error":     err,
				}).Warn("Unable to configure asset.")
			}

		default:
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
				"Serial":    msg.Asset.Serial,
				"AssetType": msg.Asset.Type,
				"Vendor":    msg.Asset.Vendor, //at this point the vendor may or may not be known.
				"Location":  msg.Asset.Location,
			}).Warn("Unknown action request on asset.")
		} //switch
	} //for
}

// Sets up the connection to the asset
// Attempts login with current, if that fails tries with default passwords.
// Returns a connection interface that can be type cast to devices.Bmc or devices.BmcChassis
func (b *Butler) setupConnection(asset *asset.Asset, dontCheckCredentials bool) (connection interface{}, err error) {

	log := b.log
	component := "setupConnection"

	bmcUser := viper.GetString("bmcUser")
	bmcPassword := viper.GetString("bmcPassword")

	client, err := discover.ScanAndConnect(asset.IpAddress, bmcUser, bmcPassword)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"IP":        asset.IpAddress,
			"butler-id": b.id,
			"Error":     err,
		}).Warn("Unable to connect to bmc.")
		return connection, err
	}

	switch client.(type) {
	case devices.Bmc:

		bmc := client.(devices.Bmc)
		asset.Model = bmc.BmcType()

		if !dontCheckCredentials {
			//attempt to login with credentials
			err := bmc.CheckCredentials()
			if err == bmcerros.ErrLoginFailed {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
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
						"butler-id": b.id,
						"Asset":     fmt.Sprintf("%+v", asset),
						"Error":     err,
					}).Warn("Unable to login to bmc with default credentials.")
					return bmc, err
				}
			} else if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"butler-id": b.id,
					"Asset":     fmt.Sprintf("%+v", asset),
					"Error":     err,
				}).Warn("Failed to login to bmc.")
				return bmc, err
			}

			//login successful
			//At this point bmc lib can tell us the vendor.
			asset.Vendor = bmc.Vendor()
		}
		return bmc, err

	case devices.BmcChassis:
		chassis := client.(devices.BmcChassis)
		asset.Model = chassis.BmcType()

		err := chassis.CheckCredentials()
		if err == bmcerros.ErrLoginFailed {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
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
					"butler-id": b.id,
					"Asset":     fmt.Sprintf("%+v", asset),
					"Error":     err,
				}).Warn("Unable to login to bmc with default credentials.")
				return chassis, err
			}
		} else if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
				"Asset":     fmt.Sprintf("%+v", asset),
				"Error":     err,
			}).Warn("failed to login to bmc chassis.")
			return chassis, err
		}

		//login successful
		//At this point we know the vendor.
		asset.Vendor = chassis.Vendor()
		return chassis, err
	default:
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": b.id,
			"Asset":     fmt.Sprintf("%+v", asset),
		}).Warn("Unkown device type.")
		return connection, errors.New("Unknown asset type.")
	}

	return connection, err
}
