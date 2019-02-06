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

func validateConfigureArgs() {

	//one of these args are required
	if !runConfig.FilterParams.All &&
		!runConfig.FilterParams.Chassis &&
		!runConfig.FilterParams.Servers &&
		runConfig.FilterParams.Serials == "" &&
		runConfig.FilterParams.Ips == "" {

		log.Error("Expected flag missing --all/--chassis/--servers/--serials/--ips (try --help)")
		os.Exit(1)
	}

	if runConfig.FilterParams.All && (runConfig.FilterParams.Serials != "" || runConfig.FilterParams.Ips != "") {
		log.Error("--all --serial --ip are mutually exclusive args.")
		os.Exit(1)
	}

}

func configure() {

	runConfig.Configure = true
	validateConfigureArgs()

	inventoryChan, butlerChan, stopChan := pre()

	//Read in BMC configuration data
	assetConfigDir := viper.GetString("bmcCfgDir")
	assetConfigFile := fmt.Sprintf("%s/%s", assetConfigDir, "configuration.yml")

	//returns the file read as a slice of bytes
	//config may contain templated values.
	assetConfig, err := resource.ReadYamlTemplate(assetConfigFile)
	if err != nil {
		log.Fatal("Unable to read BMC configuration: ", assetConfigFile, " Error: ", err)
		os.Exit(1)
	}

	//iterate over the inventory channel for assets,
	//create a butler message for each asset along with the configuration,
	//at this point templated values in the config are not yet rendered.
loop:
	for {
		select {
		case assetList, ok := <-inventoryChan:
			if !ok {
				break loop
			}
			for _, asset := range assetList {
				asset.Configure = true
				butlerMsg := butler.Msg{Asset: asset, AssetConfig: assetConfig}
				if interrupt {
					break loop
				}

				butlerChan <- butlerMsg
			}
		case <-stopChan:
			interrupt = true
		}
	}

	post(butlerChan)
}
