package ilo

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/hp"

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

const (
	// BmcType defines the bmc model that is supported by this package
	BmcType = "ilo"

	// Ilo2 is the constant for iLO2
	Ilo2 = "ilo2"
	// Ilo3 is the constant for iLO3
	Ilo3 = "ilo3"
	// Ilo4 is the constant for iLO4
	Ilo4 = "ilo4"
	// Ilo5 is the constant for iLO5
	Ilo5 = "ilo5"
)

// Ilo holds the status and properties of a connection to an iLO device
type Ilo struct {
	ip         string
	username   string
	password   string
	sessionKey string
	httpClient *http.Client
	sshClient  *sshclient.SSHClient
	serial     string
	loginURL   *url.URL
	rimpBlade  *hp.RimpBlade
}

// New returns a new Ilo ready to be used
func New(ip string, username string, password string) (ilo *Ilo, err error) {
	loginURL, err := url.Parse(fmt.Sprintf("https://%s/json/login_session", ip))
	if err != nil {
		return nil, err
	}

	client, err := httpclient.Build()
	if err != nil {
		return ilo, err
	}

	xmlURL := fmt.Sprintf("https://%s/xmldata?item=all", ip)
	resp, err := client.Get(xmlURL)
	if err != nil {
		return ilo, err
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ilo, err
	}
	defer resp.Body.Close()

	rimpBlade := &hp.RimpBlade{}
	err = xml.Unmarshal(payload, rimpBlade)
	if err != nil {
		httpclient.DumpInvalidPayload(xmlURL, ip, payload)
		return ilo, err
	}

	return &Ilo{ip: ip, username: username, password: password, loginURL: loginURL, rimpBlade: rimpBlade}, err
}

// CheckCredentials verify whether the credentials are valid or not
func (i *Ilo) CheckCredentials() (err error) {
	err = i.httpLogin()
	if err != nil {
		return err
	}
	return err
}

// get calls a given json endpoint of the iLO and returns the data
func (i *Ilo) get(endpoint string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": hp.VendorID, "ip": i.ip, "endpoint": endpoint}).Debug("retrieving data from bmc")

	bmcURL := fmt.Sprintf("https://%s", i.ip)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return payload, err
	}

	u, err := url.Parse(bmcURL)
	if err != nil {
		return payload, err
	}

	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "sessionKey" {
			req.AddCookie(cookie)
		}
	}
	if log.GetLevel() == log.DebugLevel {
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
		return payload, err
	}
	defer resp.Body.Close()
	if log.GetLevel() == log.DebugLevel {
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

// posts the payload to the given endpoint
func (i *Ilo) post(endpoint string, data []byte, debug bool) (statusCode int, body []byte, err error) {

	u, err := url.Parse(fmt.Sprintf("https://%s/%s", i.ip, endpoint))
	if err != nil {
		return 0, []byte{}, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
	if err != nil {
		return 0, []byte{}, err
	}

	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "sessionKey" {
			req.AddCookie(cookie)
		}
	}

	if debug {
		fmt.Println(fmt.Sprintf("%s/%s", i.ip, endpoint))
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			fmt.Printf("%s\n\n", dump)
		}
	}

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return 0, []byte{}, err
	}
	defer resp.Body.Close()
	if debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			fmt.Printf("%s\n\n", dump)
		}
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, []byte{}, err
	}

	//fmt.Printf("%s\n", body)
	return resp.StatusCode, body, err
}

// Serial returns the device serial
func (i *Ilo) Serial() (serial string, err error) {
	return strings.ToLower(strings.TrimSpace(i.rimpBlade.HSI.Sbsn)), err
}

// Model returns the device model
func (i *Ilo) Model() (model string, err error) {
	return i.rimpBlade.HSI.Spn, err
}

// BmcType returns the type of bmc we are talking to
func (i *Ilo) BmcType() (bmcType string) {
	switch i.rimpBlade.MP.Pn {
	case "Integrated Lights-Out 2 (iLO 2)":
		return Ilo2
	case "Integrated Lights-Out 3 (iLO 3)":
		return Ilo3
	case "Integrated Lights-Out 4 (iLO 4)":
		return Ilo4
	case "Integrated Lights-Out 5 (iLO 5)":
		return Ilo5
	default:
		return i.rimpBlade.MP.Pn
	}
}

// BmcVersion returns the version of the bmc we are running
func (i *Ilo) BmcVersion() (bmcVersion string, err error) {
	return i.rimpBlade.MP.Fwri, err
}

