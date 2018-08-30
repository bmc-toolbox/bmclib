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

package config

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type Params struct {
	ButlersToSpawn       int
	BmcPrimaryUser       string
	BmcPrimaryPassword   string
	BmcSecondaryUser     string
	BmcSecondaryPassword string
	BmcDefaultUser       string
	BmcDefaultPassword   string
	CfgFile              string
	FilterParams         *FilterParams
	InventoryParams      *InventoryParams
	IgnoreLocation       bool
	Locations            []string
	MetricsParams        *MetricsParams
	Version              string
	Verbose              bool
}

type InventoryParams struct {
	Source string //dora, csv
	ApiUrl string
	File   string
}

type MetricsParams struct {
	Target string
	Server string
	Port   int
	Prefix string
}

type FilterParams struct {
	Chassis  bool
	Blade    bool
	Discrete bool
	All      bool
	Serial   string //can be one or more serials separated by commas.
	Ip       string
}

//Config params constructor
func (p *Params) Load(cfgFile string) {

	//FilterParams holds the configure/setup/execute related host filter cli args.
	p.FilterParams = &FilterParams{}
	p.MetricsParams = &MetricsParams{}
	p.InventoryParams = &InventoryParams{}

	//read in config file with viper
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.SetConfigName("bmcbutler")
		viper.AddConfigPath("/etc/bmcbutler")
		viper.AddConfigPath(fmt.Sprintf("%s/.bmcbutler", home))
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading config:", viper.ConfigFileUsed())
		fmt.Println("  ->", err)
		os.Exit(1)
	}

	//The config file viper is using.
	p.CfgFile = viper.ConfigFileUsed()

	//Read in metrics config
	p.MetricsParams.Target = viper.GetString("metrics.receiver.target")
	switch p.MetricsParams.Target {
	case "graphite":
		p.MetricsParams.Server = viper.GetString("metrics.receiver.graphite.host")
		p.MetricsParams.Port = viper.GetInt("metrics.receiver.graphite.port")
		p.MetricsParams.Prefix = viper.GetString("metrics.receiver.graphite.prefix")
	}

	//Inventory to read assets from
	p.InventoryParams.Source = viper.GetString("inventory.configure.source")
	switch p.InventoryParams.Source {
	case "dora":
		p.InventoryParams.ApiUrl = viper.GetString("inventory.configure.dora.apiUrl")
	case "csv":
		p.InventoryParams.File = viper.GetString("inventory.configure.csv.file")
	}

	//Butlers to spawn
	p.ButlersToSpawn = viper.GetInt("butlersToSpawn")
	if p.ButlersToSpawn == 0 {
		p.ButlersToSpawn = 5
	}

	//Locations this bmcbutler will action assets for,
	//assets in locations not in this slice are ignored.
	p.Locations = viper.GetStringSlice("locations")

	//BMC user account credentials
	p.BmcPrimaryUser = viper.GetString("bmcPrimaryUser")
	p.BmcPrimaryPassword = viper.GetString("bmcPrimaryPassword")
	p.BmcSecondaryUser = viper.GetString("bmcSecondaryUser")
	p.BmcSecondaryPassword = viper.GetString("bmcSecondaryPassword")

}

//Reads in vendor default credentials based on given vendor.
func (p *Params) LoadBmcDefaultCredentials(vendor string) (err error) {
	p.BmcDefaultUser = viper.GetString(fmt.Sprintf("bmcDefaults.%s.user", vendor))
	p.BmcDefaultPassword = viper.GetString(fmt.Sprintf("bmcDefaults.%s.password", vendor))

	if p.BmcDefaultUser == "" || p.BmcDefaultPassword == "" {
		return errors.New(fmt.Sprintf("No vendor default credentials in config for: %s", vendor))
	}

	return err
}
