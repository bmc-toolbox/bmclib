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
	// validate firmware update mechanism is supported
	err = c.firmwareUpdateCompatible(ctx)
	if err != nil {
		return "", err
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
		return "", err
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
	// record the http client time to be restored
	httpClientTimeout := c.redfishwrapper.HttpClientTimeout()
	defer func() {
		c.redfishwrapper.SetHttpClientTimeout(httpClientTimeout)
	}()

	c.redfishwrapper.SetHttpClientTimeout(time.Until(ctxDeadline))

	payload := map[string]io.Reader{
		"UpdateParameters": bytes.NewReader(updateParameters),
		"UpdateFile":       reader,
	}

	updateService, err := c.redfishwrapper.UpdateService()

	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, err.Error())
	}

	var resp *http.Response
	if updateService.MultipartHTTPPushURI != "" {
		// TODO: should use updateService.MultipartHTTPPushURI rather than hardcoded path
		// but should be tested when modified
		resp, err = c.runRequestWithMultipartPayload(http.MethodPost, "/redfish/v1/UpdateService/MultipartUpload", payload)
	} else if updateService.HTTPPushURI != "" {
		resp, err = c.runRequestWithPayload(http.MethodPost, updateService.HTTPPushURI, payload["UpdateFile"])
	} else {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, "No URI available for push updates")
	}

	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, err.Error())
	}

	if resp.StatusCode != http.StatusAccepted {
		return "", errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"non 202 status code returned: "+strconv.Itoa(resp.StatusCode),
		)
	}

	return location2TaskID(resp.Header.Get("Location"))
}

func location2TaskID(location string) (taskID string, err error) {
	if strings.Contains(location, "JID_") {
		// The response contains a location header pointing to the task URI
		// Location: /redfish/v1/TaskService/Tasks/JID_467696020275
		taskID = strings.Split(location, "JID_")[1]
	} else if strings.Contains(location, "/Monitor") {
		// OpenBMC returns a monitor URL in Location
		// Location: /redfish/v1/TaskService/Tasks/12/Monitor
		splits := strings.Split(location, "/")
		taskID = splits[5]
	} else {
		return "", bmclibErrs.ErrTaskNotFound
	}

	return taskID, nil
}

// FirmwareInstallStatus returns the status of the firmware install task queued
func (c *Conn) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (state string, err error) {
	vendor, _, err := c.DeviceVendorModel(ctx)
	if err != nil {
		return state, errors.Wrap(err, "unable to determine device vendor, model attributes")
	}

	// component is not used, we hack it for tests
	if component == "testOpenbmc" {
		vendor = constants.Packet
	}

	var task *gofishrf.Task
	switch {
	case strings.Contains(vendor, constants.Dell):
		task, err = c.dellJobAsRedfishTask(taskID)
		if task == nil {
			return state, errors.New("failed to lookup task status for task ID: " + taskID)
		}

		state = strings.ToLower(string(task.TaskState))

	case strings.Contains(vendor, constants.Packet):
		resp, _ := c.redfishwrapper.Get("/redfish/v1/TaskService/Tasks/" + taskID)
		if resp.StatusCode != 200 {
			err = errors.Wrap(
				bmclibErrs.ErrFirmwareInstall,
				"HTTP Error: "+fmt.Sprint(resp.StatusCode),
			)

			state = "failed"
			break
		}

		//task, err := gofishrf.GetTask(c.redfishwrapper, "/redfish/v1/TaskService/Tasks/" + taskID)
		//fmt.Printf("task:", task);

		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		state, err = c.openbmcGetStatus(data)

	default:
		err = errors.Wrap(
			bmclibErrs.ErrNotImplemented,
			"FirmwareInstallStatus() for vendor: "+vendor,
		)
	}

	if err != nil {
		return state, err
	}

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

	// for now we expect multipart HTTP push update support,
	// or at least the unstructured HTTP push update support
	if updateService.MultipartHTTPPushURI == "" && updateService.HTTPPushURI == "" {
		return errors.Wrap(bmclibErrs.ErrRedfishUpdateService, "No HTTP push updates supported (multipart or unstructured)")
	}

	return nil
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
func (c *Conn) runRequestWithMultipartPayload(method, url string, payload map[string]io.Reader) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("unable to execute request, no target provided")
	}

	var payloadBuffer bytes.Buffer
	var err error
	payloadWriter := multipart.NewWriter(&payloadBuffer)
	for key, reader := range payload {
		var partWriter io.Writer
		if file, ok := reader.(*os.File); ok {
			// Add a file stream
			if partWriter, err = payloadWriter.CreateFormFile(key, filepath.Base(file.Name())); err != nil {
				return nil, err
			}
		} else {
			// Add other fields
			if partWriter, err = updateParametersFormField(key, payloadWriter); err != nil {
				return nil, err
			}
		}
		if _, err = io.Copy(partWriter, reader); err != nil {
			return nil, err
		}
	}
	payloadWriter.Close()

	return c.redfishwrapper.RunRawRequestWithHeaders(method, url, bytes.NewReader(payloadBuffer.Bytes()), payloadWriter.FormDataContentType(), nil)
}

// Updates using an unstrctured HTTP updates
func (c *Conn) runRequestWithPayload(method, url string, payload io.Reader) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("unable to execute request, no target provided")
	}

	b, _ := io.ReadAll(payload)
	payloadReadSeeker := bytes.NewReader(b)
	return c.redfishwrapper.RunRawRequestWithHeaders(method, url, payloadReadSeeker, "application/octet-stream", nil)
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
	case strings.Contains(vendor, constants.Packet):
		//task, err = c.getDellFirmwareInstallTaskScheduled(component)
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
