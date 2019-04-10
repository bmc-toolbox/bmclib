package m1000e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

const (
	// BMCType defines the bmc model that is supported by this package
	BMCType = "m1000e"
)

var (
	macFinder          = regexp.MustCompile("([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})")
	findBmcIP          = regexp.MustCompile("bladeIpAddress\">((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})")
	findRedundancyMode = regexp.MustCompile("selected=\"selected\">(.+)</option>")
)

// M1000e holds the status and properties of a connection to a CMC device
type M1000e struct {
	ip           string
	username     string
	password     string
	httpClient   *http.Client
	sshClient    *sshclient.SSHClient
	cmcJSON      *dell.CMC
	cmcTemp      *dell.CMCTemp
	cmcWWN       *dell.CMCWWN
	SessionToken string //required to set config
}

// New returns a connection to M1000e
func New(ip string, username string, password string) (chassis *M1000e, err error) {
	return &M1000e{ip: ip, username: username, password: password}, err
}

// CheckCredentials verify whether the credentials are valid or not
func (m *M1000e) CheckCredentials() (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}
	return err
}

func (m *M1000e) get(endpoint string) (payload []byte, err error) {
	log.WithFields(log.Fields{"step": "chassis connection", "vendor": dell.VendorID, "ip": m.ip, "endpoint": endpoint}).Debug("retrieving data from chassis")

	resp, err := m.httpClient.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", m.ip, endpoint))
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
	payload = bytes.Replace(payload, []byte("\"nic\":[],"), []byte("\"nic\":{},"), -1)
	payload = bytes.Replace(payload, []byte("N\\/A"), []byte("0"), -1)

	return payload, err
}

// Name returns the hostname of the machine
func (m *M1000e) Name() (name string, err error) {
	err = m.httpLogin()
	if err != nil {
		return name, err
	}
	return m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.ChassisStatus.CHASSISName, err
}

// BmcType returns just Model id string - m1000e
func (m *M1000e) BmcType() (model string) {
	return BMCType
}

// Model returns the full device model string
func (m *M1000e) Model() (model string, err error) {
	err = m.httpLogin()
	if err != nil {
		return model, err
	}
	return strings.TrimSpace(m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.ChassisStatus.ROChassisProductname), err
}

// Serial returns the device serial
func (m *M1000e) Serial() (serial string, err error) {
	err = m.httpLogin()
	if err != nil {
		return serial, err
	}
	return strings.ToLower(m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.ChassisStatus.ROChassisServiceTag), err
}

// PowerKw returns the current power usage in Kw
func (m *M1000e) PowerKw() (power float64, err error) {
	err = m.httpLogin()
	if err != nil {
		return power, err
	}
	p, err := strconv.Atoi(strings.TrimRight(m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.PsuStatus.AcPower, " W"))
	if err != nil {
		return power, err
	}
	return float64(p) / 1000.00, err
}

// TempC returns the current temperature of the machine
func (m *M1000e) TempC() (temp int, err error) {
	err = m.httpLogin()
	if err != nil {
		return temp, err
	}

	url := "json?method=temp-sensors"
	payload, err := m.get(url)
	if err != nil {
		return temp, err
	}

	dellCMCTemp := &dell.CMCTemp{}
	err = json.Unmarshal(payload, dellCMCTemp)
	if err != nil {
		return temp, err
	}

	if dellCMCTemp.ChassisTemp != nil {
		return dellCMCTemp.ChassisTemp.TempCurrentValue, err
	}

	return temp, err
}

// Fans returns all found fans in the device
func (m *M1000e) Fans() (fans []*devices.Fan, err error) {
	serial, err := m.Serial()
	if err != nil {
		return fans, err
	}

	for pos, fan := range m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.Fans {
		if fans == nil {
			fans = make([]*devices.Fan, 0)
		}

		if fan.Presence == 1 {
			status := "OK"
			if fan.ActiveError != "No Errors" {
				status = fan.ActiveError
			}

			p, err := strconv.Atoi(pos)
			if err != nil && pos != "ECM" {
				return fans, fmt.Errorf("unable to read: %s", pos)
			}

			f := &devices.Fan{
				Serial:     fmt.Sprintf("%d_%s", p, serial),
				Position:   p,
				Status:     status,
				CurrentRPM: fan.FanRPM,
			}
			fans = append(fans, f)
		}
	}

	return fans, err
}

// Status returns health string status from the bmc
func (m *M1000e) Status() (status string, err error) {
	err = m.httpLogin()
	if err != nil {
		return "", err
	}
	if m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.CMCStatus.CMCActiveError == "No Errors" {
		status = "OK"
	} else {
		status = m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.CMCStatus.CMCActiveError
	}
	return status, err
}

