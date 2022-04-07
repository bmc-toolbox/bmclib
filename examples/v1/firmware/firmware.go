package main

import (
	"context"
	"fmt"
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
	host := ""
	port := ""
	user := "root"
	pass := ""

	l := logrus.New()
	l.Level = logrus.DebugLevel
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

	for _, update := range []string{"/tmp/iDRAC-with-Lifecycle-Controller_Firmware_F87RP_WN64_5.00.00.00_A00.EXE"} {
		fh, err := os.Open(update)
		if err != nil {
			log.Fatal(err)
		}

		taskID, err := cl.FirmwareInstall(ctx, devices.SlugBMC, devices.FirmwareApplyOnReset, true, fh)
		if err != nil {
			l.Error(err)
		}

		state, err := cl.FirmwareInstallStatus(ctx, "", taskID, "5.00.00.00")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("state: %s\n", state)
	}

}
