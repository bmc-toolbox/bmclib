package supermicro

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
)

func (c *Client) firmwareInstallBIOS(ctx context.Context, reader io.Reader, fileSize int64) error {
	var err error

	c.log.V(2).Info("set firmware install mode", "ip", c.host, "component", "BIOS", "model", c.model)

	// 0. pre flash mode requisite
	if err := c.checkComponentUpdateMisc(ctx, "preUpdate"); err != nil {
		return err
	}

	// 1. set the device to flash mode - prepares the flash
	err = c.setBIOSFirmwareInstallMode(ctx)
	if err != nil {
		return errors.Wrap(err, ErrFirmwareInstallMode.Error())
	}

	err = c.setBiosUpdateStart(ctx)
	if err != nil {
		return err
	}

	c.log.V(2).Info("uploading firmware", "ip", c.host, "component", "BIOS", "model", c.model)

	// 2. upload firmware image file
	err = c.uploadBIOSFirmware(ctx, reader)
	if err != nil {
		return err
	}

	c.log.V(2).Info("verifying uploaded firmware", "ip", c.host, "component", "BIOS", "model", c.model)

	// 3. BMC verifies the uploaded firmware version
	err = c.verifyBIOSFirmwareVersion(ctx)
	if err != nil {
		return err
	}

	c.log.V(2).Info("initiating firmware install", "ip", c.host, "component", "BIOS", "model", c.model)

	// pre install requisite
	err = c.setBIOSOp(ctx)
	if err != nil {
		return err
	}

	// 4. Run the firmware install process
	err = c.initiateBIOSFirmwareInstall(ctx)
	if err != nil {
		return err
	}

	return nil
}

// checks component update status
func (c *Client) checkComponentUpdateMisc(ctx context.Context, stage string) error {
	var payload, expectResponse []byte

	switch stage {
	case "preUpdate":
		payload = []byte(`op=COMPONENT_UPDATE_MISC.XML&r=(0,0)&_=`)
		// RES=-1 indicates the BMC is not in BIOS update mode
		expectResponse = []byte(`<MISC_INFO RES="-1" SYSOFF="0"/>`)

	case "postUpdate":
		payload = []byte(`op=COMPONENT_UPDATE_MISC.XML&r=(1,0)&_=`)
		// RES=0 indicates the BMC is in BIOS update mode
		expectResponse = []byte(`<MISC_INFO RES="0" SYSOFF="0"/>`)

	// When SYSOFF=1 the system requires a power cycle
	default:
		return errors.New("unknown stage: " + stage)

	}

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return err
	}

	if status != http.StatusOK || !bytes.Contains(body, expectResponse) {
		// this indicates the BMC is in firmware update mode and now requires a reset
		// calling BIOS_UNLOCK.xml doesn't help here
		if stage == "preUpdate" && bytes.Contains(body, []byte(`<MISC_INFO RES="0" SYSOFF="0"/>`)) {
			return bmclibErrs.ErrBMCColdResetRequired
		}

		if bytes.Contains(body, []byte(`<MISC_INFO RES="0" SYSOFF="1"/>`)) {
			return bmclibErrs.ErrHostPowercycleRequired
		}

		if stage == "postUpdate" && bytes.Contains(body, []byte(`<html>`)) {
			return bmclibErrs.ErrSessionExpired
		}

		return unexpectedResponseErr(payload, body, status)
	}

	return nil
}

func (c *Client) setBIOSFirmwareInstallMode(ctx context.Context) error {

	payload := []byte(`op=BIOS_UPLOAD.XML&r=(0,0)&_=`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return unexpectedResponseErr(payload, body, status)
	}

	switch {
	case bytes.Contains(body, []byte(`LOCK_FW_UPLOAD RES="0"`)):
		// This response indicates another web session that initiated the firmware upload has the lock,
		// the BMC cannot be reset through a web session, nor can any other user obtain the firmware upload lock.
		// Since the firmware upload lock is associated with the cookie that initiated the request only the initiating session can cancel it.
		//
		// The only way to get out of this situation is through an IPMI (or redfish?) based BMC cold reset.
		///
		// The caller must check if a firmware update is in progress before proceeding with the reset.
		//
		// If after multiple calls to check the install progress - the progress seems stalled at 1%
		// it indicates no update was active, and the BMC can be reset.
		//
		// <IPMI><percent>1</percent></IPMI>
		return errors.Wrap(
			bmclibErrs.ErrBMCColdResetRequired,
			"firmware upload mode active, another session may have initiated an install",
		)

	case bytes.Contains(body, []byte(`LOCK_FW_UPLOAD RES="1"`)):
		return nil
	default:
		return unexpectedResponseErr(payload, body, status)
	}

}

