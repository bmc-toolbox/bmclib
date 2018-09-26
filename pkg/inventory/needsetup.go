package inventory

import (
	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
)

// A inventory source is required to have a type with these fields
type NeedSetup struct {
	Log             *logrus.Logger
	BatchSize       int                  //number of inventory assets to return per iteration
	Channel         chan<- []asset.Asset //the channel to send inventory assets over
	Config          *config.Params       //bmcbutler config
	FilterAssetType []string
}

//AssetRetrieve looks at n.Config.FilterParams
//and returns the appropriate function that will retrieve assets.
func (n *NeedSetup) AssetRetrieve() func() {

	//needsetup is only supported for chassis actions.
	n.FilterAssetType = append(n.FilterAssetType, "chassis")

	//Based on the filter param given, return the asset iterator method.
	switch {
	case n.Config.FilterParams.Serials != "":
		return n.AssetIterBySerial
	default:
		return n.AssetIter
	}

}

// A routine that returns data to iter over
func (n *NeedSetup) AssetIter() {
	//Asset needs to be an inventory asset
	//Iter stuffs assets into an array of Assets
	//Iter writes the assets array to the channel

	n.Log.Println("This needs to be implemented, see csv.go for examples.")
	close(n.Channel)
}

func (n *NeedSetup) AssetIterBySerial() {

	n.Log.Println("This needs to be implemented.")
	close(n.Channel)
}
