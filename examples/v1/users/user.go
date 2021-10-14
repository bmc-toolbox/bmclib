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
	host := ""
	port := ""
	user := ""
	pass := ""

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.NewLogger(l)

	var err error

	cl := bmclib.NewClient(host, port, user, pass, bmclib.WithLogger(logger))

	// we may want to specify multiple protocols here
	cl.Registry.Drivers = cl.Registry.Using("redfish")

	err = cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	_, err = cl.CreateUser(ctx, "foobar", "sekurity101", "Administrator")
	if err != nil {
		l.Error(err)
	}

}
