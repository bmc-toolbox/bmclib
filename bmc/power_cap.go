package bmc

import "context"

// PowerSupplyInfo describes a single power supply.
type PowerSupplyInfo struct {
	// Name is the power supply name.
	Name string
	// Health is the power supply health (e.g. "OK", "Warning", "Critical").
	Health string
	// CapacityWatts is the maximum output capacity in watts.
	CapacityWatts float64
}

// PowerInfo aggregates a device's power metrics and supplies.
type PowerInfo struct {
	// ConsumedWatts is the current power draw in watts.
	ConsumedWatts float64
	// CapacityWatts is the total power capacity in watts.
	CapacityWatts float64
	// LimitInWatts is the configured power cap in watts; nil means no cap is
	// configured (power capping disabled).
	LimitInWatts *float64
	// PowerSupplies lists the device's power supplies.
	PowerSupplies []PowerSupplyInfo
}

// PowerReader is implemented by providers that can read a device's power
// metrics.
type PowerReader interface {
	// ReadPower returns the device's power metrics and supplies.
	ReadPower(ctx context.Context) (PowerInfo, error)
}

// PowerCapSetter is implemented by providers that can set a chassis power limit
// (power capping).
type PowerCapSetter interface {
	// SetPowerCap sets the power limit in watts. A nil limitWatts disables power
	// capping.
	SetPowerCap(ctx context.Context, limitWatts *float64) error
}
