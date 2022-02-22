package devices

type PowerSensor struct {
	ID              string
	Name            string
	InputWatts      float32
	OutputWatts     float32
	LastOutputWatts float32
}

type TemperatureSensor struct {
	ID              string
	Name            string
	ReadingCelsius  float32
	PhysicalContext string
}

type FanSensor struct {
	ID              string
	Name            string
	Reading         float32
	PhysicalContext string
}

type ChassisHealth struct {
	ID     string
	Name   string
	State  string
	Health string
}
