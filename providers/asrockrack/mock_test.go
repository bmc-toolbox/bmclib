package asrockrack

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"
)

var (
	loginPayload           = []byte(`username=foo&password=bar&certlogin=0`)
	loginResponse          = []byte(`{ "ok": 0, "privilege": 4, "extendedpriv": 259, "racsession_id": 10, "remote_addr": "136.144.50.145", "server_name": "10.230.148.171", "server_addr": "10.230.148.171", "HTTPSEnabled": 1, "CSRFToken": "l5L29IP7" }`)
	fwinfoResponse         = []byte(`{ "BMC_fw_version": "0.01.00", "BIOS_fw_version": "L2.07B", "ME_fw_version": "5.1.3.78", "Micro_Code_version": "000000ca", "CPLD_version": "N\/A", "CM_version": "0.13.01", "BPB_version": "0.0.002.0", "Node_id": "2" }`)
	fwUploadResponse       = []byte(`{"cc": 0}`)
	fwVerificationResponse = []byte(`[ { "id": 1, "current_image_name": "ast2500e", "current_image_version1": "0.01.00", "current_image_version2": "", "new_image_version": "0.03.00", "section_status": 0, "verification_status": 5 } ]`)
	fwUpgradeProgress      = []byte(`{ "id": 1, "action": "Flashing...", "progress": "__PERCENT__% done         ", "state": __STATE__ }`)
	usersPayload           = []byte(`[ { "id": 1, "name": "anonymous", "access": 0, "kvm": 1, "vmedia": 1, "snmp": 0, "prev_snmp": 0, "network_privilege": "administrator", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "none", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "ami_format", "ssh_key": "Not Available", "creation_time": 4802 }, { "id": 2, "name": "admin", "access": 1, "kvm": 1, "vmedia": 1, "snmp": 0, "prev_snmp": 0, "network_privilege": "administrator", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "none", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "ami_format", "ssh_key": "Not Available", "creation_time": 188 }, { "id": 3, "name": "foo", "access": 1, "kvm": 1, "vmedia": 1, "snmp": 0, "prev_snmp": 0, "network_privilege": "administrator", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "none", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "ami_format", "ssh_key": "Not Available", "creation_time": 4802 }, { "id": 4, "name": "", "access": 0, "kvm": 0, "vmedia": 0, "snmp": 0, "prev_snmp": 0, "network_privilege": "", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "", "ssh_key": "Not Available", "creation_time": 0 }, { "id": 5, "name": "", "access": 0, "kvm": 0, "vmedia": 0, "snmp": 0, "prev_snmp": 0, "network_privilege": "", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "", "ssh_key": "Not Available", "creation_time": 0 }, { "id": 6, "name": "", "access": 0, "kvm": 0, "vmedia": 0, "snmp": 0, "prev_snmp": 0, "network_privilege": "", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "", "ssh_key": "Not Available", "creation_time": 0 }, { "id": 7, "name": "", "access": 0, "kvm": 0, "vmedia": 0, "snmp": 0, "prev_snmp": 0, "network_privilege": "", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "", "ssh_key": "Not Available", "creation_time": 0 }, { "id": 8, "name": "", "access": 0, "kvm": 0, "vmedia": 0, "snmp": 0, "prev_snmp": 0, "network_privilege": "", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "", "ssh_key": "Not Available", "creation_time": 0 }, { "id": 9, "name": "", "access": 0, "kvm": 0, "vmedia": 0, "snmp": 0, "prev_snmp": 0, "network_privilege": "", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "", "ssh_key": "Not Available", "creation_time": 0 }, { "id": 10, "name": "", "access": 0, "kvm": 0, "vmedia": 0, "snmp": 0, "prev_snmp": 0, "network_privilege": "", "fixed_user_count": 2, "snmp_access": "", "OEMProprietary_level_Privilege": 1, "privilege_limit_serial": "", "snmp_authentication_protocol": "", "snmp_privacy_protocol": "", "email_id": "", "email_format": "", "ssh_key": "Not Available", "creation_time": 0 } ]`)
	inventoryinfoResponse  = []byte(`[ { "device_id": 1, "device_name": "CPU1", "device_type": "CPU", "product_manufacturer_name": "Intel(R) Corporation", "product_name": "Intel(R) Xeon(R) E-2278G CPU @ 3.40GHz", "product_part_number": "N\/A", "product_version": "N\/A", "product_serial_number": "N\/A", "product_asset_tag": "N\/A", "product_extra": "N\/A" }, { "device_id": 5, "device_name": "DDR4_A1", "device_type": "Memory", "product_manufacturer_name": "Micron", "product_name": "SODIMM", "product_part_number": "18ASF2G72HZ-2G6E1   ", "product_version": "N\/A", "product_serial_number": "2724B52D", "product_asset_tag": "N\/A", "product_extra": "2666 MT\/s  16GB" }, { "device_id": 7, "device_name": "DDR4_B1", "device_type": "Memory", "product_manufacturer_name": "Micron", "product_name": "SODIMM", "product_part_number": "18ASF2G72HZ-2G6E1   ", "product_version": "N\/A", "product_serial_number": "2724B58A", "product_asset_tag": "N\/A", "product_extra": "2666 MT\/s  16GB" }, { "device_id": 37, "device_name": "PCIe card 1", "device_type": "PCIe & OCP Card", "product_manufacturer_name": "8086(Intel Corporation)", "product_name": "020000(Ethernet controller)", "product_part_number": "1572", "product_version": "N\/A", "product_serial_number": "N\/A", "product_asset_tag": "PCIE7", "product_extra": "N\/A" }, { "device_id": 105, "device_name": "Storage ", "device_type": "Storage device", "product_manufacturer_name": "N\/A", "product_name": "N\/A", "product_part_number": "INTEL SSDSC2KB480G8", "product_version": "N\/A", "product_serial_number": "PHYF001303ED480BGN", "product_asset_tag": "SATA_4", "product_extra": "N\/A" }, { "device_id": 106, "device_name": "Storage ", "device_type": "Storage device", "product_manufacturer_name": "N\/A", "product_name": "N\/A", "product_part_number": "INTEL SSDSC2KB480G8", "product_version": "N\/A", "product_serial_number": "BTYF01940L38480BGN", "product_asset_tag": "SATA_5", "product_extra": "N\/A" } ]`)
	fruinfoResponse        = []byte(`[ { "device": { "id": 0, "name": "BMC_FRU" }, "common_header": { "version": 1, "internal_use_area_start_offset": 0, "chassis_info_area_start_offset": 1, "board_info_area_start_offset": 4, "product_info_area_start_offset": 11, "multi_record_area_start_offset": 0 }, "chassis": { "version": 1, "length": 3, "type": "Main Server Chassis", "part_number": "", "serial_number": "K61206147700263", "custom_fields": "" }, "board": { "version": 1, "length": 7, "language": 0, "date": "Mon Jul 20 06:04:00 2020\\n", "manufacturer": "ASRockRack", "product_name": "E3C246D4I-NL", "serial_number": "197965920000514", "part_number": "", "fru_file_id": "", "custom_fields": "" }, "product": { "version": 1, "length": 7, "language": 0, "manufacturer": "Packet", "product_name": "c3.small.x86", "part_number": "Open19", "product_version": "R1.00", "serial_number": "D6S0R8000736", "asset_tag": "", "fru_file_id": "", "custom_fields": "" } } ]`)
	biosPOSTCodeResponse   = []byte(`{ "poststatus": 1, "postdata": 160 }`)
	chassisStatusResponse  = []byte(`{ "power_status": 1, "led_status": 0 }`)

	// TODO: implement under rw mutex
	httpRequestTestVar *http.Request
)

