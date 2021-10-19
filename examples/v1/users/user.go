package main

import (
	"context"
	"log"
	"time"

	"github.com/bmc-toolbox/bmclib"
	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// set BMC parameters here
	host := ""
	port := ""
	user := ""
	pass := ""

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.NewLogger(l)

	if host == "" || user == "" || pass == "" {
		log.Fatal("required host/user/pass parameters not defined")
	}

	cl := bmclib.NewClient(host, port, user, pass, bmclib.WithLogger(logger))

	cl.Registry.Drivers = cl.Registry.Using("redfish")
	// cl.Registry.Drivers = cl.Registry.Using("vendorapi")

	err := cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	_, err = cl.CreateUser(ctx, "foobar", "sekurity101", "Administrator")
	if err != nil {
		l.Error(err)
	}

}
