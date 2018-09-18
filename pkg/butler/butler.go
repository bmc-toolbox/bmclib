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
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
	"github.com/bmc-toolbox/bmcbutler/pkg/metrics"

	bmclibLogger "github.com/bmc-toolbox/bmclib/logging"
)

var (
	ErrBmcConnectionFail = errors.New("Unable to login to bmc") //could be a timeout or just bad credentials.
	ErrUnkownAsset       = errors.New("Unknown asset type")
)

type ButlerMsg struct {
	Asset        asset.Asset //Asset to be configured
	AssetConfig  []byte      //The BMC configuration read in from configuration.yml
	AssetSetup   []byte      //The One time setup configuration read from setup.yml
	AssetExecute string      //Commands to be executed on the BMC
}

type ButlerManager struct {
	Config         *config.Params //bmcbutler config, cli params
	ButlerChan     <-chan ButlerMsg
	Log            *logrus.Logger
	MetricsEmitter *metrics.Emitter
	SyncWG         *sync.WaitGroup
}

// spawn a pool of butlers, wait until they are done.
func (bm *ButlerManager) SpawnButlers() {

	log := bm.Log
	component := "Butler Manager - SpawnButlers()"
	doneChan := make(chan int)
	interruptChan := make(chan struct{})

	defer bm.SyncWG.Done()

	var b int

	//setup interrupt handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		_ = <-sigChan
		interruptChan <- struct{}{}

		log.WithFields(logrus.Fields{
			"component": component,
		}).Warn("Interrupt SIGINT/SIGTERM recieved, butlers will exit gracefully.")
		return
	}()

	//Spawn butlers
	for b = 1; b <= bm.Config.ButlersToSpawn; b++ {
		butlerInstance := Butler{
			butlerChan:     bm.ButlerChan,
			config:         bm.Config,
			doneChan:       doneChan,
			interruptChan:  interruptChan,
			id:             b,
			log:            bm.Log,
			metricsEmitter: bm.MetricsEmitter,
		}
		go butlerInstance.Run()
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"Count":     bm.Config.ButlersToSpawn,
	}).Info("Spawned butlers.")

	bm.MetricsEmitter.UpdateGauge(
		[]string{"butler", "spawned"},
		float32(bm.Config.ButlersToSpawn))

	//wait until butlers are done.
	for b > 1 {
		done := <-doneChan
		log.WithFields(logrus.Fields{
			"component": component,
			"butler-id": done,
		}).Debug("Butler exited.")
		b--
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"Count":     bm.Config.ButlersToSpawn,
	}).Info("All butlers exited.")

}

type Butler struct {
	id             int
	butlerChan     <-chan ButlerMsg
	config         *config.Params //bmcbutler config, cli params
	doneChan       chan<- int
	interruptChan  <-chan struct{}
	log            *logrus.Logger
	metricsEmitter *metrics.Emitter
}

func (b *Butler) myLocation(location string) bool {
	for _, l := range b.config.Locations {
		if l == location {
			return true
		}
	}

	return false
}

// butler recieves bmc config, assets over channel
// iterate over assets and apply config
func (b *Butler) Run() {

	var err error
	log := b.log
	component := "Butler Run"
	metric := b.metricsEmitter

	//set bmclib logger params
	bmclibLogger.SetFormatter(&logrus.TextFormatter{})
	if log.Level == logrus.DebugLevel {
		bmclibLogger.SetLevel(logrus.DebugLevel)
	}

	//flag when a signal is received
	var exitFlag bool

	go func() {
		select {
		case <-b.interruptChan:
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
			}).Debug("Butler recieved interrupt.. will exit.")

			exitFlag = true
			return
		}
	}()

	defer func() { b.doneChan <- b.id }()

	for {
		msg, ok := <-b.butlerChan
		if !ok {
			return
		}

		if exitFlag {
			return
		}

		metric.IncrCounter([]string{"butler", "asset_recvd"}, 1)

		//if asset has no IPAddress, we can't do anything about it
		if msg.Asset.IpAddress == "" {
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
				"Serial":    msg.Asset.Serial,
				"AssetType": msg.Asset.Type,
			}).Debug("Asset was retrieved without any IP address info, skipped.")

			metric.IncrCounter([]string{"butler", "asset_recvd_noip"}, 1)
			continue
		}

		//if asset has a location defined, we may want to filter it
		if msg.Asset.Location != "" {
			if !b.myLocation(msg.Asset.Location) && !b.config.IgnoreLocation {
				log.WithFields(logrus.Fields{
					"component":     component,
					"butler-id":     b.id,
					"Serial":        msg.Asset.Serial,
					"AssetType":     msg.Asset.Type,
					"AssetLocation": msg.Asset.Location,
				}).Debug("Butler wont manage asset based on its current location.")

				metric.IncrCounter([]string{"butler", "asset_recvd_location_unmanaged"}, 1)
				continue
			}
		}

		switch {
		case msg.Asset.Setup == true:
			err = b.setupAsset(msg.AssetSetup, &msg.Asset)
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
				metric.IncrCounter([]string{"butler", "setup_fail"}, 1)
				continue
			}

			metric.IncrCounter([]string{"butler", "setup_success"}, 1)
			continue
		case msg.Asset.Execute == true:
			err = b.executeCommand(msg.AssetExecute, &msg.Asset)
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
				metric.IncrCounter([]string{"butler", "execute_fail"}, 1)
				continue
			}

			metric.IncrCounter([]string{"butler", "execute_success"}, 1)
			continue
		case msg.Asset.Configure == true:
			err = b.configureAsset(msg.AssetConfig, &msg.Asset)
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

				metric.IncrCounter([]string{"butler", "configure_fail"}, 1)
				continue
			}

			metric.IncrCounter([]string{"butler", "configure_success"}, 1)
			continue
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
