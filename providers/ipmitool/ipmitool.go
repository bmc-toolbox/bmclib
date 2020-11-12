package ipmitool

import (
	"github.com/go-logr/logr"
)

const DriverName = "ipmitool"

// Conn for Ipmitool connection details
type Conn struct {
	Host string
	Port string
	User string
	Pass string
	Log  logr.Logger
}
