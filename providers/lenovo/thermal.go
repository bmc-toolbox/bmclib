package lenovo

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/stmcginnis/gofish"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.ThermalReader = (*Conn)(nil)

// Thermal returns the temperature and fan readings from the first chassis that
// exposes a Thermal resource.
//
// Implements bmc.ThermalReader.
func (c *Conn) Thermal(ctx context.Context) (bmc.ThermalReading, error) {
	chassis, err := c.redfishwrapper.Chassis(ctx)
	if err != nil {
		return bmc.ThermalReading{}, fmt.Errorf("reading chassis collection: %w", err)
	}

	for _, ch := range chassis {
		thermal, err := ch.Thermal()
		if err != nil || thermal == nil {
			continue
		}

		reading := bmc.ThermalReading{}
		for _, t := range thermal.Temperatures {
			reading.Temperatures = append(reading.Temperatures, bmc.TemperatureReading{
				Name:           t.Name,
				ReadingCelsius: gofish.Deref(t.ReadingCelsius),
				Health:         string(t.Status.Health),
			})
		}
		for _, f := range thermal.Fans {
			name := f.Name
			if name == "" {
				// Fall back to the deprecated FanName for pre-1.1.0 Fan schemas.
				name = f.FanName //nolint:staticcheck
			}
			reading.Fans = append(reading.Fans, bmc.FanReading{
				Name:    name,
				Reading: gofish.Deref(f.Reading),
				Health:  string(f.Status.Health),
			})
		}

		return reading, nil
	}

	return bmc.ThermalReading{}, fmt.Errorf("no chassis exposes a Thermal resource")
}
