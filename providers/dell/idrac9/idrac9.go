package idrac9

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	multierror "github.com/hashicorp/go-multierror"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/httpclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

const (
	// BMCModel defines the bmc model that is supported by this package
	BMCModel = "iDRAC9"
)

// IDrac9 holds the status and properties of a connection to an iDrac device
type IDrac9 struct {
	ip             string
	username       string
	password       string
	xsrfToken      string
	client         *http.Client
	iDracInventory *dell.IDracInventory
}

// New returns a new IDrac9 ready to be used
func New(ip string, username string, password string) (iDrac *IDrac9, err error) {
	client, err := httpclient.Build()
	if err != nil {
		return iDrac, err
	}

	return &IDrac9{ip: ip, username: username, password: password, client: client}, err
}

// Login initiates the connection to a bmc device
func (i *IDrac9) Login() (err error) {
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip}).Debug("connecting to bmc")

	url := fmt.Sprintf("https://%s/sysmgmt/2015/bmc/session", i.ip)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("user", fmt.Sprintf("\"%s\"", i.username))
	req.Header.Add("password", fmt.Sprintf("\"%s\"", i.password))

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return errors.ErrPageNotFound
	}

	i.xsrfToken = resp.Header.Get("XSRF-TOKEN")

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	iDracAuth := &dell.IDracAuth{}
	err = json.Unmarshal(payload, iDracAuth)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return err
	}

	if iDracAuth.AuthResult != 0 {
		return errors.ErrLoginFailed
	}

	err = i.loadHwData()
	if err != nil {
		return err
	}

	return err
}

// loadHwData load the full hardware information from the iDrac
func (i *IDrac9) loadHwData() (err error) {
	url := "sysmgmt/2012/server/inventory/hardware"
	payload, err := i.get(url, nil)
	if err != nil {
		return err
	}

	iDracInventory := &dell.IDracInventory{}
	err = xml.Unmarshal(payload, iDracInventory)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return err
	}

	if iDracInventory == nil || iDracInventory.Component == nil {
		return errors.ErrUnableToReadData
	}

	i.iDracInventory = iDracInventory

	return err
}

// get calls a given json endpoint of the ilo and returns the data
func (i *IDrac9) get(endpoint string, extraHeaders *map[string]string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip, "endpoint": endpoint}).Debug("retrieving data from bmc")

	bmcURL := fmt.Sprintf("https://%s", i.ip)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return payload, err
	}

	req.Header.Add("XSRF-TOKEN", i.xsrfToken)

	if extraHeaders != nil {
		for key, value := range *extraHeaders {
			req.Header.Add(key, value)
		}
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	if resp.StatusCode == 404 {
		return payload, errors.ErrPageNotFound
	}

	return payload, err
}

// Nics returns all found Nics in the device
func (i *IDrac9) Nics() (nics []*devices.Nic, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_NICView" {
			for _, property := range component.Properties {
				if property.Name == "ProductName" && property.Type == "string" {
					data := strings.Split(property.Value, " - ")
					if len(data) == 2 {
						if nics == nil {
							nics = make([]*devices.Nic, 0)
						}

						n := &devices.Nic{
							Name:       data[0],
							MacAddress: strings.ToLower(data[1]),
						}
						nics = append(nics, n)
					} else {
						err = multierror.Append(err, fmt.Errorf("invalid network card %s, please review", data))
					}
				}
			}
		} else if component.Classname == "DCIM_iDRACCardView" {
			for _, property := range component.Properties {
				if property.Name == "PermanentMACAddress" && property.Type == "string" {
					if nics == nil {
						nics = make([]*devices.Nic, 0)
					}

					n := &devices.Nic{
						Name:       "bmc",
						MacAddress: strings.ToLower(property.Value),
					}
					nics = append(nics, n)
				}
			}
		}
	}
	return nics, err
}

// Serial returns the device serial
func (i *IDrac9) Serial() (serial string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "NodeID" && property.Type == "string" {
					return strings.ToLower(property.Value), err
				}
			}
		}
	}
	return serial, err
}

// Status returns health string status from the bmc
func (i *IDrac9) Status() (status string, err error) {
	extraHeaders := &map[string]string{
		"X-SYSMGMT-OPTIMIZE": "true",
	}

	url := "sysmgmt/2016/server/extended_health"
	payload, err := i.get(url, extraHeaders)
	if err != nil {
		return status, err
	}

	iDracHealthStatus := &dell.IDracHealthStatus{}
	err = json.Unmarshal(payload, iDracHealthStatus)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return status, err
	}

	for _, entry := range iDracHealthStatus.HealthStatus {
		if entry != 0 && entry != 2 {
			return "Degraded", err
		}
	}

	return "OK", err
}

