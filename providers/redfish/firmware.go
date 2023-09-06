package redfish

import (
	"bytes"
	"context"
	"encoding/json"
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
	gofishrf "github.com/stmcginnis/gofish/redfish"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal"
)

var (
	errInsufficientCtxTimeout = errors.New("remaining context timeout insufficient to install firmware")
	errMultiPartPayload       = errors.New("error preparing multipart payload")
)

// SupportedFirmwareApplyAtValues returns the supported redfish firmware applyAt values
func SupportedFirmwareApplyAtValues() []string {
	return []string{
		constants.FirmwareApplyImmediate,
		constants.FirmwareApplyOnReset,
	}
}

// FirmwareInstall uploads and initiates the firmware install process
func (c *Conn) FirmwareInstall(ctx context.Context, component, applyAt string, forceInstall bool, reader io.Reader) (taskID string, err error) {
	// limit to *os.File until theres a need for other types of readers
	updateFile, ok := reader.(*os.File)
	if !ok {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "method expects an *os.File object")
	}

	// validate firmware update mechanism is supported
	err = c.firmwareUpdateCompatible(ctx)
	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
	}

	// validate applyAt parameter
	if !internal.StringInSlice(applyAt, SupportedFirmwareApplyAtValues()) {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "invalid applyAt parameter: "+applyAt)
	}

	// expect atleast 10 minutes left in the deadline to proceed with the update
	//
	// this gives the BMC enough time to have the firmware uploaded and return a response to the client.
	ctxDeadline, _ := ctx.Deadline()
	if time.Until(ctxDeadline) < 10*time.Minute {
		return "", errors.Wrap(errInsufficientCtxTimeout, " "+time.Until(ctxDeadline).String())
	}

	// list redfish firmware install task if theres one present
	task, err := c.GetFirmwareInstallTaskQueued(ctx, component)
	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
	}

	if task != nil {
		msg := fmt.Sprintf("task for %s firmware install present: %s", component, task.ID)
		c.Log.V(2).Info("warn", msg)

		if forceInstall {
			err = c.purgeQueuedFirmwareInstallTask(ctx, component)
			if err != nil {
				return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
			}
		} else {
			return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, msg)
		}
	}

	updateParameters, err := json.Marshal(struct {
		Targets            []string `json:"Targets"`
		RedfishOpApplyTime string   `json:"@Redfish.OperationApplyTime"`
		Oem                struct{} `json:"Oem"`
	}{
		[]string{},
		applyAt,
		struct{}{},
	})

	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
	}

	// override the gofish HTTP client timeout,
	// since the context timeout is set at Open() and is at a lower value than required for this operation.
	//
	// record the http client timeout to be restored
	httpClientTimeout := c.redfishwrapper.HttpClientTimeout()
	defer func() {
		c.redfishwrapper.SetHttpClientTimeout(httpClientTimeout)
	}()

	c.redfishwrapper.SetHttpClientTimeout(time.Until(ctxDeadline))

	payload := &multipartPayload{
		updateParameters: bytes.NewReader(updateParameters),
		updateFile:       updateFile,
	}

	resp, err := c.runRequestWithMultipartPayload(http.MethodPost, "/redfish/v1/UpdateService/MultipartUpload", payload)
	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, err.Error())
	}

	if resp.StatusCode != http.StatusAccepted {
		return "", errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"non 202 status code returned: "+strconv.Itoa(resp.StatusCode),
		)
	}

	// The response contains a location header pointing to the task URI
	// Location: /redfish/v1/TaskService/Tasks/JID_467696020275
	if strings.Contains(resp.Header.Get("Location"), "JID_") {
		taskID = strings.Split(resp.Header.Get("Location"), "JID_")[1]
	}

	return taskID, nil
}

type multipartPayload struct {
	updateParameters io.Reader
	updateFile       *os.File
}

