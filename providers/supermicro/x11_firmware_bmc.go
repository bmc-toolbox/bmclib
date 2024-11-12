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

	"github.com/pkg/errors"

	"github.com/metal-toolbox/bmclib/constants"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
)

var (
	ErrFirmwareInstallMode = errors.New("firmware install mode error")
	ErrMultipartForm       = errors.New("multipart form error")
)

func (c *x11) firmwareUploadBMC(ctx context.Context, reader io.Reader) error {
	c.log.V(2).Info("setting device to firmware install mode", "ip", c.host, "component", "BMC", "model", c.model)

	// 1. set the device to flash mode - prepares the flash
	err := c.setBMCFirmwareInstallMode(ctx)
	if err != nil {
		return err
	}

	c.log.V(2).Info("uploading firmware", "ip", c.host, "component", "BMC", "model", c.model)

	// 2. upload firmware image file
	err = c.uploadBMCFirmware(ctx, reader)
	if err != nil {
		return err
	}

	c.log.V(2).Info("verifying uploaded firmware", "ip", c.host, "component", "BMC", "model", c.model)

	// 3. BMC verifies the uploaded firmware version
	return c.verifyBMCFirmwareVersion(ctx)
}

func (c *x11) setBMCFirmwareInstallMode(ctx context.Context) error {
	payload := []byte(`op=LOCK_UPLOAD_FW.XML&r=(0,0)&_=`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return errors.Wrap(ErrFirmwareInstallMode, err.Error())
	}

	if status != http.StatusOK {
		return errors.Wrap(ErrFirmwareInstallMode, "Unexpected status code: "+strconv.Itoa(status))
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
			"unable to acquire lock for firmware upload, check if an update is in progress",
		)

	case bytes.Contains(body, []byte(`LOCK_FW_UPLOAD RES="1"`)):
		return nil
	default:
		return errors.Wrap(ErrFirmwareInstallMode, "set firmware install mode returned unexpected response body")
	}

}

// -----------------------------212212001131894333502018521064
// Content-Disposition: form-data; name="fw_image"; filename="BMC_X11AST2500-4101MS_20221020_01.74.09_STDsp.bin"
// Content-Type: application/macbinary
//
// ... contents...
//
// -----------------------------348113760313214626342869148824
// Content-Disposition: form-data; name="CSRF-TOKEN"
//
// JhVe1BUiWzOVQdvXUKn7ClsQ5xffq8StMOxG7ZNlpKs
// -----------------------------348113760313214626342869148824--
func (c *x11) uploadBMCFirmware(ctx context.Context, fwReader io.Reader) error {
	var payloadBuffer bytes.Buffer
	var err error

	type form struct {
		name string
		data io.Reader
	}

	formParts := []form{
		{
			name: "fw_image",
			data: fwReader,
		},
	}

	if c.csrfToken != "" {
		formParts = append(formParts, form{
			name: "csrf-token",
			data: bytes.NewBufferString(c.csrfToken),
		})
	}

	payloadWriter := multipart.NewWriter(&payloadBuffer)

	for _, part := range formParts {
		var partWriter io.Writer

		switch part.name {
		case "fw_image":
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
			// BMCs with newer firmware (>=1.74.09) accept the form with this name value
			// h.Set("Content-Disposition", `form-data; name="CSRF-TOKEN"`)
			//
			// the BMCs running older firmware (<=1.23.06) versions expects the name value in this format
			// this seems to be backwards compatible with the newer firmware.
			h.Set("Content-Disposition", `form-data; name="CSRF_TOKEN"`)

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
		"cgi/oem_firmware_upload.cgi",
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

func (c *x11) verifyBMCFirmwareVersion(ctx context.Context) error {
	errUnexpectedResponse := errors.New("unexpected response")

	payload := []byte(`op=UPLOAD_FW_VERSION.XML&r=(0,0)&_=`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return err
	}

	if status != 200 {
		return errors.Wrap(ErrFirmwareInstallMode, "Unexpected status code: "+strconv.Itoa(status))
	}

	if !bytes.Contains(body, []byte(`FW_VERSION NEW`)) {
		return errors.Wrap(errUnexpectedResponse, string(body))
	}

	return nil
}

// initiate BMC firmware install process
func (c *x11) initiateBMCFirmwareInstall(ctx context.Context) error {
	// preserve all configuration, sensor data and SSL certs(?) during upgrade
	payload := "op=main_fwupdate&preserve_config=1&preserve_sdr=1&preserve_ssl=1"

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}

	// don't spend much time on this call since it doesn't return and holds the connection.
	sctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, status, err := c.query(sctx, "cgi/op.cgi", http.MethodPost, bytes.NewBufferString(payload), headers, 0)
	if err != nil {
		// this operation causes the BMC to go AWOL and not send any response
		// so we ignore the error here, the caller can invoke FirmwareInstallStatus in the same session,
		// to check the install status to determine install progress.

		// whats returned is a *url.Error{} and errors.Is(err, context.DeadlineExceeded) doesn't seem to match
		// so a string contains it is.
		if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "operation timed out") {
			return nil
		}

		return err
	}

	if status != 200 {
		return errors.Wrap(ErrFirmwareInstallMode, "Unexpected status code: "+strconv.Itoa(status))
	}

	return nil
}

// statusBMCFirmwareInstall returns the status of the firmware install process
func (c *x11) statusBMCFirmwareInstall(ctx context.Context) (state constants.TaskState, status string, err error) {
	payload := []byte(`fwtype=0&_`)

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}
	resp, httpStatus, err := c.query(ctx, "cgi/upgrade_process.cgi", http.MethodPost, bytes.NewReader(payload), headers, 0)
	if err != nil {
		return constants.Unknown, "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, err.Error())
	}

	if httpStatus != http.StatusOK {
		return constants.Unknown, "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "Unexpected http status code: "+strconv.Itoa(httpStatus))
	}

	// if theres html or no <percent> xml in the response, the session expired
	// at the end of the install the BMC resets itself and the response is in HTML.
	if bytes.Contains(resp, []byte(`<html>`)) || !bytes.Contains(resp, []byte(`<percent>`)) {
		// reopen session here, check firmware install status
		return constants.Unknown, "session expired/unexpected response", bmclibErrs.ErrSessionExpired
	}

	// as long as the response is xml, the firmware install is running
	part := strings.Split(string(resp), "<percent>")[1]
	percent := strings.Split(part, "</percent>")[0]
	percent += "%"

	switch percent {
	// TODO:
	// X11DPH-T - returns percent 0 all the time
	//
	// 0% indicates its either not running or complete
	case "0%", "100%":
		return constants.Complete, percent, nil
	// until 2% its initializing
	case "1%", "2%":
		return constants.Initializing, percent, nil
	// any other percent value indicates its active
	default:
		return constants.Running, percent, nil
	}
}
