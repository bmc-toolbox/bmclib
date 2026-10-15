package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// GetChassisPowerMetrics returns the chassis power readings and the configured power cap.
//
// Implements bmc.ChassisPowerMetricsGetter.
func (c *Conn) GetChassisPowerMetrics(ctx context.Context) (metrics bmc.ChassisPowerMetrics, err error) {
	return c.redfishwrapper.GetChassisPowerMetrics(ctx)
}

// SetChassisPowerCap sets the chassis power cap in watts; a nil limitWatts clears the cap.
//
// Implements bmc.ChassisPowerCapSetter.
func (c *Conn) SetChassisPowerCap(ctx context.Context, limitWatts *float64) (err error) {
	return c.redfishwrapper.SetChassisPowerCap(ctx, limitWatts)
}
