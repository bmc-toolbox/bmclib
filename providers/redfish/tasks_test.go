package redfish

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// handler registered in redfish_test.go
func dellJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusNotFound)
	}

	_, _ = w.Write(jsonResponse(r.RequestURI))
}

func Test_dellFirmwareUpdateTask(t *testing.T) {
	// see fixtures/v1/dell/jobs.json for the job IDs
	// completed job
	status, err := mockClient.dellJobAsRedfishTask("467767920358")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, status)
	assert.Equal(t, "2022-03-08T16:02:33", status.EndTime)
	assert.Equal(t, "2022-03-08T15:59:52", status.StartTime)
	assert.Equal(t, 100, status.PercentComplete)
	assert.Equal(t, "Completed", string(status.TaskState))
	assert.Equal(t, "Job completed successfully.", string(status.TaskStatus))
}

func Test_dellPurgeScheduledFirmwareInstallJob(t *testing.T) {
	err := mockClient.dellPurgeScheduledFirmwareInstallJob("bios")
	if err != nil {
		t.Fatal(err)
	}
}
