package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bmc-toolbox/bmclib"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// set BMC parameters here
	host := "10.247.150.161"
	port := ""
	user := "admin"
	pass := "RmrJ56BFUarn6g"

	l := logrus.New()
	l.Level = logrus.TraceLevel
	logger := logrusr.New(l)

	if host == "" || user == "" || pass == "" {
		log.Fatal("required host/user/pass parameters not defined")
	}

	cl := bmclib.NewClient(host, port, user, pass, bmclib.WithLogger(logger))

	err := cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	for _, update := range []string{"/tmp/E6D4INL2.09C.ima"} {
		fh, err := os.Open(update)
		if err != nil {
			log.Fatal(err)
		}

		_, err = cl.FirmwareInstall(ctx, devices.SlugBMC, devices.FirmwareApplyOnReset, true, fh)
		if err != nil {
			l.Error(err)
		}

	}

}