// setup test BMC
var server *httptest.Server
var bmcURL *url.URL
var fwUpgradeState *testFwUpgradeState

type testFwUpgradeState struct {
	FlashModeSet     bool
	FirmwareUploaded bool
	FirmwareVerified bool
	UpgradeInitiated bool
	UpgradePercent   int
	ResetDone        bool
}

// the bmc lib client
var aClient *ASRockRack

func TestMain(m *testing.M) {
	var err error
	// setup mock server
	server = mockASRockBMC()
	bmcURL, _ = url.Parse(server.URL)

	l := logrus.New()
	l.Level = logrus.DebugLevel
	// setup bmc client
	tLog := logrusr.New(l)
	aClient, err = New(bmcURL.Host, "foo", "bar", tLog)
	if err != nil {
		log.Fatal(err.Error())
	}

	// firmware update test state
	fwUpgradeState = &testFwUpgradeState{}
	os.Exit(m.Run())
}

/////////////// mock bmc service ///////////////////////////
func mockASRockBMC() *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/", index)
	handler.HandleFunc("/api/session", session)
	handler.HandleFunc("/api/asrr/fw-info", fwinfo)
	handler.HandleFunc("/api/fru", fruinfo)
	handler.HandleFunc("/api/asrr/inventory_info", inventoryinfo)
	handler.HandleFunc("/api/sensors", sensorsinfo)
	handler.HandleFunc("/api/asrr/getbioscode", biosPOSTCodeinfo)
	handler.HandleFunc("/api/chassis-status", chassisStatusInfo)

	// fw update endpoints - in order of invocation
	handler.HandleFunc("/api/maintenance/flash", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/firmware", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/firmware/verification", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/firmware/upgrade", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/firmware/flash-progress", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/reset", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/asrr/maintenance/BIOS/firmware", biosFirmwareUpgrade)

	// user accounts endpoints
	handler.HandleFunc("/api/settings/users", userAccountList)
	handler.HandleFunc("/api/settings/users/3", userAccountList)
	return httptest.NewTLSServer(handler)
}

func index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_, _ = w.Write([]byte(`ASRockRack`))
	}
}

