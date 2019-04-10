package idrac8

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	multierror "github.com/hashicorp/go-multierror"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

const (
	// BMCType defines the bmc model that is supported by this package
	BMCType = "idrac8"
)

// IDrac8 holds the status and properties of a connection to an iDrac device
type IDrac8 struct {
	ip             string
	username       string
	password       string
	httpClient     *http.Client
	sshClient      *sshclient.SSHClient
	st1            string
	st2            string
	iDracInventory *dell.IDracInventory
}

// New returns a new IDrac8 ready to be used
func New(ip string, username string, password string) (iDrac *IDrac8, err error) {
	return &IDrac8{ip: ip, username: username, password: password}, err
}

// CheckCredentials verify whether the credentials are valid or not
func (i *IDrac8) CheckCredentials() (err error) {
	err = i.httpLogin()
	if err != nil {
		return err
	}
	return nil
}

// PUTs data
func (i *IDrac8) put(endpoint string, payload []byte) (statusCode int, response []byte, err error) {
	bmcURL := fmt.Sprintf("https://%s", i.ip)

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s", bmcURL, endpoint), bytes.NewReader(payload))
	if err != nil {
		return statusCode, response, err
	}
	req.Header.Add("ST2", i.st2)

	u, err := url.Parse(bmcURL)
	if err != nil {
		return statusCode, response, err
	}

	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "-http-session-" || cookie.Name == "tokenvalue" {
			req.AddCookie(cookie)
		}
	}

	if log.GetLevel() == log.TraceLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] %s/%s", bmcURL, endpoint))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return statusCode, response, err
	}
	defer resp.Body.Close()

	if log.GetLevel() == log.TraceLevel {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("[Response]")
			log.Println("<<<<<<<<<<<<<<")
			log.Printf("%s\n\n", dump)
			log.Println("<<<<<<<<<<<<<<")
		}
	}

	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return statusCode, response, err
	}

	if resp.StatusCode == 500 {
		return resp.StatusCode, response, errors.Err500
	}

	return resp.StatusCode, response, err
}

// posts the payload to the given endpoint
func (i *IDrac8) post(endpoint string, data []byte, formDataContentType string) (statusCode int, body []byte, err error) {

	u, err := url.Parse(fmt.Sprintf("https://%s/%s", i.ip, endpoint))
	if err != nil {
		return 0, []byte{}, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
	if err != nil {
		return 0, []byte{}, err
	}

	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "-http-session-" || cookie.Name == "tokenvalue" {
			req.AddCookie(cookie)
		}
	}

	req.Header.Add("ST2", i.st2)

	if formDataContentType == "" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	} else {

		// Set multipart form content type
		req.Header.Set("Content-Type", formDataContentType)

		// set the token value
		c := new(http.Cookie)
		c.Name = "tokenvalue"
		c.Value = i.st1

		req.AddCookie(c)
	}

	if log.GetLevel() == log.TraceLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] https://%s/%s", i.ip, endpoint))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return 0, []byte{}, err
	}
	defer resp.Body.Close()
	if log.GetLevel() == log.TraceLevel {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("[Response]")
			log.Println("<<<<<<<<<<<<<<")
			log.Printf("%s\n\n", dump)
			log.Println("<<<<<<<<<<<<<<")
		}
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, []byte{}, err
	}

	//fmt.Printf("%s\n", body)
	return resp.StatusCode, body, err
}

