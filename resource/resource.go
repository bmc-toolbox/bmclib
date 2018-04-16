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
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Resource struct {
	Log *logrus.Logger
}

type User struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
	Role     string `yaml:"role"`
}

type Syslog struct {
	Server string `yaml:"server"`
	Port   int    `yaml:"port"`
	Enable bool   `yaml:"enable"`
}

type Ldap struct {
	Server  string `yaml:"server"`
	Port    int    `yaml:"port"`
	Enable  bool   `yaml:"enable"`
	Role    string `yaml:"role"`
	BaseDn  string `yaml:"baseDn"`
	GroupDn string `yaml:"groupDn"`
}

type Network struct {
	Hostname    string `yaml:"hostname"`
	DNSFromDHCP bool   `yaml:"dnsfromdhcp"`
}

type CommonCfg struct {
	Syslog  Syslog      `yaml:"syslog"`
	Ldap    Ldap        `yaml:"ldap"`
	Network Network     `yaml:"network"`
	User    interface{} `yaml:"user'`
}

func (r *Resource) getUserCfg(cfg []interface{}) []User {

	var users []User

	for _, userCfgInterface := range cfg {
		userCfg := userCfgInterface.(map[interface{}]interface{})
		user := User{
			Name:     userCfg["name"].(string),
			Password: userCfg["password"].(string),
			Role:     userCfg["role"].(string),
		}

		users = append(users, user)
	}

	return users

}

func (r *Resource) getSyslogCfg(cfg map[interface{}]interface{}) Syslog {

	syslog := Syslog{
		Server: cfg["server"].(string),
		Port:   cfg["port"].(int),
		Enable: cfg["enable"].(bool),
	}

	return syslog

}

func (r *Resource) getLdapCfg(cfg map[interface{}]interface{}) Ldap {

	//how do we validate all required params are given?
	//this func will fail if any field is nil.

	ldap := Ldap{
		Server:  cfg["server"].(string),
		Port:    cfg["port"].(int),
		Enable:  cfg["enable"].(bool),
		Role:    cfg["role"].(string),
		BaseDn:  cfg["baseDn"].(string),
		GroupDn: cfg["groupDn"].(string),
	}

	return ldap

}

func (r *Resource) ReadResources() []interface{} {
	// returns an slice of resources to be applied,
	// in the order they need to be applied

	config := make([]interface{}, 0)
	component := "resource"
	log := r.Log

	cfgDir := viper.GetString("bmcCfgDir")
	cfgFile := fmt.Sprintf("%s/%s", cfgDir, "common.yml")

	yamlData, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"cfgFile":   cfgFile,
			"error":     err,
		}).Fatal("Unable to read bmc cfg yaml.")
	}

	commonCfg := make(map[string]interface{})
	err = yaml.Unmarshal([]byte(yamlData), &commonCfg)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"cfgFile":   cfgFile,
			"error":     err,
		}).Fatal("Unable to Unmarshal common.yml.")
	}

	// config is appended to the config slice
	// in the order its defined
	for resource, resourceCfg := range commonCfg {
		switch resource {
		case "syslog":
			syslog := r.getSyslogCfg(resourceCfg.(map[interface{}]interface{}))
			config = append(config, syslog)
		case "ldap":
			ldap := r.getLdapCfg(resourceCfg.(map[interface{}]interface{}))
			config = append(config, ldap)
		case "user":
			user := r.getUserCfg(resourceCfg.([]interface{}))
			config = append(config, user)
		default:
			log.WithFields(logrus.Fields{
				"component": component,
				"cfgFile":   cfgFile,
				"resource":  resource,
			}).Warn("Unknown resource declared.")

		}
	}

	return config
}
