package main

import (
	"os"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
)

// bmc lib takes in its opts a logger (https://github.com/go-logr/logr).
// If you do not define one, by default, it uses logrus (https://github.com/go-logr/logr)
// See the logr docs for more details, but the following implementations already exist:
// github.com/google/glog: glogr
// k8s.io/klog: klogr
// go.uber.org/zap: zapr
// log (the Go standard library logger): stdr
// github.com/sirupsen/logrus: logrusr
// github.com/wojas/genericr: genericr
func main() {

	ip := "<bmc_ip>"
	user := "user"
	pass := "password"

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	//logger.SetFormatter(&logrus.JSONFormatter{})

	logger.Info("printing status with a user defined logger")
	conn, err := withUserDefinedLogger(ip, user, pass, logger)
	if err != nil {
		logger.Fatal(err)
	}
	printStatus(conn, logger)

	logger.Info("printing status with the default builtin logger")
	os.Setenv("BMCLIB_LOG_LEVEL", "debug")
	conn, err = withDefaultBuiltinLogger(ip, user, pass)
	if err != nil {
		logger.Fatal(err)
	}
	printStatus(conn, logger)
}

func withUserDefinedLogger(ip, user, pass string, logger *logrus.Logger) (interface{}, error) {
	myLog := logrusr.NewLogger(logger)
	opts := func(o *discover.Options) {
		o.Logger = myLog
	}

	return discover.ScanAndConnect(ip, user, pass, opts)
}

func withDefaultBuiltinLogger(ip, user, pass string) (interface{}, error) {
	return discover.ScanAndConnect(ip, user, pass)
}

func printStatus(connection interface{}, logger *logrus.Logger) {
	switch connection.(type) {
	case devices.Bmc:
		conn := connection.(devices.Bmc)
		defer conn.Close()

		sr, err := conn.Serial()
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"serial": sr}).Info("serial")

		md, err := conn.Model()
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"model": md}).Info("model")

		mm, err := conn.Memory()
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"memory": mm}).Info("memory")

		st, err := conn.Status()
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"status": st}).Info("status")

		hw := conn.HardwareType()
		logger.WithFields(logrus.Fields{"hwType": hw}).Info("hwType")

		state, err := conn.PowerState()
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"state": state}).Info("state")

	case devices.Cmc:
		cmc := connection.(devices.Cmc)
		sts, err := cmc.Status()
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"status": sts}).Info("status")
	default:
		logger.Fatal("Unknown device")
	}
}
