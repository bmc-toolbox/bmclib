package main

/*
 This utilizes the 'v2' bmclib interface methods
*/

import (
	"context"
	"fmt"
	"time"

	"github.com/bmc-toolbox/bmclib"
	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/bmc-toolbox/bmclib/providers/asrockrack"
	"github.com/bombsimon/logrusr"
	"github.com/jacobweinstock/registrar"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Hour)
	defer cancel()
	host := ""
	port := ""
	user := ""
	pass := ""

	log := logging.DefaultLogger()

	asrockRack, err := asrockrack.New(ctx, host, user, pass, log)
	if err != nil {
		log.Error(err, "failed to init asrockrack instance")
	}

	// how do I get logr to log at trace/debug level
	l := logrus.New()
	l.Level = logrus.InfoLevel
	logg := logrusr.NewLogger(l)

	// define and register driver
	drivers := registrar.NewRegistry()

	// The driver interface being provided could implement methods for ProviderName, ProviderProtocol etc
	// that way we don't have to pass all of these in individually?

	regOptions := bmclib.WithRegistry(drivers)

	// bmclib client - setting the logger as a param doesn't work - its overwritten by the default logger
	cl := bmclib.NewClient(host, port, user, pass, regOptions)
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
	//	err = cl.UpdateBMCFirmware(ctx, "/tmp/E3C246D4I-NL_L0.03.00.ima")
	//	if err != nil {
	//		log.Error(err, "error updating BMC firmware")
	//	}

	//	v, err = cl.GetBIOSVersion(ctx)
	//	if err != nil {
	//		log.Error(err, "unable to retrieve BIOS version")
	//	}
	//
	//	fmt.Println("BIOS version: " + v)
	//
	//	err = cl.UpdateBIOSFirmware(ctx, "/tmp/E6D4INL2.07B")
	//	if err != nil {
	//		log.Error(err, "error updating BIOS firmware")
	//	}
}
