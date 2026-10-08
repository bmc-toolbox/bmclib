package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// GetPowerMetrics returns the chassis power readings and the configured power cap.
//
// Implements bmc.PowerMetricsGetter.
func (c *Conn) GetPowerMetrics(ctx context.Context) (metrics bmc.PowerMetrics, err error) {
	return c.redfishwrapper.GetPowerMetrics(ctx)
}

// SetPowerCap sets the chassis power cap in watts; a nil limitWatts clears the cap.
//
// Implements bmc.PowerCapSetter.
func (c *Conn) SetPowerCap(ctx context.Context, limitWatts *float64) (err error) {
	return c.redfishwrapper.SetPowerCap(ctx, limitWatts)
}
