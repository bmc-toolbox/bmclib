package inventory

// An example inventory source, a csv file.
// to use this source, set source: csv in bmcbutler.yml

import (
	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/gocarina/gocsv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// A inventory source is required to have a type with these fields
type Csv struct {
	Log       *logrus.Logger
	BatchSize int                  //number of inventory assets to return per iteration
	Channel   chan<- []asset.Asset //the channel to send inventory assets over
}

type CsvAsset struct {
	BmcAddress string `csv:"bmcaddress"`
	Serial     string `csv:"serial"` //optional
	Vendor     string `csv:"vendor"` //optional
	Type       string `csv:"type"`   //optional
}

func (c *Csv) readCsv() []*CsvAsset {

	log := c.Log
	csvFile_ := viper.GetString("inventory.configure.csv.file")

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

func (c *Csv) AssetIterBySerial(serial string) {

	log := c.Log
	serials := strings.Split(serial, ",")

	csvAssets := c.readCsv()

	assets := make([]asset.Asset, 0)
	for _, serial := range serials {

		log.Debug("Fetching asset from csv by serial: ", serial)
		for _, item := range csvAssets {
			if item == nil {
				continue
			}
			if item.BmcAddress == "" {
				continue
			}

			if item.Serial == serial {
				assets = append(assets, asset.Asset{IpAddress: item.BmcAddress,
					Serial: item.Serial,
					Vendor: item.Vendor,
					Type:   item.Type})
			}
		}
	}

	//pass the asset to the channel
	c.Channel <- assets
	close(c.Channel)

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

		assets = append(assets, asset.Asset{IpAddress: item.BmcAddress,
			Serial: item.Serial,
			Vendor: item.Vendor,
			Type:   item.Type})

	}

	c.Channel <- assets
	close(c.Channel)
}
