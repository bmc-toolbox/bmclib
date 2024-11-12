package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bombsimon/logrusr/v2"
	bmclib "github.com/metal-toolbox/bmclib"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	incompatibleRedfishVersions := flag.String("incompatible-redfish-versions", "", "Comma separated list of redfish versions to deem incompatible")
	withSecureTLS := flag.Bool("secure-tls", false, "Enable secure TLS")
	certPoolFile := flag.String("cert-pool", "", "Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true")
	flag.Parse()

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}

	if *withSecureTLS {
		var pool *x509.CertPool
		if *certPoolFile != "" {
			pool = x509.NewCertPool()
			data, err := os.ReadFile(*certPoolFile)
			if err != nil {
				l.Fatal(err)
			}
			pool.AppendCertsFromPEM(data)
		}
		// a nil pool uses the system certs
		clientOpts = append(clientOpts, bmclib.WithSecureTLS(pool))
	}

	if len(*incompatibleRedfishVersions) > 0 {
		// blacklist a redfish version
		clientOpts = append(
			clientOpts,
			bmclib.WithRedfishVersionsNotCompatible(strings.Split(*incompatibleRedfishVersions, ",")))
	}

	cl := bmclib.NewClient(*host, *user, *pass, clientOpts...)

	cl.Registry.Drivers = cl.Registry.FilterForCompatible(ctx)

	err := cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	inv, err := cl.Inventory(ctx)
	if err != nil {
		l.Error(err)
	}

	b, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		l.Error(err)
	}

	fmt.Println(string(b))
}