func (c *Client) setBiosUpdateStart(ctx context.Context) error {
	payload := []byte(`op=BIOS_UPDATE_START.XML&r=(1,0)&_=`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return err
	}

	// yep, the endpoint returns 500 even when successful
	if status != http.StatusOK && status != 500 {
		return unexpectedResponseErr(payload, body, status)
	}

	return nil
}

// ------WebKitFormBoundaryXIAavwG4xzohdB6k
// Content-Disposition: form-data; name="bios_rom"; filename="BIOS_X11SCM-1B0F_20220916_1.9_STDsp.bin"
// Content-Type: application/macbinary
//
// ------WebKitFormBoundaryXIAavwG4xzohdB6k
// Content-Disposition: form-data; name="CSRF-TOKEN"
//
// OO8+cjamaZZOMf6ZiGDY3Lw+7O20r5lR8aI8ByuTo3E
// ------WebKitFormBoundaryXIAavwG4xzohdB6k--
func (c *Client) uploadBIOSFirmware(ctx context.Context, fwReader io.Reader) error {
	var payloadBuffer bytes.Buffer
	var err error

	formParts := []struct {
		name string
		data io.Reader
	}{
		{
			name: "bios_rom",
			data: fwReader,
		},
		{
			name: "csrf-token",
			data: bytes.NewBufferString(c.csrfToken),
		},
	}

	payloadWriter := multipart.NewWriter(&payloadBuffer)

	for _, part := range formParts {
		var partWriter io.Writer

		switch part.name {
		case "bios_rom":
			file, ok := part.data.(*os.File)
			if !ok {
				return errors.Wrap(ErrMultipartForm, "expected io.Reader on firmware image file")
			}

			if partWriter, err = payloadWriter.CreateFormFile(part.name, filepath.Base(file.Name())); err != nil {
				return errors.Wrap(ErrMultipartForm, err.Error())
			}

		case "csrf-token":
			// Add csrf token field
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition", `form-data; name="CSRF-TOKEN"`)

			if partWriter, err = payloadWriter.CreatePart(h); err != nil {
				return errors.Wrap(ErrMultipartForm, err.Error())
			}
		default:
			return errors.Wrap(ErrMultipartForm, "unexpected form part: "+part.name)
		}

		if _, err = io.Copy(partWriter, part.data); err != nil {
			return err
		}
	}
	payloadWriter.Close()

	resp, statusCode, err := c.query(
		ctx,
		"cgi/bios_upload.cgi",
		http.MethodPost,
		bytes.NewReader(payloadBuffer.Bytes()),
		map[string]string{"Content-Type": payloadWriter.FormDataContentType()},
		0,
	)

	if err != nil {
		return errors.Wrap(ErrMultipartForm, err.Error())
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d %s", statusCode, resp)
	}

	return nil
}

func (c *Client) verifyBIOSFirmwareVersion(ctx context.Context) error {
	payload := []byte(`op=BIOS_UPDATE_CHECK.XML&r=(0,0)&_=`)
	expectResponse := []byte(`<BIOS_UPDATE_CHECK RES="00"/>`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return err
	}

	if status != http.StatusOK || !bytes.Contains(body, expectResponse) {
		return unexpectedResponseErr(payload, body, status)
	}

	payload = []byte(`op=BIOS_REV.XML&_=`)
	expectResponse = []byte(`<BIOS_Rev OldRev`)

	body, status, err = c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return err
	}

	if status != http.StatusOK || !bytes.Contains(body, expectResponse) {
		return unexpectedResponseErr(payload, body, status)
	}

	return nil
}

