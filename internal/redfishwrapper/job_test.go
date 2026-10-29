package redfishwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// newDellJobClient serves the Dell service root (with the given handler) and the JobService
// fixtures, and returns an opened client.
func newDellJobClient(t *testing.T, serviceRoot http.HandlerFunc) *Client {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", serviceRoot)
	mux.HandleFunc("/redfish/v1/JobService", endpointFunc(t, "dell/jobservice.json"))
	mux.HandleFunc("/redfish/v1/JobService/Jobs", endpointFunc(t, "dell/jobs.json"))
	mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_123456789012", endpointFunc(t, "dell/job_jid_123456789012.json"))

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(context.Background()))
	t.Cleanup(func() { _ = client.Close(context.Background()) })

	return client
}

func TestGetJobs(t *testing.T) {
	client := newDellJobClient(t, endpointFunc(t, "dell/serviceroot.json"))

	jobs, err := client.GetJobs(context.Background())
	require.NoError(t, err)
	require.Len(t, jobs, 1)

	percent := 0
	assert.Equal(t, bmc.Job{
		ID:              "JID_123456789012",
		Name:            "Configure: BIOS.Setup.1-1",
		JobState:        "Pending",
		JobStatus:       "OK",
		PercentComplete: &percent,
		StartTime:       "TIME_NOW",
		Messages:        []string{"Task successfully scheduled."},
	}, jobs[0])
}

func TestGetJobsNoJobService(t *testing.T) {
	client := newDellJobClient(t, func(w http.ResponseWriter, _ *http.Request) {
		var root map[string]any
		require.NoError(t, json.Unmarshal(mustReadFile(t, "dell/serviceroot.json"), &root))
		delete(root, "JobService")
		_ = json.NewEncoder(w).Encode(root)
	})

	_, err := client.GetJobs(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrJobServiceNotFound)
}
