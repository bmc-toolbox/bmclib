package main

import (
	"context"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	bmclib "github.com/bmc-toolbox/bmclib/v2"
	"github.com/bmc-toolbox/bmclib/v2/constants"
	"github.com/bmc-toolbox/common"
	"github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	port := flag.Int("port", 443, "BMC port to connect to")
	withSecureTLS := flag.Bool("secure-tls", false, "Enable secure TLS")
	certPoolPath := flag.String("cert-pool", "", "Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true")
	firmwarePath := flag.String("firmware", "", "The local path of the firmware to install")
	firmwareVersion := flag.String("version", "", "The firmware version being installed")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	if *host == "" || *user == "" || *pass == "" {
		l.Fatal("required host/user/pass parameters not defined")
	}
	clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}

	if *withSecureTLS {
		var pool *x509.CertPool
		if *certPoolPath != "" {
			pool = x509.NewCertPool()
			data, err := ioutil.ReadFile(*certPoolPath)
			if err != nil {
				l.Fatal(err)
			}
			pool.AppendCertsFromPEM(data)
		}
		// a nil pool uses the system certs
		clientOpts = append(clientOpts, bmclib.WithSecureTLS(pool))
	}

	cl := bmclib.NewClient(*host, strconv.Itoa(*port), *user, *pass, clientOpts...)
	err := cl.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	// collect inventory
	inventory, err := cl.Inventory(ctx)
	if err != nil {
		l.Fatal(err)
	}

	l.WithField("bmc-version", inventory.BMC.Firmware.Installed).Info()

	// open file handle
	fh, err := os.Open(*firmwarePath)
	if err != nil {
		l.Fatal(err)
	}
	defer fh.Close()

	// SlugBMC hardcoded here, this can be any of the existing component slugs from devices/constants.go
	// assuming that the BMC provider implements the required component firmware update support
	taskID, err := cl.FirmwareInstall(ctx, common.SlugBMC, constants.FirmwareApplyOnReset, true, fh)
	if err != nil {
		l.Error(err)
	}

	state, err := cl.FirmwareInstallStatus(ctx, taskID, common.SlugBMC, *firmwareVersion)
	if err != nil {
		log.Fatal(err)
	}

	l.WithField("state", state).Info("BMC firmware install state")
}
