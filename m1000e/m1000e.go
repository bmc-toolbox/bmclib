package m1000e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/ncode/bmc/dell"
	"github.com/ncode/bmc/devices"
	"github.com/ncode/bmc/errors"
	"github.com/ncode/bmc/httpclient"
)

const (
	// BMCModel defines the bmc model that is supported by this package
	BMCModel = "m1000e"
)

var (
	macFinder = regexp.MustCompile("([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})")
	findBmcIP = regexp.MustCompile("bladeIpAddress\">((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})")
)

// M1000e holds the status and properties of a connection to a CMC device
type M1000e struct {
	ip           string
	username     string
	password     string
	client       *http.Client
	cmcJSON      *dell.CMC
	cmcTemp      *dell.CMCTemp
	cmcWWN       *dell.CMCWWN
	SessionToken string //required to set config
}

// New returns a connection to M1000e
func New(ip string, username string, password string) (chassis *M1000e, err error) {
	return &M1000e{ip: ip, username: username, password: password}, err
}

// Login initiates the connection to a chassis device
func (m *M1000e) Login() (err error) {
	log.WithFields(log.Fields{"step": "chassis connection", "vendor": dell.VendorID, "ip": m.ip}).Debug("connecting to chassis")

	form := url.Values{}
	form.Add("user", m.username)
	form.Add("password", m.password)

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", m.ip))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	m.client, err = httpclient.Build()
	if err != nil {
		return err
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	auth, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(string(auth), "Try Again") {
		return errors.ErrLoginFailed
	}

	if resp.StatusCode == 404 {
		return errors.ErrPageNotFound
	}

	err = m.loadHwData()
	if err != nil {
		return err
	}

	return err
}

func (m *M1000e) loadHwData() (err error) {
	url := "json?method=groupinfo"
	payload, err := m.get(url)
	if err != nil {
		return err
	}

	m.cmcJSON = &dell.CMC{}
	err = json.Unmarshal(payload, m.cmcJSON)
	if err != nil {
		httpclient.DumpInvalidPayload(url, m.ip, payload)
		return err
	}

	if m.cmcJSON.Chassis == nil {
		return errors.ErrUnableToReadData
	}

	url = "json?method=blades-wwn-info"
	payload, err = m.get(url)
	if err != nil {
		return err
	}

	m.cmcWWN = &dell.CMCWWN{}
	err = json.Unmarshal(payload, m.cmcWWN)
	if err != nil {
		httpclient.DumpInvalidPayload(url, m.ip, payload)
		return err
	}

	return err
}

// Logout logs out and close the chassis connection
func (m *M1000e) Logout() (err error) {
	_, err = m.client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/logout", m.ip))
	if err != nil {
		return err
	}
	return err
}

func (m *M1000e) get(endpoint string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "chassis connection", "vendor": dell.VendorID, "ip": m.ip, "endpoint": endpoint}).Debug("retrieving data from chassis")

	resp, err := m.client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", m.ip, endpoint))
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

	// Dell has a really shitty consistency of the data type returned, here we fix what's possible
	payload = bytes.Replace(payload, []byte("\"bladeTemperature\":-1"), []byte("\"bladeTemperature\":\"0\""), -1)
	payload = bytes.Replace(payload, []byte("\"nic\": [],"), []byte("\"nic\": {},"), -1)
	payload = bytes.Replace(payload, []byte("N\\/A"), []byte("0"), -1)

	return payload, err
}

// Name returns the hostname of the machine
func (m *M1000e) Name() (name string, err error) {
	return m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.ChassisStatus.CHASSISName, err
}

// Model returns the device model
func (m *M1000e) Model() (model string, err error) {
	return strings.TrimSpace(m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.ChassisStatus.ROChassisProductname), err
}

// Serial returns the device serial
func (m *M1000e) Serial() (serial string, err error) {
	return strings.ToLower(m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.ChassisStatus.ROChassisServiceTag), err
}

// PowerKw returns the current power usage in Kw
func (m *M1000e) PowerKw() (power float64, err error) {
	p, err := strconv.Atoi(strings.TrimRight(m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.PsuStatus.AcPower, " W"))
	if err != nil {
		return power, err
	}
	return float64(p) / 1000.00, err
}

