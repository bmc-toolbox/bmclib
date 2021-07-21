package main

import (
	"context"
	"fmt"
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
	cl.Registry.Drivers = cl.Registry.Using("redfish")
	err = cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	p, err := cl.GetPowerSensors(ctx)
	if err != nil {
		fmt.Println(err, "unable to retrieve power sensor data")
	}

	fmt.Printf("%+v\n", p)

	t, err := cl.GetTemperatureSensors(ctx)
	if err != nil {
		fmt.Println(err, "unable to retrieve temperature sensor data")
	}

	fmt.Printf("%+v\n", t)

	f, err := cl.GetFanSensors(ctx)
	if err != nil {
		fmt.Println(err, "unable to retrieve fan sensor data")
	}

	fmt.Printf("%+v\n", f)

	c, err := cl.GetChassisHealth(ctx)
	if err != nil {
		fmt.Println(err, "unable to retrieve chassis health data")
	}

	fmt.Printf("%+v\n", c)

}
