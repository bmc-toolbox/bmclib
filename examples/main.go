package main

import (
	"context"
	"os"
	"time"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/go-logr/zapr"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
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
	//logger.SetFormatter(&logrus.JSONFormatter{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info("printing status with the default builtin logger")
	os.Setenv("BMCLIB_LOG_LEVEL", "info")
	conn, err := withDefaultBuiltinLogger(ctx, ip, user, pass)
	if err != nil {
		logger.Fatal(err)
	}
	printStatus(ctx, conn, logger)

	logger.Info("printing status with a user defined logger")
	conn, err = withUserDefinedLogger(ctx, ip, user, pass, logger)
	if err != nil {
		logger.Fatal(err)
	}
	printStatus(ctx, conn, logger)

}

func withUserDefinedLogger(ctx context.Context, ip, user, pass string, logger *logrus.Logger) (interface{}, error) {
	z, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	log := zapr.NewLogger(z)
	return discover.ScanAndConnect(ip, user, pass, discover.WithLogger(log), discover.WithContext(ctx))
}

func withDefaultBuiltinLogger(ctx context.Context, ip, user, pass string) (interface{}, error) {
	return discover.ScanAndConnect(ip, user, pass, discover.WithContext(ctx))
}

func printStatus(ctx context.Context, connection interface{}, logger *logrus.Logger) {
	switch con := connection.(type) {
	case devices.Bmc:
		conn := con
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
		cmc := con
		defer cmc.Close()
		sts, err := cmc.Status()
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"status": sts}).Info("status")
	case devices.BmcWorker:
		conn := con
		state, err := conn.DataRequest(ctx, devices.SystemState)
		if err != nil {
			logger.Fatal(err)
		}
		logger.WithFields(logrus.Fields{"state": state.Value}).Info("state")
	default:
		logger.Fatal("Unknown device")
	}
}
