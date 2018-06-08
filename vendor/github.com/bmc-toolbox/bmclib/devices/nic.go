package devices

// Nic represents a network interface devices
type Nic struct {
	MacAddress string
	Name       string
	Up         bool
	Speed      string
}
