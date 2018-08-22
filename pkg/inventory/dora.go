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

package inventory

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/metrics"
)

type Dora struct {
	Log            *logrus.Logger
	BatchSize      int
	AssetsChan     chan<- []asset.Asset
	MetricsEmitter metrics.Emitter
}

type DoraAssetAttributes struct {
	Serial         string `json:"serial"`
	BmcAddress     string `json:"bmc_address"`
	Vendor         string `json:"vendor"`
	ScannedAddress string `json:"ip"`   // set when we unmarshal the scanned_ports data
	Site           string `json:"site"` // set when we unmarshal the scanned_ports data
}

type DoraAssetData struct {
	Attributes DoraAssetAttributes `json:"attributes"`
}

type DoraLinks struct {
	First string `json:first`
	Last  string `json:last`
	Next  string `json:next`
}

type DoraAsset struct {
	Data  []DoraAssetData `json:"data"`
	Links DoraLinks       `json:"links"`
}

// for a list of assets, update its location value
func (d *Dora) setLocation(doraInventoryAssets []asset.Asset) (err error) {

	component := "inventory"
	log := d.Log

	apiUrl := viper.GetString("inventory.configure.dora.apiUrl")
	queryUrl := fmt.Sprintf("%s/v1/scanned_ports?filter[port]=22&filter[ip]=", apiUrl)

	//collect IpAddresses used to look up the location
	ips := make([]string, 0)

	for _, asset := range doraInventoryAssets {
		ips = append(ips, asset.IpAddress)
	}

	queryUrl += strings.Join(ips, ",")
	resp, err := http.Get(queryUrl)
	if err != nil || resp.StatusCode != 200 {
		log.WithFields(logrus.Fields{
			"component":   component,
			"url":         queryUrl,
			"error":       err,
			"Status code": resp.StatusCode,
		}).Warn("Unable to query Dora for IP location info.")
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	var doraScannedPortAssets DoraAsset
	err = json.Unmarshal(body, &doraScannedPortAssets)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"url":       queryUrl,
			"error":     err,
		}).Warn("Unable to unmarshal Dora scanned IP info.")
		return err
	}

	// for each scanned IP update respective asset Location
	for _, scannedPortAsset := range doraScannedPortAssets.Data {
		for idx, inventoryAsset := range doraInventoryAssets {
			if scannedPortAsset.Attributes.ScannedAddress == inventoryAsset.IpAddress {
				doraInventoryAssets[idx].Location = scannedPortAsset.Attributes.Site
			}
		}
	}

	return err
}

func (d *Dora) AssetIterBySerial(serials string, assetType string) {

	var path string

	//assetTypes := []string{"blade", "chassis", "discrete"}
	component := "inventory"
	log := d.Log
	defer close(d.AssetsChan)

	apiUrl := viper.GetString("inventory.configure.dora.apiUrl")

	if assetType == "blade" {
		path = "blades"
	} else if assetType == "discrete" {
		path = "discretes"
	} else {
		path = assetType
	}

	queryUrl := fmt.Sprintf("%s/v1/%s?filter[serial]=", apiUrl, path)
	queryUrl += strings.ToLower(serials)

	assets := make([]asset.Asset, 0)

	resp, err := http.Get(queryUrl)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"url":       queryUrl,
			"error":     err,
		}).Fatal("Failed to query dora for serial(s).")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	//dora returns a list of assets
	var doraAssets DoraAsset
	err = json.Unmarshal(body, &doraAssets)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"url":       queryUrl,
			"error":     err,
		}).Fatal("Unable to unmarshal data returned from dora.")
	}

	if len(doraAssets.Data) == 0 {
		log.WithFields(logrus.Fields{
			"component": component,
		}).Warn("No data for serial(s) in dora.")
		return
	}

	for _, item := range doraAssets.Data {
		if item.Attributes.BmcAddress == "" {
			log.WithFields(logrus.Fields{
				"component": component,
				"DoraAsset": fmt.Sprintf("%+v", item),
			}).Warn("Asset location could not be determined, since the asset has no IP.")
			continue
		}

		assets = append(assets, asset.Asset{IpAddress: item.Attributes.BmcAddress,
			Serial: item.Attributes.Serial,
			Vendor: item.Attributes.Vendor,
			Type:   assetType})

	}

	//set the location for the assets
	err = d.setLocation(assets)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"Error":     err,
		}).Warn("Unable to determine location of assets.")
		return
	}

	//pass the asset to the channel
	d.AssetsChan <- assets

}

