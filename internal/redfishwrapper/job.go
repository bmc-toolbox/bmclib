package redfishwrapper

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// GetJobs returns the members of the BMC job collection (JobService/Jobs). It returns
// ErrJobServiceNotFound when the service root has no JobService.
func (c *Client) GetJobs(ctx context.Context) ([]bmc.Job, error) {
	service, err := c.client.Service.JobService()
	if err != nil {
		return nil, fmt.Errorf("JobService: %w", err)
	}
	if service == nil {
		return nil, bmclibErrs.ErrJobServiceNotFound
	}

	jobs, err := c.Jobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("JobService jobs: %w", err)
	}

	out := make([]bmc.Job, 0, len(jobs))
	for _, job := range jobs {
		if job == nil {
			continue
		}
		out = append(out, bmc.Job{
			ID:              job.ID,
			Name:            job.Name,
			JobState:        string(job.JobState),
			JobStatus:       string(job.JobStatus),
			PercentComplete: percentComplete(job.PercentComplete),
			StartTime:       job.StartTime,
			EndTime:         job.EndTime,
			Messages:        messageTexts(job.Messages),
		})
	}

	return out, nil
}
