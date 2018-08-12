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

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	logrusSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	log              *logrus.Logger
	verbose          bool
	cfgFile          string
	inventorySource  string
	classifierSource string
	version          string
	assetType        string
	serial           string
	execCommand      string
	ipList           string
	isChassis        bool
	isBlade          bool
	isDiscrete       bool
	all              bool
	ignoreLocation   bool
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
		validateArgs()
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

	if verbose == true {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
}

func validateArgs() {

	//one of these args are required
	if all == false && serial == "" && ipList == "" {
		log.Error("Either --all OR --serial OR --ip expected.")
		os.Exit(1)
	}

	if all == true && serial != "" && ipList != "" {
		log.Error("--all, --serial, --ip are mutually exclusive args.")
		os.Exit(1)
	}

	//if --all or --ipList was passed
	if all == true || ipList != "" {
		return
	}

	if serial != "" && (isChassis == false && isBlade == false && isDiscrete == false) {
		log.Error("--serial requires one of --chassis | --blade | -discrete")
		os.Exit(1)
	}

	if isChassis == true && isBlade == true {
		log.Error("Either --chassis or --blade may be specified.")
		os.Exit(1)
	}

	if isChassis == true && isDiscrete == true {
		log.Error("Either --chassis or --discrete may be specified.")
		os.Exit(1)
	}

	if isDiscrete == true && isBlade == true {
		log.Error("Either --discrete or --blade may be specified.")
		os.Exit(1)
	}

	if isChassis == true {
		assetType = "chassis"
	} else if isBlade == true {
		assetType = "blade"
	} else if isDiscrete == true {
		assetType = "discrete"
	} else {
		log.Error("Asset type not known, see --help.")
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
	rootCmd.PersistentFlags().StringVarP(&serial, "serial", "", "", "Serial(s) of the asset to setup config (separated by commas - no spaces).")
	rootCmd.PersistentFlags().StringVarP(&ipList, "iplist", "", "", "IP Address(s) of the asset to setup config (separated by commas - no spaces).")
	rootCmd.PersistentFlags().StringVarP(&execCommand, "command", "", "", "Command to execute on BMCs.")
	rootCmd.PersistentFlags().BoolVarP(&isChassis, "chassis", "", false, "Use in conjuction with --serial, declare asset to be a chassis")
	rootCmd.PersistentFlags().BoolVarP(&isBlade, "blade", "", false, "Use in conjuction with --serial, declare asset to be a blade")
	rootCmd.PersistentFlags().BoolVarP(&isDiscrete, "discrete", "", false, "Use in conjuction with --serial, declare asset to be a discrete")
	rootCmd.PersistentFlags().BoolVarP(&all, "all", "", false, "Runs configuration/setup on all assets")
	rootCmd.PersistentFlags().BoolVarP(&ignoreLocation, "ignorelocation", "", false, "Ignores location")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "Configuration file for bmcbutler (default: /etc/bmcbutler/bmcbutler.yml)")

	cobra.OnInitialize(initConfig)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
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

}
