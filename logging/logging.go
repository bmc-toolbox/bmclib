package logging

import (
	"os"

	"github.com/bombsimon/logrusr/v2"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
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

	return logrusr.New(logrusLog)
}

// ZeroLogger is a logr.Logger implementation that uses zerolog.
// This logger handles nested structs better than the logrus implementation.
func ZeroLogger(level string) logr.Logger {
	zl := zerolog.New(os.Stdout)
	zl = zl.With().Caller().Timestamp().Logger()
	var l zerolog.Level
	switch level {
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}
	zl = zl.Level(l)

	return zerologr.New(&zl)
}
