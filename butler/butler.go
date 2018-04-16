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
	"github.com/ncode/bmc/devices"
	"github.com/ncode/bmc/discover"
	"github.com/ncode/bmcbutler/asset"
	"github.com/ncode/bmcbutler/resource"
	//"reflect"
	"runtime"
	"sync"
	"time"
)

type ButlerMsg struct {
	Assets []asset.Asset
	Config []interface{}
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
			}).Info("Channel returned !ok, goodbye.")
			return
		}

		for _, asset := range msg.Assets {
			fmt.Printf("Applying config for: %+v\n", asset)
			b.applyConfig(id, msg.Config, asset)
		}

	}
}

// applyConfig setups up the bmc connection,
// and iterates over the config to be applied.
func (b *Butler) applyConfig(id int, config []interface{}, asset asset.Asset) {

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
		}).Warn("Unable to connect to bmc.")
		return
	}

	bmc, ok := client.(devices.Bmc)
	if ok {
		for _, configResource := range config {

			switch resourceType := configResource.(type) {
			case resource.Ldap:
				return
			case resource.Syslog:
				syslogCfg := configResource.(resource.Syslog)
				_, err := bmc.SyslogSet(syslogCfg.Server, syslogCfg.Port, syslogCfg.Enable)
				if err != nil {
					log.WithFields(logrus.Fields{
						"component": component,
						"butler-id": id,
						"Error":     err,
						"Asset":     fmt.Sprintf("%+v", asset),
					}).Warn("Unable to set syslog resource.")
					return
				} else {
					log.WithFields(logrus.Fields{
						"component": component,
						"butler-id": id,
						"Asset":     fmt.Sprintf("%+v", asset),
					}).Info("Applied syslog config.")
				}
			case []resource.User:
				return
			default:
				log.WithFields(logrus.Fields{
					"component":    component,
					"butler-id":    id,
					"resourceType": resourceType,
					"Config":       fmt.Sprintf("%+v", configResource),
				}).Warn("Unsupported resource type.")
				return
			}
		}

		time.Sleep(1000 * time.Millisecond)
		bmc.SshClose()
	}
	//		status, err := bmc.IsOn()

}
