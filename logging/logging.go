package logging

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	switch os.Getenv("BMCLIB_LOG_LEVEL") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

// SetFormatter allows to format logrus formater
func SetFormatter(formater log.Formatter) {
	log.SetFormatter(formater)
}

// SetOutput allows to set logrus output
func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

// SetLevel allows to set logrus loglevel
func SetLevel(level log.Level) {
	log.SetLevel(level)
}
