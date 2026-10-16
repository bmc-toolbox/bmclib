package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.ChassisThermalGetter = (*Conn)(nil)

// GetChassisThermal returns the chassis temperature sensor and fan readings.
//
// Implements bmc.ChassisThermalGetter.
func (c *Conn) GetChassisThermal(ctx context.Context) (thermal bmc.ChassisThermal, err error) {
	return c.redfishwrapper.GetChassisThermal(ctx)
}
