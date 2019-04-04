package devices

// Chassis contains all the chassis the information we will expose across different vendors
type Chassis struct {
	Serial        string
	Name          string
	BmcAddress    string
	Blades        []*Blade
	StorageBlades []*StorageBlade
	Fans          []*Fan
	Nics          []*Nic
	Psus          []*Psu
	TempC         int
	PassThru      string
	Status        string
	PowerKw       float64
	Model         string
	Vendor        string
	FwVersion     string
}
