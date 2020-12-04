package openbmc

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"

	"github.com/go-logr/logr"
)

type OpenBmc struct {
	ip         string
	username   string
	password   string
	httpClient *http.Client
	ctx        context.Context
	log        logr.Logger
}

func (b *OpenBmc) http_get(endpoint string) (payload []byte, err error) {
	url := fmt.Sprintf("https://%s:%s@%s/%s", b.username, b.password, b.ip, endpoint)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return payload, err
	}

	reqDump, _ := httputil.DumpRequestOut(req, true)
	b.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", url)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return payload, err
	}
	defer resp.Body.Close()

	respDump, _ := httputil.DumpResponse(resp, true)
	b.log.V(2).Info("responseTRace", "responseDump", string(respDump))

	payload, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return payload, err
	}

	return payload, err
}

func (b *OpenBmc) http_post(endpoint string, data string) (response []byte, err error) {
	url := fmt.Sprintf("https://%s:%s@%s/%s", b.username, b.password, b.ip, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return response, err
	}
	req.Header.Set("Content-Type", "application/json")

	reqDump, _ := httputil.DumpRequestOut(req, true)
	b.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", url)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	respDump, _ := httputil.DumpResponse(resp, true)
	b.log.V(2).Info("responseTRace", "responseDump", string(respDump))

	if resp.StatusCode != 200 {
		err = fmt.Errorf("HTTP POST to %s failed, status: %s", endpoint, resp.Status)
		return response, err
	}

	return ioutil.ReadAll(resp.Body)
}

func (b *OpenBmc) redfish_get(endpoint string) (payload []byte, err error) {
	return b.http_get("redfish/v1/" + endpoint)
}

func (b *OpenBmc) redfish_post(endpoint string, data string) (response []byte, err error) {
	return b.http_post("redfish/v1/" + endpoint, data)
}

func (b *OpenBmc) Bios(cfg *cfgresources.Bios) (err error) {
	return err
}

func (b *OpenBmc) BiosVersion() (version string, err error) {
	return version, err
}

func (b *OpenBmc) CPU() (cpu string, cpuCount int, coreCount int, hyperthreadCount int, err error) {
	return cpu, cpuCount, coreCount, hyperthreadCount, err
}

func (b *OpenBmc) ChassisSerial() (serial string, err error) {
	return serial, err
}

func (b *OpenBmc) CheckCredentials() (err error) {
	return err
}

func (b *OpenBmc) Close() (err error) {
	return err
}

