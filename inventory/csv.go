package inventory

// An example inventory source, a csv file.
// to use this source, set source: csv in bmcbutler.yml

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/joelrebel/bmcbutler/asset"
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

func readCsv() []*CsvAsset {

	csvFile_ := viper.GetString("inventory.csv.file")

	var csvAssets []*CsvAsset
	csvFile, err := os.Open(csvFile_)
	if err != nil {
		fmt.Println("Unable to read csv: ", csvFile)
		fmt.Println(err)
		os.Exit(1)
	}

	err = gocsv.UnmarshalFile(csvFile, &csvAssets)
	if err != nil {
		fmt.Println("Unable to unmarshal data from: ", csvFile)
		fmt.Println(err)
		os.Exit(1)
	}

	return csvAssets
}

func (c *Csv) AssetIterBySerial(serial string) {

	serials := strings.Split(serial, ",")

	csvAssets := readCsv()

	assets := make([]asset.Asset, 0)
	for _, serial := range serials {
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
	csvAssets := readCsv()

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