func (c *Client) setBIOSOp(ctx context.Context) error {
	payload := []byte(`op=BIOS_OPTION.XML&_=`)
	expectResponse := []byte(`<BIOS_OP Res="0"/>`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return err
	}

	if status != http.StatusOK || !bytes.Contains(body, expectResponse) {
		return unexpectedResponseErr(payload, body, status)
	}

	return nil
}

func (c *Client) initiateBIOSFirmwareInstall(ctx context.Context) error {
	// save all current SMBIOS, NVRAM, ME configuration
	payload := []byte(`op=main_biosupdate&_=`)
	expectResponse := []byte(`ok`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	// don't spend much time on this call since it doesn't return and holds the connection.
	sctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	body, status, err := c.query(sctx, "cgi/op.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		// this endpoint generally times out - its expected
		if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "operation timed out") {
			return nil
		}

		return err
	}

	if status != http.StatusOK || !bytes.Contains(body, expectResponse) {
		return unexpectedResponseErr(payload, body, status)
	}

	return nil
}

func (c *Client) setBIOSUpdateDone(ctx context.Context) error {
	payload := []byte(`op=BIOS_UPDATE_DONE.XML&r=(1,0)&_=`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return err
	}

	// yep, the endpoint returns 500 even when successful
	if status != http.StatusOK && status != 500 {
		return unexpectedResponseErr(payload, body, status)
	}

	return nil
}

// statusBIOSFirmwareInstall returns the status of the firmware install process
func (c *Client) statusBIOSFirmwareInstall(ctx context.Context) (string, error) {
	payload := []byte(`fwtype=1&_`)

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}
	resp, status, err := c.query(ctx, "cgi/upgrade_process.cgi", http.MethodPost, bytes.NewReader(payload), headers, 0)
	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, err.Error())
	}

	if status != http.StatusOK {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "Unexpected status code: "+strconv.Itoa(status))
	}

	switch {
	// 1% indicates the file has been uploaded and the firmware install is not yet initiated
	case bytes.Contains(resp, []byte("<status>0</status>")) && bytes.Contains(resp, []byte("<percent>1</percent>")):
		return constants.FirmwareInstallFailed, bmclibErrs.ErrBMCColdResetRequired

	// 0% along with the check on the component endpoint indicates theres no update in progress
	case (bytes.Contains(resp, []byte("<status>0</status>")) && bytes.Contains(resp, []byte("<percent>0</percent>"))):
		if err := c.checkComponentUpdateMisc(ctx, "postUpdate"); err != nil {
			if errors.Is(err, bmclibErrs.ErrHostPowercycleRequired) {
				return constants.FirmwareInstallPowerCyleHost, nil
			}
		}

		return constants.FirmwareInstallComplete, nil

	// status 0 and 100% indicates the update is complete and requires a few post update calls
	case bytes.Contains(resp, []byte("<status>0</status>")) && bytes.Contains(resp, []byte("<percent>100</percent>")):
		// notifies the BMC the BIOS update is done
		if err := c.setBIOSUpdateDone(ctx); err != nil {
			return "", err
		}

		// tells the BMC it can get out of the BIOS update mode
		if err := c.checkComponentUpdateMisc(ctx, "postUpdate"); err != nil {
			if errors.Is(err, bmclibErrs.ErrHostPowercycleRequired) {
				return constants.FirmwareInstallPowerCyleHost, nil
			}

			return constants.FirmwareInstallPowerCyleHost, err
		}

		return constants.FirmwareInstallPowerCyleHost, nil

	// status 8 and percent 0 indicates its initializing the update
	case bytes.Contains(resp, []byte("<status>8</status>")) && bytes.Contains(resp, []byte("<percent>0</percent>")):
		return constants.FirmwareInstallRunning, nil

	// status 8 and any other percent value indicates its running
	case bytes.Contains(resp, []byte("<status>8</status>")) && bytes.Contains(resp, []byte("<percent>")):
		return constants.FirmwareInstallRunning, nil

	case bytes.Contains(resp, []byte(`<html>`)):
		return constants.FirmwareInstallUnknown, bmclibErrs.ErrSessionExpired

	default:
		return constants.FirmwareInstallUnknown, nil
	}
}
