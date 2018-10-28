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
	"errors"
	"fmt"
	"os"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

type Params struct {
	ButlersToSpawn       int
	Credentials          []map[string]string
	BmcPrimaryUser       string
	BmcPrimaryPassword   string
	BmcSecondaryUser     string
	BmcSecondaryPassword string
	BmcDefaultUser       string
	BmcDefaultPassword   string
	CfgFile              string
	Configure            bool //indicates configure was invoked
	DryRun               bool //when set, don't carry out any actions, just log.
	Setup                bool //indicates setup was invoked
	Execute              bool //indicates execute was invoked
	FilterParams         *FilterParams
	InventoryParams      *InventoryParams
	IgnoreLocation       bool
	Locations            []string
	MetricsParams        *MetricsParams
	Version              string
	Verbose              bool
}

type InventoryParams struct {
	Source        string //dora, csv, enc
	EncExecutable string
	ApiUrl        string
	File          string
}

type MetricsParams struct {
	Client        string //The metrics client.
	Host          string
	Port          int
	Prefix        string
	FlushInterval time.Duration
}

type FilterParams struct {
	Chassis   bool
	Blades    bool
	Discretes bool
	Servers   bool //Server is to replace Discrete, Blade.
	All       bool
	Serials   string //can be one or more serials separated by commas.
	Ips       string
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
	p.MetricsParams.Client = viper.GetString("metrics.clients.client")
	switch p.MetricsParams.Client {
	case "graphite":
		p.MetricsParams.Host = viper.GetString("metrics.clients.graphite.host")
		p.MetricsParams.Port = viper.GetInt("metrics.clients.graphite.port")
		p.MetricsParams.Prefix = viper.GetString("metrics.clients.graphite.prefix")
		p.MetricsParams.FlushInterval = viper.GetDuration("metrics.clients.graphite.flushinterval")
	}

	//Inventory to read assets from
	p.InventoryParams.Source = viper.GetString("inventory.configure.source")
	switch p.InventoryParams.Source {
	case "dora":
		p.InventoryParams.ApiUrl = viper.GetString("inventory.configure.dora.apiUrl")
	case "csv":
		p.InventoryParams.File = viper.GetString("inventory.configure.csv.file")
	case "enc":
		p.InventoryParams.EncExecutable = viper.GetString("inventory.configure.enc.bin")
	}

	//Butlers to spawn
	p.ButlersToSpawn = viper.GetInt("butlersToSpawn")
	if p.ButlersToSpawn == 0 {
		p.ButlersToSpawn = 5
	}

	//Locations this bmcbutler will action assets for,
	//assets in locations not in this slice are ignored.
	p.Locations = viper.GetStringSlice("locations")

	//store credentials, the way bmclogin expects them.
	credentials := viper.GetStringMap("credentials")

	_, keyExists := credentials["accounts"]
	if !keyExists {
		fmt.Println("Error: expected credentials -> accounts config not declared.")
		os.Exit(1)
	}

	for _, m := range credentials["accounts"].([]interface{}) {
		for k, v := range m.(map[interface{}]interface{}) {
			p.Credentials = append(p.Credentials, map[string]string{k.(string): v.(string)})
		}

	}

	//BMC user account credentials
	p.BmcPrimaryUser = viper.GetString("bmcPrimaryUser")
	p.BmcPrimaryPassword = viper.GetString("bmcPrimaryPassword")
	p.BmcSecondaryUser = viper.GetString("bmcSecondaryUser")
	p.BmcSecondaryPassword = viper.GetString("bmcSecondaryPassword")
}

//Reads in vendor default credentials based on given vendor.
func (p *Params) GetDefaultCredentials(vendor string) (err error) {
	p.BmcDefaultUser = viper.GetString(fmt.Sprintf("bmcDefaults.%s.user", vendor))
	p.BmcDefaultPassword = viper.GetString(fmt.Sprintf("bmcDefaults.%s.password", vendor))

	if p.BmcDefaultUser == "" || p.BmcDefaultPassword == "" {
		return errors.New(fmt.Sprintf("No vendor default credentials in config for: %s", vendor))
	}

	return err
}
