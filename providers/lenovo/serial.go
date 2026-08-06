package lenovo

import (
	"context"
	"net/url"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.SerialInterfaceGetter = (*Conn)(nil)
	_ bmc.SerialInterfaceSetter = (*Conn)(nil)
)

// SerialInterfaces returns the BMC serial interfaces.
//
// Implements bmc.SerialInterfaceGetter.
func (c *Conn) SerialInterfaces(ctx context.Context) ([]bmc.SerialInterface, error) {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return nil, err
	}

	ifaces, err := manager.SerialInterfaces()
	if err != nil {
		return nil, err
	}

	out := make([]bmc.SerialInterface, 0, len(ifaces))
	for _, s := range ifaces {
		out = append(out, bmc.SerialInterface{
			ID:          s.ID,
			Enabled:     s.InterfaceEnabled,
			BitRate:     string(s.BitRate),
			FlowControl: string(s.FlowControl),
			Parity:      string(s.Parity),
			StopBits:    string(s.StopBits),
			DataBits:    string(s.DataBits),
		})
	}

	return out, nil
}

// SetSerialInterface PATCHes a BMC serial interface.
//
// Implements bmc.SerialInterfaceSetter.
func (c *Conn) SetSerialInterface(ctx context.Context, id string, attrs map[string]any) error {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return err
	}

	target, err := url.JoinPath(manager.ODataID, "SerialInterfaces/"+id)
	if err != nil {
		return err
	}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, target, attrs, nil))
}
