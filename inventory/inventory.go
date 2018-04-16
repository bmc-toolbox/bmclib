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
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/ncode/bmcbutler/asset"
	"io/ioutil"
	"net/http"
	"strings"
)

type Inventory struct {
	Log       *logrus.Logger
	BatchSize int
	Channel   chan<- []asset.Asset
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
func (i *Inventory) setLocation(doraInventoryAssets []asset.Asset) {

	component := "inventory"
	log := i.Log

	apiUrl := viper.GetString("inventory.dora.apiUrl")
	queryUrl := fmt.Sprintf("%s/v1/scanned_ports?filter[port]=22&filter[ip]=", apiUrl)

	//collect IpAddresses used to look up the location
	ips := make([]string, 0)

	for _, asset := range doraInventoryAssets {
		ips = append(ips, asset.IpAddress)
	}

	queryUrl += strings.Join(ips, ",")
	resp, err := http.Get(queryUrl)
	if err != nil {
		log.WithFields(logrus.Fields{
			"component": component,
			"url":       queryUrl,
			"error":     err,
		}).Fatal("Failed to query dora for networks.")
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
		}).Fatal("Unable to unmarshal data returned from dora.")
	}

	// for each scanned IP update respective asset Location
	for _, scannedPortAsset := range doraScannedPortAssets.Data {
		for idx, inventoryAsset := range doraInventoryAssets {
			if scannedPortAsset.Attributes.ScannedAddress == inventoryAsset.IpAddress {
				doraInventoryAssets[idx].Location = scannedPortAsset.Attributes.Site
			}
		}
	}
}

// A routine that returns data to iter over
func (i *Inventory) AssetIter() {

	//Asset needs to be an inventory asset
	//Iter stuffs assets into an array of Assets
	//Iter writes the assets array to the channel
	// split out dora code into dora.go

	component := "inventory"
	log := i.Log

	apiUrl := viper.GetString("inventory.dora.apiUrl")
	//queryUrl := fmt.Sprintf("%s/v1/blades?page[offset]=%d&page[limit]=%d", apiUrl, 0, i.BatchSize)
	// TESTING - return only HP
	queryUrl := fmt.Sprintf("%s/v1/blades?page[offset]=%d&page[limit]=%d&filter[vendor]!=Dell", apiUrl, 0, i.BatchSize)

	for {
		assets := make([]asset.Asset, 0)

		resp, err := http.Get(queryUrl)
		if err != nil {
			log.WithFields(logrus.Fields{
				"component": component,
				"url":       queryUrl,
				"error":     err,
			}).Fatal("Failed to query dora for assets.")
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
			}).Fatal("Unable to unmarshal data returned from dora.")
		}

		// for each asset, get its location
		// store in the assets slice
		// if an asset has no bmcAddress we log and skip it.
		for _, item := range doraAssets.Data {

			if item.Attributes.BmcAddress == "" {
				log.WithFields(logrus.Fields{
					"component": component,
					"DoraAsset": fmt.Sprintf("%+v", item),
				}).Warn("Asset location could not be determined, since the asset has no IP.")
				continue
			}

			assets = append(assets,
				asset.Asset{IpAddress: item.Attributes.BmcAddress,
					Serial: item.Attributes.Serial,
					Vendor: item.Attributes.Vendor,
					Type:   "blade"})

		}

		//set the location for the assets
		i.setLocation(assets)

		//pass the asset to the channel
		i.Channel <- assets

		// if we reached the end of dora assets
		if doraAssets.Links.Next == "" {
			log.WithFields(logrus.Fields{
				"component": component,
				"url":       queryUrl,
			}).Info("Reached end of assets in dora")
			close(i.Channel)
			break
		}

		// next url to query
		queryUrl = fmt.Sprintf("%s%s", apiUrl, doraAssets.Links.Next)
		//fmt.Printf("--> %s\n", queryUrl)

	}

}
