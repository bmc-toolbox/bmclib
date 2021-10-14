package asrockrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
)

var (
	loginPayload           = []byte(`username=foo&password=bar&certlogin=0`)
	loginResponse          = []byte(`{ "ok": 0, "privilege": 4, "extendedpriv": 259, "racsession_id": 10, "remote_addr": "136.144.50.145", "server_name": "10.230.148.171", "server_addr": "10.230.148.171", "HTTPSEnabled": 1, "CSRFToken": "l5L29IP7" }`)
	fwinfoResponse         = []byte(`{ "BMC_fw_version": "0.01.00", "BIOS_fw_version": "L2.07B", "ME_fw_version": "5.1.3.78", "Micro_Code_version": "000000ca", "CPLD_version": "N\/A", "CM_version": "0.13.01", "BPB_version": "0.0.002.0", "Node_id": "2" }`)
	fwUploadResponse       = []byte(`{"cc": 0}`)
	fwVerificationResponse = []byte(`[ { "id": 1, "current_image_name": "ast2500e", "current_image_version1": "0.01.00", "current_image_version2": "", "new_image_version": "0.03.00", "section_status": 0, "verification_status": 5 } ]`)
	fwUpgradeProgress      = []byte(`{ "id": 1, "action": "Flashing...", "progress": "__PERCENT__% done         ", "state": __STATE__ }`)
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
	tLog := logrusr.NewLogger(l)
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

	// fw update endpoints - in order of invocation
	handler.HandleFunc("/api/maintenance/flash", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/firmware", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/firmware/verification", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/firmware/upgrade", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/firmware/flash-progress", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/maintenance/reset", bmcFirmwareUpgrade)
	handler.HandleFunc("/api/asrr/maintenance/BIOS/firmware", biosFirmwareUpgrade)

	return httptest.NewTLSServer(handler)
}

func index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_, _ = w.Write([]byte(`ASRockRack`))
	}
}

func biosFirmwareUpgrade(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s -> %s\n", r.Method, r.RequestURI)
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
	fmt.Printf("%s -> %s\n", r.Method, r.RequestURI)
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