// Name returns the name of this server from the iLO point of view
func (i *Ilo) Name() (name string, err error) {
	err = i.httpLogin()
	if err != nil {
		return name, err
	}

	url := "json/overview"
	payload, err := i.get(url)
	if err != nil {
		return name, err
	}

	overview := &hp.Overview{}
	err = json.Unmarshal(payload, overview)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return name, err
	}

	return overview.ServerName, err
}

// Status returns health string status from the bmc
func (i *Ilo) Status() (health string, err error) {
	err = i.httpLogin()
	if err != nil {
		return health, err
	}

	url := "json/overview"
	payload, err := i.get(url)
	if err != nil {
		return health, err
	}

	overview := &hp.Overview{}
	err = json.Unmarshal(payload, overview)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return health, err
	}

	if overview.SystemHealth == "OP_STATUS_OK" {
		return "OK", err
	}

	return overview.SystemHealth, err
}

// Memory returns the total amount of memory of the server
func (i *Ilo) Memory() (mem int, err error) {
	err = i.httpLogin()
	if err != nil {
		return mem, err
	}

	url := "json/mem_info"
	payload, err := i.get(url)
	if err != nil {
		return mem, err
	}

	hpMemData := &hp.Mem{}
	err = json.Unmarshal(payload, hpMemData)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return mem, err
	}

	if hpMemData.MemTotalMemSize != 0 {
		return hpMemData.MemTotalMemSize / 1024, err
	}

	for _, slot := range hpMemData.Memory {
		mem = mem + slot.MemSize
	}

	return mem / 1024, err
}

// CPU returns the cpu, cores and hyperthreads of the server
func (i *Ilo) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	err = i.httpLogin()
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	url := "json/proc_info"
	payload, err := i.get(url)
	if err != nil {
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	hpProcData := &hp.Procs{}
	err = json.Unmarshal(payload, hpProcData)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return cpu, cpuCount, coreCount, hyperthreadCount, err
	}

	for _, proc := range hpProcData.Processors {
		return httpclient.StandardizeProcessorName(proc.ProcName), len(hpProcData.Processors), proc.ProcNumCores, proc.ProcNumThreads, err
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// BiosVersion returns the current version of the bios
func (i *Ilo) BiosVersion() (version string, err error) {
	err = i.httpLogin()
	if err != nil {
		return version, err
	}

	url := "json/overview"
	payload, err := i.get(url)
	if err != nil {
		return version, err
	}

	overview := &hp.Overview{}
	err = json.Unmarshal(payload, overview)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return version, err
	}

	if overview.SystemRom != "" {
		return overview.SystemRom, err
	}

	return version, errors.ErrBiosNotFound
}

// PowerKw returns the current power usage in Kw
func (i *Ilo) PowerKw() (power float64, err error) {
	err = i.httpLogin()
	if err != nil {
		return power, err
	}

	url := "json/power_summary"
	payload, err := i.get(url)
	if err != nil {
		return power, err
	}

	hpPowerSummary := &hp.PowerSummary{}
	err = json.Unmarshal(payload, hpPowerSummary)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return power, err
	}

	return float64(hpPowerSummary.PowerSupplyInputPower) / 1024, err
}

// PowerState returns the current power state of the machine
func (i *Ilo) PowerState() (state string, err error) {
	err = i.httpLogin()
	if err != nil {
		return state, err
	}

	url := "json/power_summary"
	payload, err := i.get(url)
	if err != nil {
		return state, err
	}

	hpPowerSummary := &hp.PowerSummary{}
	err = json.Unmarshal(payload, hpPowerSummary)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return state, err
	}

	return strings.ToLower(hpPowerSummary.HostpwrState), err
}

// TempC returns the current temperature of the machine
func (i *Ilo) TempC() (temp int, err error) {
	err = i.httpLogin()
	if err != nil {
		return temp, err
	}

	url := "json/health_temperature"
	payload, err := i.get(url)
	if err != nil {
		return temp, err
	}

	hpHealthTemperature := &hp.HealthTemperature{}
	err = json.Unmarshal(payload, hpHealthTemperature)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return temp, err
	}

	for _, item := range hpHealthTemperature.Temperature {
		if item.Location == "Ambient" {
			return item.Currentreading, err
		}
	}

	return temp, err
}

