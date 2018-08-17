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
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmcbutler/butler"
	"github.com/bmc-toolbox/bmcbutler/inventory"
	"github.com/bmc-toolbox/bmcbutler/metrics"
	"github.com/bmc-toolbox/bmcbutler/resource"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup onetime configuration for BMCs.",
	Long: `Some BMC configuration options must be set just once,
and this config can cause the BMC and or its dependencies to power reset,
for example: disabling/enabling flex addresses on the Dell m1000e,
this requires all the blades in the chassis to be power cycled.`,
	Run: func(cmd *cobra.Command, args []string) {
		setup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)

}

func setup() {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	//flag when its time to exit.
	var exitFlag bool

	go func() {
		_ = <-sigChan
		exitFlag = true
	}()

	//A sync waitgroup for routines spawned here.
	var setupWG sync.WaitGroup

	// A channel butlers sends metrics to the metrics sender
	metricsChan := make(chan []metrics.MetricsMsg, 5)

	//the metrics forwarder routine
	metricsForwarder := metrics.Metrics{
		Logger:  log,
		Channel: metricsChan,
		SyncWG:  &setupWG,
	}

	//metrics emitter instance, used by methods to emit metrics to the forwarder.
	metricsEmitter := metrics.Emitter{Channel: metricsChan}

	//spawn metrics forwarder routine
	go metricsForwarder.Run()
	setupWG.Add(1)

	// A channel to recieve inventory assets
	inventoryChan := make(chan []asset.Asset)

	butlersToSpawn := viper.GetInt("butlersToSpawn")

	if butlersToSpawn == 0 {
		butlersToSpawn = 5
	}

	inventorySource := viper.GetString("inventory.setup.source")

	//if --iplist was passed, set inventorySource
	if ipList != "" {
		inventorySource = "iplist"
	}

	switch inventorySource {
	case "needSetup":
		inventoryInstance := inventory.NeedSetup{Log: log, BatchSize: 10, Channel: inventoryChan}
		// Spawn a goroutine that returns a slice of assets over inventoryChan
		// the number of assets in the slice is determined by the batch size.
		if serial == "" {
			go inventoryInstance.AssetIter()
		} else {
			go inventoryInstance.AssetIterBySerial(serial, assetType)
		}
	case "iplist":
		inventoryInstance := inventory.IpList{Log: log, BatchSize: 1, Channel: inventoryChan}

		// invoke goroutine that passes assets by IP to the butler,
		// here we declare setup = true since this is a setup action.
		go inventoryInstance.AssetIter(ipList)
	default:
		fmt.Println("Unknown inventory source declared in cfg: ", inventorySource)
		os.Exit(1)
	}

	// Spawn butlers to work
	butlerChan := make(chan butler.ButlerMsg, 10)
	butlerManager := butler.ButlerManager{
		Log:            log,
		SpawnCount:     butlersToSpawn,
		ButlerChan:     butlerChan,
		MetricsEmitter: metricsEmitter,
	}

	// let butler run from any location on any given BMC
	if serial != "" || ipList != "" || ignoreLocation {
		butlerManager.IgnoreLocation = true
	}

	go butlerManager.SpawnButlers()

	//Read in BMC configuration data
	configDir := viper.GetString("bmcCfgDir")
	configFile := fmt.Sprintf("%s/%s", configDir, "setup.yml")

	//returns the file read as a slice of bytes
	//config may contain templated values.
	config, err := resource.ReadYamlTemplate(configFile)
	if err != nil {
		log.Fatal("Unable to read BMC setup configuration: ", configFile, " Error: ", err)
		os.Exit(1)
	}

	//over inventory channel and pass asset lists recieved to bmcbutlers
	for assetList := range inventoryChan {
		for _, asset := range assetList {
			//if signal was received, break out.
			if exitFlag {
				break
			}

			asset.Setup = true

			//NOTE: if all butlers exit, and we're trying to write to butlerChan
			//      this loop is going to be stuck waiting for the butlerMsg to be read,
			//      make sure to break out of this loop or have butlerChan closed in such a case,
			//      for now, we fix this by setting exitFlag to break out of the loop.

			butlerMsg := butler.ButlerMsg{Asset: asset, Setup: config}
			butlerChan <- butlerMsg
		}

		//if sigterm is received, break out.
		if exitFlag {
			break
		}
	}

	close(butlerChan)

	//wait until butlers are done.
	butlerManager.Wait()

	close(metricsChan)
	setupWG.Wait()
}
