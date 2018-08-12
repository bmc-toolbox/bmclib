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

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"sync"
	"time"

	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmcbutler/butler"
	"github.com/bmc-toolbox/bmcbutler/inventory"
	"github.com/bmc-toolbox/bmcbutler/metrics"
	"github.com/bmc-toolbox/bmcbutler/resource"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Apply config to bmcs.",
	Run: func(cmd *cobra.Command, args []string) {
		configure()
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

func configure() {

	var configureWG sync.WaitGroup

	// A channel to recieve inventory assets
	inventoryChan := make(chan []asset.Asset, 5)

	butlersToSpawn := viper.GetInt("butlersToSpawn")

	if butlersToSpawn == 0 {
		butlersToSpawn = 5
	}

	inventorySource := viper.GetString("inventory.configure.source")

	//if --iplist was passed, set inventorySource
	if ipList != "" {
		inventorySource = "iplist"
	}

	switch inventorySource {
	case "csv":
		inventoryInstance := inventory.Csv{Log: log, Channel: inventoryChan}
		if all {
			go inventoryInstance.AssetIter()
		} else {
			go inventoryInstance.AssetIterBySerial(serial)
		}
	case "dora":
		inventoryInstance := inventory.Dora{Log: log, BatchSize: 10, Channel: inventoryChan}
		// Spawn a goroutine that returns a slice of assets over inventoryChan
		// the number of assets in the slice is determined by the batch size.
		if all {
			go inventoryInstance.AssetIter()
		} else {
			go inventoryInstance.AssetIterBySerial(serial, assetType)
		}
	case "iplist":
		inventoryInstance := inventory.IpList{Log: log, BatchSize: 1, Channel: inventoryChan}

		// invoke goroutine that passes assets by IP to spawned butlers,
		// here we declare setup = false since this is a configure action.
		go inventoryInstance.AssetIter(ipList)

	default:
		fmt.Println("Unknown/no inventory source declared in cfg: ", inventorySource)
		os.Exit(1)
	}

	// A channel butlers sends metrics to the metrics sender
	metricsChan := make(chan []metrics.MetricsMsg, 5)
	configureWG.Add(1)

	//the metrics sender
	metricsSender := metrics.Metrics{
		Logger:  log,
		Channel: metricsChan,
		SyncWG:  &configureWG,
	}

	//spawm metrics sender
	go metricsSender.Run()

	// Spawn butlers to work
	butlerChan := make(chan butler.ButlerMsg, 5)
	butlerManager := butler.ButlerManager{
		Log:         log,
		SpawnCount:  butlersToSpawn,
		ButlerChan:  butlerChan,
		MetricsChan: metricsChan,
	}

	if serial != "" {
		butlerManager.IgnoreLocation = true
	}

	go butlerManager.SpawnButlers()

	//give the butlers a second to spawn.
	time.Sleep(1 * time.Second)

	//Read in BMC configuration data
	configDir := viper.GetString("bmcCfgDir")
	configFile := fmt.Sprintf("%s/%s", configDir, "configuration.yml")

	//returns the file read as a slice of bytes
	//config may contain templated values.
	config, err := resource.ReadYamlTemplate(configFile)
	if err != nil {
		log.Fatal("Unable to read BMC configuration: ", configFile, " Error: ", err)
		os.Exit(1)
	}

	//iterate over the inventory channel for assets,
	//create a butler message for each asset along with the configuration,
	//at this point templated values in the config are not yet rendered.
	for assetList := range inventoryChan {
		for _, asset := range assetList {
			asset.Configure = true
			butlerMsg := butler.ButlerMsg{Asset: asset, Config: config}
			butlerChan <- butlerMsg
		}
	}

	close(butlerChan)

	//wait until butlers are done.
	butlerManager.Wait()

	close(metricsChan)
	configureWG.Wait()

}