// PowerKw returns the current power usage in Kw
func (i *IDrac9) PowerKw() (power float64, err error) {
	url := "sysmgmt/2015/server/sensor/power"
	payload, err := i.get(url, nil)
	if err != nil {
		return power, err
	}

	iDracPowerData := &dell.IDracPowerData{}
	err = json.Unmarshal(payload, iDracPowerData)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return power, err
	}

	return iDracPowerData.Root.Powermonitordata.PresentReading.Reading.Reading / 1000.00, err
}

// PowerState returns the current power state of the machine
func (i *IDrac9) PowerState() (state string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "PowerState" && property.Type == "uint16" {
					return strings.ToLower(property.DisplayValue), err
				}
			}
		}
	}
	return state, err
}

// BiosVersion returns the current version of the bios
func (i *IDrac9) BiosVersion() (version string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "BIOSVersionString" && property.Type == "string" {
					return property.Value, err
				}
			}
		}
	}

	return version, err
}

// Name returns the name of this server from the bmc point of view
func (i *IDrac9) Name() (name string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "HostName" && property.Type == "string" {
					return property.Value, err
				}
			}
		}
	}

	return name, err
}

// BmcVersion returns the version of the bmc we are running
func (i *IDrac9) BmcVersion() (bmcVersion string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_iDRACCardView" {
			for _, property := range component.Properties {
				if property.Name == "FirmwareVersion" && property.Type == "string" {
					return property.Value, err
				}
			}
		}
	}
	return bmcVersion, err
}

// Model returns the device model
func (i *IDrac9) Model() (model string, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "Model" && property.Type == "string" {
					return property.Value, err
				}
			}
		}
	}
	return model, err
}

// BmcType returns the type of bmc we are talking to
func (i *IDrac9) BmcType() (bmcType string, err error) {
	return "iDrac9", err
}

// License returns the bmc license information
func (i *IDrac9) License() (name string, licType string, err error) {
	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	url := "sysmgmt/2012/server/license"
	payload, err := i.get(url, extraHeaders)
	if err != nil {
		return name, licType, err
	}

	iDracLicense := &dell.IDracLicense{}
	err = json.Unmarshal(payload, iDracLicense)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return name, licType, err
	}

	if iDracLicense.License.VConsole == 1 {
		return "Enterprise", "Licensed", err
	}
	return "-", "Unlicensed", err
}

// Memory return the total amount of memory of the server
func (i *IDrac9) Memory() (mem int, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "SysMemTotalSize" && property.Type == "uint32" {
					size, err := strconv.Atoi(property.Value)
					if err != nil {
						return mem, err
					}
					return size / 1024, err
				}
			}
		}
	}
	return mem, err
}

// TempC returns the current temperature of the machine
func (i *IDrac9) TempC() (temp int, err error) {
	extraHeaders := &map[string]string{
		"X-SYSMGMT-OPTIMIZE": "true",
	}

	url := "sysmgmt/2012/server/temperature"
	payload, err := i.get(url, extraHeaders)
	if err != nil {
		return temp, err
	}

	iDracTemp := &dell.IDracTemp{}
	err = json.Unmarshal(payload, iDracTemp)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return temp, err
	}

	return iDracTemp.Temperatures.IDRACEmbedded1SystemBoardInletTemp.Reading, err
}

// CPU return the cpu, cores and hyperthreads the server
func (i *IDrac9) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_CPUView" {
			cpuCount++
			if component.Key == "CPU.Socket.1" {
				var e error
				for _, property := range component.Properties {
					if property.Name == "Model" && property.Type == "string" {
						cpu = httpclient.StandardizeProcessorName(property.DisplayValue)
					} else if property.Name == "NumberOfProcessorCores" && property.Type == "uint32" {
						if coreCount, e = strconv.Atoi(property.Value); e != nil {
							err = multierror.Append(err, fmt.Errorf("invalid core count %s", e))
						}
					} else if property.Name == "NumberOfEnabledThreads" && property.Type == "uint32" {
						if hyperthreadCount, e = strconv.Atoi(property.Value); e != nil {
							err = multierror.Append(err, fmt.Errorf("invalid thread count %s", e))
						}
					}
				}
			}
		}
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// Logout logs out and close the bmc connection
func (i *IDrac9) Logout() (err error) {
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip}).Debug("logout from bmc")

	resp, err := i.client.Get(fmt.Sprintf("https://%s/sysmgmt/2015/bmc/session/logout", i.ip))
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return err
}

// IsBlade returns if the current hardware is a blade or not
func (i *IDrac9) IsBlade() (isBlade bool, err error) {
	model, err := i.Model()
	if err != nil {
		return isBlade, err
	}

	if strings.HasPrefix(model, "PowerEdge M") {
		isBlade = true
	}

	return isBlade, err
}