func userAccountList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if os.Getenv("TEST_FAIL_QUERY") != "" {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			_, _ = w.Write(usersPayload)
		}
	case "PUT":
		httpRequestTestVar = r
	}
}

func biosFirmwareUpgrade(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		switch r.RequestURI {
		case "/api/asrr/maintenance/BIOS/firmware":

			// validate content type
			if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
				w.WriteHeader(http.StatusBadRequest)
			}

			// parse multipart form
			err := r.ParseMultipartForm(100)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
	}
}

func bmcFirmwareUpgrade(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		switch r.RequestURI {
		// 3. bmc verifies uploaded firmware image
		case "/api/maintenance/firmware/verification":
			if !fwUpgradeState.FirmwareUploaded {
				w.WriteHeader(http.StatusBadRequest)
			}
			fwUpgradeState.FirmwareVerified = true
			_, _ = w.Write(fwVerificationResponse)
		// 5. flash progress
		case "/api/maintenance/firmware/flash-progress":
			if !fwUpgradeState.UpgradeInitiated {
				w.WriteHeader(http.StatusBadRequest)
			}

			resp := fwUpgradeProgress
			if fwUpgradeState.UpgradePercent >= 100 {
				fwUpgradeState.UpgradePercent = 100
				// state: 2  indicates firmware flash complete
				resp = bytes.Replace(resp, []byte("__STATE__"), []byte(strconv.Itoa(2)), 1)
			} else {
				// state: 0 indicates firmware flash in progress
				resp = bytes.Replace(resp, []byte("__STATE__"), []byte(strconv.Itoa(0)), 1)
				fwUpgradeState.UpgradePercent += 50
			}

			resp = bytes.Replace(resp, []byte("__PERCENT__"), []byte(strconv.Itoa(fwUpgradeState.UpgradePercent)), 1)
			_, _ = w.Write(resp)
		}
	case "PUT":

		switch r.RequestURI {
		// 1. set device to flash mode
		case "/api/maintenance/flash":
			fwUpgradeState.FlashModeSet = true
			w.WriteHeader(http.StatusOK)
		// 4. run the upgrade
		case "/api/maintenance/firmware/upgrade":
			if !fwUpgradeState.FirmwareVerified {
				w.WriteHeader(http.StatusBadRequest)
			}

			if r.Header.Get("Content-Type") != "application/json" {
				w.WriteHeader(http.StatusBadRequest)
			}

			p := &preserveConfig{}
			err := json.NewDecoder(r.Body).Decode(&p)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// config should be preserved
			if p.PreserveConfig != 1 {
				w.WriteHeader(http.StatusBadRequest)
			}

			// full firmware flash
			if p.FlashStatus != 1 {
				w.WriteHeader(http.StatusBadRequest)
			}

			fwUpgradeState.UpgradeInitiated = true
			// respond with request body
			b := new(bytes.Buffer)
			_, _ = b.ReadFrom(r.Body)
			_, _ = w.Write(b.Bytes())
		}
	case "POST":
		switch r.RequestURI {
		case "/api/maintenance/reset":
			w.WriteHeader(http.StatusOK)

		// 2. upload firmware
		case "/api/maintenance/firmware":

			// validate flash mode set
			if !fwUpgradeState.FlashModeSet {
				w.WriteHeader(http.StatusBadRequest)
			}

			// validate content type
			if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
				w.WriteHeader(http.StatusBadRequest)
			}

			// parse multipart form
			err := r.ParseMultipartForm(100)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}

			fwUpgradeState.FirmwareUploaded = true
			_, _ = w.Write(fwUploadResponse)
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func fwinfo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_, _ = w.Write(fwinfoResponse)
	}
}

func fruinfo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_, _ = w.Write(fruinfoResponse)
	}
}

func inventoryinfo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_, _ = w.Write(inventoryinfoResponse)
	}
}

func sensorsinfo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		fh, err := os.Open("./fixtures/E3C246D4I-NL/sensors.json")
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(fh)
		if err != nil {
			log.Fatal(err)
		}
		_, _ = w.Write(b)
	}
}

func biosPOSTCodeinfo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_, _ = w.Write(biosPOSTCodeResponse)
	}
}

func chassisStatusInfo(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_, _ = w.Write(chassisStatusResponse)
	}
}

func session(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// login to BMC
		b, _ := ioutil.ReadAll(r.Body)
		if string(b) == string(loginPayload) {
			// login request needs to be of the right content-typ
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				w.WriteHeader(http.StatusBadRequest)
			}

			w.Header().Set("Content-Type", "application/json")
			http.SetCookie(w, &http.Cookie{Name: "QSESSIONID", Value: "94ed00f482249dd77arIcp6eBBJaik", Path: "/"})
			_, _ = w.Write(loginResponse)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	case "DELETE":
		//1for h, values := range r.Header {
		//1	for _, v := range values {
		//1		fmt.Println(h, v)
		//1	}
		//1}
		if r.Header.Get("X-Csrftoken") != "l5L29IP7" {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
