package main

import (
	"context"
	"fmt"
	"os"

	bmclib "github.com/bmc-toolbox/bmclib/v2"
	logrusr "github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	// setup logger
	l := logrus.New()
	l.Level = logrus.TraceLevel
	logger := logrusr.New(l)

	clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}

	host := os.Getenv("BMC_HOST")
	bmcPass := os.Getenv("BMC_PASSWORD")
	bmcUser := os.Getenv("BMC_USERNAME")
	// init client
	client := bmclib.NewClient(host, "", bmcUser, bmcPass, clientOpts...)

	ctx := context.TODO()
	// open BMC session
	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer client.Close(ctx)

	// retrieve bios configuration
	biosConfig, err := client.GetBiosConfiguration(ctx)
	if err != nil {
		l.Error(err)
	}

	fmt.Println(biosConfig)
}
