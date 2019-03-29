package devices

// Fan represents que current status a fan
type Fan struct {
	Status     string
	Position   int
	Model      string
	CurrentRPM int
	MaxRPM     int
	MinRPM     int
	PowerKw    float64
}
