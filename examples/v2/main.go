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
	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	host := "127.0.0.1"
	port := ""
	user := "foo"
	pass := "bar"

	cl := bmclib.NewClient(host, port, user, pass)
	l := logrus.New()
	l.Level = logrus.TraceLevel
	cl.Logger = logrusr.NewLogger(l)

	err := cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	v, err := cl.GetBMCVersion(ctx)
	if err != nil {
		log.Fatal(err, "unable to retrieve BMC version")
	}

	fmt.Println("BMC version: " + v)

}