func (b *OpenBmc) CurrentHTTPSCert() ([]*x509.Certificate, bool, error) {

	dialer := &net.Dialer{
		Timeout: time.Duration(10) * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", b.ip+":"+"443", &tls.Config{InsecureSkipVerify: true})

	if err != nil {
		return []*x509.Certificate{{}}, false, err
	}

	defer conn.Close()

	return conn.ConnectionState().PeerCertificates, false, nil

}

func (b *OpenBmc) Disks() (disks []*devices.Disk, err error) {
	return disks, err
}

func (b *OpenBmc) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {
	return []byte{}, nil
}

func (b *OpenBmc) HardwareType() (model string) {
	return "OpenBMC" // FIXME: what's this supposed to mean?
}

func (b *OpenBmc) IsBlade() (isBlade bool, err error) {
	return isBlade, err
}

func (b *OpenBmc) IsOn() (status bool, err error) {
	return status, err
}

func (b *OpenBmc) Ldap(cfgLdap *cfgresources.Ldap) error {
	return nil
}

func (b *OpenBmc) LdapGroup(cfgGroup []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {
	return err
}

func (b *OpenBmc) License() (name string, licType string, err error) {
	return name, licType, err
}

func (b *OpenBmc) Memory() (mem int, err error) {
	type Memory struct {
		TotalSystemMemoryGiB int `json:"TotalSystemMemoryGiB"`
	}

	type OBmcSystemMemorySummary struct {
		MemorySummary Memory `json:"MemorySummary"`
	}

	data, err := b.redfish_get("Systems/system")
	if err != nil {
		return mem, err
	}

	memsum := &OBmcSystemMemorySummary{}
	err = json.Unmarshal(data, memsum)
	if err != nil {
		return mem, err
	}

	return memsum.MemorySummary.TotalSystemMemoryGiB * 1024, err
}

func (b *OpenBmc) Model() (model string, err error) {
	type OBmcSystemModel struct {
		Model string `json:"Model"`
	}

	data, err := b.redfish_get("Systems/system")
	if err != nil {
		return model, err
	}

	mod := &OBmcSystemModel{}
	err = json.Unmarshal(data, mod)
	if err != nil {
		return model, err
	}

	return mod.Model, err
}

func (b *OpenBmc) Name() (name string, err error) {
	return name, err
}

func (b *OpenBmc) Network(cfg *cfgresources.Network) (reset bool, err error) {
	return reset, err
}

func (b *OpenBmc) Nics() (nics []*devices.Nic, err error) {
	return nics, err
}

func (b *OpenBmc) Ntp(cfg *cfgresources.Ntp) (err error) {
	return err
}

func (b *OpenBmc) Power(cfg *cfgresources.Power) (err error) {
	return err
}

func (b *OpenBmc) do_reset(action string) (err error) {
	_, err = b.redfish_post("Systems/system/Action/ComputerSystem.Reset",
	                        fmt.Sprintf(`{"ResetType":"%s"}`, action))
	return err
}

func (b *OpenBmc) do_bmc_reset(action string) (err error) {
	_, err = b.redfish_post("Managers/bmc/Action/Manager.Reset",
	                        fmt.Sprintf(`{"ResetType":"%s"}`, action))
	return err
}

func (b *OpenBmc) PowerCycle() (status bool, err error) {
	err = b.do_reset("PowerCycle")
	return err == nil, err
}

func (b *OpenBmc) PowerCycleBmc() (status bool, err error) {
	err = b.do_bmc_reset("GracefulRestart")
	return err == nil, err
}

func (b *OpenBmc) PowerKw() (power float64, err error) {
	return power, err
}

func (b *OpenBmc) PowerOn() (status bool, err error) {
	err = b.do_reset("On")
	return err == nil, err
}

func (b *OpenBmc) PowerOff() (status bool, err error) {
	err = b.do_reset("ForceOff")
	return err == nil, err
}

func (b *OpenBmc) PowerState() (state string, err error) {
	type OBmcSystemPowerState struct {
		PowerState string `json:"PowerState"`
	}

	data, err := b.redfish_get("Systems/system")
	if err != nil {
		return state, err
	}

	ps := &OBmcSystemPowerState{}
	err = json.Unmarshal(data, ps)
	if err != nil {
		return state, err
	}

	return ps.PowerState, err
}

func (b *OpenBmc) PxeOnce() (status bool, err error) {
	return status, err
}

func (b *OpenBmc) Resources() []string {
	return []string{}
}

func (b *OpenBmc) Screenshot() (response []byte, extension string, err error) {
	return response, extension, err
}

func (b *OpenBmc) Serial() (serial string, err error) {
	type OBmcSystemSerial struct {
		SerialNumber string `json:"SerialNumber"`
	}

	data, err := b.redfish_get("Systems/system")
	if err != nil {
		return serial, err
	}

	ser := &OBmcSystemSerial{}
	err = json.Unmarshal(data, ser)
	if err != nil {
		return serial, err
	}

	return ser.SerialNumber, err
}

func (b *OpenBmc) ServerSnapshot() (server interface{}, err error) {
	return server, err
}

func (b *OpenBmc) SetLicense(cfg *cfgresources.License) (err error) {
	return err
}

func (b *OpenBmc) Slot() (slot int, err error) {
	return slot, err
}

func (b *OpenBmc) Status() (health string, err error) {
	type StatusHealth struct {
		Health string `json:"Health"`
	}

	type OBmcSystemStatus struct {
		Status StatusHealth `json:"Status"`
	}

	data, err := b.redfish_get("Systems/system")
	if err != nil {
		return health, err
	}

	stat := &OBmcSystemStatus{}
	err = json.Unmarshal(data, stat)
	if err != nil {
		return health, err
	}

	return stat.Status.Health, err
}

func (b *OpenBmc) Syslog(cfg *cfgresources.Syslog) (err error) {
	return err
}

func (b *OpenBmc) TempC() (temp int, err error) {
	return temp, err
}

func (b *OpenBmc) UpdateCredentials(username string, password string) {
	b.username = username
	b.password = password
}

func (b *OpenBmc) UpdateFirmware(source, file string) (status bool, err error) {
	return true, fmt.Errorf("NYI")
}

func (b *OpenBmc) UploadHTTPSCert(cert []byte, certFileName string, key []byte, keyFileName string) (bool, error) {
	return false, fmt.Errorf("NYI")
}

func (b *OpenBmc) User(users []*cfgresources.User) (err error) {
	return err
}

func (b *OpenBmc) Vendor() (vendor string) {
	return "OpenBMC"
}

func (b *OpenBmc) Version() (bmcVersion string, err error) {
	return bmcVersion, err
}

func New(ctx context.Context, ip string, username string, password string, log logr.Logger) (obmc *OpenBmc, err error) {
	client, err := httpclient.Build()
	if err != nil {
		return obmc, err
	}

	return &OpenBmc{
		ip:         ip,
		username:   username,
		password:   password,
		httpClient: client,
		ctx:        ctx,
		log:        log,
	}, err
}
