package devices

type PowerSensor struct {
	Name            string
	ID              string
	InputWatts      float32
	OutputWatts     float32
	LastOutputWatts float32
}

type TemperatureSensor struct {
	ID              string
	ReadingCelsius  float32
	PhysicalContext string
}

type FanSensor struct {
	ID              string
	Reading         float32
	PhysicalContext string
}

type ChassisHealth struct {
	ID     string
	State  string
	Health string
}
