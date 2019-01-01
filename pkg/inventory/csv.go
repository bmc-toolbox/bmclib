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

// Csv inventory struct holds attributes required to read in assets from a csv file.
type Csv struct {
	Config          *config.Params
	Log             *logrus.Logger
	BatchSize       int //number of inventory assets to return per iteration
	AssetsChan      chan<- []asset.Asset
	FilterAssetType []string
}

// CsvAsset struct holds attributes of an asset listed in a csv file.
type CsvAsset struct {
	BmcAddress string `csv:"bmcaddress"`
	Serial     string `csv:"serial"` //optional
	Vendor     string `csv:"vendor"` //optional
	Type       string `csv:"type"`   //optional
}

func (c *Csv) readCsv() []*CsvAsset {

	log := c.Log

	var csvAssets []*CsvAsset
	csvFile, err := os.Open(c.Config.InventoryParams.File)
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
		c.FilterAssetType = append(c.FilterAssetType, "servers")
	case c.Config.FilterParams.Discretes:
		c.FilterAssetType = append(c.FilterAssetType, "servers")
	case !c.Config.FilterParams.Chassis && !c.Config.FilterParams.Blades && !c.Config.FilterParams.Discretes:
		c.FilterAssetType = []string{"chassis", "servers"}
	}

	//Based on the filter param given, return the asset iterator method.
	switch {
	case c.Config.FilterParams.Serials != "":
		return c.AssetIterBySerial
	case c.Config.FilterParams.Ips != "":
		return c.AssetIterByIP
	default:
		return c.AssetIter
	}

}

// AssetIterBySerial iterates over Assets and passes on the inventory channel.
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
				assets = append(assets, asset.Asset{IPAddresses: []string{item.BmcAddress},
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

// AssetIterByIP reads in list of ips passed in via cli,
// attempts to lookup any attributes for the IP in the inventory,
// and sends an asset for each attribute over the asset channel
func (c *Csv) AssetIterByIP() {

	defer close(c.AssetsChan)

	csvAssets := c.readCsv()

	ips := c.Config.FilterParams.Ips

	//query csv inventory for asset attributes
	assets := make([]asset.Asset, 0)
	for _, ip := range strings.Split(ips, ",") {

		a := asset.Asset{IPAddresses: []string{ip}}

		c.Log.Debug("looking up attributes for IP: ", ip)
		for _, item := range csvAssets {
			if item == nil {
				continue
			}
			if item.BmcAddress == "" {
				continue
			}

			if item.BmcAddress == ip {

				a.Serial = item.Serial
				a.Vendor = item.Vendor
				a.Type = item.Type

			}
		}

		assets = append(assets, a)
	}

	c.AssetsChan <- assets
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

		assets = append(assets, asset.Asset{IPAddresses: []string{item.BmcAddress},
			Serial: item.Serial,
			Vendor: item.Vendor,
			Type:   item.Type})

	}

	c.AssetsChan <- assets
	close(c.AssetsChan)
}