// Psus returns a list of psus installed on the device
func (i *IDrac9) Psus() (psus []*devices.Psu, err error) {
	url := "data?get=powerSupplies"
	payload, err := i.get(url, nil)
	if err != nil {
		return psus, err
	}

	iDracRoot := &dell.IDracRoot{}
	err = xml.Unmarshal(payload, iDracRoot)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return psus, err
	}

	serial, _ := i.Serial()

	for _, psu := range iDracRoot.PsSensorList {
		if psus == nil {
			psus = make([]*devices.Psu, 0)
		}
		var status string
		if psu.SensorHealth == 2 {
			status = "OK"
		} else {
			status = "BROKEN"
		}

		p := &devices.Psu{
			Serial:     fmt.Sprintf("%s_%s", serial, strings.Split(psu.Name, " ")[0]),
			Status:     status,
			CapacityKw: float64(psu.MaxWattage) / 1000.00,
		}

		psus = append(psus, p)
	}

	return psus, err
}

// Vendor returns bmc's vendor
func (i *IDrac9) Vendor() (vendor string) {
	return dell.VendorID
}

// ServerSnapshot do best effort to populate the server data and returns a blade or discrete
func (i *IDrac9) ServerSnapshot() (server interface{}, err error) {
	if isBlade, _ := i.IsBlade(); isBlade {
		blade := &devices.Blade{}
		blade.Serial, _ = i.Serial()
		blade.BmcAddress = i.ip
		blade.BmcType, _ = i.BmcType()
		blade.BmcVersion, _ = i.BmcVersion()
		blade.Model, _ = i.Model()
		blade.Vendor = i.Vendor()
		blade.Nics, _ = i.Nics()
		blade.Disks, _ = i.Disks()
		blade.BiosVersion, _ = i.BiosVersion()
		blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, _ = i.CPU()
		blade.Memory, _ = i.Memory()
		blade.Status, _ = i.Status()
		blade.Name, _ = i.Name()
		blade.TempC, _ = i.TempC()
		blade.PowerKw, _ = i.PowerKw()
		blade.BmcLicenceType, blade.BmcLicenceStatus, _ = i.License()
		server = blade
	} else {
		discrete := &devices.Discrete{}
		discrete.Serial, _ = i.Serial()
		discrete.BmcAddress = i.ip
		discrete.BmcType, _ = i.BmcType()
		discrete.BmcVersion, _ = i.BmcVersion()
		discrete.Model, _ = i.Model()
		discrete.Vendor = i.Vendor()
		discrete.Nics, _ = i.Nics()
		discrete.Disks, _ = i.Disks()
		discrete.BiosVersion, _ = i.BiosVersion()
		discrete.Processor, discrete.ProcessorCount, discrete.ProcessorCoreCount, discrete.ProcessorThreadCount, _ = i.CPU()
		discrete.Memory, _ = i.Memory()
		discrete.Status, _ = i.Status()
		discrete.Name, _ = i.Name()
		discrete.TempC, _ = i.TempC()
		discrete.PowerKw, _ = i.PowerKw()
		discrete.BmcLicenceType, discrete.BmcLicenceStatus, _ = i.License()
		discrete.Psus, _ = i.Psus()
		server = discrete
	}

	return server, err
}

// Disks returns a list of disks installed on the device
func (i *IDrac9) Disks() (disks []*devices.Disk, err error) {
	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_PhysicalDiskView" {
			if disks == nil {
				disks = make([]*devices.Disk, 0)
			}
			disk := &devices.Disk{}

			for _, property := range component.Properties {
				if property.Name == "Model" {
					disk.Model = strings.ToLower(property.Value)
				} else if property.Name == "SerialNumber" {
					disk.Serial = strings.ToLower(property.Value)
				} else if property.Name == "MediaType" {
					if property.DisplayValue == "Solid State Drive" {
						disk.Type = "SSD"
					} else if property.DisplayValue == "Hard Disk Drive" {
						disk.Type = "HDD"
					} else {
						disk.Type = property.DisplayValue
					}
				} else if property.Name == "PrimaryStatus" {
					disk.Status = property.DisplayValue
				} else if property.Name == "SizeInBytes" {
					size, err := strconv.Atoi(property.Value)
					if err != nil {
						return disks, err
					}
					disk.Size = fmt.Sprintf("%d GB", size/1024/1024/1024)
				} else if property.Name == "Revision" {
					disk.FwVersion = strings.ToLower(property.Value)
				}
			}

			if disk.Serial != "" {
				disks = append(disks, disk)
			}
		}
	}
	return disks, err
}

// UpdateCredentials updates login credentials
func (i *IDrac9) UpdateCredentials(username string, password string) {
	i.username = username
	i.password = password
}
