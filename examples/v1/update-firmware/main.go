package main

/*
 This utilizes what is to tbe the 'v1' bmclib interface methods to flash a firmware image
*/

import (
	"context"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"os"
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
	certPoolPath := flag.String("cert-pool", "", "Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true")
	firmwarePath := flag.String("firmware", "", "The firmware path to read")
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

	v, err := cl.GetBMCVersion(ctx)
	if err != nil {
		l.Fatal(err, "unable to retrieve BMC version")
	}
	logger.Info("BMC version", v)

	// open file handle
	fh, err := os.Open(*firmwarePath)
	if err != nil {
		l.Fatal(err)
	}
	defer fh.Close()

	fi, err := fh.Stat()
	if err != nil {
		l.Fatal(err)
	}

	err = cl.UpdateBMCFirmware(ctx, fh, fi.Size())
	if err != nil {
		l.Fatal(err)
	}
	logger.WithValues("host", *host).Info("Updated BMC firmware")
}
