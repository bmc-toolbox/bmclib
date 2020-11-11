package generic

import (
	"context"
	"strings"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/pkg/errors"
)

var (
	// Ensure the compiler errors if the interfaces are not implemented properly
	_ devices.PowerBootRequester = (*Generic)(nil)

	powerCommands = map[devices.PowerCommand]powerCommandHandler{
		devices.PowerOn:  powerOn,
		devices.PowerOff: powerOff,
	}
)

type powerCommandHandler func(context.Context, *Generic) (bool, error)

// PowerRequest for power actions against a BMC
func (g *Generic) PowerRequest(ctx context.Context, command devices.PowerCommand) (bool, error) {
	fn, ok := powerCommands[command]
	if ok {
		return fn(ctx, g)
	}

	return false, errors.New("not implemented")
}

func powerOn(ctx context.Context, g *Generic) (bool, error) {
	i, err := ipmi.New(g.Username, g.Password, g.Host)
	if err != nil {
		return false, err
	}
	return i.PowerOn()
}

func powerOff(ctx context.Context, g *Generic) (bool, error) {
	i, err := ipmi.New(g.Username, g.Password, g.Host)
	if err != nil {
		return false, err
	}
	return i.PowerOff()
}

// BootDeviceRequest sets the next boot device and options
func (g *Generic) BootDeviceRequest(ctx context.Context, bo devices.BootOptions) (bool, error) {
	i, err := ipmi.New(g.Username, g.Password, g.Host)
	if err != nil {
		return false, err
	}
	var options []string
	if bo.Persistent {
		options = append(options, "persistent")
	}
	if bo.EfiBoot {
		options = append(options, "efiboot")
	}

	return i.BootDevice(ctx, strings.ToLower(string(bo.Device)), options)
}
