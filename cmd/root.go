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
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
	logrusSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/cobra"

	"github.com/bmc-toolbox/bmcbutler/pkg/config"
)

var (
	log            *logrus.Logger
	verbose        bool
	butlersToSpawn int
	cfgFile        string
	version        string
	execCommand    string
	locations      string
	runConfig      *config.Params
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "bmcbutler",
	Short:            "A bmc config manager",
	TraverseChildren: true,
	//setup logger before we run our code, but after init()
	//so cli flags are evaluated
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogger()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupLogger() {

	//setup logging
	log = logrus.New()
	log.Out = os.Stdout

	hook, err := logrusSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "BMCbutler")
	if err != nil {
		log.Error("Unable to connect to local syslog daemon.")
	} else {
		log.AddHook(hook)
	}

	if runConfig.Verbose == true {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}

func init() {

	//bmcbutler runtime configuration.
	runConfig = &config.Params{}
	runConfig.Load(cfgFile)

	rootCmd.PersistentFlags().BoolVarP(&runConfig.Verbose, "verbose", "v", false, "verbose logging")

	//Asset filter params.
	rootCmd.PersistentFlags().BoolVarP(&runConfig.FilterParams.All, "all", "", false, "Action all assets")
	rootCmd.PersistentFlags().BoolVarP(&runConfig.FilterParams.Blades, "blades", "", false, "Action just Blade(s) assets")
	rootCmd.PersistentFlags().BoolVarP(&runConfig.FilterParams.Chassis, "chassis", "", false, "Action just Chassis assets.")
	rootCmd.PersistentFlags().BoolVarP(&runConfig.FilterParams.Servers, "servers", "", false, "Action just Server assets.")
	rootCmd.PersistentFlags().BoolVarP(&runConfig.DryRun, "dryrun", "", false, "Only log assets that will be actioned.")
	rootCmd.PersistentFlags().BoolVarP(&runConfig.FilterParams.Discretes, "discretes", "", false, "Action just Discrete(s) assets")
	rootCmd.PersistentFlags().StringVarP(&runConfig.FilterParams.Serials, "serials", "", "", "Serial(s) of the asset to setup config (separated by commas - no spaces).")
	rootCmd.PersistentFlags().StringVarP(&runConfig.FilterParams.Ips, "ips", "", "", "IP Address(s) of the asset to setup config (separated by commas - no spaces).")

	rootCmd.PersistentFlags().BoolVarP(&runConfig.IgnoreLocation, "ignorelocation", "", false, "Action assets in all locations (ignore locations directive in config)")
	rootCmd.PersistentFlags().IntVarP(&butlersToSpawn, "butlers", "b", 0, "Number of butlers to spawn (overide butlersToSpawn directive in config)")
	rootCmd.PersistentFlags().StringVarP(&locations, "locations", "l", "", "Action assets by given location(s). (overide locations directive in config)")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Configuration file for bmcbutler (default: /etc/bmcbutler/bmcbutler.yml)")

	//move to exec
	rootCmd.PersistentFlags().StringVarP(&execCommand, "command", "", "", "Command to execute on BMCs.")

	//NOTE: to override any config from the flags declared here, see overrideConfigFromFlags in common.go
}
