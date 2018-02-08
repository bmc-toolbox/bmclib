package devices

// Discrete contains all the blade information we will expose across different vendors
type Discrete struct {
	Serial               string
	Name                 string
	BiosVersion          string
	BmcType              string
	BmcAddress           string
	BmcVersion           string
	BmcSSHReachable      bool
	BmcWEBReachable      bool
	BmcIpmiReachable     bool
	BmcLicenceType       string
	BmcLicenceStatus     string
	BmcAuth              bool
	Nics                 []*Nic
	Psus                 []*Psu
	Model                string
	TempC                int
	PowerKw              float64
	Status               string
	Vendor               string
	Processor            string
	ProcessorCount       int
	ProcessorCoreCount   int
	ProcessorThreadCount int
	Memory               int
}
