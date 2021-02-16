package main

/*
 This utilizes the 'v2' bmclib interface methods
*/

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bmc-toolbox/bmclib"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/bmc-toolbox/bmclib/providers/asrockrack"
	"github.com/bombsimon/logrusr"
	"github.com/jacobweinstock/registrar"
	"github.com/sirupsen/logrus"
)

// Here we describe two ways to connect to bmc device over https (as of now)
// and retrieve the BMC version
func main() {
	// invokes the ScanAndConnectv2() method which does the work to register and return
	// a bmclib client based on the BMC vendor/model
	scanAndConnectv2()

	// this method describes the steps ScanAndConnectv2 wraps around
	// to connect to a BMC device, the vendor/model is left to the caller to identify
	// and register the appropriate driver
	setupRegistryAndConnect()
}

func scanAndConnectv2() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Hour)
	defer cancel()
	host := "127.0.0.1"
	user := "foo"
	pass := "bar"

	// Connect and identify BMC, returning a bmc client
	client, err := discover.ScanAndConnectv2(host, user, pass, discover.WithContext(ctx))
	if err != nil {
		log.Fatal(err, "connect failed")
	}

	// open connection
	err = client.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}
	defer client.Close(ctx)

	v, err := client.GetBMCVersion(ctx)
	if err != nil {
		log.Fatal(err, "unable to retrieve BMC version")
	}

	fmt.Println("BMC version: " + v)
}

// This method lists the steps to register and connect to a BMC using a given provider
func setupRegistryAndConnect() {

	host := "127.0.0.1"
	user := "foo"
	pass := "bar"

	log := logging.DefaultLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Hour)
	defer cancel()

	asrockRack, err := asrockrack.New(ctx, host, user, pass, log)
	if err != nil {
		log.Error(err, "failed to init asrockrack instance")
	}

	l := logrus.New()
	l.Level = logrus.InfoLevel
	logg := logrusr.NewLogger(l)

	// define and register driver
	drivers := registrar.NewRegistry()

	// The driver interface being provided could implement methods for ProviderName, ProviderProtocol etc
	// that way we don't have to pass all of these in individually?
	drivers.Register(asrockrack.ProviderName, asrockrack.ProviderProtocol, asrockrack.Features, nil, asrockRack)
	regOptions := bmclib.WithRegistry(drivers)

	// bmclib client - setting the logger as a param doesn't work - its overwritten by the default logger
	cl := bmclib.NewClient(host, "", user, pass, regOptions)
	cl.Registry.Logger = logg

	// open connection
	err = cl.Open(ctx)
	if err != nil {
		log.Error(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	v, err := cl.GetBMCVersion(ctx)
	if err != nil {
		log.Error(err, "unable to retrieve BMC version")
	}

	fmt.Println("BMC version: " + v)

}
