package inventory

// An example inventory source, a csv file.
// to use this source, set source: csv in bmcbutler.yml

import (
	"os"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
)

// A inventory source is required to have a type with these fields
type Csv struct {
	Config          *config.Params
	Log             *logrus.Logger
	BatchSize       int //number of inventory assets to return per iteration
	AssetsChan      chan<- []asset.Asset
	FilterAssetType []string
}

type CsvAsset struct {
	BmcAddress string `csv:"bmcaddress"`
	Serial     string `csv:"serial"` //optional
	Vendor     string `csv:"vendor"` //optional
	Type       string `csv:"type"`   //optional
}

func (c *Csv) readCsv() []*CsvAsset {

	log := c.Log
	csvFile_ := c.Config.InventoryParams.File

	var csvAssets []*CsvAsset
	csvFile, err := os.Open(csvFile_)
	if err != nil {
		log.Error("Error: ", err)
		os.Exit(1)
	}

	err = gocsv.UnmarshalFile(csvFile, &csvAssets)
	if err != nil {
		log.Error("Error: ", err)
		os.Exit(1)
	}

	return csvAssets
}

//AssetRetrieve looks at c.Config.FilterParams
//and returns the appropriate function that will retrieve assets.
func (c *Csv) AssetRetrieve() func() {

	//setup the asset types we want to retrieve data for.
	switch {
	case c.Config.FilterParams.Chassis:
		c.FilterAssetType = append(c.FilterAssetType, "chassis")
	case c.Config.FilterParams.Blades:
		c.FilterAssetType = append(c.FilterAssetType, "blade")
	case c.Config.FilterParams.Discretes:
		c.FilterAssetType = append(c.FilterAssetType, "discrete")
	case !c.Config.FilterParams.Chassis && !c.Config.FilterParams.Blades && !c.Config.FilterParams.Discretes:
		c.FilterAssetType = []string{"chassis", "blade", "discrete"}
	}

	//Based on the filter param given, return the asset iterator method.
	switch {
	case c.Config.FilterParams.Serials != "":
		return c.AssetIterBySerial
	default:
		return c.AssetIter
	}

}

func (c *Csv) AssetIterBySerial() {

	log := c.Log
	csvAssets := c.readCsv()

	serials := c.Config.FilterParams.Serials
	assets := make([]asset.Asset, 0)
	for _, serial := range strings.Split(serials, ",") {

		log.Debug("Fetching asset from csv by serial: ", serial)
		for _, item := range csvAssets {
			if item == nil {
				continue
			}
			if item.BmcAddress == "" {
				continue
			}

			if item.Serial == serial {
				assets = append(assets, asset.Asset{IpAddresses: []string{item.BmcAddress},
					Serial: item.Serial,
					Vendor: item.Vendor,
					Type:   item.Type})
			}
		}
	}

	//pass the asset to the channel
	c.AssetsChan <- assets
	close(c.AssetsChan)

}

// AssetIter reads in assets and passes them to the inventory channel.
func (c *Csv) AssetIter() {

	//Asset needs to be an inventory asset
	csvAssets := c.readCsv()

	assets := make([]asset.Asset, 0)
	for _, item := range csvAssets {

		if item == nil {
			continue
		}

		if item.BmcAddress == "" {
			continue
		}

		assets = append(assets, asset.Asset{IpAddresses: []string{item.BmcAddress},
			Serial: item.Serial,
			Vendor: item.Vendor,
			Type:   item.Type})

	}

	c.AssetsChan <- assets
	close(c.AssetsChan)
}