// get calls a given json endpoint of the ilo and returns the data
func (i *IDrac8) get(endpoint string, extraHeaders *map[string]string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip, "endpoint": endpoint}).Debug("retrieving data from bmc")

	bmcURL := fmt.Sprintf("https://%s", i.ip)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return payload, err
	}
	req.Header.Add("ST2", i.st2)
	if extraHeaders != nil {
		for key, value := range *extraHeaders {
			req.Header.Add(key, value)
		}
	}

	u, err := url.Parse(bmcURL)
	if err != nil {
		return payload, err
	}

	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "-http-session-" {
			req.AddCookie(cookie)
		}
	}
	if log.GetLevel() == log.TraceLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] https://%s/%s", bmcURL, endpoint))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()
	if log.GetLevel() == log.TraceLevel {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("[Response]")
			log.Println("<<<<<<<<<<<<<<")
			log.Printf("%s\n\n", dump)
			log.Println("<<<<<<<<<<<<<<")
		}
	}

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
func (i *IDrac8) Nics() (nics []*devices.Nic, err error) {
	err = i.loadHwData()
	if err != nil {
		return nics, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_NICView" {
			var speed string
			var up bool
			for _, property := range component.Properties {
				if property.Name == "LinkSpeed" && property.Type == "uint8" && property.DisplayValue != "Unknown" {
					speed = property.DisplayValue
					up = true
				} else if property.Name == "ProductName" && property.Type == "string" {
					data := strings.Split(property.Value, " - ")
					if len(data) == 2 {
						if nics == nil {
							nics = make([]*devices.Nic, 0)
						}

						n := &devices.Nic{
							Name:       data[0],
							Speed:      speed,
							Up:         up,
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
func (i *IDrac8) Serial() (serial string, err error) {
	err = i.loadHwData()
	if err != nil {
		return serial, err
	}

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
func (i *IDrac8) Status() (status string, err error) {
	err = i.httpLogin()
	if err != nil {
		return status, err
	}

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
func (i *IDrac8) PowerKw() (power float64, err error) {
	err = i.httpLogin()
	if err != nil {
		return power, err
	}

	url := "data?get=powermonitordata"
	payload, err := i.get(url, nil)
	if err != nil {
		return power, err
	}

	iDracRoot := &dell.IDracRoot{}
	err = xml.Unmarshal(payload, iDracRoot)
	if err != nil {
		return power, err
	}

	if iDracRoot.Powermonitordata != nil && iDracRoot.Powermonitordata.PresentReading != nil && iDracRoot.Powermonitordata.PresentReading.Reading != nil {
		value, err := strconv.Atoi(iDracRoot.Powermonitordata.PresentReading.Reading.Reading)
		if err != nil {
			return power, err
		}
		return float64(value) / 1000.00, err
	}

	return power, err
}

// PowerState returns the current power state of the machine
func (i *IDrac8) PowerState() (state string, err error) {
	err = i.loadHwData()
	if err != nil {
		return state, err
	}

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
func (i *IDrac8) BiosVersion() (version string, err error) {
	err = i.loadHwData()
	if err != nil {
		return version, err
	}

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
func (i *IDrac8) Name() (name string, err error) {
	err = i.loadHwData()
	if err != nil {
		return name, err
	}

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
func (i *IDrac8) BmcVersion() (bmcVersion string, err error) {
	err = i.loadHwData()
	if err != nil {
		return bmcVersion, err
	}

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
func (i *IDrac8) Model() (model string, err error) {
	err = i.loadHwData()
	if err != nil {
		return model, err
	}

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
func (i *IDrac8) BmcType() (bmcType string) {
	return BMCType
}

// License returns the bmc license information
func (i *IDrac8) License() (name string, licType string, err error) {
	err = i.httpLogin()
	if err != nil {
		return name, licType, err
	}

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
		return name, licType, err
	}

	if iDracLicense.License.VConsole == 1 {
		return "Enterprise", "Licensed", err
	}
	return "-", "Unlicensed", err
}

// Memory return the total amount of memory of the server
func (i *IDrac8) Memory() (mem int, err error) {
	err = i.loadHwData()
	if err != nil {
		return mem, err
	}

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

// Disks returns a list of disks installed on the device
func (i *IDrac8) Disks() (disks []*devices.Disk, err error) {
	err = i.loadHwData()
	if err != nil {
		return disks, err
	}

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
				} else if property.Name == "DeviceDescription" {
					disk.Location = property.DisplayValue
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

// TempC returns the current temperature of the machine
func (i *IDrac8) TempC() (temp int, err error) {
	err = i.httpLogin()
	if err != nil {
		return temp, err
	}

	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	url := "sysmgmt/2012/server/temperature"
	payload, err := i.get(url, extraHeaders)
	if err != nil {
		return temp, err
	}

	iDracTemp := &dell.IDracTemp{}
	err = json.Unmarshal(payload, iDracTemp)
	if err != nil {
		return temp, err
	}

	return iDracTemp.Temperatures.IDRACEmbedded1SystemBoardInletTemp.Reading, err
}

// CPU return the cpu, cores and hyperthreads the server
func (i *IDrac8) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	err = i.httpLogin()
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	url := "sysmgmt/2012/server/processor"
	payload, err := i.get(url, extraHeaders)
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	dellBladeProc := &dell.BladeProcessorEndpoint{}
	err = json.Unmarshal(payload, dellBladeProc)
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	for _, proc := range dellBladeProc.Proccessors {
		hasHT := 0
		for _, ht := range proc.HyperThreading {
			if ht.Capable == 1 {
				hasHT = 2
			}
		}
		return httpclient.StandardizeProcessorName(proc.Brand), len(dellBladeProc.Proccessors), proc.CoreCount, proc.CoreCount * hasHT, err
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// IsBlade returns if the current hardware is a blade or not
func (i *IDrac8) IsBlade() (isBlade bool, err error) {
	err = i.httpLogin()
	if err != nil {
		return isBlade, err
	}

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
func (i *IDrac8) Psus() (psus []*devices.Psu, err error) {
	err = i.httpLogin()
	if err != nil {
		return psus, err
	}

	url := "data?get=powerSupplies"
	payload, err := i.get(url, nil)
	if err != nil {
		return psus, err
	}

	iDracRoot := &dell.IDracRoot{}
	err = xml.Unmarshal(payload, iDracRoot)
	if err != nil {
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
			PowerKw:    0.00,
			CapacityKw: float64(psu.MaxWattage) / 1000.00,
		}

		psus = append(psus, p)
	}

	return psus, err
}

// Vendor returns bmc's vendor
func (i *IDrac8) Vendor() (vendor string) {
	return dell.VendorID
}

// ServerSnapshot do best effort to populate the server data and returns a blade or discrete
func (i *IDrac8) ServerSnapshot() (server interface{}, err error) {
	err = i.httpLogin()
	if err != nil {
		return server, err
	}

	if isBlade, _ := i.IsBlade(); isBlade {
		blade := &devices.Blade{}
		blade.Vendor = i.Vendor()
		blade.BmcAddress = i.ip
		blade.BmcType = i.BmcType()

		blade.Serial, _ = i.Serial()
		if err != nil {
			return nil, err
		}
		blade.BmcVersion, err = i.BmcVersion()
		if err != nil {
			return nil, err
		}
		blade.Model, err = i.Model()
		if err != nil {
			return nil, err
		}
		blade.Nics, err = i.Nics()
		if err != nil {
			return nil, err
		}
		blade.Disks, err = i.Disks()
		if err != nil {
			return nil, err
		}
		blade.BiosVersion, err = i.BiosVersion()
		if err != nil {
			return nil, err
		}
		blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = i.CPU()
		if err != nil {
			return nil, err
		}
		blade.Memory, err = i.Memory()
		if err != nil {
			return nil, err
		}
		blade.Status, err = i.Status()
		if err != nil {
			return nil, err
		}
		blade.Name, err = i.Name()
		if err != nil {
			return nil, err
		}
		blade.TempC, err = i.TempC()
		if err != nil {
			return nil, err
		}
		blade.PowerKw, err = i.PowerKw()
		if err != nil {
			return nil, err
		}
		blade.PowerState, err = i.PowerState()
		if err != nil {
			return nil, err
		}
		blade.BmcLicenceType, blade.BmcLicenceStatus, err = i.License()
		if err != nil {
			return nil, err
		}
		server = blade
	} else {
		discrete := &devices.Discrete{}
		discrete.Vendor = i.Vendor()
		discrete.BmcAddress = i.ip
		discrete.BmcType = i.BmcType()

		discrete.Serial, err = i.Serial()
		if err != nil {
			return nil, err
		}
		discrete.BmcVersion, err = i.BmcVersion()
		if err != nil {
			return nil, err
		}
		discrete.Model, err = i.Model()
		if err != nil {
			return nil, err
		}
		discrete.Nics, err = i.Nics()
		if err != nil {
			return nil, err
		}
		discrete.Disks, err = i.Disks()
		if err != nil {
			return nil, err
		}
		discrete.BiosVersion, err = i.BiosVersion()
		if err != nil {
			return nil, err
		}
		discrete.Processor, discrete.ProcessorCount, discrete.ProcessorCoreCount, discrete.ProcessorThreadCount, err = i.CPU()
		if err != nil {
			return nil, err
		}
		discrete.Memory, err = i.Memory()
		if err != nil {
			return nil, err
		}
		discrete.Status, err = i.Status()
		if err != nil {
			return nil, err
		}
		discrete.Name, err = i.Name()
		if err != nil {
			return nil, err
		}
		discrete.TempC, err = i.TempC()
		if err != nil {
			return nil, err
		}
		discrete.PowerKw, err = i.PowerKw()
		if err != nil {
			return nil, err
		}
		discrete.PowerState, err = i.PowerState()
		if err != nil {
			return nil, err
		}
		discrete.BmcLicenceType, discrete.BmcLicenceStatus, err = i.License()
		if err != nil {
			return nil, err
		}
		discrete.Psus, err = i.Psus()
		if err != nil {
			return nil, err
		}
		server = discrete
	}

	return server, err
}

// UpdateCredentials updates login credentials
func (i *IDrac8) UpdateCredentials(username string, password string) {
	i.username = username
	i.password = password
}
