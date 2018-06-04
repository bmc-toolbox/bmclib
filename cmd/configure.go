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
	"github.com/joelrebel/bmcbutler/asset"
	"github.com/joelrebel/bmcbutler/butler"
	"github.com/joelrebel/bmcbutler/inventory"
	"github.com/joelrebel/bmcbutler/resource"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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

	// A channel to recieve inventory assets
	inventoryChan := make(chan []asset.Asset)

	inventorySource := viper.GetString("inventory.configure.source")
	butlersToSpawn := viper.GetInt("butlersToSpawn")

	if butlersToSpawn == 0 {
		butlersToSpawn = 5
	}

	switch inventorySource {
	case "csv":
		inventoryInstance := inventory.Csv{Log: log, Channel: inventoryChan}
		if serial == "" {
			go inventoryInstance.AssetIter()
		} else {
			go inventoryInstance.AssetIterBySerial(serial)
		}
	case "dora":
		inventoryInstance := inventory.Dora{Log: log, BatchSize: 10, Channel: inventoryChan}
		// Spawn a goroutine that returns a slice of assets over inventoryChan
		// the number of assets in the slice is determined by the batch size.
		if serial == "" {
			go inventoryInstance.AssetIter()
		} else {
			go inventoryInstance.AssetIterBySerial(serial, assetType)
		}
	default:
		fmt.Println("Unknown inventory source declared in cfg: ", inventorySource)
		os.Exit(1)
	}

	// Read in declared resources
	resourceInstance := resource.Resource{Log: log}
	config := resourceInstance.ReadResources()

	// Spawn butlers to work
	butlerChan := make(chan butler.ButlerMsg, 10)
	butlerInstance := butler.Butler{Log: log, SpawnCount: butlersToSpawn, Channel: butlerChan}
	if serial != "" {
		butlerInstance.IgnoreLocation = true
	}
	go butlerInstance.Spawn()

	//over inventory channel and pass asset lists recieved to bmcbutlers
	for asset := range inventoryChan {
		butlerMsg := butler.ButlerMsg{Assets: asset, Config: config}
		butlerChan <- butlerMsg
	}

	close(butlerChan)
	//wait until butlers are done.
	butlerInstance.Wait()
}
