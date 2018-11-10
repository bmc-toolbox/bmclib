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

	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
	"github.com/bmc-toolbox/bmcbutler/pkg/metrics"
)

// Dora struct holds attributes required to retrieve assets from Dora,
// and pass them to the butlers.
type Dora struct {
	Log             *logrus.Logger
	BatchSize       int
	AssetsChan      chan<- []asset.Asset
	MetricsEmitter  *metrics.Emitter
	Config          *config.Params
	FilterAssetType []string
}

// DoraAssetAttributes struct is used to unmarshal Dora data.
type DoraAssetAttributes struct {
	Serial         string `json:"serial"`
	BmcAddress     string `json:"bmc_address"`
	Vendor         string `json:"vendor"`
	ScannedAddress string `json:"ip"`   // set when we unmarshal the scanned_ports data
	Site           string `json:"site"` // set when we unmarshal the scanned_ports data
}

// DoraAssetData struct is used to unmarshal Dora data.
type DoraAssetData struct {
	Attributes DoraAssetAttributes `json:"attributes"`
}

// DoraLinks struct is used to unmarshal Dora data.
type DoraLinks struct {
	First string `json:"first"`
	Last  string `json:"last"`
	Next  string `json:"next"`
}

// DoraAsset struct is used to unmarshal Dora data.
type DoraAsset struct {
	Data  []DoraAssetData `json:"data"`
	Links DoraLinks       `json:"links"`
}

// for a list of assets, update its location value
func (d *Dora) setLocation(doraInventoryAssets []asset.Asset) (err error) {

	component := "inventory"
	log := d.Log

	apiURL := d.Config.InventoryParams.APIURL
	queryURL := fmt.Sprintf("%s/v1/scanned_ports?filter[port]=22&filter[ip]=", apiURL)

	//collect IPAddresses used to look up the location
	ips := make([]string, 0)

	for _, asset := range doraInventoryAssets {
		ips = append(ips, asset.IPAddress)
	}

	queryURL += strings.Join(ips, ",")
	resp, err := http.Get(queryURL)
	if err != nil || resp.StatusCode != 200 {
		log.WithFields(logrus.Fields{
			"component":   component,
			"url":         queryURL,
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
			"url":       queryURL,
			"error":     err,
		}).Warn("Unable to unmarshal Dora scanned IP info.")
		return err
	}

	// for each scanned IP update respective asset Location
	for _, scannedPortAsset := range doraScannedPortAssets.Data {
		for idx, inventoryAsset := range doraInventoryAssets {
			if scannedPortAsset.Attributes.ScannedAddress == inventoryAsset.IPAddress {
				doraInventoryAssets[idx].Location = scannedPortAsset.Attributes.Site
			}
		}
	}

	return err
}

//AssetRetrieve looks at d.Config.FilterParams
//and returns the appropriate function that will retrieve assets.
func (d *Dora) AssetRetrieve() func() {

	//setup the asset types we want to retrieve data for.
	switch {
	case d.Config.FilterParams.Chassis:
		d.FilterAssetType = append(d.FilterAssetType, "chassis")
	case d.Config.FilterParams.Blades:
		d.FilterAssetType = append(d.FilterAssetType, "blade")
	case d.Config.FilterParams.Discretes:
		d.FilterAssetType = append(d.FilterAssetType, "discrete")
	case !d.Config.FilterParams.Chassis && !d.Config.FilterParams.Blades && !d.Config.FilterParams.Discretes:
		d.FilterAssetType = []string{"chassis", "blade", "discrete"}
	}

	//Based on the filter param given, return the asset iterator method.
	switch {
	case d.Config.FilterParams.Serials != "":
		return d.AssetIterBySerial
	default:
		return d.AssetIter
	}

}

