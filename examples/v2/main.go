package main

/*
 This utilizes the 'v2' bmclib interface methods to flash a firmware image
*/

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/bmc-toolbox/bmclib"
	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	host := ""
	port := ""
	user := ""
	pass := ""

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.NewLogger(l)

	var err error

	cl := bmclib.NewClient(host, port, user, pass, bmclib.WithLogger(logger))
	err = cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	v, err := cl.GetBMCVersion(ctx)
	if err != nil {
		log.Fatal(err, "unable to retrieve BMC version")
	}

	fmt.Println("BMC version: " + v)

	// open file handle
	fh, err := os.Open("/tmp/E3C246D4I-NL_L0.03.00.ima")
	if err != nil {
		log.Fatal(err)
	}
	defer fh.Close()

	err = cl.UpdateBMCFirmware(ctx, fh)
	if err != nil {
		log.Fatal(err)
	}

}
