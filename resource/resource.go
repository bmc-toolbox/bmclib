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

package resource

import (
	"fmt"
	"github.com/ncode/bmclib/cfgresources"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Resource struct {
	Log *logrus.Logger
}

func (r *Resource) ReadResources() (config *cfgresources.ResourcesConfig) {
	// returns an slice of resources to be applied,
	// in the order they need to be applied

	component := "resource"
	log := r.Log

	cfgDir := viper.GetString("bmcCfgDir")
	cfgFile := fmt.Sprintf("%s/%s", cfgDir, "common.yml")

	_, err := os.Stat(cfgFile)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"cfgFile":   cfgFile,
			"error":     err,
		}).Fatal("Declared cfg file not found.")
	}

	yamlData, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"cfgFile":   cfgFile,
			"error":     err,
		}).Fatal("Unable to read bmc cfg yaml.")
	}

	//1. read in data from common.yaml
	err = yaml.Unmarshal([]byte(yamlData), &config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"cfgFile":   cfgFile,
			"error":     err,
		}).Fatal("Unable to Unmarshal common.yml.")
	}

	//read in data from vendor directories,
	//update config

	//read in data from dc directories

	//read in data from environment directories,
	return config
}
