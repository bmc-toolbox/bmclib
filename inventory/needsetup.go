package inventory

import (
	"github.com/bmc-toolbox/bmcbutler/asset"
	"github.com/sirupsen/logrus"
)

// A inventory source is required to have a type with these fields
type NeedSetup struct {
	Log       *logrus.Logger
	BatchSize int                  //number of inventory assets to return per iteration
	Channel   chan<- []asset.Asset //the channel to send inventory assets over
}

// A routine that returns data to iter over
func (n *NeedSetup) AssetIter() {
	//Asset needs to be an inventory asset
	//Iter stuffs assets into an array of Assets
	//Iter writes the assets array to the channel

	n.Log.Println("This needs to be implemented, see csv.go for examples.")
	close(n.Channel)
}

func (n *NeedSetup) AssetIterBySerial(serial string, assetType string) {
	n.Log.Println("This needs to be implemented.")
	close(n.Channel)
}
