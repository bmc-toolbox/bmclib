package main

import (
	"context"
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
	host := "10.211.132.157"
	user := "root"
	pass := "yxvZdxAQ38ZWlZ"

	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	if host == "" || user == "" || pass == "" {
		log.Fatal("required host/user/pass parameters not defined")
	}

	os.Setenv("DEBUG_BMCLIB", "true")
	defer os.Unsetenv("DEBUG_BMCLIB")

	cl := bmclib.NewClient(host, user, pass, bmclib.WithLogger(logger))

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