// Nics returns all found Nics in the device
func (i *Ilo) Nics() (nics []*devices.Nic, err error) {
	if i.rimpBlade.HSI != nil && i.rimpBlade.HSI.NICS != nil {
		for _, nic := range i.rimpBlade.HSI.NICS {
			var name string
			if strings.HasPrefix(nic.Description, "iLO") {
				name = "bmc"
			} else {
				name = nic.Description
			}

			if nics == nil {
				nics = make([]*devices.Nic, 0)
			}

			n := &devices.Nic{
				Name:       name,
				MacAddress: strings.ToLower(nic.MacAddr),
			}
			nics = append(nics, n)
		}
	}
	return nics, err
}

// License returns the iLO's license information
func (i *Ilo) License() (name string, licType string, err error) {
	err = i.httpLogin()
	if err != nil {
		return name, licType, err
	}

	url := "json/license"
	payload, err := i.get(url)
	if err != nil {
		return name, licType, err
	}

	hpIloLicense := &hp.IloLicense{}
	err = json.Unmarshal(payload, hpIloLicense)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return name, licType, err
	}

	return hpIloLicense.Name, hpIloLicense.Type, err
}

// Psus returns a list of psus installed on the device
func (i *Ilo) Psus() (psus []*devices.Psu, err error) {
	err = i.httpLogin()
	if err != nil {
		return psus, err
	}

	url := "json/power_supplies"
	payload, err := i.get(url)
	if err != nil {
		return psus, err
	}

	hpIloPowerSupply := &hp.IloPowerSupply{}
	err = json.Unmarshal(payload, hpIloPowerSupply)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return psus, err
	}

	for _, psu := range hpIloPowerSupply.Supplies {
		if psus == nil {
			psus = make([]*devices.Psu, 0)
		}
		var status string
		if psu.PsCondition == "PS_OK" {
			status = "OK"
		} else {
			status = psu.PsCondition
		}

		p := &devices.Psu{
			Serial:     strings.ToLower(psu.PsSerialNum),
			Status:     status,
			PowerKw:    float64(psu.PsOutputWatts) / 1000.00,
			CapacityKw: float64(psu.PsMaxCapWatts) / 1000.00,
		}

		psus = append(psus, p)
	}

	return psus, err
}

// Disks returns a list of disks installed on the device
func (i *Ilo) Disks() (disks []*devices.Disk, err error) {
	err = i.httpLogin()
	if err != nil {
		return disks, err
	}

	url := "json/health_phy_drives"
	payload, err := i.get(url)
	if err != nil {
		return disks, err
	}

	hpIloDisks := &hp.IloDisks{}
	err = json.Unmarshal(payload, hpIloDisks)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return disks, err
	}

	for _, disksArray := range hpIloDisks.PhyDriveArrays {
		for _, physicalDrive := range disksArray.PhysicalDrives {
			if disks == nil {
				disks = make([]*devices.Disk, 0)
			}
			var status string
			if physicalDrive.Status == "OP_STATUS_OK" {
				status = "OK"
			} else {
				status = physicalDrive.Status
			}

			var diskType string
			if strings.Contains(physicalDrive.DriveMediatype, "HDD") {
				diskType = "HDD"
			} else if strings.Contains(physicalDrive.DriveMediatype, "SSD") {
				diskType = "SSD"
			} else {
				diskType = physicalDrive.DriveMediatype
			}

			disk := &devices.Disk{
				Serial:    strings.ToLower(physicalDrive.SerialNo),
				Status:    status,
				Model:     strings.ToLower(physicalDrive.Model),
				Size:      physicalDrive.Capacity,
				Location:  physicalDrive.Location,
				Type:      diskType,
				FwVersion: strings.ToLower(physicalDrive.FwVersion),
			}

			disks = append(disks, disk)
		}
	}

	return disks, err
}

<<<<<<< HEAD
=======
// Logout logs out and close the iLO connection
func (i *Ilo) Logout() (err error) {
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": hp.VendorID, "ip": i.ip}).Debug("logout from bmc")

	data := []byte(`{"method":"logout"}`)

	req, err := http.NewRequest("POST", i.loginURL.String(), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.client.Do(req)
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return err
}

>>>>>>> master
// IsBlade returns if the current hardware is a blade or not
func (i *Ilo) IsBlade() (isBlade bool, err error) {
	if i.rimpBlade.BladeSystem != nil {
		isBlade = true
	} else {
		isBlade = false
	}

	return isBlade, err
}

// Vendor returns bmc's vendor
func (i *Ilo) Vendor() (vendor string) {
	return hp.VendorID
}

// ServerSnapshot do best effort to populate the server data and returns a blade or discrete
func (i *Ilo) ServerSnapshot() (server interface{}, err error) {
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
func (i *Ilo) UpdateCredentials(username string, password string) {
	i.username = username
	i.password = password
}