// TempC returns the current temperature of the machine
func (m *M1000e) TempC() (temp int, err error) {
	url := "json?method=temp-sensors"
	payload, err := m.get(url)
	if err != nil {
		return temp, err
	}

	dellCMCTemp := &dell.CMCTemp{}
	err = json.Unmarshal(payload, dellCMCTemp)
	if err != nil {
		httpclient.DumpInvalidPayload(url, m.ip, payload)
		return temp, err
	}

	if dellCMCTemp.ChassisTemp != nil {
		return dellCMCTemp.ChassisTemp.TempCurrentValue, err
	}

	return temp, err
}

// Status returns health string status from the bmc
func (m *M1000e) Status() (status string, err error) {
	if m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.CMCStatus.CMCActiveError == "No Errors" {
		status = "OK"
	} else {
		status = m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.CMCStatus.CMCActiveError
	}
	return status, err
}

// FwVersion returns the current firmware version of the bmc
func (m *M1000e) FwVersion() (version string, err error) {
	return m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.ChassisStatus.ROCmcFwVersionString, err
}

// Nics returns all found Nics in the device
func (m *M1000e) Nics() (nics []*devices.Nic, err error) {
	payload, err := m.get("cmc_status?cat=C01&tab=T11&id=P31")
	if err != nil {
		return nics, err
	}

	mac := macFinder.FindString(string(payload))
	if mac != "" {
		nics = make([]*devices.Nic, 0)
		n := &devices.Nic{
			Name:       "OA1",
			MacAddress: strings.ToLower(mac),
		}
		nics = append(nics, n)
	}

	return nics, err
}

// IsActive returns health string status from the bmc
func (m *M1000e) IsActive() bool {
	return true
}

// PassThru returns the type of switch we have for this chassis
func (m *M1000e) PassThru() (passthru string, err error) {
	passthru = "1G"
	for _, dellBlade := range m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.Blades {
		if dellBlade.BladePresent == 1 && dellBlade.IsStorageBlade == 0 {
			for _, nic := range dellBlade.Nics {
				if strings.Contains(nic.BladeNicName, "10G") {
					passthru = "10G"
				} else {
					passthru = "1G"
				}
				return passthru, err
			}
		}
	}
	return passthru, err
}

// Psus returns a list of psus installed on the device
func (m *M1000e) Psus() (psus []*devices.Psu, err error) {
	serial, _ := m.Serial()
	for _, psu := range m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.PsuStatus.Psus {
		if psu.PsuPresent == 0 {
			continue
		}

		i, err := strconv.ParseFloat(strings.TrimSuffix(psu.PsuAcCurrent, " A"), 64)
		if err != nil {
			return psus, err
		}

		e, err := strconv.ParseFloat(psu.PsuAcVolts, 64)
		if err != nil {
			return psus, err
		}

		var status string
		if psu.PsuActiveError == "No Errors" {
			status = "OK"
		} else {
			status = psu.PsuActiveError
		}

		p := &devices.Psu{
			Serial:     fmt.Sprintf("%s_%s", serial, psu.PsuPosition),
			CapacityKw: float64(psu.PsuCapacity) / 1000.00,
			PowerKw:    (i * e) / 1000.00,
			Status:     status,
		}

		psus = append(psus, p)
	}

	return psus, err
}

// StorageBlades returns all StorageBlades found in this chassis
func (m *M1000e) StorageBlades() (storageBlades []*devices.StorageBlade, err error) {
	for _, dellBlade := range m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.Blades {
		if dellBlade.BladePresent == 1 && dellBlade.IsStorageBlade == 1 {
			storageBlade := devices.StorageBlade{}
			storageBlade.BladePosition = dellBlade.BladeMasterSlot
			storageBlade.Serial = strings.ToLower(dellBlade.BladeSvcTag)
			storageBlade.Model = dellBlade.BladeModel
			storageBlade.PowerKw = float64(dellBlade.ActualPwrConsump) / 1000
			temp, err := strconv.Atoi(dellBlade.BladeTemperature)
			if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": m.ip, "position": storageBlade.BladePosition, "type": "chassis", "error": err}).Warning("Auditing blade")
				continue
			}
			storageBlade.TempC = temp
			if dellBlade.BladeLogDescription == "No Errors" {
				storageBlade.Status = "OK"
			} else {
				storageBlade.Status = dellBlade.BladeLogDescription
			}
			storageBlade.Vendor = dell.VendorID
			storageBlade.FwVersion = dellBlade.BladeBIOSver
			storageBlades = append(storageBlades, &storageBlade)
		}
	}
	return storageBlades, err
}

