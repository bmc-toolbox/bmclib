package generic

import (
	"context"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/pkg/errors"
)

var (
	// Ensure the compiler errors if the interfaces are not implemented properly
	_ devices.Configurer = (*Generic)(nil)

	configureRequests = map[devices.Setting]configureHandler{
		devices.User: userCreate,
	}
)

type configureHandler func(context.Context, *Generic, devices.Config) (*devices.Configuration, error)

// Configure a BMC
func (g *Generic) Configure(ctx context.Context, config devices.Config) (*devices.Configuration, error) {
	fn, ok := configureRequests[config.Type]
	if ok {
		return fn(ctx, g, config)
	}

	return nil, errors.New("not implemented")
}

func userCreate(ctx context.Context, g *Generic, c devices.Config) (*devices.Configuration, error) {
	return nil, errors.New("not implemented")
}
