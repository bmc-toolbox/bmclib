package devices

// Blade contains all the blade information we will expose across different vendors
type Blade struct {
	Serial               string
	Name                 string
	BiosVersion          string
	BmcType              string
	BmcAddress           string
	BmcVersion           string
	BmcLicenceType       string
	BmcLicenceStatus     string
	Nics                 []*Nic
	BladePosition        int
	Model                string
	TempC                int
	PowerKw              float64
	Status               string
	Vendor               string
	ChassisSerial        string
	Processor            string
	ProcessorCount       int
	ProcessorCoreCount   int
	ProcessorThreadCount int
	StorageBlade         StorageBlade
	Memory               int
}
