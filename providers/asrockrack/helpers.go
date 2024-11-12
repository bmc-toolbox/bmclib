package asrockrack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"os"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/constants"
	brrs "github.com/metal-toolbox/bmclib/errors"
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

type biosPOSTCode struct {
	PostStatus int `json:"poststatus"`
	PostData   int `json:"postdata"`
}

// component is part of a payload returned by the inventory info endpoint
type component struct {
	DeviceID                int    `json:"device_id"`
	DeviceName              string `json:"device_name"`
	DeviceType              string `json:"device_type"`
	ProductManufacturerName string `json:"product_manufacturer_name"`
	ProductName             string `json:"product_name"`
	ProductPartNumber       string `json:"product_part_number"`
	ProductVersion          string `json:"product_version"`
	ProductSerialNumber     string `json:"product_serial_number"`
	ProductAssetTag         string `json:"product_asset_tag"`
	ProductExtra            string `json:"product_extra"`
}

// fru is part of a payload returned by the fru info endpoint
type fru struct {
	Component      string
	Version        int    `json:"version"`
	Length         int    `json:"length"`
	Language       int    `json:"language"`
	Manufacturer   string `json:"manufacturer"`
	ProductName    string `json:"product_name"`
	PartNumber     string `json:"part_number"`
	ProductVersion string `json:"product_version"`
	SerialNumber   string `json:"serial_number"`
	AssetTag       string `json:"asset_tag"`
	FruFileID      string `json:"fru_file_id"`
	Type           string `json:"type"`
	CustomFields   string `json:"custom_fields"`
}

// sensor is part of the payload returned by the sensors endpoint
type sensor struct {
	ID                            int     `json:"id"`
	SensorNumber                  int     `json:"sensor_number"`
	Name                          string  `json:"name"`
	OwnerID                       int     `json:"owner_id"`
	OwnerLun                      int     `json:"owner_lun"`
	RawReading                    float64 `json:"raw_reading"`
	Type                          string  `json:"type"`
	TypeNumber                    int     `json:"type_number"`
	Reading                       float64 `json:"reading"`
	SensorState                   int     `json:"sensor_state"`
	DiscreteState                 int     `json:"discrete_state"`
	SettableReadableThreshMask    int     `json:"settable_readable_threshMask"`
	LowerNonRecoverableThreshold  float64 `json:"lower_non_recoverable_threshold"`
	LowerCriticalThreshold        float64 `json:"lower_critical_threshold"`
	LowerNonCriticalThreshold     float64 `json:"lower_non_critical_threshold"`
	HigherNonCriticalThreshold    float64 `json:"higher_non_critical_threshold"`
	HigherCriticalThreshold       float64 `json:"higher_critical_threshold"`
	HigherNonRecoverableThreshold float64 `json:"higher_non_recoverable_threshold"`
	Accessible                    int     `json:"accessible"`
	Unit                          string  `json:"unit"`
}

// Payload to preseve config when updating the BMC firmware
type preserveConfig struct {
	FlashStatus     int `json:"flash_status"` // 1 = full firmware flash, 2 = section based flash, 3 - version compare flash
	PreserveConfig  int `json:"preserve_config"`
	PreserveNetwork int `json:"preserve_network"`
	PreserveUser    int `json:"preserve_user"`
}

// Firmware flash progress
// { "id": 1, "action": "Flashing...", "progress": "12% done         ", "state": 0 }
// { "id": 1, "action": "Flashing...", "progress": "100% done", "state": 0 }
type upgradeProgress struct {
	ID       int    `json:"id,omitempty"`
	Action   string `json:"action,omitempty"`
	Progress string `json:"progress,omitempty"`
	State    int    `json:"state,omitempty"`
}

// Chassis status struct
type chassisStatus struct {
	PowerStatus int `json:"power_status"`
	LEDStatus   int `json:"led_status"`
}

// BIOS upgrade commands
// 2 == configure
// 3 == apply upgrade
type biosUpdateAction struct {
	Action int `json:"action"`
}

var (
	knownPOSTCodes = map[int]string{
		160: constants.POSTStateOS,
		2:   constants.POSTStateBootINIT, // no differentiation between BIOS init and PXE boot
		144: constants.POSTStateUEFI,
		154: constants.POSTStateUEFI,
		178: constants.POSTStateUEFI,
	}
)