// FirmwareInstallStatus returns the status of the firmware install task queued
func (c *Conn) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (state string, err error) {
	vendor, _, err := c.DeviceVendorModel(ctx)
	if err != nil {
		return state, errors.Wrap(err, "unable to determine device vendor, model attributes")
	}

	var task *gofishrf.Task
	switch {
	case strings.Contains(vendor, constants.Dell):
		task, err = c.dellJobAsRedfishTask(taskID)
	default:
		err = errors.Wrap(
			bmclibErrs.ErrNotImplemented,
			"FirmwareInstallStatus() for vendor: "+vendor,
		)
	}

	if err != nil {
		return state, err
	}

	if task == nil {
		return state, errors.New("failed to lookup task status for task ID: " + taskID)
	}

	state = strings.ToLower(string(task.TaskState))

	// so much for standards...
	switch state {
	case "starting", "downloading", "downloaded":
		return constants.FirmwareInstallInitializing, nil
	case "running", "stopping", "cancelling", "scheduling":
		return constants.FirmwareInstallRunning, nil
	case "pending", "new":
		return constants.FirmwareInstallQueued, nil
	case "scheduled":
		return constants.FirmwareInstallPowerCyleHost, nil
	case "interrupted", "killed", "exception", "cancelled", "suspended", "failed":
		return constants.FirmwareInstallFailed, nil
	case "completed":
		return constants.FirmwareInstallComplete, nil
	default:
		return constants.FirmwareInstallUnknown + ": " + state, nil
	}

}

// firmwareUpdateCompatible retuns an error if the firmware update process for the BMC is not supported
func (c *Conn) firmwareUpdateCompatible(ctx context.Context) (err error) {
	updateService, err := c.redfishwrapper.UpdateService()
	if err != nil {
		return err
	}

	// TODO: check for redfish version

	// update service disabled
	if !updateService.ServiceEnabled {
		return errors.Wrap(bmclibErrs.ErrRedfishUpdateService, "service disabled")
	}

	// for now we expect multipart HTTP push update support
	if updateService.MultipartHTTPPushURI == "" {
		return errors.Wrap(bmclibErrs.ErrRedfishUpdateService, "Multipart HTTP push updates not supported")
	}

	return nil
}

// pipeReaderFakeSeeker wraps the io.PipeReader and implements the io.Seeker interface
// to meet the API requirements for the Gofish client https://github.com/stmcginnis/gofish/blob/46b1b33645ed1802727dc4df28f5d3c3da722b15/client.go#L434
//
// The Gofish method linked does not currently perform seeks and so a PR will be suggested
// to change the method signature to accept an io.Reader instead.
type pipeReaderFakeSeeker struct {
	*io.PipeReader
}

// Seek impelements the io.Seeker interface only to panic if called
func (p pipeReaderFakeSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.New("Seek() not implemented for fake pipe reader seeker.")
}

// multipartPayloadSize prepares a temporary multipart form to determine the form size
//
// It creates a temporary form without reading in the update file payload and returns
// sizeOf(form) + sizeOf(update file)
func multipartPayloadSize(payload *multipartPayload) (int64, *bytes.Buffer, error) {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	// Add UpdateParameters field part
	part, err := updateParametersFormField("UpdateParameters", form)
	if err != nil {
		return 0, body, err
	}

	// a buffer to save the contents of the updateParameters reader
	buf := bytes.Buffer{}
	teeReader := io.TeeReader(payload.updateParameters, &buf)

	if _, err = io.Copy(part, teeReader); err != nil {
		return 0, body, err
	}

	// restore the reader
	payload.updateParameters = bytes.NewReader(buf.Bytes())

	// Add updateFile form
	_, err = form.CreateFormFile("UpdateFile", filepath.Base(payload.updateFile.Name()))
	if err != nil {
		return 0, body, err
	}

	// determine update file size
	finfo, err := payload.updateFile.Stat()
	if err != nil {
		return 0, body, err
	}

	// add terminating boundary to multipart form
	err = form.Close()
	if err != nil {
		return 0, body, err
	}

	return int64(body.Len()) + finfo.Size(), body, nil
}

// runRequestWithMultipartPayload is a copy of https://github.com/stmcginnis/gofish/blob/main/client.go#L349
// with a change to add the UpdateParameters multipart form field with a json content type header
// the resulting form ends up in this format
//
// Content-Length: 416
// Content-Type: multipart/form-data; boundary=--------------------
// ----1771f60800cb2801

// --------------------------1771f60800cb2801
// Content-Disposition: form-data; name="UpdateParameters"
// Content-Type: application/json

// {"Targets": [], "@Redfish.OperationApplyTime": "OnReset", "Oem":
//  {}}
// --------------------------1771f60800cb2801
// Content-Disposition: form-data; name="UpdateFile"; filename="dum
// myfile"
// Content-Type: application/octet-stream

