package ilo

import (
	"bytes"
	"context"
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
	"github.com/go-logr/logr"
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
	loginURL   *url.URL
	rimpBlade  *hp.RimpBlade
	ctx        context.Context
	log        logr.Logger
}

// New returns a new Ilo ready to be used
func New(ctx context.Context, host string, username string, password string, log logr.Logger) (*Ilo, error) {
	loginURL, err := url.Parse(fmt.Sprintf("https://%s/json/login_session", host))
	if err != nil {
		return nil, err
	}

	client, err := httpclient.Build()
	if err != nil {
		return nil, err
	}

	xmlURL := fmt.Sprintf("https://%s/xmldata?item=all", host)
	resp, err := client.Get(xmlURL)
	if err != nil {
		return nil, err
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rimpBlade := &hp.RimpBlade{}
	err = xml.Unmarshal(payload, rimpBlade)
	if err != nil {
		return nil, err
	}

	sshClient, err := sshclient.New(host, username, password)
	if err != nil {
		return nil, err
	}

	ilo := &Ilo{
		ip:        host,
		username:  username,
		password:  password,
		loginURL:  loginURL,
		rimpBlade: rimpBlade,
		sshClient: sshClient,
		ctx:       ctx,
		log:       log,
	}
	return ilo, nil
}

// CheckCredentials verify whether the credentials are valid or not
func (i *Ilo) CheckCredentials() (err error) {
	err = i.httpLogin()
	if err != nil {
		return err
	}
	return err
}

// Calls a given JSON ILO endpoint and returns the status code and the data.
func (i *Ilo) get(endpoint string, useSession bool) (int, []byte, error) {
	i.log.V(1).Info("Retrieving data from ILO...",
		"step", "bmc connection",
		"vendor", hp.VendorID,
		"ip", i.ip,
		"endpoint", endpoint,
	)

	bmcURL := fmt.Sprintf("https://%s", i.ip)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return 0, nil, err
	}

	u, err := url.Parse(bmcURL)
	if err != nil {
		return 0, nil, err
	}

	if useSession {
		for _, cookie := range i.httpClient.Jar.Cookies(u) {
			if cookie.Name == "sessionKey" {
				req.AddCookie(cookie)
			}
		}
	} else {
		req.SetBasicAuth(i.username, i.password)
	}

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", fmt.Sprintf("%s/%s", bmcURL, endpoint))

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respDump, _ := httputil.DumpResponse(resp, true)
	i.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	if resp.StatusCode == 404 {
		return 404, nil, errors.ErrPageNotFound
	}

	return resp.StatusCode, payload, nil
}

// posts the payload to the given endpoint
func (i *Ilo) post(endpoint string, data []byte) (statusCode int, body []byte, err error) {
	u, err := url.Parse(fmt.Sprintf("https://%s/%s", i.ip, endpoint))
	if err != nil {
		return 0, []byte{}, err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(data))
	if err != nil {
		return 0, []byte{}, err
	}

	req.Header.Add("Content-Type", "application/json")
	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "sessionKey" {
			req.AddCookie(cookie)
		}
	}

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", fmt.Sprintf("%s/%s", i.ip, endpoint))

	resp, err := i.httpClient.Do(req)
	if err != nil {
		return 0, []byte{}, err
	}

	defer resp.Body.Close()

	respDump, _ := httputil.DumpResponse(resp, true)
	i.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, []byte{}, err
	}

	return resp.StatusCode, body, err
}

// Serial returns the device serial
func (i *Ilo) Serial() (serial string, err error) {
	return strings.ToLower(strings.TrimSpace(i.rimpBlade.HSI.Sbsn)), err
}

// Returns the serial number of the chassis where the blade is attached.
func (i *Ilo) ChassisSerial() (string, error) {
	err := i.httpLogin()
	if err != nil {
		return "", err
	}

	endpoint := "json/rck_info"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", err
	}

	rckInfo := &hp.RckInfo{}
	err = json.Unmarshal(payload, rckInfo)
	if err != nil {
		return "", err
	}

	if rckInfo.EncSn == "Unknown" {
		chassisInfo, err := i.parseChassisInfo()
		if err != nil {
			return "", err
		}

		return strings.ToLower(chassisInfo.ChassisSn), nil
	}

	return strings.ToLower(rckInfo.EncSn), err
}

// Model returns the device model
func (i *Ilo) Model() (model string, err error) {
	return i.rimpBlade.HSI.Spn, err
}

// HardwareType returns the type of bmc we are talking to
func (i *Ilo) HardwareType() (bmcType string) {
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

// Version returns the version of the bmc we are running
func (i *Ilo) Version() (bmcVersion string, err error) {
	return i.rimpBlade.MP.Fwri, err
}

// Returns the name of this server from the ILO point of view.
func (i *Ilo) Name() (name string, err error) {
	err = i.httpLogin()
	if err != nil {
		return name, err
	}

	endpoint := "json/overview"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}
		return "", err
	}

	overview := &hp.Overview{}
	err = json.Unmarshal(payload, overview)
	if err != nil {
		return "", err
	}

	return overview.ServerName, err
}

