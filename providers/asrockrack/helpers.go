package asrockrack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/bmc-toolbox/bmclib/errors"
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

// BIOS upgrade commands
// 2 == configure
// 3 == apply upgrade
type biosUpdateAction struct {
	Action int `json:"action"`
}

// 1 Set BMC to flash mode and prepare flash area
// at this point all logged in sessions are terminated
// and no logins are permitted
func (a *ASRockRack) setFlashMode() error {

	endpoint := "api/maintenance/flash"

	_, statusCode, err := a.queryHTTPS(endpoint, "PUT", nil, nil, 0)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	a.resetRequired = true

	return nil
}

func multipartSize(fieldname, filename string) int64 {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)
	_, _ = form.CreateFormFile(fieldname, filename)
	_ = form.Close()
	return int64(body.Len())
}

// 2 Upload the firmware file
func (a *ASRockRack) uploadFirmware(endpoint string, fwReader io.Reader, fileSize int64) error {

	fieldName, fileName := "fwimage", "image"
	contentLength := multipartSize(fieldName, fileName) + fileSize

	// setup pipe
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	// initiate a mulitpart writer
	form := multipart.NewWriter(pipeWriter)

	errCh := make(chan error, 1)
	go func() {
		defer pipeWriter.Close()

		// create form part
		part, err := form.CreateFormFile(fieldName, fileName)
		if err != nil {
			errCh <- err
			return
		}

		// copy from source into form part writer
		_, err = io.Copy(part, fwReader)
		if err != nil {
			errCh <- err
			return
		}

		// add terminating boundary to multipart form
		errCh <- form.Close()
	}()

	// multi-part content type
	headers := map[string]string{
		"Content-Type": form.FormDataContentType(),
	}

	// POST payload
	_, statusCode, err := a.queryHTTPS(endpoint, "POST", pipeReader, headers, contentLength)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	return nil
}

// 3. Verify uploaded firmware file - to be invoked after uploadFirmware()
func (a *ASRockRack) verifyUploadedFirmware() error {

	endpoint := "api/maintenance/firmware/verification"

	_, statusCode, err := a.queryHTTPS(endpoint, "GET", nil, nil, 0)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("non 200 response: %d", statusCode)
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
	_, statusCode, err := a.queryHTTPS(endpoint, "PUT", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	return nil

}

// 4. reset BMC
func (a *ASRockRack) reset() error {

	endpoint := "api/maintenance/reset"

	_, statusCode, err := a.queryHTTPS(endpoint, "POST", nil, nil, 0)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	return nil

}

// 5. firmware flash progress
func (a *ASRockRack) flashProgress(endpoint string) (*upgradeProgress, error) {

	resp, statusCode, err := a.queryHTTPS(endpoint, "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
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

	resp, statusCode, err := a.queryHTTPS(endpoint, "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
	}

	f := &firmwareInfo{}
	err = json.Unmarshal(resp, f)
	if err != nil {
		return nil, err
	}

	return f, nil

}

// Set the BIOS upgrade configuration
//  - preserve current configuration
func (a *ASRockRack) biosUpgradeConfiguration() error {

	endpoint := "api/asrr/maintenance/BIOS/configuration"

	// Preserve existing configuration?
	p := biosUpdateAction{Action: 2}
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resp, statusCode, err := a.queryHTTPS(endpoint, "POST", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	f := &firmwareInfo{}
	err = json.Unmarshal(resp, f)
	if err != nil {
		return err
	}

	return nil

}

// Run BIOS upgrade
func (a *ASRockRack) biosUpgrade() error {

	endpoint := "api/asrr/maintenance/BIOS/upgrade"

	// Run upgrade
	p := biosUpdateAction{Action: 3}
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resp, statusCode, err := a.queryHTTPS(endpoint, "POST", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return err
	}

	if statusCode != 200 {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	f := &firmwareInfo{}
	err = json.Unmarshal(resp, f)
	if err != nil {
		return err
	}

	return nil

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

	resp, statusCode, err := a.queryHTTPS(urlEndpoint, "POST", bytes.NewReader(payload), headers, 0)
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

	_, statusCode, err := a.queryHTTPS(urlEndpoint, "DELETE", nil, nil, 0)
	if err != nil {
		return fmt.Errorf("Error logging out: " + err.Error())
	}

	if err != nil {
		return fmt.Errorf("Error logging out: " + err.Error())
	}

	if statusCode != 200 {
		return fmt.Errorf("non 200 response at https logout: %d", statusCode)
	}

	return nil
}

// queryHTTPS run the HTTPS query passing in the required headers
// the / suffix should be excluded from the URLendpoint
// returns - response body, http status code, error if any
func (a *ASRockRack) queryHTTPS(URLendpoint, method string, payload io.Reader, headers map[string]string, contentLength int64) ([]byte, int, error) {

	var body []byte
	var err error
	var req *http.Request

	URL := fmt.Sprintf("https://%s/%s", a.ip, URLendpoint)
	req, err = http.NewRequest(method, URL, payload)
	if err != nil {
		return nil, 0, err
	}

	// add headers
	req.Header.Add("X-CSRFTOKEN", a.loginSession.CSRFToken)
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// Content-Length headers are ignored, unless defined in this manner
	// https://go.googlesource.com/go/+/go1.16/src/net/http/request.go#161
	// https://go.googlesource.com/go/+/go1.16/src/net/http/request.go#88
	if contentLength > 0 {
		req.ContentLength = contentLength
	}

	// debug dump request
	if os.Getenv("BMCLIB_LOG_LEVEL") == "trace" {
		reqDump, _ := httputil.DumpRequestOut(req, true)
		a.log.V(3).Info("trace", "url", URL, "requestDump", string(reqDump))
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return body, 0, err
	}

	// debug dump response
	if os.Getenv("BMCLIB_LOG_LEVEL") == "trace" {
		respDump, _ := httputil.DumpResponse(resp, true)
		a.log.V(3).Info("trace", "responseDump", string(respDump))
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, 0, err
	}

	defer resp.Body.Close()

	return body, resp.StatusCode, nil

}
