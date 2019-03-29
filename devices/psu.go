package devices

// Psu represents a power supply device
type Psu struct {
	Serial        string
	CapacityKw    float64
	PowerKw       float64
	Status        string
	PartnerNumber string
}
