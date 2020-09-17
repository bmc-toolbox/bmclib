package logging

import (
	"os"

	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

// DefaultLogger if no client logger is defined
func DefaultLogger() logr.Logger {
	logrusLog := logrus.New()
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	logrusLog.SetOutput(os.Stdout)

	switch os.Getenv("BMCLIB_LOG_LEVEL") {
	case "debug":
		logrusLog.SetLevel(logrus.DebugLevel)
	case "trace":
		logrusLog.SetLevel(logrus.TraceLevel)
		logrusLog.SetReportCaller(true)
	default:
		logrusLog.SetLevel(logrus.InfoLevel)
	}

	return logrusr.NewLogger(logrusLog)
}
