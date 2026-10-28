package redfishwrapper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// GetChassisFanSpeeds returns the fan readings of the first chassis exposing fans.
//
// Fans are read from the chassis ThermalSubsystem Fans collection; chassis whose
// firmware still serves only the deprecated Thermal resource are read from
// Thermal.Fans. Errors raised while collecting are returned to the caller.
func (c *Client) GetChassisFanSpeeds(ctx context.Context) ([]bmc.ChassisFan, error) {
	chassis, err := c.Chassis(ctx)
	if err != nil {
		return nil, err
	}

	for _, ch := range chassis {
		if ch == nil {
			continue
		}

		subsystem, err := thermalSubsystem(ch)
		if err != nil {
			return nil, err
		}

		if subsystem != nil {
			fans, found, err := c.thermalSubsystemFans(ch, subsystem)
			if err != nil {
				return nil, err
			}
			if found {
				return fans, nil
			}
		}

		thermal, err := legacyThermal(ch)
		if err != nil {
			return nil, err
		}
		if thermal != nil {
			return legacyFans(thermal), nil
		}
	}

	return nil, bmclibErrs.ErrChassisFansNotFound
}

// GetChassisTemperatures returns the temperature readings of the first chassis
// exposing temperature sensors.
//
// Temperatures are read from the chassis ThermalSubsystem ThermalMetrics; chassis
// whose firmware still serves only the deprecated Thermal resource are read from
// Thermal.Temperatures. Errors raised while collecting are returned to the caller.
func (c *Client) GetChassisTemperatures(ctx context.Context) ([]bmc.ChassisTemperature, error) {
	chassis, err := c.Chassis(ctx)
	if err != nil {
		return nil, err
	}

	for _, ch := range chassis {
		if ch == nil {
			continue
		}

		subsystem, err := thermalSubsystem(ch)
		if err != nil {
			return nil, err
		}

		if subsystem != nil {
			metrics, err := subsystem.ThermalMetrics()
			if err != nil {
				return nil, fmt.Errorf("chassis %s ThermalMetrics: %w", ch.ID, err)
			}
			if metrics != nil {
				return metricsTemperatures(metrics), nil
			}
		}

		thermal, err := legacyThermal(ch)
		if err != nil {
			return nil, err
		}
		if thermal != nil {
			return legacyTemperatures(thermal), nil
		}
	}

	return nil, bmclibErrs.ErrChassisTemperaturesNotFound
}

// thermalSubsystem returns the chassis ThermalSubsystem, or nil when the chassis
// has no such link or the BMC answers 404 for it (firmware advertising the link
// without implementing the resource). Any other failure is returned.
func thermalSubsystem(ch *schemas.Chassis) (*schemas.ThermalSubsystem, error) {
	subsystem, err := ch.ThermalSubsystem()
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("chassis %s ThermalSubsystem: %w", ch.ID, err)
	}

	return subsystem, nil
}

// legacyThermal returns the deprecated Thermal resource of the chassis, or nil when
// the chassis has no such link. Any failure is returned.
func legacyThermal(ch *schemas.Chassis) (*schemas.Thermal, error) {
	thermal, err := ch.Thermal()
	if err != nil {
		return nil, fmt.Errorf("chassis %s Thermal: %w", ch.ID, err)
	}

	return thermal, nil
}

// thermalSubsystemFans reads the Fans collection of a ThermalSubsystem. found is false
// when the subsystem carries no Fans link.
//
// The link is taken from the raw resource because gofish's ThermalSubsystem.Fans()
// decodes the members as the deprecated Thermal fan and drops SpeedPercent.
func (c *Client) thermalSubsystemFans(ch *schemas.Chassis, subsystem *schemas.ThermalSubsystem) ([]bmc.ChassisFan, bool, error) {
	var links struct {
		Fans schemas.Link `json:"Fans"`
	}
	if err := json.Unmarshal(subsystem.RawData, &links); err != nil {
		return nil, false, fmt.Errorf("chassis %s ThermalSubsystem: %w", ch.ID, err)
	}
	if links.Fans == "" {
		return nil, false, nil
	}

	members, err := schemas.ListReferencedFans(c.client, string(links.Fans))
	if err != nil {
		return nil, false, fmt.Errorf("chassis %s ThermalSubsystem Fans: %w", ch.ID, err)
	}

	fans := make([]bmc.ChassisFan, 0, len(members))
	for _, member := range members {
		if member == nil {
			continue
		}
		fans = append(fans, bmc.ChassisFan{
			Name:            member.Name,
			PhysicalContext: string(member.PhysicalContext),
			SpeedRPM:        member.SpeedPercent.SpeedRPM,
			SpeedPercent:    member.SpeedPercent.Reading,
			Health:          string(member.Status.Health),
			State:           string(member.Status.State),
		})
	}

	return fans, true, nil
}

// metricsTemperatures converts ThermalMetrics temperature excerpts. Excerpts carry
// readings only, so Health and State stay empty.
func metricsTemperatures(metrics *schemas.ThermalMetrics) []bmc.ChassisTemperature {
	temperatures := make([]bmc.ChassisTemperature, 0, len(metrics.TemperatureReadingsCelsius))
	for i := range metrics.TemperatureReadingsCelsius {
		reading := &metrics.TemperatureReadingsCelsius[i]
		temperatures = append(temperatures, bmc.ChassisTemperature{
			Name:            reading.DeviceName,
			PhysicalContext: string(reading.PhysicalContext),
			ReadingCelsius:  reading.Reading,
		})
	}

	return temperatures
}

// legacyFans converts the fans of a deprecated Thermal resource. Reading is placed
// in SpeedRPM or SpeedPercent according to ReadingUnits.
func legacyFans(thermal *schemas.Thermal) []bmc.ChassisFan {
	fans := make([]bmc.ChassisFan, 0, len(thermal.Fans))
	for i := range thermal.Fans {
		fan := &thermal.Fans[i]
		name := fan.Name
		if name == "" {
			// Fan schemas before v1.1.0 only carry the deprecated FanName.
			name = fan.FanName //nolint:staticcheck // still the only fan name on older firmware
		}

		converted := bmc.ChassisFan{
			Name:            name,
			PhysicalContext: string(fan.PhysicalContext),
			Health:          string(fan.Status.Health),
			State:           string(fan.Status.State),
		}
		if fan.Reading != nil {
			reading := float64(*fan.Reading)
			if fan.ReadingUnits == schemas.PercentReadingUnits {
				converted.SpeedPercent = &reading
			} else {
				converted.SpeedRPM = &reading
			}
		}

		fans = append(fans, converted)
	}

	return fans
}

// legacyTemperatures converts the temperature sensors of a deprecated Thermal resource.
func legacyTemperatures(thermal *schemas.Thermal) []bmc.ChassisTemperature {
	temperatures := make([]bmc.ChassisTemperature, 0, len(thermal.Temperatures))
	for i := range thermal.Temperatures {
		sensor := &thermal.Temperatures[i]
		temperatures = append(temperatures, bmc.ChassisTemperature{
			Name:            sensor.Name,
			PhysicalContext: string(sensor.PhysicalContext),
			ReadingCelsius:  sensor.ReadingCelsius,
			Health:          string(sensor.Status.Health),
			State:           string(sensor.Status.State),
		})
	}

	return temperatures
}

// isNotFound reports whether err is a Redfish error with HTTP status 404.
func isNotFound(err error) bool {
	var redfishErr *schemas.Error
	return errors.As(err, &redfishErr) && redfishErr.HTTPReturnedStatusCode == http.StatusNotFound
}
