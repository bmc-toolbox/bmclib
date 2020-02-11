package discover

import (
	"os"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers/dummy/ibmc"

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

// ScanAndConnect will scan the bmc trying to learn the device type and return a working connection
func ScanAndConnect(host string, username string, password string) (bmcConnection interface{}, err error) {
	log.WithFields(log.Fields{"step": "ScanAndConnect", "host": host}).Debug("detecting vendor")

	// return a connection to our dummy device.
	if os.Getenv("BMCLIB_TEST") == "1" {
		log.WithFields(log.Fields{"step": "ScanAndConnect", "host": host}).Debug("returning connection to dummy ibmc device.")
		bmc, err := ibmc.New(host, username, password)
		return bmc, err
	}

	client, err := httpclient.Build()
	if err != nil {
		return bmcConnection, err
	}

	var probe = Probe{client: client, username: username, password: password, host: host}
	var devices = []func() (interface{}, error){
		probe.hpIlo,
		probe.idrac8,
		probe.idrac9,
		probe.hpC7000, //doesn't work
		probe.m1000e,
		probe.supermicrox,
		probe.hpCl100,
	}

	for _, probeDevice := range devices {

		log.WithFields(log.Fields{"step": "ScanAndConnect", "host": host}).Debug("probing to identify device")

		bmcConnection, err := probeDevice()

		// if the device didn't match continue to probe
		if err != nil && (err == errors.ErrDeviceNotMatched) {
			continue
		}

		// if theres a bmc connection error, return the error
		if err != nil {
			return bmcConnection, err
		}

		// return a bmcConnection
		return bmcConnection, err

	}

	return bmcConnection, errors.ErrVendorUnknown
}
