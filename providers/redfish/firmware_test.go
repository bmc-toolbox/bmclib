package redfish

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/common"
)

// handler registered in mock_test.go
func multipartUpload(w http.ResponseWriter, r *http.Request) {
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
		if !strings.Contains(string(body), want) {
			fmt.Println(string(body))
			log.Fatal("expected value not in multipartUpload payload: " + string(want))
		}
	}

	if r.Header.Get("Content-Length") != expectedContentLength {
		log.Fatal("Header Content-Length does not match expected")
	}

	w.Header().Add("Location", "/redfish/v1/TaskService/Tasks/JID_467696020275")
	w.WriteHeader(http.StatusAccepted)
}

func TestFirmwareInstall(t *testing.T) {
	// curl -Lv -s -k -u root:calvin \
	// -F 'UpdateParameters={"Targets": [], "@Redfish.OperationApplyTime": "OnReset", "Oem": {}};type=application/json' \
	// -F'foo.bin=@/tmp/dummyfile;application/octet-stream'
	// https://192.168.1.1/redfish/v1/UpdateService/MultipartUpload --trace-ascii /dev/stdout

	tmpdir := t.TempDir()
	binPath := filepath.Join(tmpdir, "test.bin")
	err := os.WriteFile(binPath, []byte(`HELLOWORLD`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	fh, err := os.Open(binPath)
	if err != nil {
		t.Fatalf("%s -> %s", err.Error(), binPath)
	}

	defer os.Remove(binPath)

	tests := []struct {
		component          string
		applyAt            string
		forceInstall       bool
		setRequiredTimeout bool
		reader             io.Reader
		expectTaskID       string
		expectErr          error
		expectErrSubStr    string
		testName           string
	}{
		{
			common.SlugBIOS,
			constants.FirmwareApplyOnReset,
			false,
			false,
			nil,
			"",
			bmclibErrs.ErrFirmwareInstall,
			"method expects an *os.File object",
			"expect *os.File object",
		},
		{
			common.SlugBIOS,
			constants.FirmwareApplyOnReset,
			false,
			false,
			&os.File{},
			"",
			errInsufficientCtxTimeout,
			"",
			"remaining context deadline",
		},
		{
			common.SlugBIOS,
			"invalidApplyAt",
			false,
			true,
			&os.File{},
			"",
			bmclibErrs.ErrFirmwareInstall,
			"invalid applyAt parameter",
			"applyAt parameter invalid",
		},
		{
			common.SlugBIOS,
			constants.FirmwareApplyOnReset,
			false,
			true,
			fh,
			"467696020275",
			bmclibErrs.ErrFirmwareInstall,
			"task for BIOS firmware install present",
			"task ID exists",
		},
		{
			common.SlugBIOS,
			constants.FirmwareApplyOnReset,
			true,
			true,
			fh,
			"467696020275",
			nil,
			"task for BIOS firmware install present",
			"task created (previous task purged with force)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.TODO(), 1*time.Second)
			if tc.setRequiredTimeout {
				ctx, cancel = context.WithTimeout(context.TODO(), 20*time.Minute)
			}

			taskID, err := mockClient.FirmwareInstall(ctx, tc.component, tc.applyAt, tc.forceInstall, tc.reader)
			if tc.expectErr != nil {
				assert.ErrorIs(t, err, tc.expectErr)
				if tc.expectErrSubStr != "" {
					assert.True(t, strings.Contains(err.Error(), tc.expectErrSubStr))
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expectTaskID, taskID)
			}

			defer cancel()
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

func TestFirmwareUpdateCompatible(t *testing.T) {
	err := mockClient.firmwareUpdateCompatible(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
}

// referenced in main_test.go
func openbmcStatus(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/redfish/v1/TaskService/Tasks/15" {
		// return an HTTP error, don't care to return correct data after
		http.Error(w, "404 page not found:"+r.URL.Path, http.StatusNotFound)
	}

	mytask := `{
  "@odata.id": "/redfish/v1/TaskService/Tasks/15",
  "@odata.type": "#Task.v1_4_3.Task",
  "Id": "15",
  "Messages": [
    {
      "@odata.type": "#Message.v1_1_1.Message",
      "Message": "The task with Id '15' has started.",
      "MessageArgs": [
        "15"
      ],
      "MessageId": "TaskEvent.1.0.3.TaskStarted",
      "MessageSeverity": "OK",
      "Resolution": "None."
    }
  ],
  "Name": "Task 15",
  "TaskState": "TestState",
  "TaskStatus": "TestStatus"
}
`
	_, _ = w.Write([]byte(mytask))

}

func Test_FirmwareInstall2(t *testing.T) {
	state, err := mockClient.FirmwareInstallStatus(context.TODO(), "", "testOpenbmc", "15")
	if err != nil {
		t.Fatal(err)
	}
	if state != "unknown: teststate" {
		t.Fatal("Wrong test state:", state)
	}
}