// AssetIterBySerial is an iterator method,
// to retrieve assets from Dora by the given serial numbers,
// assets are then sent over the inventory channel.
func (d *Dora) AssetIterBySerial() {

	serials := d.Config.FilterParams.Serials
	apiURL := d.Config.InventoryParams.APIURL

	component := "inventory"

	log := d.Log
	defer close(d.AssetsChan)

	for _, assetType := range d.FilterAssetType {
		var path string
		//setup the right dora query path
		switch assetType {
		case "blade":
			path = "blades"
		case "discrete":
			path = "discretes"
		default:
			path = assetType
		}

		queryURL := fmt.Sprintf("%s/v1/%s?filter[serial]=", apiURL, path)
		queryURL += strings.ToLower(serials)
		assets := make([]asset.Asset, 0)

		resp, err := http.Get(queryURL)
		if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"url":       queryURL,
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
				"url":       queryURL,
				"error":     err,
			}).Fatal("Unable to unmarshal data returned from dora.")
		}

		if len(doraAssets.Data) == 0 {
			log.WithFields(logrus.Fields{
				"component": component,
				"Query url": queryURL,
			}).Debug("Asset was not located in dora inventory.")
			continue
		} else {
			log.WithFields(logrus.Fields{
				"component": component,
				"Query url": queryURL,
			}).Debug("Asset located in dora inventory.")
		}

		for _, item := range doraAssets.Data {
			if item.Attributes.BmcAddress == "" {
				log.WithFields(logrus.Fields{
					"component": component,
					"DoraAsset": fmt.Sprintf("%+v", item),
				}).Warn("Asset location could not be determined, since the asset has no IP.")
				continue
			}

			assets = append(assets, asset.Asset{IPAddress: item.Attributes.BmcAddress,
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
}

// AssetIter routine that returns data to iter over
func (d *Dora) AssetIter() {

	//Asset needs to be an inventory asset
	//Iter stuffs assets into an array of Assets
	//Iter writes the assets array to the channel

	apiURL := d.Config.InventoryParams.APIURL
	component := "retrieveInventoryAssetsDora"

	metric := d.MetricsEmitter

	defer close(d.AssetsChan)
	//defer d.MetricsEmitter.MeasureSince(component, time.Now())

	log := d.Log

	for _, assetType := range d.FilterAssetType {
		var path string

		//since this asset type in dora is plural.
		if assetType == "blade" {
			path = "blades"
		} else if assetType == "discrete" {
			path = "discretes"
		} else {
			path = assetType
		}

		queryURL := fmt.Sprintf("%s/v1/%s?page[offset]=%d&page[limit]=%d", apiURL, path, 0, d.BatchSize)
		for {
			assets := make([]asset.Asset, 0)

			resp, err := http.Get(queryURL)
			if err != nil || resp.StatusCode != 200 {
				log.WithFields(logrus.Fields{
					"component": component,
					"url":       queryURL,
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
					"url":       queryURL,
					"error":     err,
				}).Fatal("Error unmarshaling data returned from Dora.")
			}

			metric.IncrCounter(
				[]string{"inventory", "assets_fetched_dora"},
				float32(len(doraAssets.Data)))

			// for each asset, get its location
			// store in the assets slice
			// if an asset has no bmcAddress we log and skip it.
			for _, item := range doraAssets.Data {

				if item.Attributes.BmcAddress == "" || item.Attributes.BmcAddress == "0.0.0.0" {
					log.WithFields(logrus.Fields{
						"component": component,
						"DoraAsset": fmt.Sprintf("%+v", item),
					}).Warn("Asset location could not be determined, since the asset has no IP.")

					metric.IncrCounter([]string{"inventory", "assets_noip_dora"}, 1)
					continue
				}

				assets = append(assets,
					asset.Asset{IPAddress: item.Attributes.BmcAddress,
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

				metric.IncrCounter([]string{"inventory", "assets_nolocation_dora"}, 1)
				continue
			}

			metric.IncrCounter(
				[]string{"inventory", "assets_returned_dora"},
				float32(len(assets)))

			//pass the asset to the channel
			d.AssetsChan <- assets

			// if we reached the end of dora assets
			if doraAssets.Links.Next == "" {
				log.WithFields(logrus.Fields{
					"component": component,
					"url":       queryURL,
				}).Info("Reached end of assets in dora")
				break
			}

			// next url to query
			queryURL = fmt.Sprintf("%s%s", apiURL, doraAssets.Links.Next)
		}
	}
}