// FwVersion returns the current firmware version of the bmc
func (m *M1000e) FwVersion() (version string, err error) {
	err = m.httpLogin()
	if err != nil {
		return version, err
	}
	return m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.ChassisStatus.ROCmcFwVersionString, err
}

// Nics returns all found Nics in the device
func (m *M1000e) Nics() (nics []*devices.Nic, err error) {
	err = m.httpLogin()
	if err != nil {
		return nics, err
	}

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
	err = m.httpLogin()
	if err != nil {
		return passthru, err
	}
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
	serial, err := m.Serial()
	if err != nil {
		return psus, err
	}

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

		psuPosition, err := strconv.Atoi(strings.Split(psu.PsuPosition, "_")[1])
		if err != nil {
			return psus, err
		}

		p := &devices.Psu{
			Serial:     fmt.Sprintf("%s_%s", serial, psu.PsuPosition),
			CapacityKw: float64(psu.PsuCapacity) / 1000.00,
			PowerKw:    (i * e) / 1000.00,
			Status:     status,
			PartNumber: psu.PsuPartNum,
			Position:   psuPosition,
		}

		psus = append(psus, p)
	}

	return psus, err
}

// StorageBlades returns all StorageBlades found in this chassis
func (m *M1000e) StorageBlades() (storageBlades []*devices.StorageBlade, err error) {
	err = m.httpLogin()
	if err != nil {
		return storageBlades, err
	}
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
	err = m.httpLogin()
	if err != nil {
		return blades, err
	}
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
					Name: "bmc",
				}
				if bmcData.IsNotDoubleHeight.IsSelected == 1 {
					blade.FlexAddressEnabled = true
					n.MacAddress = strings.ToLower(bmcData.IsNotDoubleHeight.PortPMAC)
				} else {
					n.MacAddress = strings.ToLower(bmcData.IsNotDoubleHeight.PortFMAC)
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
	chassis.Name, err = m.Name()
	if err != nil {
		return nil, err
	}
	chassis.Serial, err = m.Serial()
	if err != nil {
		return nil, err
	}
	chassis.Model, err = m.Model()
	if err != nil {
		return nil, err
	}
	chassis.PowerKw, err = m.PowerKw()
	if err != nil {
		return nil, err
	}
	chassis.TempC, err = m.TempC()
	if err != nil {
		return nil, err
	}
	chassis.Status, err = m.Status()
	if err != nil {
		return nil, err
	}
	chassis.FwVersion, err = m.FwVersion()
	if err != nil {
		return nil, err
	}
	chassis.PassThru, err = m.PassThru()
	if err != nil {
		return nil, err
	}
	chassis.Blades, err = m.Blades()
	if err != nil {
		return nil, err
	}
	chassis.StorageBlades, err = m.StorageBlades()
	if err != nil {
		return nil, err
	}
	chassis.Nics, err = m.Nics()
	if err != nil {
		return nil, err
	}
	chassis.Psus, err = m.Psus()
	if err != nil {
		return nil, err
	}
	chassis.Fans, err = m.Fans()
	if err != nil {
		return nil, err
	}

	return chassis, err
}

// UpdateCredentials updates login credentials
func (m *M1000e) UpdateCredentials(username string, password string) {
	m.username = username
	m.password = password
}

// IsPsuRedundant informs whether or not the power is currently redundant
func (m *M1000e) IsPsuRedundant() (status bool, err error) {
	err = m.httpLogin()
	if err != nil {
		return status, err
	}

	if m.cmcJSON.Chassis.ChassisGroupMemberHealthBlob.PsuStatus.PsuRedundancy == 1 {
		return true, err
	}
	return false, err
}

// PsuRedundancyMode returns the current redundancy mode is configured for the chassis
func (m *M1000e) PsuRedundancyMode() (mode string, err error) {
	err = m.httpLogin()
	if err != nil {
		return mode, err
	}

	payload, err := m.get("pwr_redundancy?cat=C00&tab=T03&id=P10")
	if err != nil {
		return mode, err
	}

	fnd := findRedundancyMode.FindStringSubmatch(string(payload))
	var rm string
	if len(fnd) > 0 {
		rm = fnd[1]
	}
	switch rm {
	case "Grid Redundancy":
		return devices.Grid, err
	case "Power Supply Redundancy":
		return devices.PowerSupply, err
	case "No Redundancy":
		return devices.NoRedundancy, err
	default:
		return devices.Unknown, err
	}
}