// hey.
// --------------------------1771f60800cb2801--
func (c *Conn) runRequestWithMultipartPayload(method, url string, payload *multipartPayload) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("unable to execute request, no target provided")
	}

	// A content-length header is passed in to indicate the payload size
	//
	// The Content-length is set explicitly since the payload is an io.Reader,
	// https://github.com/golang/go/blob/ddad9b618cce0ed91d66f0470ddb3e12cfd7eeac/src/net/http/request.go#L861
	//
	// Without the content-length header the http client will set the Transfer-Encoding to 'chunked'
	// and that does not work for some BMCs (iDracs).
	contentLength, _, err := multipartPayloadSize(payload)
	if err != nil {
		return nil, errors.Wrap(err, "error determining multipart payload size")
	}

	headers := map[string]string{
		"Content-Length": strconv.FormatInt(contentLength, 10),
	}

	// setup pipe
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	// initiate a mulitpart writer
	form := multipart.NewWriter(pipeWriter)

	// go routine blocks on the io.Copy until the http request is made
	go func() {
		var err error
		defer func() {
			if err != nil {
				c.Log.Error(err, "multipart upload error occurred")
			}
		}()

		defer pipeWriter.Close()

		// Add UpdateParameters part
		parametersPart, err := updateParametersFormField("UpdateParameters", form)
		if err != nil {
			c.Log.Error(errMultiPartPayload, err.Error()+": UpdateParameters part copy error")

			return
		}

		if _, err = io.Copy(parametersPart, payload.updateParameters); err != nil {
			c.Log.Error(errMultiPartPayload, err.Error()+": UpdateParameters part copy error")

			return
		}

		// Add UpdateFile part
		updateFilePart, err := form.CreateFormFile("UpdateFile", filepath.Base(payload.updateFile.Name()))
		if err != nil {
			c.Log.Error(errMultiPartPayload, err.Error()+": UpdateFile part create error")

			return
		}

		if _, err = io.Copy(updateFilePart, payload.updateFile); err != nil {
			c.Log.Error(errMultiPartPayload, err.Error()+": UpdateFile part copy error")

			return
		}

		// add terminating boundary to multipart form
		form.Close()
	}()

	// pipeReader wrapped as a io.ReadSeeker to satisfy the gofish method signature
	reader := pipeReaderFakeSeeker{pipeReader}

	return c.redfishwrapper.RunRawRequestWithHeaders(method, url, reader, form.FormDataContentType(), headers)
}

// sets up the UpdateParameters MIMEHeader for the multipart form
// the Go multipart writer CreateFormField does not currently let us set Content-Type on a MIME Header
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.8:src/mime/multipart/writer.go;l=151
func updateParametersFormField(fieldName string, writer *multipart.Writer) (io.Writer, error) {
	if fieldName != "UpdateParameters" {
		return nil, errors.New("")
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="UpdateParameters"`)
	h.Set("Content-Type", "application/json")

	return writer.CreatePart(h)
}

// GetFirmwareInstallTaskQueued returns the redfish task object for a queued update task
func (c *Conn) GetFirmwareInstallTaskQueued(ctx context.Context, component string) (*gofishrf.Task, error) {
	vendor, _, err := c.DeviceVendorModel(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine device vendor, model attributes")
	}

	var task *gofishrf.Task

	// check an update task for the component is currently scheduled
	switch {
	case strings.Contains(vendor, constants.Dell):
		task, err = c.getDellFirmwareInstallTaskScheduled(component)
	default:
		err = errors.Wrap(
			bmclibErrs.ErrNotImplemented,
			"GetFirmwareInstallTask() for vendor: "+vendor,
		)
	}

	if err != nil {
		return nil, err
	}

	return task, nil
}

// purgeQueuedFirmwareInstallTask removes any existing queued firmware install task for the given component slug
func (c *Conn) purgeQueuedFirmwareInstallTask(ctx context.Context, component string) error {
	vendor, _, err := c.DeviceVendorModel(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to determine device vendor, model attributes")
	}

	// check an update task for the component is currently scheduled
	switch {
	case strings.Contains(vendor, constants.Dell):
		err = c.dellPurgeScheduledFirmwareInstallJob(component)
	default:
		err = errors.Wrap(
			bmclibErrs.ErrNotImplemented,
			"purgeFirmwareInstallTask() for vendor: "+vendor,
		)
	}

	return err
}
