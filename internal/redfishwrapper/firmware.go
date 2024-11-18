package redfishwrapper

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
	"github.com/stmcginnis/gofish/redfish"

	"github.com/metal-toolbox/bmclib/constants"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
)

type installMethod string

const (
	unstructuredHttpPush installMethod = "unstructuredHttpPush"
	multipartHttpUpload  installMethod = "multipartUpload"
)

var (
	// the URI for starting a firmware update via StartUpdate is defined in the Redfish Resource and
	// Schema Guide (2024.1)
	startUpdateURI = "/redfish/v1/UpdateService/Actions/UpdateService.StartUpdate"
)

var (
	errMultiPartPayload   = errors.New("error preparing multipart payload")
	errUpdateParams       = errors.New("error in redfish UpdateParameters payload")
	errTaskIdFromRespBody = errors.New("failed to identify firmware install taskID from response body")
)

type RedfishUpdateServiceParameters struct {
	Targets            []string                     `json:"Targets"`
	OperationApplyTime constants.OperationApplyTime `json:"@Redfish.OperationApplyTime"`
	Oem                json.RawMessage              `json:"Oem"`
}

// FirmwareUpload uploads and initiates the firmware install process
func (c *Client) FirmwareUpload(ctx context.Context, updateFile *os.File, params *RedfishUpdateServiceParameters) (taskID string, err error) {
	parameters, err := json.Marshal(params)
	if err != nil {
		return "", errors.Wrap(errUpdateParams, err.Error())
	}

	installMethod, installURI, err := c.firmwareInstallMethodURI()
	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, err.Error())
	}

	// override the gofish HTTP client timeout,
	// since the context timeout is set at Open() and is at a lower value than required for this operation.
	//
	// record the http client timeout to be restored when this method returns
	httpClientTimeout := c.HttpClientTimeout()
	defer func() {
		c.SetHttpClientTimeout(httpClientTimeout)
	}()

	ctxDeadline, _ := ctx.Deadline()
	c.SetHttpClientTimeout(time.Until(ctxDeadline))

	var resp *http.Response

	switch installMethod {
	case multipartHttpUpload:
		var uploadErr error
		resp, uploadErr = c.multipartHTTPUpload(installURI, updateFile, parameters)
		if uploadErr != nil {
			return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, uploadErr.Error())
		}

	case unstructuredHttpPush:
		var uploadErr error
		resp, uploadErr = c.unstructuredHttpUpload(installURI, updateFile)
		if uploadErr != nil {
			return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, uploadErr.Error())
		}

	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, "unsupported install method: "+string(installMethod))
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"unexpected status code returned: "+resp.Status,
		)
	}

	// The response contains a location header pointing to the task URI
	// Location: /redfish/v1/TaskService/Tasks/JID_467696020275
	var location = resp.Header.Get("Location")
	if strings.Contains(location, "/TaskService/Tasks/") {
		return taskIDFromLocationHeader(location)
	}

	rfTask := &redfish.Task{}
	if err := rfTask.UnmarshalJSON(response); err != nil {
		// we got invalid JSON
		return "", fmt.Errorf("unmarshaling redfish response: %w", err)
	}
	// it's possible to get well-formed JSON that isn't a Task (thanks SMC). Test that we have something sensible.
	if strings.Contains(rfTask.ODataType, "Task") {
		return rfTask.ID, nil
	}

	return taskIDFromResponseBody(response)
}

// StartUpdateForUploadedFirmware starts an update for a firmware file previously uploaded and returns the taskID
func (c *Client) StartUpdateForUploadedFirmware(ctx context.Context) (taskID string, err error) {
	errStartUpdate := errors.New("error in starting update for uploaded firmware")
	updateService, err := c.client.Service.UpdateService()
	if err != nil {
		return "", errors.Wrap(err, "error querying redfish update service")
	}

	// Start update the hard way. We do this to get back the task object from the response body so that
	// we can parse the task id out of it.
	resp, err := updateService.GetClient().PostWithHeaders(startUpdateURI, nil, nil)
	if err != nil {
		return "", errors.Wrap(err, "error querying redfish start update endpoint")
	}

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "error reading redfish start update response body")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", errors.Wrap(errStartUpdate, "unexpected status code returned: "+resp.Status)
	}

	var location = resp.Header.Get("Location")
	if strings.Contains(location, "/TaskService/Tasks/") {
		return taskIDFromLocationHeader(location)
	}

	rfTask := &redfish.Task{}
	if err := rfTask.UnmarshalJSON(response); err != nil {
		// we got invalid JSON
		return "", fmt.Errorf("unmarshaling redfish response: %w", err)
	}
	if strings.Contains(rfTask.ODataType, "Task") {
		return rfTask.ID, nil
	}

	return taskIDFromResponseBody(response)
}

// StartUpdateForUploadedFirmware starts an update for a firmware file previously uploaded
func (c *Client) StartUpdateForUploadedFirmwareNoTaskID(ctx context.Context) error {
	updateService, err := c.client.Service.UpdateService()
	if err != nil {
		return errors.Wrap(err, "error querying redfish update service")
	}

	err = updateService.StartUpdate()
	if err != nil {
		return errors.Wrap(err, "error querying redfish start update endpoint")
	}

	return nil
}

type TaskAccepted struct {
	Accepted struct {
		Code                string `json:"code"`
		Message             string `json:"Message"`
		MessageExtendedInfo []struct {
			MessageID         string   `json:"MessageId"`
			Severity          string   `json:"Severity"`
			Resolution        string   `json:"Resolution"`
			Message           string   `json:"Message"`
			MessageArgs       []string `json:"MessageArgs"`
			RelatedProperties []string `json:"RelatedProperties"`
		} `json:"@Message.ExtendedInfo"`
	} `json:"Accepted"`
}

