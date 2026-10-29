package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.ChassisFanSpeedsGetter    = (*Conn)(nil)
	_ bmc.ChassisTemperaturesGetter = (*Conn)(nil)
)

// GetChassisFanSpeeds returns the chassis fan readings.
//
// Implements bmc.ChassisFanSpeedsGetter.
func (c *Conn) GetChassisFanSpeeds(ctx context.Context) (fans []bmc.ChassisFan, err error) {
	return c.redfishwrapper.GetChassisFanSpeeds(ctx)
}

// GetChassisTemperatures returns the chassis temperature sensor readings.
//
// Implements bmc.ChassisTemperaturesGetter.
func (c *Conn) GetChassisTemperatures(ctx context.Context) (temperatures []bmc.ChassisTemperature, err error) {
	return c.redfishwrapper.GetChassisTemperatures(ctx)
}