// A routine that returns data to iter over
func (d *Dora) AssetIter() {

	//Asset needs to be an inventory asset
	//Iter stuffs assets into an array of Assets
	//Iter writes the assets array to the channel

	component := "retrieveInventoryAssetsDora"

	//setup metrics related things

	metricsData := make(map[string]int)

	defer close(d.AssetsChan)
	defer d.MetricsEmitter.MeasureRunTime(
		time.Now().Unix(), fmt.Sprintf("assets.dora.%s", component))

	assetTypes := []string{"blade", "chassis", "discrete"}
	log := d.Log
	apiUrl := viper.GetString("inventory.configure.dora.apiUrl")

	for _, assetType := range assetTypes {
		var path string

		//setup default metric values
		metricTotalAssets := fmt.Sprintf("assets.dora.%s.total", assetType)
		metricsData[metricTotalAssets] = 0

		metricNoIp := fmt.Sprintf("assets.dora.%s.noip", assetType)
		metricsData[metricNoIp] = 0

		metricNoLocation := fmt.Sprintf("assets.dora.%s.nolocation", assetType)
		metricsData[metricNoLocation] = 0

		metricAssetsToConfigure := fmt.Sprintf("assets.dora.%s.configure", assetType)
		metricsData[metricAssetsToConfigure] = 0

		//since this asset type in dora is plural.
		if assetType == "blade" {
			path = "blades"
		} else if assetType == "discrete" {
			path = "discretes"
		} else {
			path = assetType
		}

		queryUrl := fmt.Sprintf("%s/v1/%s?page[offset]=%d&page[limit]=%d", apiUrl, path, 0, d.BatchSize)
		for {
			assets := make([]asset.Asset, 0)

			resp, err := http.Get(queryUrl)
			if err != nil || resp.StatusCode != 200 {
				log.WithFields(logrus.Fields{
					"component": component,
					"url":       queryUrl,
					"error":     err,
				}).Fatal("Error querying Dora for assets.")
			}

			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			var doraAssets DoraAsset
			err = json.Unmarshal(body, &doraAssets)
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"url":       queryUrl,
					"error":     err,
				}).Fatal("Error unmarshaling data returned from Dora.")
			}

			metricsData[metricTotalAssets] += len(doraAssets.Data)

			// for each asset, get its location
			// store in the assets slice
			// if an asset has no bmcAddress we log and skip it.
			for _, item := range doraAssets.Data {

				if item.Attributes.BmcAddress == "" || item.Attributes.BmcAddress == "0.0.0.0" {
					log.WithFields(logrus.Fields{
						"component": component,
						"DoraAsset": fmt.Sprintf("%+v", item),
					}).Warn("Asset location could not be determined, since the asset has no IP.")

					metricsData[metricNoIp] += 1
					continue
				}

				assets = append(assets,
					asset.Asset{IpAddress: item.Attributes.BmcAddress,
						Serial: item.Attributes.Serial,
						Vendor: item.Attributes.Vendor,
						Type:   assetType})

			}

			//set the location for the assets
			err = d.setLocation(assets)
			if err != nil {
				log.WithFields(logrus.Fields{
					"component": component,
					"Error":     err,
					"Assets":    fmt.Sprintf("%+v", assets),
				}).Warn("Asset location could not be determined, ignoring assets")

				metricsData[metricNoLocation] += 1
				continue
			}

			metricsData[metricAssetsToConfigure] += len(assets)

			//pass the asset to the channel
			d.AssetsChan <- assets

			// if we reached the end of dora assets
			if doraAssets.Links.Next == "" {
				log.WithFields(logrus.Fields{
					"component": component,
					"url":       queryUrl,
				}).Info("Reached end of assets in dora")
				break
			}

			// next url to query
			queryUrl = fmt.Sprintf("%s%s", apiUrl, doraAssets.Links.Next)
		}

		d.MetricsEmitter.EmitMetricMap(metricsData)
	}

}