func (a *ASRockRack) listUsers(ctx context.Context) ([]*UserAccount, error) {
	resp, statusCode, err := a.queryHTTPS(ctx, "api/settings/users", "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
	}

	accounts := []*UserAccount{}

	err = json.Unmarshal(resp, &accounts)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (a *ASRockRack) createUpdateUser(ctx context.Context, account *UserAccount) error {
	endpoint := "api/settings/users/" + fmt.Sprintf("%d", account.ID)

	payload, err := json.Marshal(account)
	if err != nil {
		return err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	_, statusCode, err := a.queryHTTPS(ctx, endpoint, "PUT", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	return nil
}

// 1 Set BMC to flash mode and prepare flash area
//
// with the BMC set in flash mode, no new logins are accepted
// and only a few endpoints can be queried with the existing session
// one of the few being the install progress/flash status endpoint.
func (a *ASRockRack) setFlashMode(ctx context.Context) error {
	device := common.NewDevice()
	device.Metadata = map[string]string{}
	_ = a.fruAttributes(ctx, &device)

	pConfig := &preserveConfig{}
	// preserve config is needed by e3c256d4i
	switch device.Model {
	case E3C256D4ID_NL:
		pConfig = &preserveConfig{PreserveConfig: 1}
	}

	payload, err := json.Marshal(pConfig)
	if err != nil {
		return err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	_, statusCode, err := a.queryHTTPS(ctx, "api/maintenance/flash", "PUT", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
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
func (a *ASRockRack) uploadFirmware(ctx context.Context, endpoint string, file *os.File) error {
	var size int64
	finfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("unable to determine file size: %w", err)
	}

	size = finfo.Size()

	fieldName, fileName := "fwimage", "image"
	contentLength := multipartSize(fieldName, fileName) + size

	// Before reading the file, rewind to the beginning
	_, _ = file.Seek(0, 0)

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
		_, err = io.Copy(part, file)
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
	_, statusCode, err := a.queryHTTPS(ctx, endpoint, "POST", pipeReader, headers, contentLength)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	return nil
}

// 3. Verify uploaded firmware file - to be invoked after uploadFirmware()
func (a *ASRockRack) verifyUploadedFirmware(ctx context.Context) error {
	_, statusCode, err := a.queryHTTPS(ctx, "api/maintenance/firmware/verification", "GET", nil, nil, 0)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	return nil
}

// 4. Start firmware flashing process - to be invoked after verifyUploadedFirmware
func (a *ASRockRack) upgradeBMC(ctx context.Context) error {
	endpoint := "api/maintenance/firmware/upgrade"

	// preserve all configuration during upgrade, full flash
	pConfig := &preserveConfig{FlashStatus: 1, PreserveConfig: 1, PreserveNetwork: 1, PreserveUser: 1}
	payload, err := json.Marshal(pConfig)
	if err != nil {
		return err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	_, statusCode, err := a.queryHTTPS(ctx, endpoint, "PUT", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	return nil
}

// 5. firmware flash progress
func (a *ASRockRack) flashProgress(ctx context.Context, endpoint string) (*upgradeProgress, error) {
	resp, statusCode, err := a.queryHTTPS(ctx, endpoint, "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
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
func (a *ASRockRack) firmwareInfo(ctx context.Context) (*firmwareInfo, error) {
	resp, statusCode, err := a.queryHTTPS(ctx, "api/asrr/fw-info", "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
	}

	f := &firmwareInfo{}
	err = json.Unmarshal(resp, f)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Query BIOS/UEFI POST code information from the BMC
func (a *ASRockRack) postCodeInfo(ctx context.Context) (*biosPOSTCode, error) {
	resp, statusCode, err := a.queryHTTPS(ctx, "/api/asrr/getbioscode", "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
	}

	b := &biosPOSTCode{}
	err = json.Unmarshal(resp, b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Query the inventory info endpoint
func (a *ASRockRack) inventoryInfoE3C246D41D(ctx context.Context) ([]*component, error) {
	resp, statusCode, err := a.queryHTTPS(ctx, "api/asrr/inventory_info", "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
	}

	components := []*component{}
	err = json.Unmarshal(resp, &components)
	if err != nil {
		return nil, err
	}

	return components, nil
}

// Query the fru info endpoint
func (a *ASRockRack) fruInfo(ctx context.Context) ([]*fru, error) {
	resp, statusCode, err := a.queryHTTPS(ctx, "api/fru", "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
	}

	data := []map[string]*fru{}
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no FRU data returned")
	}

	frus := []*fru{}
	for key, f := range data[0] {
		switch key {
		case "chassis", "board", "product":
			frus = append(frus, &fru{
				Component:      key,
				Version:        f.Version,
				Length:         f.Length,
				Language:       f.Language,
				Manufacturer:   f.Manufacturer,
				ProductName:    f.ProductName,
				PartNumber:     f.PartNumber,
				ProductVersion: f.ProductVersion,
				SerialNumber:   f.SerialNumber,
				AssetTag:       f.SerialNumber,
				FruFileID:      f.FruFileID,
				CustomFields:   f.CustomFields,
				Type:           f.Type,
			})
		}
	}

	return frus, nil
}

// Query the sensors  endpoint
func (a *ASRockRack) sensors(ctx context.Context) ([]*sensor, error) {
	resp, statusCode, err := a.queryHTTPS(ctx, "api/sensors", "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
	}

	sensors := []*sensor{}
	err = json.Unmarshal(resp, &sensors)
	if err != nil {
		return nil, err
	}

	return sensors, nil
}

// Set the BIOS upgrade configuration
//   - preserve current configuration
func (a *ASRockRack) biosUpgradeConfiguration(ctx context.Context) error {
	endpoint := "api/asrr/maintenance/BIOS/configuration"

	// Preserve existing configuration?
	p := biosUpdateAction{Action: 2}
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resp, statusCode, err := a.queryHTTPS(ctx, endpoint, "POST", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	f := &firmwareInfo{}
	return json.Unmarshal(resp, f)
}

// Run BIOS upgrade
func (a *ASRockRack) upgradeBIOS(ctx context.Context) error {
	endpoint := "api/asrr/maintenance/BIOS/upgrade"

	// Run upgrade
	p := biosUpdateAction{Action: 3}
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resp, statusCode, err := a.queryHTTPS(ctx, endpoint, "POST", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	f := &firmwareInfo{}
	return json.Unmarshal(resp, f)
}

// Returns the chassis status object which includes the power state
func (a *ASRockRack) chassisStatusInfo(ctx context.Context) (*chassisStatus, error) {
	resp, statusCode, err := a.queryHTTPS(ctx, "/api/chassis-status", "GET", nil, nil, 0)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("non 200 response: %d", statusCode)
	}

	chassisStatus := chassisStatus{}
	err = json.Unmarshal(resp, &chassisStatus)
	if err != nil {
		return nil, err
	}

	return &chassisStatus, nil
}

// Aquires a session id cookie and a csrf token
func (a *ASRockRack) httpsLogin(ctx context.Context) error {
	urlEndpoint := "api/session"

	// login payload
	payload := []byte(
		fmt.Sprintf("username=%s&password=%s&certlogin=0",
			a.username,
			a.password,
		),
	)

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	resp, statusCode, err := a.queryHTTPS(ctx, urlEndpoint, "POST", bytes.NewReader(payload), headers, 0)
	if err != nil {
		return fmt.Errorf("logging in: %w", err)
	}

	if statusCode == 401 {
		return brrs.ErrLoginFailed
	}

	// Unmarshal login session
	err = json.Unmarshal(resp, a.loginSession)
	if err != nil {
		return fmt.Errorf("unmarshalling response payload: %w", err)
	}

	return nil
}

// Close ends the BMC session
func (a *ASRockRack) httpsLogout(ctx context.Context) error {
	_, statusCode, err := a.queryHTTPS(ctx, "api/session", "DELETE", nil, nil, 0)
	if err != nil {
		return fmt.Errorf("logging out: %w", err)
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response at https logout: %d", statusCode)
	}

	return nil
}

// queryHTTPS run the HTTPS query passing in the required headers
// the / suffix should be excluded from the URLendpoint
// returns - response body, http status code, error if any
func (a *ASRockRack) queryHTTPS(ctx context.Context, endpoint, method string, payload io.Reader, headers map[string]string, contentLength int64) ([]byte, int, error) {
	var body []byte
	var err error
	var req *http.Request

	URL := fmt.Sprintf("https://%s/%s", a.ip, endpoint)
	req, err = http.NewRequestWithContext(ctx, method, URL, payload)
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
	if os.Getenv(constants.EnvEnableDebug) == "true" {
		reqDump, _ := httputil.DumpRequestOut(req, true)
		a.log.V(3).Info("trace", "url", URL, "requestDump", string(reqDump))
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return body, 0, err
	}

	// debug dump response
	if os.Getenv(constants.EnvEnableDebug) == "true" {
		respDump, _ := httputil.DumpResponse(resp, true)
		a.log.V(3).Info("trace", "responseDump", string(respDump))
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return body, 0, err
	}

	defer resp.Body.Close()

	return body, resp.StatusCode, nil
}
