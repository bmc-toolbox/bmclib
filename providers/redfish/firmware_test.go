package redfish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bmc-toolbox/bmclib/devices"
	bmclibErrs "github.com/bmc-toolbox/bmclib/errors"
)

// handler registered in mock_test.go
func multipartUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusNotFound)
	}

	body, err := ioutil.ReadAll(r.Body)
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

func Test_FirmwareInstall(t *testing.T) {
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
		component       string
		applyAt         string
		forceInstall    bool
		reader          io.Reader
		expectTaskID    string
		expectErr       error
		expectErrSubStr string
		testName        string
	}{
		{
			devices.SlugBIOS,
			"invalidApplyAt",
			false,
			nil,
			"",
			bmclibErrs.ErrFirmwareInstall,
			"invalid applyAt parameter",
			"applyAt parameter invalid",
		},
		{
			devices.SlugBIOS,
			devices.FirmwareApplyOnReset,
			false,
			fh,
			"467696020275",
			bmclibErrs.ErrFirmwareInstall,
			"task for BIOS firmware install present",
			"task ID exists",
		},
		{
			devices.SlugBIOS,
			devices.FirmwareApplyOnReset,
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
			taskID, err := mockClient.FirmwareInstall(context.TODO(), tc.component, tc.applyAt, tc.forceInstall, tc.reader)
			if tc.expectErr != nil {
				assert.ErrorIs(t, err, tc.expectErr)
				if tc.expectErrSubStr != "" {
					assert.True(t, strings.Contains(err.Error(), tc.expectErrSubStr))
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.expectTaskID, taskID)
			}
		})
	}

}

func Test_multipartPayloadSize(t *testing.T) {
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
		payload      []map[string]io.Reader
		expectedSize int64
		errorMsg     string
	}{
		{
			"content length as expected",
			[]map[string]io.Reader{
				{"UpdateParameters": bytes.NewReader(updateParameters)},
				{"UpdateFile": testfileFH},
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

func Test_firmwareUpdateCompatible(t *testing.T) {
	err := mockClient.firmwareUpdateCompatible(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
}
