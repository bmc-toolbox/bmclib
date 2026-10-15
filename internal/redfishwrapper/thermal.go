package redfishwrapper

import (
	"context"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// thermal returns the Thermal resource of the first chassis that exposes one with
// at least one temperature sensor or fan. Chassis without a Thermal resource are skipped.
func (c *Client) thermal(ctx context.Context) (*schemas.Thermal, error) {
	chassis, err := c.Chassis(ctx)
	if err != nil {
		return nil, err
	}

	for _, ch := range chassis {
		if ch == nil {
			continue
		}

		thermal, err := ch.Thermal()
		if err != nil || thermal == nil || (len(thermal.Temperatures) == 0 && len(thermal.Fans) == 0) {
			continue
		}

		return thermal, nil
	}

	return nil, bmclibErrs.ErrThermalNotFound
}

// GetChassisThermal returns the temperature sensor and fan readings of the first
// chassis exposing a Thermal resource.
func (c *Client) GetChassisThermal(ctx context.Context) (thermal bmc.ChassisThermal, err error) {
	t, err := c.thermal(ctx)
	if err != nil {
		return thermal, err
	}

	for i := range t.Temperatures {
		sensor := &t.Temperatures[i]
		thermal.Temperatures = append(thermal.Temperatures, bmc.ChassisTemperature{
			Name:            sensor.Name,
			PhysicalContext: string(sensor.PhysicalContext),
			ReadingCelsius:  sensor.ReadingCelsius,
			Health:          string(sensor.Status.Health),
			State:           string(sensor.Status.State),
		})
	}

	for i := range t.Fans {
		fan := &t.Fans[i]
		name := fan.Name
		if name == "" {
			// Fan schemas before v1.1.0 only carry the deprecated FanName.
			name = fan.FanName //nolint:staticcheck // still the only fan name on older firmware
		}
		thermal.Fans = append(thermal.Fans, bmc.ChassisFan{
			Name:         name,
			Reading:      fan.Reading,
			ReadingUnits: string(fan.ReadingUnits),
			Health:       string(fan.Status.Health),
			State:        string(fan.Status.State),
		})
	}

	return thermal, nil
}
