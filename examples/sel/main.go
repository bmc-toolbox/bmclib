package main

import (
	"context"
	"crypto/x509"
	"flag"
	"io/ioutil"
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
	withSecureTLS := flag.Bool("secure-tls", false, "Enable secure TLS")
	certPoolFile := flag.String("cert-pool", "", "Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true")
	action := flag.String("action", "get", "Action to perform on the System Event Log (clear|get)")
	flag.Parse()

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	if *host == "" || *user == "" || *pass == "" {
		l.Fatal("required host/user/pass parameters not defined")
	}

	clientOpts := []bmclib.Option{
		bmclib.WithLogger(logger),
		bmclib.WithRedfishUseBasicAuth(true),
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
	cl.Registry.Drivers = cl.Registry.Supports(providers.FeatureClearSystemEventLog)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := cl.Open(ctx)
	if err != nil {
		l.WithError(err).Fatal(err, "BMC login failed")
	}
	defer cl.Close(ctx)

	switch *action {
	case "get":
		entries, err := cl.GetSystemEventLog(ctx)
		if err != nil {
			l.WithError(err).Fatal(err, "failed to get System Event Log")
		}
		l.Info("System Event Log entries", "entries", entries)
		return
	case "get-raw":
		eventlog, err := cl.GetSystemEventLogRaw(ctx)
		if err != nil {
			l.WithError(err).Fatal(err, "failed to get System Event Log Raw")
		}
		l.Info("System Event Log", "eventlog", eventlog)
		return
	case "clear":
		err = cl.ClearSystemEventLog(ctx)
		if err != nil {
			l.WithError(err).Fatal(err, "failed to clear System Event Log")
		}
		l.Info("System Event Log cleared")
		return
	default:
		l.Fatal("invalid action")
	}
}
