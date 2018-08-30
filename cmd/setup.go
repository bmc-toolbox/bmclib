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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bmc-toolbox/bmcbutler/pkg/butler"
	"github.com/bmc-toolbox/bmcbutler/pkg/resource"
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
	runConfig.Setup = true
	inventoryChan, butlerChan := pre()

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

			butlerMsg := butler.ButlerMsg{Asset: asset, AssetSetup: config}
			butlerChan <- butlerMsg
		}

		//if sigterm is received, break out.
		if exitFlag {
			break
		}
	}

	post(butlerChan)
}
