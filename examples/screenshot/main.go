package main

import (
	"context"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/bombsimon/logrusr/v2"
	"github.com/metal-toolbox/bmclib"
	"github.com/metal-toolbox/bmclib/providers"
	"github.com/sirupsen/logrus"
)

func main() {
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	port := flag.String("port", "443", "BMC port to connect to")
	withSecureTLS := flag.Bool("secure-tls", false, "Enable secure TLS")
	certPoolFile := flag.String("cert-pool", "", "Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true")
	flag.Parse()

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	if *host == "" || *user == "" || *pass == "" {
		l.Fatal("required host/user/pass parameters not defined")
	}

	clientOpts := []bmclib.Option{
		bmclib.WithLogger(logger),
		bmclib.WithRedfishPort(*port),
	}

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

	cl := bmclib.NewClient(*host, *user, *pass, clientOpts...)
	cl.Registry.Drivers = cl.Registry.Supports(providers.FeatureScreenshot)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := cl.Open(ctx)
	if err != nil {
		l.WithError(err).Fatal(err, "BMC login failed")

		return
	}
	defer cl.Close(ctx)

	image, fileType, err := cl.Screenshot(ctx)
	if err != nil {
		l.WithError(err).Error()

		return
	}

	filename := fmt.Sprintf("screenshot." + fileType)
	fh, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		l.WithError(err).Error()

		return
	}

	defer fh.Close()

	_, err = fh.Write(image)
	if err != nil {
		l.WithError(err).Error()

		return
	}

	l.Info("screenshot saved as: " + filename)
}
