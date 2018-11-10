package inventory

import (
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
)

// IPList struct is in inventory source,
// this struct holds attributes to setup the IP list source.
type IPList struct {
	Log       *logrus.Logger
	BatchSize int                  //number of inventory assets to return per iteration
	Channel   chan<- []asset.Asset //the channel to send inventory assets over
	Config    *config.Params       //bmcbutler config
}

//AssetRetrieve looks at d.Config.FilterParams
//and returns the appropriate function that will retrieve assets.
func (i *IPList) AssetRetrieve() func() {
	return i.AssetIter
}

// AssetIter is an iterator method that sends assets to configure
// over the inventory channel.
func (i *IPList) AssetIter() {
	ips := strings.Split(i.Config.FilterParams.Ips, ",")

	assets := make([]asset.Asset, 0)
	for _, ip := range ips {
		assets = append(assets, asset.Asset{IPAddress: ip})
	}

	//pass the asset to the channel
	i.Channel <- assets
	close(i.Channel)

}
