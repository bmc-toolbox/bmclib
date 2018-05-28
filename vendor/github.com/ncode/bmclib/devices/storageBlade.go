package devices

// StorageBlade contains all the storage blade information we will expose across different vendors
type StorageBlade struct {
	Serial        string
	FwVersion     string
	BladePosition int
	Model         string
	TempC         int
	PowerKw       float64
	Status        string
	Vendor        string
	ChassisSerial string
	BladeSerial   string
}
