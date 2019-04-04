package devices

// Fan represents que current status a fan
type Fan struct {
	Serial     string
	Status     string
	Position   int
	Model      string
	CurrentRPM int64
	PowerKw    float64
}
