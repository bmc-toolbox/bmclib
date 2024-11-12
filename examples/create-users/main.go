package main

import (
	"context"
	"crypto/x509"
	"encoding/csv"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/bombsimon/logrusr/v2"
	bmclib "github.com/metal-toolbox/bmclib"
	"github.com/sirupsen/logrus"
)

func main() {
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	withSecureTLS := flag.Bool("secure-tls", false, "Enable secure TLS")
	certPoolFile := flag.String("cert-pool", "", "Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true")
	userCSV := flag.String("user-csv", "", "A CSV file of users to create containing 3 columns: username, password, role")
	dryRun := flag.Bool("dry-run", false, "Connect to the BMC but do not create users")
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

	cl := bmclib.NewClient(*host, *user, *pass, clientOpts...)
	cl.Registry.Drivers = cl.Registry.Using("redfish")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := cl.Open(ctx)
	if err != nil {
		l.WithError(err).Fatal(err, "BMC login failed")
	}
	defer cl.Close(ctx)

	fh, err := os.Open(*userCSV)
	if err != nil {
		l.WithError(err).WithField("file", *userCSV).Fatal()
	}
	defer fh.Close()
	reader := csv.NewReader(fh)
	i := 0
	for {
		record, err := reader.Read()
		i++
		if err == io.EOF {
			break
		}
		if err != nil {
			l.WithError(err).Fatal()
		}
		if len(record) != 3 {
			l.WithField("line", i).WithField("length", len(record)).Infof("line did not have 3 columns")
			continue
		}
		if !*dryRun {
			_, err = cl.CreateUser(ctx, record[0], record[1], record[2])
			if err != nil {
				l.WithError(err).Error("error creating user")
				continue
			}
		}
		l.WithFields(logrus.Fields(map[string]interface{}{
			"user": record[0],
			"role": record[2],
		})).Info("created user")
	}

	l.WithField("count", i).Info("created users")

}
