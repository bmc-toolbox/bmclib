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
	"io/ioutil"
	"os"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

type Resource struct {
	Log    *logrus.Logger
	Vendor string
}

// Reads the given config .yml file, returns it as a slice of bytes.
func ReadYamlTemplate(yamlFile string) (yamlTemplate []byte, err error) {

	//check file exists
	_, err = os.Stat(yamlFile)
	if err != nil {
		return []byte{}, err
	}

	//read in file
	yamlTemplate, err = ioutil.ReadFile(yamlFile)
	if err != nil {
		return []byte{}, err
	}

	return yamlTemplate, nil
}

// Renders templated values in the given config .yml, returns it as a slice of bytes.
func (r *Resource) RenderYamlTemplate(yamlTemplate []byte) (yamlData []byte) {

	log := r.Log
	component := "RenderYamlTemplate"

	//render any templated data
	ctx := plush.NewContext()

	//assign variables that may be used in the template.
	ctx.Set("vendor", strings.ToLower(r.Vendor))

	//render, plush is awesome!
	s, err := plush.Render(string(yamlTemplate), ctx)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component":    component,
			"yamlTemplate": yamlTemplate,
			"error":        err,
		}).Fatal("Error rendering yaml template.")
	}

	return []byte(s)
}

// Config resources are configuration parameters applied periodically,
// Given a yaml template this method gets the template rendered and returns Unmarshalled yaml.
func (r *Resource) LoadConfigResources(yamlTemplate []byte) (config *cfgresources.ResourcesConfig) {

	component := "LoadConfigResources"
	log := r.Log

	yamlData := r.RenderYamlTemplate(yamlTemplate)

	err := yaml.Unmarshal(yamlData, &config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
		}).Fatal("Unable to Unmarshal config resources template.")
	}

	return config
}

// Setup resources are one time setup parameters,
// it may involve resetting/power cycling various dependencies,
//  - e.g blades in a chassis that need to be power cycled
//    if the flex addresses have been enabled/disabled.
// Given a yaml template this method gets the template rendered and returns Unmarshalled yaml.
func (r *Resource) LoadSetupResources(yamlTemplate []byte) (config *cfgresources.ResourcesSetup) {

	component := "LoadSetupResources"
	log := r.Log

	yamlData := r.RenderYamlTemplate(yamlTemplate)

	err := yaml.Unmarshal(yamlData, &config)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"error":     err,
		}).Fatal("Unable to Unmarshal setup resources template.")
	}

	return config
}
