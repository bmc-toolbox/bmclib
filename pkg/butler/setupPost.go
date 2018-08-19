package butler

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"/jrebello/go-serverdb/serverdb"
)

//Actions to be taken once a chassis was setup successfully.
func (s *SetupAction) Post(asset *asset.Asset) {
	log := s.Log
	component := "Post setup"

	extra := asset.Extra
	serverId, _ := strconv.Atoi(extra["serverDbId"])

	apiUser := viper.GetString("inventory.setup.serverdb.apiUser")
	apiKey := viper.GetString("inventory.setup.serverdb.apiKey")

	sdb := serverdb.New(apiUser, apiKey)
	resp, err := sdb.UpdateChassisStatus(serverId, "installed")
	if err != nil {
		log.WithFields(logrus.Fields{
			"component":         component,
			"butler-id":         s.Id,
			"Asset":             fmt.Sprintf("%+v", asset),
			"Error":             err,
			"ServerDb response": resp,
		}).Warn("Unable to update chassis status to installed in ServerDb.")
	}

	log.WithFields(logrus.Fields{
		"component":         component,
		"butler-id":         s.Id,
		"Asset":             fmt.Sprintf("%+v", s.Asset),
		"ServerDb response": resp,
	}).Info("Updated asset state to installed in serverDb")

	return
}
