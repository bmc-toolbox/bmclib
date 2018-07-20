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
	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/bmc-toolbox/bmcbutler/butler"
	"github.com/bmc-toolbox/bmcbutler/inventory"
	"github.com/bmc-toolbox/bmcbutler/resource"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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
		go inventoryInstance.AssetIter(ipList, true)
	default:
		fmt.Println("Unknown inventory source declared in cfg: ", inventorySource)
		os.Exit(1)
	}

	// Spawn butlers to work
	butlerChan := make(chan butler.ButlerMsg, 10)
	butlerInstance := butler.Butler{Log: log, SpawnCount: butlersToSpawn, Channel: butlerChan}

	// let butler run from any location on any given BMC
	if serial != "" || ipList != "" || ignoreLocation {
		butlerInstance.IgnoreLocation = true
	}

	go butlerInstance.Spawn()

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
			butlerMsg := butler.ButlerMsg{Asset: asset, Setup: config}
			butlerChan <- butlerMsg
		}
	}

	close(butlerChan)
	//wait until butlers are done.
	butlerInstance.Wait()
}
