package generic

import (
	"context"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/pkg/errors"
)

var (
	// Ensure the compiler errors if the interfaces are not implemented properly
	_ devices.DataRequester = (*Generic)(nil)

	dataRequests = map[devices.DataType]dataRequestHandler{
		devices.SystemState: systemState,
	}
)

type dataRequestHandler func(context.Context, *Generic) (*devices.DataPoint, error)

// DataRequest for retrieving info data about a BMC
func (g *Generic) DataRequest(ctx context.Context, dType devices.DataType) (*devices.DataPoint, error) {
	fn, ok := dataRequests[dType]
	if ok {
		return fn(ctx, g)
	}

	return nil, errors.New("not implemented")
}

func systemState(ctx context.Context, g *Generic) (*devices.DataPoint, error) {
	i, err := ipmi.New(g.Username, g.Password, g.Host)
	if err != nil {
		return nil, err
	}
	result, err := i.IsOn()
	if err != nil {
		return nil, err
	}
	var v string
	if result {
		v = "on"
	} else {
		v = "off"
	}

	state := &devices.DataPoint{
		Name:  "state",
		Value: v,
		Type:  devices.SystemState,
	}
	return state, nil
}
