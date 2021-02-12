package asrockrack

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"
	"path"

	"github.com/bmc-toolbox/bmclib/errors"
	log "github.com/sirupsen/logrus"
)

// API session setup response payload
type loginSession struct {
	CSRFToken         string `json:"csrftoken,omitempty"`
	Privilege         int    `json:"privilege,omitempty"`
	RACSessionID      int    `json:"racsession_id,omitempty"`
	ExtendedPrivilege int    `json:"extendedpriv,omitempty"`
}

// Firmware info endpoint response payload
type firmwareInfo struct {
	BMCVersion       string `json:"BMC_fw_version"`
	BIOSVersion      string `json:"BIOS_fw_version"`
	MEVersion        string `json:"ME_fw_version"`
	MicrocodeVersion string `json:"Micro_Code_version"`
	CPLDVersion      string `json:"CPLD_version"`
	CMVersion        string `json:"CM_version"`
	BPBVersion       string `json:"BPB_version"`
	NodeID           string `json:"Node_id"`
}

// Payload to preseve config when updating the BMC firmware
type preserveConfig struct {
	FlashStatus     int `json:"flash_status"` // 1 = full firmware flash, 2 = section based flash, 3 - version compare flash
	PreserveConfig  int `json:"preserve_config"`
	PreserveNetwork int `json:"preserve_network"`
	PreserveUser    int `json:"preserve_user"`
}

// Firmware flash progress
//{ "id": 1, "action": "Flashing...", "progress": "12% done         ", "state": 0 }
//{ "id": 1, "action": "Flashing...", "progress": "100% done", "state": 0 }
type upgradeProgress struct {
	ID       int    `json:"id,omitempty"`
	Action   string `json:"action,omitempty"`
	Progress string `json:"progress,omitempty"`
	State    int    `json:"state,omitempty"`
}

// 1 Set BMC to flash mode and prepare flash area
// at this point all logged in sessions are terminated
// and no logins are permitted
func (a *ASRockRack) setFlashMode() error {

	endpoint := "api/maintenance/flash"

	_, statusCode, err := a.queryHTTPS(endpoint, "PUT", nil, nil)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("Non 200 response: %d", statusCode)
	}

	a.resetRequired = true

	return nil
}

// 2 Upload the firmware file
func (a *ASRockRack) uploadFirmware(filePath string) error {

	endpoint := "api/maintenance/firmware"

	// setup a buffer for our multipart form
	var form bytes.Buffer
	w := multipart.NewWriter(&form)

	// create form data from update image
	fwWriter, err := w.CreateFormFile("fwimage", path.Base(filePath))
	if err != nil {
		return err
	}

	// open file handle
	fh, err := os.Open(filePath)
	if err != nil {
		return err
	}

	// copy file contents into form payload
	fReader := bufio.NewReader(fh)
	_, err = io.Copy(fwWriter, fReader)
	if err != nil {
		return err
	}

	// multi-part content type
	headers := map[string]string{"Content-Type": w.FormDataContentType()}

	// close multipart writer - adds the teminating boundary.
	w.Close()

	// POST payload
	_, statusCode, err := a.queryHTTPS(endpoint, "POST", form.Bytes(), headers)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("Non 200 response: %d", statusCode)
	}

	return nil
}

// 3. Verify uploaded firmware file - to be invoked after uploadFirmware()
func (a *ASRockRack) verifyUploadedFirmware() error {

	endpoint := "api/maintenance/firmware/verification"

	_, statusCode, err := a.queryHTTPS(endpoint, "GET", nil, nil)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("Non 200 response: %d", statusCode)
	}

	return nil

}

// 4. Start firmware flashing process - to be invoked after verifyUploadedFirmware
func (a *ASRockRack) upgradeBMC() error {

	endpoint := "api/maintenance/firmware/upgrade"

	// preserve all configuration during upgrade, full flash
	pConfig := &preserveConfig{PreserveConfig: 1, FlashStatus: 1}
	payload, err := json.Marshal(pConfig)
	if err != nil {
		return err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	_, statusCode, err := a.queryHTTPS(endpoint, "PUT", payload, headers)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("Non 200 response: %d", statusCode)
	}

	return nil

}

