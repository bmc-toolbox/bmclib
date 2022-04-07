package redfish

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

	expected := []string{
		`Content-Disposition: form-data; name="UpdateParameters"`,
		`Content-Type: application/json`,
		`{"Targets": [], "@Redfish.OperationApplyTime": "OnReset", "Oem": {}}`,
		`Content-Disposition: form-data; name="UpdateFile"; filename="test.bin"`,
		`Content-Type: application/octet-stream`,
		`HELLOWORLD`,
	}

	for _, want := range expected {
		if !strings.Contains(string(body), want) {
			log.Fatal("expected value not in multipartUpload payload: " + string(want))
		}
	}

	w.Header().Add("Location", "/redfish/v1/TaskService/Tasks/JID_467696020275")
	w.WriteHeader(http.StatusAccepted)
}

func Test_FirmwareUpload(t *testing.T) {
	// curl -Lv -s -k -u root:calvin \
	// -F 'UpdateParameters={"Targets": [], "@Redfish.OperationApplyTime": "OnReset", "Oem": {}};type=application/json' \
	// -F'foo.bin=@/tmp/dummyfile;application/octet-stream'
	// https://192.168.1.1/redfish/v1/UpdateService/MultipartUpload --trace-ascii /dev/stdout
	err := ioutil.WriteFile("/tmp/test.bin", []byte(`HELLOWORLD`), 0600)
	if err != nil {
		t.Fatal(err)
	}

	fh, err := os.Open("/tmp/test.bin")
	if err != nil {
		t.Fatalf("%s -> %s", err.Error(), "/tmp/test.bin")
	}

	_, err = mockClient.FirmwareInstall(context.TODO(), "", "invalid", false, fh)
	assert.ErrorIs(t, err, bmclibErrs.ErrFirmwareInstall)

	jobID, err := mockClient.FirmwareInstall(context.TODO(), "", devices.FirmwareApplyOnReset, false, fh)
	if err != nil {
		t.Fatal("err in FirmwareUpload" + err.Error())
	}

	assert.Equal(t, "467696020275", jobID)
}

func Test_firmwareUpdateCompatible(t *testing.T) {
	err := mockClient.firmwareUpdateCompatible(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
}
