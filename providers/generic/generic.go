package generic

import (
	"context"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/go-logr/logr"
)

const (
	// BmcType defines the bmc model that is supported by this package
	BmcType = "generic"
)

// Generic for connecting to a BMC in a generic way
type Generic struct {
	Host     string
	Username string
	Password string
	Log      logr.Logger
}

// New returns a new Generic instance to be used
func New(log logr.Logger, host string, username string, password string) (*Generic, error) {
	return &Generic{
		Host:     host,
		Username: username,
		Password: password,
		Log:      log}, nil
}

var _ devices.Connection = (*Generic)(nil)

// Open connection to BMC
func (g *Generic) Open(ctx context.Context) error {
	i, err := ipmi.New(g.Username, g.Password, g.Host)
	if err != nil {
		return err
	}
	return i.Info(ctx)
}

// Close a connection to BMC
func (g *Generic) Close(ctx context.Context) error {

	return nil
}