// 4. reset BMC
func (a *ASRockRack) reset() error {

	endpoint := "api/maintenance/reset"

	_, statusCode, err := a.queryHTTPS(endpoint, "GET", nil, nil)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("Non 200 response: %d", statusCode)
	}

	return nil

}

// 5. firmware flash progress
func (a *ASRockRack) flashProgress() (*upgradeProgress, error) {

	endpoint := "api/maintenance/firmware/flash-progress"

	resp, statusCode, err := a.queryHTTPS(endpoint, "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("Non 200 response: %d", statusCode)
	}

	p := &upgradeProgress{}
	err = json.Unmarshal(resp, p)
	if err != nil {
		return nil, err
	}

	return p, nil

}

// Query firmware information from the BMC
func (a *ASRockRack) firmwareInfo() (*firmwareInfo, error) {

	endpoint := "api/asrr/fw-info"

	resp, statusCode, err := a.queryHTTPS(endpoint, "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("Non 200 response: %d", statusCode)
	}

	f := &firmwareInfo{}
	err = json.Unmarshal(resp, f)
	if err != nil {
		return nil, err
	}

	return f, nil

}

// Aquires a session id cookie and a csrf token
func (a *ASRockRack) httpsLogin() error {

	urlEndpoint := "api/session"

	// login payload
	payload := []byte(
		fmt.Sprintf("username=%s&password=%s&certlogin=0",
			a.username,
			a.password,
		),
	)

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	resp, statusCode, err := a.queryHTTPS(urlEndpoint, "POST", payload, headers)
	if err != nil {
		return fmt.Errorf("Error logging in: " + err.Error())
	}

	if statusCode == 401 {
		return errors.ErrLoginFailed
	}

	// Unmarshal login session
	err = json.Unmarshal(resp, a.loginSession)
	if err != nil {
		return fmt.Errorf("error unmarshalling response payload: " + err.Error())
	}

	return nil
}

// Close ends the BMC session
func (a *ASRockRack) httpsLogout() error {

	urlEndpoint := "api/session"
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": VendorID, "ip": a.ip}).Debug("logout from bmc")

	_, statusCode, err := a.queryHTTPS(urlEndpoint, "DELETE", nil, nil)
	if err != nil {
		return fmt.Errorf("Error logging out: " + err.Error())
	}

	if err != nil {
		return fmt.Errorf("Error logging out: " + err.Error())
	}

	if statusCode != 200 {
		return fmt.Errorf("Non 200 response at https logout: %d", statusCode)
	}

	return nil
}

// queryHTTPS run the HTTPS query passing in the required headers
// the / suffix should be excluded from the URLendpoint
// returns - response body, http status code, error if any
func (a *ASRockRack) queryHTTPS(URLendpoint, method string, payload []byte, headers map[string]string) ([]byte, int, error) {

	var body []byte
	var err error
	var req *http.Request

	URL := fmt.Sprintf("https://%s/%s", a.ip, URLendpoint)
	if len(payload) > 0 {
		req, err = http.NewRequest(method, URL, bytes.NewReader(payload))
	} else {
		req, err = http.NewRequest(method, URL, nil)
	}

	//
	if err != nil {
		return nil, 0, err
	}

	// add headers
	req.Header.Add("X-CSRFTOKEN", a.loginSession.CSRFToken)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// debug dump request
	reqDump, _ := httputil.DumpRequestOut(req, true)
	a.log.V(3).Info("trace", "url", URL, "requestDump", string(reqDump))

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return body, 0, err
	}

	// debug dump response
	respDump, _ := httputil.DumpResponse(resp, true)
	a.log.V(3).Info("trace", "responseDump", string(respDump))

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, 0, err
	}

	defer resp.Body.Close()

	return body, resp.StatusCode, nil

}