// Blades returns all StorageBlades found in this chassis
func (m *M1000e) Blades() (blades []*devices.Blade, err error) {
	for _, dellBlade := range m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.Blades {
		if dellBlade.BladePresent == 1 && dellBlade.IsStorageBlade == 0 {
			blade := devices.Blade{}
			blade.BladePosition = dellBlade.BladeMasterSlot
			blade.Serial = strings.ToLower(dellBlade.BladeSvcTag)
			blade.Model = dellBlade.BladeModel
			if dellBlade.BladePowerState == 1 {
				blade.PowerState = "on"
			} else if dellBlade.BladePowerState == 8 {
				blade.PowerState = "off"
			} else {
				blade.PowerState = "unknow"
			}

			blade.PowerKw = float64(dellBlade.ActualPwrConsump) / 1000
			temp, err := strconv.Atoi(dellBlade.BladeTemperature)
			if err != nil {
				log.WithFields(log.Fields{"operation": "connection", "ip": m.ip, "position": blade.BladePosition, "type": "chassis"}).Warning(err)
				continue
			} else {
				blade.TempC = temp
			}
			if dellBlade.BladeLogDescription == "No Errors" {
				blade.Status = "OK"
			} else {
				blade.Status = dellBlade.BladeLogDescription
			}
			blade.Vendor = dell.VendorID
			blade.BiosVersion = dellBlade.BladeBIOSver
			blade.Name = dellBlade.BladeName
			idracURL := strings.TrimLeft(dellBlade.IdracURL, "https://")
			idracURL = strings.TrimLeft(idracURL, "http://")
			idracURL = strings.Split(idracURL, ":")[0]
			blade.BmcAddress = idracURL
			blade.BmcVersion = dellBlade.BladeUSCVer

			if bmcData, ok := m.cmcWWN.SlotMacWwn.SlotMacWwnList[blade.BladePosition]; ok {
				n := &devices.Nic{
					Name:       "bmc",
					MacAddress: strings.ToLower(bmcData.IsNotDoubleHeight.PortFMAC),
				}
				blade.Nics = append(blade.Nics, n)
			}

			if strings.HasPrefix(blade.BmcAddress, "[") {
				payload, err := m.get(fmt.Sprintf("blade_status?id=%d&cat=C10&tab=T41&id=P78", blade.BladePosition))
				if err != nil {
					log.WithFields(log.Fields{"operation": "connection", "ip": m.ip, "position": blade.BladePosition, "type": "chassis"}).Warning(err)
				} else {
					ip := findBmcIP.FindStringSubmatch(string(payload))
					if len(ip) > 0 {
						blade.BmcAddress = ip[1]
					}
				}
			}

			for _, nic := range dellBlade.Nics {
				if nic.BladeNicName == "" {
					log.WithFields(log.Fields{"operation": "connection", "ip": m.ip, "position": blade.BladePosition, "type": "chassis"}).Error("Network card information missing, please verify")
					continue
				}
				n := &devices.Nic{
					Name:       strings.ToLower(nic.BladeNicName[:len(nic.BladeNicName)-17]),
					MacAddress: strings.ToLower(nic.BladeNicName[len(nic.BladeNicName)-17:]),
				}
				blade.Nics = append(blade.Nics, n)
			}
			blades = append(blades, &blade)
		}
	}
	return blades, err
}

// Vendor returns bmc's vendor
func (m *M1000e) Vendor() (vendor string) {
	return dell.VendorID
}

// ChassisSnapshot do best effort to populate the server data and returns a blade or discrete
func (m *M1000e) ChassisSnapshot() (chassis *devices.Chassis, err error) {
	chassis = &devices.Chassis{}
	chassis.Vendor = m.Vendor()
	chassis.BmcAddress = m.ip
	chassis.Name, _ = m.Name()
	chassis.Serial, _ = m.Serial()
	chassis.Model, _ = m.Model()
	chassis.PowerKw, _ = m.PowerKw()
	chassis.TempC, _ = m.TempC()
	chassis.Status, _ = m.Status()
	chassis.FwVersion, _ = m.FwVersion()
	chassis.PassThru, _ = m.PassThru()
	chassis.Blades, _ = m.Blades()
	chassis.StorageBlades, _ = m.StorageBlades()
	chassis.Nics, _ = m.Nics()
	chassis.Psus, _ = m.Psus()
	return chassis, err
}

// UpdateCredentials updates login credentials
func (m *M1000e) UpdateCredentials(username string, password string) {
	m.username = username
	m.password = password
}
