package idrac8

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"
	"github.com/go-logr/logr"
)

const (
	// BMCType defines the bmc model that is supported by this package
	BMCType = "idrac8"
)

// IDrac8 holds the status and properties of a connection to an iDrac device
type IDrac8 struct {
	ip                   string
	username             string
	password             string
	httpClient           *http.Client
	sshClient            *sshclient.SSHClient
	st1                  string
	st2                  string
	iDracInventory       *dell.IDracInventory
	ctx                  context.Context
	log                  logr.Logger
	httpClientSetupFuncs []func(*http.Client)
}

// IDrac8Option is a type that can configure an *IDrac8
type IDrac8Option func(*IDrac8)

// WithSecureTLS enforces trusted TLS connections, with an optional CA certificate pool.
// Using this option with an nil pool uses the system CAs.
func WithSecureTLS(rootCAs *x509.CertPool) IDrac8Option {
	return func(i *IDrac8) {
		i.httpClientSetupFuncs = append(i.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

// New returns a new IDrac8 ready to be used
func New(ctx context.Context, host string, username string, password string, log logr.Logger) (*IDrac8, error) {
	return NewWithOptions(ctx, host, username, password, log)
}

// NewWithOptions returns a new IDrac8 with options ready to be used
func NewWithOptions(ctx context.Context, host string, username string, password string, log logr.Logger, opts ...IDrac8Option) (*IDrac8, error) {
	sshClient, err := sshclient.New(host, username, password)
	if err != nil {
		return nil, err
	}

	i := &IDrac8{ip: host, username: username, password: password, sshClient: sshClient, ctx: ctx, log: log}

	for _, opt := range opts {
		opt(i)
	}
	return i, nil
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
		return 0, nil, err
	}
	req.Header.Add("ST2", i.st2)

	u, err := url.Parse(bmcURL)
	if err != nil {
		return 0, nil, err
	}

	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "-http-session-" || cookie.Name == "tokenvalue" {
			req.AddCookie(cookie)
		}
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

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", fmt.Sprintf("https://%s/%s", i.ip, endpoint))

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

func (i *IDrac8) get(endpoint string, extraHeaders *map[string]string) (statusCode int, payload []byte, err error) {
	i.log.V(1).Info("retrieving data from bmc", "step", "bmc connection", "vendor", dell.VendorID, "ip", i.ip, "endpoint", endpoint)

	bmcURL := fmt.Sprintf("https://%s", i.ip)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Add("ST2", i.st2)
	if extraHeaders != nil {
		for key, value := range *extraHeaders {
			req.Header.Add(key, value)
		}
	}

	u, err := url.Parse(bmcURL)
	if err != nil {
		return 0, nil, err
	}

	for _, cookie := range i.httpClient.Jar.Cookies(u) {
		if cookie.Name == "-http-session-" {
			req.AddCookie(cookie)
		}
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

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	if resp.StatusCode == 404 {
		return 404, payload, errors.ErrPageNotFound
	}

	return resp.StatusCode, payload, err
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
			var macAddress string
			var name string
			for _, property := range component.Properties {
				if property.Name == "LinkSpeed" && property.Type == "uint8" && property.DisplayValue != "Unknown" {
					speed = property.DisplayValue
					up = true
				} else if property.Name == "PermanentMACAddress" && property.Type == "string" {
					macAddress = strings.ToLower(property.Value)
				} else if property.Name == "ProductName" && property.Type == "string" {
					name = strings.Split(property.Value, " - ")[0]
				}
			}

			if macAddress != "" {
				if nics == nil {
					nics = make([]*devices.Nic, 0)
				}
				n := &devices.Nic{
					Name:       name,
					Speed:      speed,
					Up:         up,
					MacAddress: macAddress,
				}
				nics = append(nics, n)
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
		return "", err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "NodeID" && property.Type == "string" {
					return strings.ToLower(property.Value), nil
				}
			}
		}
	}

	return "", fmt.Errorf("IDrac8 Serial(): Serial not found!")
}

// ChassisSerial returns the serial number of the chassis where the blade is attached
func (i *IDrac8) ChassisSerial() (serial string, err error) {
	err = i.loadHwData()
	if err != nil {
		return serial, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "ChassisServiceTag" && property.Type == "string" {
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
		return "", err
	}

	extraHeaders := &map[string]string{
		"X-SYSMGMT-OPTIMIZE": "true",
	}

	endpoint := "sysmgmt/2016/server/extended_health"
	statusCode, response, err := i.get(endpoint, extraHeaders)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", err
	}

	iDracHealthStatus := &dell.IDracHealthStatus{}
	err = json.Unmarshal(response, iDracHealthStatus)
	if err != nil {
		return "", err
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

	endpoint := "data?get=powermonitordata"
	statusCode, response, err := i.get(endpoint, nil)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return power, err
	}

	iDracRoot := &dell.IDracRoot{}
	err = xml.Unmarshal(response, iDracRoot)
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

// Version returns the version of the bmc we are running
func (i *IDrac8) Version() (bmcVersion string, err error) {
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

// Slot returns the current slot within the chassis
func (i *IDrac8) Slot() (slot int, err error) {
	err = i.loadHwData()
	if err != nil {
		return -1, err
	}

	for _, component := range i.iDracInventory.Component {
		if component.Classname == "DCIM_SystemView" {
			for _, property := range component.Properties {
				if property.Name == "BaseBoardChassisSlot" && property.Type == "string" {
					if property.Value == "NA" {
						return -1, err
					}
					v := strings.Split(property.Value, " ")
					if len(v) < 2 {
						return -1, fmt.Errorf("Looks like the BaseBoardChassisSlot is ill-formatted!")
					}
					slot, err = strconv.Atoi(v[1])
					if err != nil {
						return -1, err
					}

					return slot, err
				}
			}
		}
	}

	return -1, err
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

	return "", fmt.Errorf("IDrac8 Model(): Model not found!")
}

// HardwareType returns the type of bmc we are talking to
func (i *IDrac8) HardwareType() (bmcType string) {
	return BMCType
}

// License returns the bmc license information
func (i *IDrac8) License() (name string, licType string, err error) {
	err = i.httpLogin()
	if err != nil {
		return "", "", err
	}

	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	endpoint := "sysmgmt/2012/server/license"
	statusCode, response, err := i.get(endpoint, extraHeaders)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", "", err
	}

	iDracLicense := &dell.IDracLicense{}
	err = json.Unmarshal(response, iDracLicense)
	if err != nil {
		return "", "", err
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
		return 0, err
	}

	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	endpoint := "sysmgmt/2012/server/temperature"
	statusCode, response, err := i.get(endpoint, extraHeaders)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return 0, err
	}

	iDracTemp := &dell.IDracTemp{}
	err = json.Unmarshal(response, iDracTemp)
	if err != nil {
		return 0, err
	}

	return iDracTemp.Temperatures.IDRACEmbedded1SystemBoardInletTemp.Reading, err
}

// CPU return the cpu, cores and hyperthreads the server
func (i *IDrac8) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	err = i.httpLogin()
	if err != nil {
		return "", 0, 0, 0, err
	}

	extraHeaders := &map[string]string{
		"X_SYSMGMT_OPTIMIZE": "true",
	}

	endpoint := "sysmgmt/2012/server/processor"
	statusCode, response, err := i.get(endpoint, extraHeaders)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return "", 0, 0, 0, err
	}

	dellBladeProc := &dell.BladeProcessorEndpoint{}
	err = json.Unmarshal(response, dellBladeProc)
	if err != nil {
		return "", 0, 0, 0, err
	}

	for _, proc := range dellBladeProc.Proccessors {
		hasHT := 0
		for _, ht := range proc.HyperThreading {
			if ht.Capable == 1 {
				hasHT = 2
			}
		}
		return httpclient.StandardizeProcessorName(proc.Brand), len(dellBladeProc.Proccessors), proc.CoreCount, proc.CoreCount * hasHT, nil
	}

	return "", 0, 0, 0, fmt.Errorf("IDRAC8 CPU(): No CPUs?!")
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
		return nil, err
	}

	endpoint := "data?get=powerSupplies"
	statusCode, response, err := i.get(endpoint, nil)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return nil, err
	}

	iDracRoot := &dell.IDracRoot{}
	err = xml.Unmarshal(response, iDracRoot)
	if err != nil {
		return nil, err
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

	return psus, nil
}

// Vendor returns bmc's vendor
func (i *IDrac8) Vendor() (vendor string) {
	return dell.VendorID
}

// ServerSnapshot do best effort to populate the server data and returns a blade or discrete
func (i *IDrac8) ServerSnapshot() (server interface{}, err error) { // nolint: gocyclo
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

// BiosVersion returns the BIOS version from the BMC, implements the Firmware interface
func (i *IDrac8) GetBIOSVersion(ctx context.Context) (string, error) {
	return "", errors.ErrNotImplemented
}

// BMCVersion returns the BMC version, implements the Firmware interface
func (i *IDrac8) GetBMCVersion(ctx context.Context) (string, error) {
	return "", errors.ErrNotImplemented
}

// Updates the BMC firmware, implements the Firmware interface
func (i *IDrac8) FirmwareUpdateBMC(ctx context.Context, filePath string) error {
	return errors.ErrNotImplemented
}
