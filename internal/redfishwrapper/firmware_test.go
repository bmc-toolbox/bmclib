package redfishwrapper

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestRunRequestWithMultipartPayload(t *testing.T) {
	defer goleak.VerifyNone(t)

	// init things
	tmpdir := t.TempDir()
	binPath := filepath.Join(tmpdir, "test.bin")
	err := os.WriteFile(binPath, []byte(`HELLOWORLD`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	updateFile, err := os.Open(binPath)
	if err != nil {
		t.Fatalf("%s -> %s", err.Error(), binPath)
	}

	defer updateFile.Close()
	defer os.Remove(binPath)

	multipartEndpoint := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusNotFound)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		// payload size
		expectedContentLength := "476"

		expected := []string{
			`Content-Disposition: form-data; name="UpdateParameters"`,
			`Content-Type: application/json`,
			`{"Targets":[],"@Redfish.OperationApplyTime":"OnReset","Oem":{}}`,
			`Content-Disposition: form-data; name="UpdateFile"; filename="test.bin"`,
			`Content-Type: application/octet-stream`,
			`HELLOWORLD`,
		}

		for _, want := range expected {
			assert.Contains(t, string(body), want, "expected value in payload")
		}

		assert.Equal(t, expectedContentLength, r.Header.Get("Content-Length"))

		w.Header().Add("Location", "/redfish/v1/TaskService/Tasks/JID_467696020275")
		w.WriteHeader(http.StatusAccepted)
	}

	tests := map[string]struct {
		hfunc     map[string]func(http.ResponseWriter, *http.Request)
		updateURI string
		payload   *multipartPayload
		err       error
	}{
		"happy case - multipart push": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/": endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/UpdateService/MultipartUpload": multipartEndpoint,
			},
			updateURI: "/redfish/v1/UpdateService/MultipartUpload",
			payload: &multipartPayload{
				updateParameters: []byte(`{"Targets":[],"@Redfish.OperationApplyTime":"OnReset","Oem":{}}`),
				updateFile:       updateFile,
			},
			err: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			handleFunc := tc.hfunc
			for endpoint, handler := range handleFunc {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			_, err = client.runRequestWithMultipartPayload(tc.updateURI, tc.payload)
			if tc.err != nil {
				assert.ErrorContains(t, err, tc.err.Error())
				return
			}

			assert.Nil(t, err)
			client.Close(context.Background())
		})
	}
}

func TestFirmwareInstallMethodURI(t *testing.T) {
	tests := map[string]struct {
		hfunc               map[string]func(http.ResponseWriter, *http.Request)
		expectInstallMethod installMethod
		expectUpdateURI     string
		err                 error
	}{
		"happy case - multipart push": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/":              endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Systems":       endpointFunc(t, "systems.json"),
				"/redfish/v1/Managers":      endpointFunc(t, "managers.json"),
				"/redfish/v1/Managers/1":    endpointFunc(t, "managers_1.json"),
				"/redfish/v1/UpdateService": endpointFunc(t, "updateservice_with_multipart.json"),
			},
			expectInstallMethod: multipartHttpUpload,
			expectUpdateURI:     "/redfish/v1/UpdateService/MultipartUpload",
			err:                 nil,
		},
		"happy case - unstructured http push": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/":              endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Systems":       endpointFunc(t, "systems.json"),
				"/redfish/v1/Managers":      endpointFunc(t, "managers.json"),
				"/redfish/v1/Managers/1":    endpointFunc(t, "managers_1.json"),
				"/redfish/v1/UpdateService": endpointFunc(t, "updateservice_with_httppushuri.json"),
			},
			expectInstallMethod: unstructuredHttpPush,
			expectUpdateURI:     "/redfish/v1/UpdateService/update",
			err:                 nil,
		},
		"failure case - service disabled": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/":              endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Systems":       endpointFunc(t, "systems.json"),
				"/redfish/v1/Managers":      endpointFunc(t, "managers.json"),
				"/redfish/v1/Managers/1":    endpointFunc(t, "managers_1.json"),
				"/redfish/v1/UpdateService": endpointFunc(t, "updateservice_disabled.json"),
			},
			expectInstallMethod: "",
			expectUpdateURI:     "",
			err:                 bmclibErrs.ErrRedfishUpdateService,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			handleFunc := tc.hfunc
			for endpoint, handler := range handleFunc {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			gotMethod, gotURI, err := client.firmwareInstallMethodURI()
			if tc.err != nil {
				assert.ErrorContains(t, err, tc.err.Error())
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectInstallMethod, gotMethod)
			assert.Equal(t, tc.expectUpdateURI, gotURI)

			client.Close(context.Background())
		})
	}
}

