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
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmcbutler/metrics"

	bmclibLogger "github.com/bmc-toolbox/bmclib/logging"
)

var (
	ErrBmcConnectionFail = errors.New("Unable to login to bmc") //could be a timeout or just bad credentials.
	ErrUnkownAsset       = errors.New("Unknown asset type")
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
	MetricsEmitter metrics.Emitter
	IgnoreLocation bool
}

// spawn a pool of butlers
func (bm *ButlerManager) SpawnButlers() {

	log := bm.Log
	component := "SpawnButlers"

	for i := 1; i <= bm.SpawnCount; i++ {
		butlerInstance := Butler{
			id:             i,
			log:            bm.Log,
			syncWG:         &bm.SyncWG,
			butlerChan:     bm.ButlerChan,
			metricsData:    make(map[string]int),
			metricsEmitter: bm.MetricsEmitter,
			ignoreLocation: bm.IgnoreLocation,
		}
		go butlerInstance.Run()
		bm.SyncWG.Add(1)
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"Count":     bm.SpawnCount,
	}).Debug("Spawned butlers.")
}

func (bm *ButlerManager) Wait() {
	bm.SyncWG.Wait()
}

type Butler struct {
	id             int
	log            *logrus.Logger
	syncWG         *sync.WaitGroup
	butlerChan     <-chan ButlerMsg
	metricsData    map[string]int
	metricsEmitter metrics.Emitter
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
	//flag when a signal is received
	var exitFlag bool

	var metricPrefix, successMetric string

	defer b.metricsEmitter.EmitMetricMap(b.metricsData)

	log := b.log
	component := "ButlerRun"

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

	defer b.syncWG.Done()
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

		//metrics to be sent out are prefixed with this string
		// location.vendor.assetType.configure.connfail
		// e.g: lhr4.dell.bmc.configure.success
		metricPrefix = fmt.Sprintf("%s.%s.%s.configure", msg.Asset.Location, msg.Asset.Vendor, msg.Asset.Type)
		successMetric = fmt.Sprintf("%s.%s", metricPrefix, "success")

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

			b.metricsData[successMetric] += 1
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
