package main

// This snippet utilizes the older bmclib interface methods
// it connects to the bmc and retries its version

import (
	"context"
	"flag"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	flag.Parse()
	ctx := context.TODO()

	l := logrus.New()
	l.Level = logrus.TraceLevel
	logger := logrusr.New(l)
	if *host == "" || *user == "" || *pass == "" {
		l.Fatal("required host/user/pass parameters not defined")
	}

	c, err := discover.ScanAndConnect(
		*host,
		*user,
		*pass,
		discover.WithContext(ctx),
		discover.WithLogger(logger),
	)

	if err != nil {
		l.WithError(err).Fatal("Error connecting to bmc")
	}

	bmc := c.(devices.Bmc)

	err = bmc.CheckCredentials()
	if err != nil {
		l.WithError(err).Fatal("Failed to validate credentials")
	}

	defer bmc.Close(ctx)

	s, err := bmc.Serial()
	if err != nil {
		l.WithError(err).Fatal("Error getting bmc serial")
	}
	l.WithField("serial", s).Info()

}
