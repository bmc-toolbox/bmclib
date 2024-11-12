package main

import (
	"context"
	"crypto/x509"
	"flag"
	"log"
	"os"
	"time"

	"github.com/bombsimon/logrusr/v2"
	"github.com/metal-toolbox/bmclib"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	imagePath := flag.String("image", "", "The .img file to be uploaded")
	unmountImage := flag.Bool("unmount", false, "Unmount floppy image")

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

	cl := bmclib.NewClient(*host, *user, *pass, clientOpts...)
	err := cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	if *unmountImage {
		if err := cl.UnmountFloppyImage(ctx); err != nil {
			log.Fatal(err)
		}

		return
	}

	// open file handle
	fh, err := os.Open(*imagePath)
	if err != nil {
		l.Fatal(err)
	}
	defer fh.Close()

	err = cl.MountFloppyImage(ctx, fh)
	if err != nil {
		l.Fatal(err)
	}

	l.WithField("img", *imagePath).Info("image mounted successfully")
}
