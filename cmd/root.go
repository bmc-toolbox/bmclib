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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
)

var (
	log              *logrus.Logger
	verbose          bool
	cfgFile          string
	inventorySource  string
	classifierSource string
	version          string
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

	if verbose == true {
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	logFile, err := os.OpenFile("/var/log/bmcbutler.log", os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		if verbose == true {
			m := io.MultiWriter(os.Stdout, logFile)
			logrus.SetOutput(m)
		} else {
			log.Out = logFile
		}
	} else {
		log.Info("Unable to log to /var/log/bmcbutlter.log, using default stderr")
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	cfgFile = "/etc/bmcbutler/bmcbutler.yml"
	// Use config file from the flag.
	viper.SetConfigFile(cfgFile)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
