package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/bombsimon/logrusr/v2"
	"github.com/metal-toolbox/bmclib"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// set BMC parameters here
	host := flag.String("host", "", "BMC hostname to connect to")
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("pass", "", "Username to login with")
	flag.Parse()

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	os.Setenv("DEBUG_BMCLIB", "true")
	defer os.Unsetenv("DEBUG_BMCLIB")

	cl := bmclib.NewClient(*host, *user, *pass, bmclib.WithLogger(logger))

	err := cl.Open(ctx)
	if err != nil {
		log.Fatal(err, "bmc login failed")
	}

	defer cl.Close(ctx)

	_, err = cl.ResetBMC(ctx, "GracefulRestart")
	if err != nil {
		l.Error(err)
	}

}
