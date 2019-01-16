package supermicrox10

import (
	"bytes"
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

	"github.com/bmc-toolbox/bmclib/providers/supermicro"

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

const (
	// BmcType defines the bmc model that is supported by this package
	BmcType = "supermicrox10"
)

// SupermicroX10 holds the status and properties of a connection to a supermicro bmc
type SupermicroX10 struct {
	ip         string
	username   string
	password   string
	httpClient *http.Client
}

// New returns a new SupermicroX10 instance ready to be used
func New(ip string, username string, password string) (sm *SupermicroX10, err error) {
	return &SupermicroX10{ip: ip, username: username, password: password}, err
}

// CheckCredentials verify whether the credentials are valid or not
func (s *SupermicroX10) CheckCredentials() (err error) {
	err = s.httpLogin()
	if err != nil {
		return err
	}
	return err
}

// get calls a given json endpoint of the ilo and returns the data
func (s *SupermicroX10) get(endpoint string) (payload []byte, err error) {

	bmcURL := fmt.Sprintf("https://%s", s.ip)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", bmcURL, endpoint), nil)
	if err != nil {
		return payload, err
	}

	u, err := url.Parse(bmcURL)
	if err != nil {
		return payload, err
	}

	for _, cookie := range s.httpClient.Jar.Cookies(u) {
		if cookie.Name == "SID" && cookie.Value != "" {
			req.AddCookie(cookie)
		}
	}

	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] https://%s/%s", bmcURL, endpoint))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	resp, err := s.httpClient.Do(req)
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

func (s *SupermicroX10) query(requestType string) (ipmi *supermicro.IPMI, err error) {
	err = s.httpLogin()
	if err != nil {
		return ipmi, err
	}

	bmcURL := fmt.Sprintf("https://%s/cgi/ipmi.cgi", s.ip)
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": supermicro.VendorID, "ip": s.ip}).Debug("retrieving data from bmc")

	req, err := http.NewRequest("POST", bmcURL, bytes.NewBufferString(requestType))
	if err != nil {
		return ipmi, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	u, err := url.Parse(bmcURL)
	if err != nil {
		return ipmi, err
	}
	for _, cookie := range s.httpClient.Jar.Cookies(u) {
		if cookie.Name == "SID" && cookie.Value != "" {
			req.AddCookie(cookie)
		}
	}
	if log.GetLevel() == log.DebugLevel {
		log.Println(fmt.Sprintf("https://%s/cgi/%s", bmcURL, s.ip))
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println("[Request]")
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return ipmi, err
	}
	defer resp.Body.Close()

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ipmi, err
	}
	defer resp.Body.Close()
	if log.GetLevel() == log.DebugLevel {
		log.Println(fmt.Sprintf("https://%s/cgi/%s", bmcURL, s.ip))
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println("[Request]")
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	ipmi = &supermicro.IPMI{}
	err = xml.Unmarshal(payload, ipmi)
	if err != nil {
		httpclient.DumpInvalidPayload(requestType, s.ip, payload)
		return ipmi, err
	}

	return ipmi, err
}

// Serial returns the device serial
func (s *SupermicroX10) Serial() (serial string, err error) {
	ipmi, err := s.query("FRU_INFO.XML=(0,0)")
	if err != nil {
		return serial, err
	}

	if ipmi.FruInfo == nil || ipmi.FruInfo.Chassis == nil {
		return serial, errors.ErrInvalidSerial
	}

	if strings.HasPrefix(ipmi.FruInfo.Chassis.SerialNum, "S") {
		serial = strings.TrimSpace(fmt.Sprintf("%s_%s", strings.TrimSpace(ipmi.FruInfo.Chassis.SerialNum), strings.TrimSpace(ipmi.FruInfo.Board.SerialNum)))
	} else {
		serial = strings.TrimSpace(fmt.Sprintf("%s_%s", strings.TrimSpace(ipmi.FruInfo.Product.SerialNum), strings.TrimSpace(ipmi.FruInfo.Board.SerialNum)))
	}

	return strings.ToLower(serial), err
}

// BmcType returns just Model id string - supermicrox10
func (s *SupermicroX10) BmcType() (model string) {
	return BmcType
}

// Model returns the device model
func (s *SupermicroX10) Model() (model string, err error) {
	ipmi, err := s.query("FRU_INFO.XML=(0,0)")
	if err != nil {
		return model, err
	}

	if ipmi.FruInfo != nil && ipmi.FruInfo.Board != nil {
		return ipmi.FruInfo.Board.ProdName, err
	}

	return model, err
}

// BmcVersion returns the version of the bmc we are running
func (s *SupermicroX10) BmcVersion() (bmcVersion string, err error) {
	ipmi, err := s.query("GENERIC_INFO.XML=(0,0)")
	if err != nil {
		return bmcVersion, err
	}

	if ipmi.GenericInfo != nil && ipmi.GenericInfo.Generic != nil {
		return ipmi.GenericInfo.Generic.IpmiFwVersion, err
	}

	return bmcVersion, err
}

// Name returns the hostname of the machine
func (s *SupermicroX10) Name() (name string, err error) {
	ipmi, err := s.query("CONFIG_INFO.XML=(0,0)")
	if err != nil {
		return name, err
	}

	if ipmi.ConfigInfo != nil && ipmi.ConfigInfo.Hostname != nil {
		return ipmi.ConfigInfo.Hostname.Name, err
	}

	return name, err
}

// Status returns health string status from the bmc
func (s *SupermicroX10) Status() (health string, err error) {
	return "NotSupported", err
}

// Memory returns the total amount of memory of the server
func (s *SupermicroX10) Memory() (mem int, err error) {
	ipmi, err := s.query("SMBIOS_INFO.XML=(0,0)")

	for _, dimm := range ipmi.Dimm {
		dimm := strings.TrimSuffix(dimm.Size, " MB")
		size, err := strconv.Atoi(dimm)
		if err != nil {
			return mem, err
		}
		mem += size
	}

	return mem / 1024, err
}

// CPU returns the cpu, cores and hyperthreads of the server
func (s *SupermicroX10) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	ipmi, err := s.query("SMBIOS_INFO.XML=(0,0)")
	for _, entry := range ipmi.CPU {
		cpu = httpclient.StandardizeProcessorName(entry.Version)
		cpuCount = len(ipmi.CPU)

		coreCount, err = strconv.Atoi(entry.Core)
		if err != nil {
			return cpu, cpuCount, coreCount, hyperthreadCount, err
		}

		hyperthreadCount = coreCount
		break
	}

	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

// BiosVersion returns the current version of the bios
func (s *SupermicroX10) BiosVersion() (version string, err error) {
	ipmi, err := s.query("SMBIOS_INFO.XML=(0,0)")
	if err != nil {
		return version, err
	}

	if ipmi.Bios != nil {
		return ipmi.Bios.Version, err
	}

	return version, err
}

// PowerKw returns the current power usage in Kw
func (s *SupermicroX10) PowerKw() (power float64, err error) {
	ipmi, err := s.query("Get_NodeInfoReadings.XML=(0,0)")
	if err != nil {
		return power, err
	}

	if ipmi.NodeInfo != nil {
		for _, node := range ipmi.NodeInfo.Nodes {
			if node.IP == strings.Split(s.ip, ":")[0] {
				value, err := strconv.Atoi(node.Power)
				if err != nil {
					return power, err
				}

				return float64(value) / 1000.00, err
			}
		}
	}

	return power, err
}

// PowerState returns the current power state of the machine
func (s *SupermicroX10) PowerState() (state string, err error) {
	ipmi, err := s.query("POWER_INFO.XML=(0,0)")
	if err != nil {
		return state, err
	}

	if ipmi.PowerInfo != nil {
		return strings.ToLower(ipmi.PowerInfo.Power.Status), err
	}

	return "unknow", err
}

// TempC returns the current temperature of the machine
func (s *SupermicroX10) TempC() (temp int, err error) {
	ipmi, err := s.query("Get_NodeInfoReadings.XML=(0,0)")
	if err != nil {
		return temp, err
	}

	if ipmi.NodeInfo != nil {
		for _, node := range ipmi.NodeInfo.Nodes {
			if node.IP == strings.Split(s.ip, ":")[0] {
				temp, err := strconv.Atoi(node.SystemTemp)
				if err != nil {
					return temp, err
				}

				return temp, err
			}
		}
	}

	return temp, err
}

// IsBlade returns if the current hardware is a blade or not
func (s *SupermicroX10) IsBlade() (isBlade bool, err error) {
	return false, err
}

// Nics returns all found Nics in the device
func (s *SupermicroX10) Nics() (nics []*devices.Nic, err error) {
	ipmi, err := s.query("GENERIC_INFO.XML=(0,0)")
	if err != nil {
		return nics, err
	}

	bmcNic := &devices.Nic{
		Name:       "bmc",
		MacAddress: ipmi.GenericInfo.Generic.BmcMac,
	}

	nics = append(nics, bmcNic)

	ipmi, err = s.query("Get_PlatformInfo.XML=(0,0)")
	if err != nil {
		return nics, err
	}

	// TODO: (ncode) This needs to become dinamic somehow
	if ipmi.PlatformInfo != nil {
		if ipmi.PlatformInfo.MbMacAddr1 != "" {
			bmcNic := &devices.Nic{
				Name:       "eth0",
				MacAddress: ipmi.PlatformInfo.MbMacAddr1,
			}
			nics = append(nics, bmcNic)
		}

		if ipmi.PlatformInfo.MbMacAddr2 != "" {
			bmcNic := &devices.Nic{
				Name:       "eth1",
				MacAddress: ipmi.PlatformInfo.MbMacAddr2,
			}
			nics = append(nics, bmcNic)
		}

		if ipmi.PlatformInfo.MbMacAddr3 != "" {
			bmcNic := &devices.Nic{
				Name:       "eth2",
				MacAddress: ipmi.PlatformInfo.MbMacAddr3,
			}
			nics = append(nics, bmcNic)
		}

		if ipmi.PlatformInfo.MbMacAddr4 != "" {
			bmcNic := &devices.Nic{
				Name:       "eth3",
				MacAddress: ipmi.PlatformInfo.MbMacAddr4,
			}
			nics = append(nics, bmcNic)
		}
	}

	return nics, err
}

// License returns the iLO's license information
func (s *SupermicroX10) License() (name string, licType string, err error) {
	ipmi, err := s.query("BIOS_LINCENSE_ACTIVATE.XML=(0,0)")
	if err != nil {
		return name, licType, err
	}

	if ipmi.BiosLicense != nil {
		switch ipmi.BiosLicense.Check {
		case "0":
			return "oob", "Activated", err
		case "1":
			return "oob", "Not Activated", err
		}
	}

	return name, licType, err
}

// Vendor returns bmc's vendor
func (s *SupermicroX10) Vendor() (vendor string) {
	return supermicro.VendorID
}

// ServerSnapshot do best effort to populate the server data and returns a blade or discrete
func (s *SupermicroX10) ServerSnapshot() (server interface{}, err error) {
	if isBlade, _ := s.IsBlade(); isBlade {
		blade := &devices.Blade{}
		blade.Vendor = s.Vendor()
		blade.BmcAddress = s.ip
		blade.BmcType = s.BmcType()

		blade.Serial, _ = s.Serial()
		if err != nil {
			return nil, err
		}
		blade.BmcVersion, err = s.BmcVersion()
		if err != nil {
			return nil, err
		}
		blade.Model, err = s.Model()
		if err != nil {
			return nil, err
		}
		blade.Nics, err = s.Nics()
		if err != nil {
			return nil, err
		}
		blade.Disks, err = s.Disks()
		if err != nil {
			return nil, err
		}
		blade.BiosVersion, err = s.BiosVersion()
		if err != nil {
			return nil, err
		}
		blade.Processor, blade.ProcessorCount, blade.ProcessorCoreCount, blade.ProcessorThreadCount, err = s.CPU()
		if err != nil {
			return nil, err
		}
		blade.Memory, err = s.Memory()
		if err != nil {
			return nil, err
		}
		blade.Status, err = s.Status()
		if err != nil {
			return nil, err
		}
		blade.Name, err = s.Name()
		if err != nil {
			return nil, err
		}
		blade.TempC, err = s.TempC()
		if err != nil {
			return nil, err
		}
		blade.PowerKw, err = s.PowerKw()
		if err != nil {
			return nil, err
		}
		blade.PowerState, err = s.PowerState()
		if err != nil {
			return nil, err
		}
		blade.BmcLicenceType, blade.BmcLicenceStatus, err = s.License()
		if err != nil {
			return nil, err
		}
		server = blade
	} else {
		discrete := &devices.Discrete{}
		discrete.Vendor = s.Vendor()
		discrete.BmcAddress = s.ip
		discrete.BmcType = s.BmcType()

		discrete.Serial, err = s.Serial()
		if err != nil {
			return nil, err
		}
		discrete.BmcVersion, err = s.BmcVersion()
		if err != nil {
			return nil, err
		}
		discrete.Model, err = s.Model()
		if err != nil {
			return nil, err
		}
		discrete.Nics, err = s.Nics()
		if err != nil {
			return nil, err
		}
		discrete.Disks, err = s.Disks()
		if err != nil {
			return nil, err
		}
		discrete.BiosVersion, err = s.BiosVersion()
		if err != nil {
			return nil, err
		}
		discrete.Processor, discrete.ProcessorCount, discrete.ProcessorCoreCount, discrete.ProcessorThreadCount, err = s.CPU()
		if err != nil {
			return nil, err
		}
		discrete.Memory, err = s.Memory()
		if err != nil {
			return nil, err
		}
		discrete.Status, err = s.Status()
		if err != nil {
			return nil, err
		}
		discrete.Name, err = s.Name()
		if err != nil {
			return nil, err
		}
		discrete.TempC, err = s.TempC()
		if err != nil {
			return nil, err
		}
		discrete.PowerKw, err = s.PowerKw()
		if err != nil {
			return nil, err
		}
		discrete.PowerState, err = s.PowerState()
		if err != nil {
			return nil, err
		}
		discrete.BmcLicenceType, discrete.BmcLicenceStatus, err = s.License()
		if err != nil {
			return nil, err
		}
		server = discrete
	}

	return server, err
}

// Disks returns a list of disks installed on the device
func (s *SupermicroX10) Disks() (disks []*devices.Disk, err error) {
	return disks, err
}

// UpdateCredentials updates login credentials
func (s *SupermicroX10) UpdateCredentials(username string, password string) {
	s.username = username
	s.password = password
}
