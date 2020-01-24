package cfgresources

// HPE struct holds configuration parameters for HPE assets.
type HPE struct {
	// static_high
	// static_low
	// dynamic
	// os_control - note this requires the machine to be rebooted
	PowerRegulator string `yaml:"regulator"`
}
