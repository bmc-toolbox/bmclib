package configure

import (
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/sirupsen/logrus"
)

// Configure struct declares attributes required to apply configuration.
type Configure struct {
	client devices.Configure
	logger *logrus.Logger
}

// New returns a new configure struct to apply configuration.
func New(client devices.Configure, logger *logrus.Logger) *Configure {
	return &Configure{
		client: client,
		logger: logger,
	}
}

// Apply applies configuration.
func (c *Configure) Apply() {
	c.logger.WithFields(logrus.Fields{}).Info("Config applied.")
}
