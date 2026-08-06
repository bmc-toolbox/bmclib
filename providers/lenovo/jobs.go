package lenovo

import (
	"context"
	"net/url"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/stmcginnis/gofish"
)

const (
	jobServiceURI = "/redfish/v1/JobService"
	jobsURI       = jobServiceURI + "/Jobs"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.JobManager = (*Conn)(nil)

// JobService returns the XCC job-service configuration.
//
// Implements bmc.JobManager.
func (c *Conn) JobService(ctx context.Context) (bmc.JobServiceInfo, error) {
	js, err := c.redfishwrapper.JobService()
	if err != nil {
		return bmc.JobServiceInfo{}, err
	}

	return bmc.JobServiceInfo{ServiceEnabled: js.ServiceEnabled}, nil
}

// Jobs lists the XCC jobs.
//
// Implements bmc.JobManager.
func (c *Conn) Jobs(ctx context.Context) ([]bmc.JobInfo, error) {
	js, err := c.redfishwrapper.JobService()
	if err != nil {
		return nil, err
	}

	jobs, err := js.Jobs()
	if err != nil {
		return nil, err
	}

	out := make([]bmc.JobInfo, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, bmc.JobInfo{
			ID:              j.ID,
			Name:            j.Name,
			JobState:        string(j.JobState),
			PercentComplete: int(gofish.Deref(j.PercentComplete)),
			StartTime:       j.StartTime,
		})
	}

	return out, nil
}

// Job returns a job by Id.
//
// Implements bmc.JobManager.
func (c *Conn) Job(ctx context.Context, id string) (bmc.JobInfo, error) {
	target, err := url.JoinPath(jobsURI, id)
	if err != nil {
		return bmc.JobInfo{}, err
	}
	return c.jobAt(target)
}

// JobUpdateSchedule PATCHes a job's Schedule.
//
// Implements bmc.JobManager.
func (c *Conn) JobUpdateSchedule(ctx context.Context, id string, schedule map[string]any) error {
	target, err := url.JoinPath(jobsURI, id)
	if err != nil {
		return err
	}
	payload := map[string]any{"Schedule": schedule}
	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, target, payload, nil))
}

// jobAt reads a single job resource.
func (c *Conn) jobAt(url string) (bmc.JobInfo, error) {
	var doc struct {
		ID              string `json:"Id"`
		Name            string `json:"Name"`
		JobState        string `json:"JobState"`
		PercentComplete int    `json:"PercentComplete"`
		StartTime       string `json:"StartTime"`
	}
	if err := c.getJSON(url, &doc); err != nil {
		return bmc.JobInfo{}, err
	}

	return bmc.JobInfo{
		ID:              doc.ID,
		Name:            doc.Name,
		JobState:        doc.JobState,
		PercentComplete: doc.PercentComplete,
		StartTime:       doc.StartTime,
	}, nil
}