// Returns the health status from the ILO point of view.
func (i *Ilo) Status() (health string, err error) {
	err = i.httpLogin()
	if err != nil {
		return health, err
	}

	endpoint := "json/overview"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}
		return "", err
	}

	overview := &hp.Overview{}
	err = json.Unmarshal(payload, overview)
	if err != nil {
		return health, err
	}

	if overview.SystemHealth == "OP_STATUS_OK" {
		return "OK", err
	}

	return overview.SystemHealth, err
}

// Returns the total amount of memory of the server.
func (i *Ilo) Memory() (mem int, err error) {
	err = i.httpLogin()
	if err != nil {
		return 0, err
	}

	endpoint := "json/mem_info"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return 0, err
	}

	hpMemData := &hp.Mem{}
	err = json.Unmarshal(payload, hpMemData)
	if err != nil {
		return 0, err
	}

	if hpMemData.MemTotalMemSize != 0 {
		return hpMemData.MemTotalMemSize / 1024, nil
	}

	for _, slot := range hpMemData.Memory {
		mem = mem + slot.MemSize
	}

	return mem / 1024, nil
}

// Finds the CPUs.
// Returns the description, cores count, and hyperthreads count of the first CPU it finds.
// Returns also the CPU count.
// TODO: Does this make any sense?! We either return all the information about all CPUs, or just say something generic!
func (i *Ilo) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	err = i.httpLogin()
	if err != nil {
		return "", 0, 0, 0, err
	}

	endpoint := "json/proc_info"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", 0, 0, 0, err
	}

	hpProcData := &hp.Procs{}
	err = json.Unmarshal(payload, hpProcData)
	if err != nil {
		return "", 0, 0, 0, err
	}

	for _, proc := range hpProcData.Processors {
		return httpclient.StandardizeProcessorName(proc.ProcName), len(hpProcData.Processors), proc.ProcNumCores, proc.ProcNumThreads, nil
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// Returns the current version of the BIOS.
func (i *Ilo) BiosVersion() (version string, err error) {
	err = i.httpLogin()
	if err != nil {
		return "", err
	}

	endpoint := "json/overview"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", err
	}

	overview := &hp.Overview{}
	err = json.Unmarshal(payload, overview)
	if err != nil {
		return "", err
	}

	if overview.SystemRom != "" {
		return overview.SystemRom, nil
	}

	return "", errors.ErrBiosNotFound
}

// PowerKw returns the current power usage in Kw
func (i *Ilo) PowerKw() (power float64, err error) {
	err = i.httpLogin()
	if err != nil {
		return 0, err
	}

	endpoint := "json/power_summary"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return 0, err
	}

	hpPowerSummary := &hp.PowerSummary{}
	err = json.Unmarshal(payload, hpPowerSummary)
	if err != nil {
		return 0, err
	}

	return float64(hpPowerSummary.PowerSupplyInputPower) / 1000, nil
}

// PowerState returns the current power state of the machine
func (i *Ilo) PowerState() (state string, err error) {
	err = i.httpLogin()
	if err != nil {
		return "", err
	}

	endpoint := "json/power_summary"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}
		return "", err
	}

	hpPowerSummary := &hp.PowerSummary{}
	err = json.Unmarshal(payload, hpPowerSummary)
	if err != nil {
		return "", err
	}

	return strings.ToLower(hpPowerSummary.HostpwrState), nil
}

// Returns the current temperature of the server.
func (i *Ilo) TempC() (temp int, err error) {
	err = i.httpLogin()
	if err != nil {
		return 0, err
	}

	endpoint := "json/health_temperature"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}
		return 0, err
	}

	hpHealthTemperature := &hp.HealthTemperature{}
	err = json.Unmarshal(payload, hpHealthTemperature)
	if err != nil {
		return 0, err
	}

	for _, item := range hpHealthTemperature.Temperature {
		if item.Location == "Ambient" {
			return item.Currentreading, nil
		}
	}

	return 0, errors.ErrFeatureUnavailable
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

// Returns the ILO's license information.
func (i *Ilo) License() (name string, licType string, err error) {
	err = i.httpLogin()
	if err != nil {
		return "", "", err
	}

	endpoint := "json/license"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", "", err
	}

	hpIloLicense := &hp.IloLicense{}
	err = json.Unmarshal(payload, hpIloLicense)
	if err != nil {
		return "", "", err
	}

	return hpIloLicense.Name, hpIloLicense.Type, nil
}

