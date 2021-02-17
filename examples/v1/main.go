package main

// This snippet utilizes the 'v1' older bmclib interface methods
// it connects to the bmc and retries its version

import (
	"context"
	"fmt"
	"os"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
)

func main() {
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	ctx := context.TODO()
	//defer cancel()
	host := ""
	user := ""
	pass := ""

	l := logrus.New()
	l.Level = logrus.TraceLevel
	logger := logrusr.NewLogger(l)

	c, err := discover.ScanAndConnect(
		host,
		user,
		pass,
		discover.WithContext(ctx),
		discover.WithLogger(logger),
	)

	if err != nil {
		logger.Error(err, "Error connecting to bmc")
	}

	bmc := c.(devices.Bmc)

	err = bmc.CheckCredentials()
	if err != nil {
		logger.Error(err, "Failed to validate credentials")
		os.Exit(1)
	}

	defer bmc.Close(ctx)

	s, err := bmc.Serial()
	if err != nil {
		logger.Error(err, "Error getting bmc serial")
		os.Exit(1)
	}
	fmt.Println(s)

}