func taskIDFromResponseBody(resp []byte) (taskID string, err error) {
	a := &TaskAccepted{}
	if err = json.Unmarshal(resp, a); err != nil {
		return "", errors.Wrap(errTaskIdFromRespBody, err.Error())
	}

	var taskURI string

	for _, info := range a.Accepted.MessageExtendedInfo {
		for _, msg := range info.MessageArgs {
			if !strings.Contains(msg, "/TaskService/Tasks/") {
				continue
			}

			taskURI = msg
			break
		}
	}

	if taskURI == "" {
		return "", errors.Wrap(errTaskIdFromRespBody, "TaskService/Tasks/<id> URI not identified")
	}

	tokens := strings.Split(taskURI, "/")
	if len(tokens) == 0 {
		return "", errors.Wrap(errTaskIdFromRespBody, "invalid/unsupported task URI: "+taskURI)
	}

	return tokens[len(tokens)-1], nil
}

func taskIDFromLocationHeader(uri string) (taskID string, err error) {
	uri = strings.TrimSuffix(uri, "/")

	switch {
	// OpenBMC returns /redfish/v1/TaskService/Tasks/12/Monitor
	case strings.Contains(uri, "/Tasks/") && strings.HasSuffix(uri, "/Monitor"):
		taskIDPart := strings.Split(uri, "/Tasks/")[1]
		taskID := strings.TrimSuffix(taskIDPart, "/Monitor")
		return taskID, nil

	case strings.Contains(uri, "Tasks/"):
		taskIDPart := strings.Split(uri, "/Tasks/")[1]
		return taskIDPart, nil

	default:
		return "", errors.Wrap(bmclibErrs.ErrTaskNotFound, "failed to parse taskID from uri: "+uri)
	}
}

type multipartPayload struct {
	updateParameters []byte
	updateFile       *os.File
}

func (c *Client) multipartHTTPUpload(url string, update *os.File, params []byte) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("unable to execute request, no target provided")
	}

	// payload ordered in the format it ends up in the multipart form
	payload := &multipartPayload{
		updateParameters: params,
		updateFile:       update,
	}

	return c.runRequestWithMultipartPayload(url, payload)
}

func (c *Client) unstructuredHttpUpload(url string, update io.Reader) (*http.Response, error) {
	if url == "" {
		return nil, fmt.Errorf("unable to execute request, no target provided")
	}

	// TODO: transform this to read the update so that we don't hold the data in memory
	b, _ := io.ReadAll(update)
	payloadReadSeeker := bytes.NewReader(b)

	return c.RunRawRequestWithHeaders(http.MethodPost, url, payloadReadSeeker, "application/octet-stream", nil)

}

// firmwareUpdateMethodURI returns the updateMethod and URI
func (c *Client) firmwareInstallMethodURI() (method installMethod, updateURI string, err error) {
	updateService, err := c.UpdateService()
	if err != nil {
		return "", "", errors.Wrap(bmclibErrs.ErrRedfishUpdateService, err.Error())
	}

	// update service disabled
	if !updateService.ServiceEnabled {
		return "", "", errors.Wrap(bmclibErrs.ErrRedfishUpdateService, "service disabled")
	}

	switch {
	case updateService.MultipartHTTPPushURI != "":
		return multipartHttpUpload, updateService.MultipartHTTPPushURI, nil
	case updateService.HTTPPushURI != "":
		return unstructuredHttpPush, updateService.HTTPPushURI, nil
	}

	return "", "", errors.Wrap(bmclibErrs.ErrRedfishUpdateService, "unsupported update method")
}

// sets up the UpdateParameters MIMEHeader for the multipart form
// the Go multipart writer CreateFormField does not currently let us set Content-Type on a MIME Header
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.8:src/mime/multipart/writer.go;l=151
func updateParametersFormField(fieldName string, writer *multipart.Writer) (io.Writer, error) {
	if fieldName != "UpdateParameters" {
		return nil, errors.Wrap(errUpdateParams, "expected field not found to create multipart form")
	}

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="UpdateParameters"`)
	h.Set("Content-Type", "application/json")

	return writer.CreatePart(h)
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

	if _, err = io.Copy(part, bytes.NewReader(payload.updateParameters)); err != nil {
		return 0, body, err
	}

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
func (c *Client) runRequestWithMultipartPayload(url string, payload *multipartPayload) (*http.Response, error) {
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
				c.logger.Error(err, "multipart upload error occurred")
			}
		}()

		defer pipeWriter.Close()

		// Add UpdateParameters part
		parametersPart, err := updateParametersFormField("UpdateParameters", form)
		if err != nil {
			c.logger.Error(errMultiPartPayload, err.Error()+": UpdateParameters part copy error")

			return
		}

		if _, err = io.Copy(parametersPart, bytes.NewReader(payload.updateParameters)); err != nil {
			c.logger.Error(errMultiPartPayload, err.Error()+": UpdateParameters part copy error")

			return
		}

		// Add UpdateFile part
		updateFilePart, err := form.CreateFormFile("UpdateFile", filepath.Base(payload.updateFile.Name()))
		if err != nil {
			c.logger.Error(errMultiPartPayload, err.Error()+": UpdateFile part create error")

			return
		}

		if _, err = io.Copy(updateFilePart, payload.updateFile); err != nil {
			c.logger.Error(errMultiPartPayload, err.Error()+": UpdateFile part copy error")

			return
		}

		// add terminating boundary to multipart form
		form.Close()
	}()

	// pipeReader wrapped as a io.ReadSeeker to satisfy the gofish method signature
	reader := pipeReaderFakeSeeker{pipeReader}

	return c.RunRawRequestWithHeaders(http.MethodPost, url, reader, form.FormDataContentType(), headers)
}
