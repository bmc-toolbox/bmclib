package lenovo

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.PowerReader    = (*Conn)(nil)
	_ bmc.PowerCapSetter = (*Conn)(nil)
)

// ReadPower returns power metrics and supplies from the first chassis that
// exposes a Power resource.
//
// Implements bmc.PowerReader.
func (c *Conn) ReadPower(ctx context.Context) (bmc.PowerInfo, error) {
	chassis, err := c.redfishwrapper.Chassis(ctx)
	if err != nil {
		return bmc.PowerInfo{}, fmt.Errorf("reading chassis collection: %w", err)
	}

	for _, ch := range chassis {
		power, err := ch.Power()
		if err != nil || power == nil {
			continue
		}

		info := bmc.PowerInfo{}
		if len(power.PowerControl) > 0 {
			pc := power.PowerControl[0]
			info.ConsumedWatts = derefFloat32(pc.PowerConsumedWatts)
			info.CapacityWatts = derefFloat32(pc.PowerCapacityWatts)
			info.LimitInWatts = pc.PowerLimit.LimitInWatts
		}
		for _, ps := range power.PowerSupplies {
			info.PowerSupplies = append(info.PowerSupplies, bmc.PowerSupplyInfo{
				Name:          ps.Name,
				Health:        string(ps.Status.Health),
				CapacityWatts: derefFloat32(ps.PowerCapacityWatts),
			})
		}

		return info, nil
	}

	return bmc.PowerInfo{}, fmt.Errorf("no chassis exposes a Power resource")
}

// SetPowerCap sets the chassis power limit by PATCHing
// PowerControl[0].PowerLimit.LimitInWatts. A nil limitWatts disables power
// capping (sets LimitInWatts to null). Implements bmc.PowerCapSetter.
func (c *Conn) SetPowerCap(ctx context.Context, limitWatts *float64) error {
	chassis, err := c.redfishwrapper.Chassis(ctx)
	if err != nil {
		return fmt.Errorf("reading chassis collection: %w", err)
	}

	for _, ch := range chassis {
		power, err := ch.Power()
		if err != nil || power == nil {
			continue
		}

		payload := map[string]any{
			"PowerControl": []map[string]any{
				{"PowerLimit": map[string]any{"LimitInWatts": limitWatts}},
			},
		}

		return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, power.ODataID, payload, nil))
	}

	return fmt.Errorf("no chassis exposes a Power resource")
}
