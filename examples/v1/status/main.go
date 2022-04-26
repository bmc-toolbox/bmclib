package main

import (
	"context"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/bmc-toolbox/bmclib"
	"github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	port := flag.Int("port", 443, "BMC port to connect to")
	withSecureTLS := flag.Bool("secure-tls", false, "Enable secure TLS")
	certPoolFile := flag.String("cert-pool", "", "Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true")
	flag.Parse()

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	if *host == "" || *user == "" || *pass == "" {
		l.Fatal("required host/user/pass parameters not defined")
	}

	clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}

	if *withSecureTLS {
		var pool *x509.CertPool
		if *certPoolFile != "" {
			pool = x509.NewCertPool()
			data, err := ioutil.ReadFile(*certPoolFile)
			if err != nil {
				l.Fatal(err)
			}
			pool.AppendCertsFromPEM(data)
		}
		// a nil pool uses the system certs
		clientOpts = append(clientOpts, bmclib.WithSecureTLS(pool))
	}

	cl := bmclib.NewClient(*host, strconv.Itoa(*port), *user, *pass, clientOpts...)
	cl.Registry.Drivers = cl.Registry.Using("redfish")
	// cl.Registry.Drivers = cl.Registry.Using("vendorapi")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := cl.Open(ctx)
	if err != nil {
		l.WithError(err).Fatal(err, "BMC login failed")
	}
	defer cl.Close(ctx)

	inventory, err := cl.Inventory(ctx)
	if err != nil {
		l.Fatal(err)
	}

	l.WithField("bmc-version", inventory.BMC.Firmware.Installed).Info()

	state, err := cl.GetPowerState(ctx)
	if err != nil {
		l.WithError(err).Error()
	}
	l.WithField("power-state", state).Info()

	l.WithField("bios-version", inventory.BIOS.Firmware.Installed).Info()
}
