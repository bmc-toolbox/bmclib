package dummy

import (
	"context"
	"errors"

	"github.com/bmc-toolbox/bmclib/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

const (
	// VendorID represents the id of the vendor across all packages
	VendorID = "dummy"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "dummy"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "ipmi"
)

var (
	// Features implemented by ipmitool
	Features = registrar.Features{
		providers.FeaturePowerState,
	}
)

type Conn struct {
	FailOpen bool
	Log      logr.Logger
}

func (c *Conn) Name() string {
	return ProviderName
}

func (c *Conn) Open(ctx context.Context) error {
	if c.FailOpen {
		return errors.New("failed on purpose")
	}
	return nil
}

// Close a connection to a BMC
func (c *Conn) Close(ctx context.Context) (err error) {
	return nil
}

// Compatible tests whether a BMC is compatible with the ipmitool provider
func (c *Conn) Compatible(ctx context.Context) bool {
	return true
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	//panic("testing")
	return "on", errors.New("bad")
}