func TestTaskIDFromResponseBody(t *testing.T) {
	testCases := []struct {
		name        string
		body        []byte
		expectedID  string
		expectedErr error
	}{
		{
			name:        "happy case",
			body:        mustReadFile(t, "updateservice_ok_response.json"),
			expectedID:  "1234",
			expectedErr: nil,
		},
		{
			name:        "failure case",
			body:        mustReadFile(t, "updateservice_unexpected_response.json"),
			expectedID:  "",
			expectedErr: errTaskIdFromRespBody,
		},
		{
			name:        "failure case - invalid json",
			body:        []byte(`<html><head>crappy bmc is crappy<head/></html>`),
			expectedID:  "",
			expectedErr: errTaskIdFromRespBody,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskID, err := taskIDFromResponseBody(tc.body)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedID, taskID)
		})
	}
}

func TestTaskIDFromLocationHeader(t *testing.T) {
	testCases := []struct {
		name        string
		uri         string
		expectedID  string
		expectedErr error
	}{
		{
			name:        "task URI with JID",
			uri:         "http://foo/redfish/v1/TaskService/Tasks/JID_12345",
			expectedID:  "JID_12345",
			expectedErr: nil,
		},
		{
			name:        "task URI with ID",
			uri:         "http://foo/redfish/v1/TaskService/Tasks/1234",
			expectedID:  "1234",
			expectedErr: nil,
		},
		{
			name:        "task URI with Monitor suffix",
			uri:         "/redfish/v1/TaskService/Tasks/12/Monitor",
			expectedID:  "12",
			expectedErr: nil,
		},
		{
			name:        "trailing slash removed",
			uri:         "http://foo/redfish/v1/TaskService/Tasks/1/",
			expectedID:  "1",
			expectedErr: nil,
		},
		{
			name:        "invalid task URI - no task ID",
			uri:         "http://foo/redfish/v1/TaskService/Tasks/",
			expectedID:  "",
			expectedErr: bmclibErrs.ErrTaskNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			taskID, err := taskIDFromLocationHeader(tc.uri)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedID, taskID)
		})
	}
}

func TestUpdateParametersFormField(t *testing.T) {
	testCases := []struct {
		name        string
		fieldName   string
		expectedErr error
	}{
		{
			name:        "happy case",
			fieldName:   "UpdateParameters",
			expectedErr: nil,
		},
		{
			name:        "failure case",
			fieldName:   "InvalidField",
			expectedErr: errUpdateParams,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			writer := multipart.NewWriter(buf)

			output, err := updateParametersFormField(tc.fieldName, writer)
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}

			assert.NoError(t, err)
			assert.Contains(t, buf.String(), `Content-Disposition: form-data; name="UpdateParameters`)
			assert.Contains(t, buf.String(), `Content-Type: application/json`)
			assert.NotNil(t, output)

			// Validate the created multipart form content
			err = writer.Close()
			assert.NoError(t, err)

		})
	}
}

func TestMultipartPayloadSize(t *testing.T) {
	updateParameters, err := json.Marshal(struct {
		Targets            []string `json:"Targets"`
		RedfishOpApplyTime string   `json:"@Redfish.OperationApplyTime"`
		Oem                struct{} `json:"Oem"`
	}{
		[]string{},
		"foobar",
		struct{}{},
	})

	if err != nil {
		t.Fatal(err)
	}

	tmpdir := t.TempDir()
	binPath := filepath.Join(tmpdir, "test.bin")
	err = os.WriteFile(binPath, []byte(`HELLOWORLD`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	testfileFH, err := os.Open(binPath)
	if err != nil {
		t.Fatalf("%s -> %s", err.Error(), binPath)
	}

	testCases := []struct {
		testName     string
		payload      *multipartPayload
		expectedSize int64
		errorMsg     string
	}{
		{
			"content length as expected",
			&multipartPayload{
				updateParameters: updateParameters,
				updateFile:       testfileFH,
			},
			475,
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			gotSize, _, err := multipartPayloadSize(tc.payload)
			if tc.errorMsg != "" {
				assert.Contains(t, err.Error(), tc.errorMsg)
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedSize, gotSize)
		})
	}
}
