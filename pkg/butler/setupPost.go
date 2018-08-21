package butler

import (
	"fmt"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/sirupsen/logrus"
)

//Actions to be taken once a chassis was setup successfully.
func (s *SetupAction) Post(asset *asset.Asset) {
	log := s.Log
	component := "Post setup"

	log.WithFields(logrus.Fields{
		"component": component,
		"butler-id": s.Id,
		"Asset":     fmt.Sprintf("%+v", s.Asset),
	}).Info("A sample post setup action is run.")

	return
}
