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
	"github.com/spf13/cobra"

	"github.com/bmc-toolbox/bmcbutler/pkg/butler"
)

// configureCmd represents the configure command
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute actions on bmcs.",
	Run: func(cmd *cobra.Command, args []string) {
		execute()
	},
}

func init() {
	rootCmd.AddCommand(executeCmd)
}

func execute() {

	runConfig.Execute = true
	inventoryChan, butlerChan, _ := pre()

	//iterate over the inventory channel for assets,
	//create a butler message for each asset along with the configuration,
	//at this point templated values in the config are not yet rendered.
	for assetList := range inventoryChan {
		for _, asset := range assetList {
			asset.Execute = true
			butlerMsg := butler.Msg{Asset: asset, AssetExecute: execCommand}
			butlerChan <- butlerMsg
		}
	}

	post(butlerChan)
}
