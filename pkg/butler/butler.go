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
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
	"github.com/bmc-toolbox/bmcbutler/pkg/metrics"

	bmclibLogger "github.com/bmc-toolbox/bmclib/logging"
)

// Msg (butler messages) are passed over the butlerChan
// they declare assets for butlers to carry actions on.
type Msg struct {
	Asset        asset.Asset //Asset to be configured
	AssetConfig  []byte      //The BMC configuration read in from configuration.yml
	AssetSetup   []byte      //The One time setup configuration read from setup.yml
	AssetExecute string      //Commands to be executed on the BMC
}

// Manager struct holds attributes required to spawn butlers.
type Manager struct {
	Config         *config.Params //bmcbutler config, cli params
	ButlerChan     <-chan Msg
	Log            *logrus.Logger
	StopChan       <-chan struct{}
	MetricsEmitter *metrics.Emitter
	SyncWG         *sync.WaitGroup
}

// SpawnButlers spawns a pool of butlers, waits until they are done.
func (bm *Manager) SpawnButlers() {

	log := bm.Log
	component := "Butler Manager - SpawnButlers()"
	doneChan := make(chan int)

	defer bm.SyncWG.Done()

	var b int

	//Spawn butlers
	for b = 1; b <= bm.Config.ButlersToSpawn; b++ {
		butlerInstance := Butler{
			butlerChan:     bm.ButlerChan,
			config:         bm.Config,
			doneChan:       doneChan,
			stopChan:       bm.StopChan,
			SyncWG:         bm.SyncWG,
			id:             b,
			log:            bm.Log,
			metricsEmitter: bm.MetricsEmitter,
		}
		go butlerInstance.Run()
		bm.SyncWG.Add(1)
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
		}).Trace("Butler exited.")
		b--
	}

	log.WithFields(logrus.Fields{
		"component": component,
		"Count":     bm.Config.ButlersToSpawn,
	}).Info("All butlers exited.")

}

// Butler struct holds attributes required by butler to carry out tasks.
type Butler struct {
	id             int
	butlerChan     <-chan Msg
	config         *config.Params //bmcbutler config, cli params
	doneChan       chan<- int
	stopChan       <-chan struct{}
	SyncWG         *sync.WaitGroup
	log            *logrus.Logger
	metricsEmitter *metrics.Emitter
}

// Run runs a butler,
// - receives BMC config, assets over channel
// - iterates over assets and applies config
func (b *Butler) Run() {

	log := b.log
	component := "Butler Run"

	defer func() { b.doneChan <- b.id; b.SyncWG.Done() }()

	//set bmclib logger params
	bmclibLogger.SetFormatter(&logrus.TextFormatter{})
	if log.Level == logrus.DebugLevel {
		bmclibLogger.SetLevel(logrus.DebugLevel)
	}

	for {
		select {
		case msg, ok := <-b.butlerChan:
			if !ok {
				return
			}
			b.msgHandler(msg)
			//spew.Dump(msg)
		case <-b.stopChan:
			log.WithFields(logrus.Fields{
				"component": component,
				"butler-id": b.id,
			}).Debug("Butler received interrupt.. will exit.")
			return
		}
	}
}
