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

	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"

	"github.com/bmc-toolbox/bmclib/devices"
	bmclibErrs "github.com/bmc-toolbox/bmclib/errors"
)

// SupportedFirmwareApplyAtValues returns the supported redfish firmware applyAt values
func SupportedFirmwareApplyAtValues() []string {
	return []string{
		devices.FirmwareApplyImmediate,
		devices.FirmwareApplyOnReset,
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
	if !stringInSlice(applyAt, SupportedFirmwareApplyAtValues()) {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "invalid applyAt parameter: "+applyAt)
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

	payload := map[string]io.Reader{
		"UpdateParameters": bytes.NewReader(updateParameters),
		"UpdateFile":       reader,
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

// FirmwareInstallStatus returns the status of the firmware install task queued
func (c *Conn) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (state string, err error) {
	vendor, _, err := c.DeviceVendorModel(ctx)
	if err != nil {
		return state, errors.Wrap(err, "unable to determine device vendor, model attributes")
	}

	var task *rf.Task
	switch {
	case strings.Contains(vendor, devices.Dell):
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
		return devices.FirmwareInstallInitializing, nil
	case "running", "stopping", "cancelling", "scheduling":
		return devices.FirmwareInstallRunning, nil
	case "pending", "new":
		return devices.FirmwareInstallQueued, nil
	case "scheduled":
		return devices.FirmwareInstallPowerCyleHost, nil
	case "interrupted", "killed", "exception", "cancelled", "suspended", "failed":
		return devices.FirmwareInstallFailed, nil
	case "completed":
		return devices.FirmwareInstallComplete, nil
	default:
		return devices.FirmwareInstallUnknown + ": " + state, nil
	}

}

// firmwareUpdateCompatible retuns an error if the firmware update process for the BMC is not supported
func (c *Conn) firmwareUpdateCompatible(ctx context.Context) (err error) {
	updateService, err := c.conn.Service.UpdateService()
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

// multipartPayloadSize prepares a temporary multipart form to determine the form size
func multipartPayloadSize(payload []map[string]io.Reader) (int64, *bytes.Buffer, error) {
	body := &bytes.Buffer{}
	form := multipart.NewWriter(body)

	var size int64
	var err error
	for idx, elem := range payload {
		for key, reader := range elem {
			var part io.Writer
			if file, ok := reader.(*os.File); ok {
				// Add update file fields
				if _, err = form.CreateFormFile(key, filepath.Base(file.Name())); err != nil {
					return 0, body, err
				}

				// determine file size
				finfo, err := file.Stat()
				if err != nil {
					return 0, body, err
				}

				size = finfo.Size()

			} else {
				// Add other fields
				if part, err = updateParametersFormField(key, form); err != nil {
					return 0, body, err
				}

				// use a tee reader so the
				buf := bytes.Buffer{}
				teeReader := io.TeeReader(reader, &buf)

				if _, err = io.Copy(part, teeReader); err != nil {
					return 0, body, err
				}

				// place it back so its available to be read again
				payload[idx][key] = bytes.NewReader(buf.Bytes())
			}
		}
	}

	err = form.Close()
	if err != nil {
		return 0, body, err
	}

	return int64(body.Len()) + size, body, nil
}

// runRequestWithMultipartPayload sets up the mutipart payload to upload the firmware binary as a stream.
//
// The multipart form format is described below,
//
// Content-Length: 416
// Content-Type: multipart/form-data; boundary=--------------------
// ----1771f60800cb2801
//
// --------------------------1771f60800cb2801
// Content-Disposition: form-data; name="UpdateParameters"
// Content-Type: application/json
//
// {"Targets": [], "@Redfish.OperationApplyTime": "OnReset", "Oem":
//  {}}
// --------------------------1771f60800cb2801
// Content-Disposition: form-data; name="UpdateFile"; filename="dum
// myfile"
// Content-Type: application/octet-stream
//
// hey.
// --------------------------1771f60800cb2801--
func (c *Conn) runRequestWithMultipartPayload(method, url string, payload map[string]io.Reader) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("unable to execute request, no target provided")
	}

	// A content-lenght header is passed in to indicate the payload size
	contentLength, _, err := multipartPayloadSize(payload)
	if err != nil {
		return nil, errors.Wrap(err, "error determining multipart payload size")
	}

	// setup pipe to stream multipart payload
	//
	// pipeReader, the reader end of the pipe is given to the http client executing the request.
	// pipeWriter, the routine below copies to the writer end of the pipe.
	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	// form containing multipart payload
	form := multipart.NewWriter(pipeWriter)

	// routine spawned to copy the multipart payload when pipeReader is being read from
	go func() {
		var err error
		defer func() {
			if err != nil {
				c.Log.Error(err, "multipart upload error occured")
			}
		}()

		defer pipeWriter.Close()

		for _, elem := range payload {
			for key, reader := range elem {
				var part io.Writer
				// add update file multipart form header
				//  Content-Disposition: form-data; name="UpdateFile"; filename="dum
				//  myfile"
				//  Content-Type: application/octet-stream
				if file, ok := reader.(*os.File); ok {
					if part, err = form.CreateFormFile(key, filepath.Base(file.Name())); err != nil {
						return
					}
				} else {
					// add update parameters multipart form header
					//  Content-Disposition: form-data; name="UpdateParameters"
					//  Content-Type: application/json
					if part, err = updateParametersFormField(key, form); err != nil {
						return
					}
				}

				// copy multipart form data from reader
				_, err := io.Copy(part, reader)
				if err != nil {
					return
				}

			}
		}

		// add terminating boundary to multipart form
		err = form.Close()
	}()

	// pipeReader wrapped as a io.ReadSeeker to satisfy the gofish method signature
	readSeeker := pipeReaderFakeSeeker{pipeReader}

	headers := map[string]string{"Content-Length": strconv.FormatInt(contentLength, 10)}

	return c.conn.RunRawRequestWithHeaders(method, url, readSeeker, form.FormDataContentType(), headers)
}

// pipeReaderFakeSeeker wraps the io.PipeReader and implements the io.Seeker interface
// to meet the API requirements for the Gofish client https://github.com/stmcginnis/gofish/blob/main/client.go#L390
//
// The Gofish method linked does not currently perform seeks and so a PR will be suggested
// to change the method signature to accept an io.Reader instead.
type pipeReaderFakeSeeker struct {
	*io.PipeReader
}

// Seek impelements the io.Seeker interface only to panic if called
func (p pipeReaderFakeSeeker) Seek(offset int64, whence int) (int64, error) {
	panic("Seek() not implemented, runRequestWithMultipartPayload() will require a rework to split the firmware file into smaller chunks.")
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
func (c *Conn) GetFirmwareInstallTaskQueued(ctx context.Context, component string) (*rf.Task, error) {
	vendor, _, err := c.DeviceVendorModel(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine device vendor, model attributes")
	}

	var task *rf.Task

	// check an update task for the component is currently scheduled
	switch {
	case strings.Contains(vendor, devices.Dell):
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
	case strings.Contains(vendor, devices.Dell):
		err = c.dellPurgeScheduledFirmwareInstallJob(component)
	default:
		err = errors.Wrap(
			bmclibErrs.ErrNotImplemented,
			"purgeFirmwareInstallTask() for vendor: "+vendor,
		)
	}

	if err != nil {
		return err
	}

	return nil
}
