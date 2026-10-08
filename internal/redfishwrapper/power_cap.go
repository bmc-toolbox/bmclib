package redfishwrapper

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// powerControl returns the Power resource of the first chassis that exposes one
// with a non-empty PowerControl collection. Chassis without a Power resource
// (enclosures, sleds without their own power domain) are skipped.
func (c *Client) powerControl(ctx context.Context) (*schemas.Power, error) {
	chassis, err := c.Chassis(ctx)
	if err != nil {
		return nil, err
	}

	for _, ch := range chassis {
		if ch == nil {
			continue
		}

		power, err := ch.Power()
		if err != nil || power == nil || len(power.PowerControl) == 0 {
			continue
		}

		return power, nil
	}

	return nil, bmclibErrs.ErrPowerControlNotFound
}

// GetPowerMetrics returns the chassis power readings and the configured power cap
// from PowerControl[0] of the first chassis exposing a Power resource.
func (c *Client) GetPowerMetrics(ctx context.Context) (metrics bmc.PowerMetrics, err error) {
	power, err := c.powerControl(ctx)
	if err != nil {
		return metrics, err
	}

	pc := power.PowerControl[0]
	if pc.PowerConsumedWatts != nil {
		metrics.ConsumedWatts = float64(*pc.PowerConsumedWatts)
	}
	if pc.PowerCapacityWatts != nil {
		metrics.CapacityWatts = float64(*pc.PowerCapacityWatts)
	}
	if pc.PowerLimit.LimitInWatts != nil {
		limit := *pc.PowerLimit.LimitInWatts
		metrics.LimitInWatts = &limit
	}

	return metrics, nil
}

// SetPowerCap sets the chassis power cap by PATCHing
// PowerControl[0].PowerLimit.LimitInWatts on the Power resource. A nil limitWatts
// clears the cap (LimitInWatts: null). gofish exposes no Update() on the Power
// resource, hence the explicit PATCH.
func (c *Client) SetPowerCap(ctx context.Context, limitWatts *float64) (err error) {
	power, err := c.powerControl(ctx)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"PowerControl": []map[string]any{
			{"PowerLimit": map[string]any{"LimitInWatts": limitWatts}},
		},
	}

	// gofish returns an error for any non-2xx response, carrying the BMC's
	// error body; wrap it so callers can match on ErrPowerCapSet.
	resp, err := c.PatchWithHeaders(ctx, power.ODataID, payload, nil)
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrPowerCapSet, err.Error())
	}
	_ = resp.Body.Close()

	return nil
}
