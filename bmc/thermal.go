package bmc

import "context"

// TemperatureReading is a single temperature sensor reading.
type TemperatureReading struct {
	// Name is the sensor name (e.g. "Ambient Temp").
	Name string
	// ReadingCelsius is the current temperature in degrees Celsius.
	ReadingCelsius float64
	// Health is the sensor health (e.g. "OK", "Warning", "Critical").
	Health string
}

// FanReading is a single fan reading.
type FanReading struct {
	// Name is the fan name (e.g. "Fan 1 Tach").
	Name string
	// Reading is the current fan reading (RPM, unless the device reports
	// another unit).
	Reading int
	// Health is the fan health (e.g. "OK", "Warning", "Critical").
	Health string
}

// ThermalReading aggregates the thermal sensors of a device.
type ThermalReading struct {
	Temperatures []TemperatureReading
	Fans         []FanReading
}

// ThermalReader is implemented by providers that can read a device's thermal
// sensors (temperatures and fans).
type ThermalReader interface {
	// Thermal returns the device's temperature and fan readings.
	Thermal(ctx context.Context) (ThermalReading, error)
}