func (i *Ilo) parseChassisInfo() (*hp.ChassisInfo, error) {
	err := i.httpLogin()
	if err != nil {
		return nil, err
	}

	chassisInfo := &hp.ChassisInfo{}
	// We try the new way of doing things first (RedFish).
	statusCode, payload, err := i.get(hp.ChassisInfoNewURL, false)
	if err == nil && statusCode == 200 {
		err = json.Unmarshal(payload, chassisInfo)
		if err != nil {
			return nil, err
		}
		if chassisInfo.Error.Code != "" {
			e := "Code: " + chassisInfo.Error.Code + ", Message: " + chassisInfo.Error.Message
			for i, s := range chassisInfo.Error.ExtendedMessage {
				e += fmt.Sprintf(", Extended[%d]: %s", i, s)
			}
			return nil, fmt.Errorf(e)
		}

		if chassisInfo.Links.ContainedBy.ID == "/"+hp.ChassisInfoChassisURL {
			chassisInfo.ChassisType = "Blade"
			statusCode, payload, err = i.get(hp.ChassisInfoChassisURL, false)
			if err != nil || statusCode != 200 {
				if err == nil {
					err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, hp.ChassisInfoChassisURL)
				}

				return nil, err
			}

			chassisExtendedInfo := &hp.ChassisInfo{}
			err = json.Unmarshal(payload, chassisExtendedInfo)
			if err != nil {
				return nil, err
			}

			chassisInfo.ChassisSn = chassisExtendedInfo.SerialNumber
		} else {
			chassisInfo.ChassisSn = chassisInfo.SerialNumber
		}

		// Matching the new interface to the old one, since the code still drops
		//   off to the old interface in case the new interface is not available.
		chassisInfo.NodeNumber = chassisInfo.Oem.Hpe.BayNumber

		return chassisInfo, nil
	}

	if err != errors.ErrPageNotFound {
		// This is a real error, just give up...
		return nil, err
	}

	// This just means that we have to try the old way of doing things, since RedFish is not available.
	statusCode, payload, err = i.get(hp.ChassisInfoOldURL, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, hp.ChassisInfoOldURL)
		}

		return nil, err
	}

	err = json.Unmarshal(payload, chassisInfo)
	if err != nil {
		return nil, err
	}

	return chassisInfo, nil
}

// Psus returns a list of psus installed on the device
func (i *Ilo) Psus() (psus []*devices.Psu, err error) {
	err = i.httpLogin()
	if err != nil {
		return psus, err
	}

	endpoint := "json/power_supplies"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return psus, err
	}

	hpIloPowerSupply := &hp.IloPowerSupply{}
	err = json.Unmarshal(payload, hpIloPowerSupply)
	if err != nil {
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

	endpoint := "json/health_phy_drives"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return disks, err
	}

	hpIloDisks := &hp.IloDisks{}
	err = json.Unmarshal(payload, hpIloDisks)
	if err != nil {
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

// Returns whether the current hardware is a blade.
func (i *Ilo) IsBlade() (isBlade bool, err error) {
	if i.rimpBlade.BladeSystem != nil {
		return true, nil
	}

	chassisInfo, err := i.parseChassisInfo()
	if err != nil {
		return false, err
	}

	return chassisInfo.ChassisType == "Blade", nil
}

// Slot returns the current slot within the chassis
func (i *Ilo) Slot() (slot int, err error) {
	if i.rimpBlade.BladeSystem != nil {
		return i.rimpBlade.BladeSystem.Bay, nil
	}

	chassisInfo, err := i.parseChassisInfo()
	if err != nil {
		return -1, err
	}

	if chassisInfo.NodeNumber != 0 {
		return chassisInfo.NodeNumber, nil
	}

	return -1, nil
}

// Vendor returns bmc's vendor
func (i *Ilo) Vendor() (vendor string) {
	return hp.VendorID
}

// ServerSnapshot do best effort to populate the server data and returns a blade or discrete
func (i *Ilo) ServerSnapshot() (server interface{}, err error) { // nolint: gocyclo
	err = i.httpLogin()
	if err != nil {
		return server, err
	}

	if isBlade, _ := i.IsBlade(); isBlade {
		blade := &devices.Blade{}
		blade.Vendor = i.Vendor()
		blade.BmcAddress = i.ip
		blade.BmcType = i.HardwareType()

		blade.Serial, err = i.Serial()
		if err != nil {
			return nil, err
		}
		blade.BmcVersion, err = i.Version()
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
		blade.BladePosition, err = i.Slot()
		if err != nil {
			return nil, err
		}
		blade.ChassisSerial, err = i.ChassisSerial()
		if err != nil {
			return nil, err
		}
		server = blade
	} else {
		discrete := &devices.Discrete{}
		discrete.Vendor = i.Vendor()
		discrete.BmcAddress = i.ip
		discrete.BmcType = i.HardwareType()

		discrete.Serial, err = i.Serial()
		if err != nil {
			return nil, err
		}
		discrete.BmcVersion, err = i.Version()
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
		discrete.PowerState, err = i.PowerState()
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

// BiosVersion returns the BIOS version from the BMC, implements the Firmware interface
func (i *Ilo) GetBIOSVersion(ctx context.Context) (string, error) {
	return "", errors.ErrNotImplemented
}

// BMCVersion returns the BMC version, implements the Firmware interface
func (i *Ilo) GetBMCVersion(ctx context.Context) (string, error) {
	return "", errors.ErrNotImplemented
}

// Updates the BMC firmware, implements the Firmware interface
func (i *Ilo) FirmwareUpdateBMC(ctx context.Context, filePath string) error {
	return errors.ErrNotImplemented
}
