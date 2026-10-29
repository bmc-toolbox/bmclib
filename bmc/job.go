package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// Job is an entry of the BMC job collection (JobService/Jobs): scheduled or queued work such
// as a pending BIOS configuration job on iDRAC or a scheduled restart on XCC.
type Job struct {
	// ID is the Redfish Job resource ID, e.g. "JID_123456789012".
	ID string
	// Name is the job name as reported by the BMC, e.g. "Configure: BIOS.Setup.1-1".
	Name string
	// JobState is the Redfish job state, e.g. "New", "Scheduled", "Running", "Completed", "Exception".
	JobState string
	// JobStatus is the Redfish job health, e.g. "OK", "Warning", "Critical".
	JobStatus string
	// PercentComplete is the job progress in percent; nil when the BMC does not report it.
	PercentComplete *int
	// StartTime and EndTime are the job timestamps as reported by the BMC; empty when not reported.
	StartTime string
	EndTime   string
	// Messages are the non-empty message texts attached to the job.
	Messages []string
}

// JobsGetter provides retrieval of the BMC job collection (JobService/Jobs).
type JobsGetter interface {
	GetJobs(ctx context.Context) (jobs []Job, err error)
}

type jobsGetterProvider struct {
	name string
	JobsGetter
}

func getJobs(ctx context.Context, generic []jobsGetterProvider) (jobs []Job, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.JobsGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			jobs, vErr := elem.GetJobs(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return jobs, metadata, nil
		}
	}

	return nil, metadata, multierror.Append(err, errors.New("failure to get jobs"))
}

// GetJobsFromInterfaces identifies implementations of the JobsGetter interface and passes them
// to the getJobs() wrapper.
func GetJobsFromInterfaces(ctx context.Context, generic []interface{}) (jobs []Job, metadata Metadata, err error) {
	implementations := make([]jobsGetterProvider, 0)
	for _, elem := range generic {
		temp := jobsGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case JobsGetter:
			temp.JobsGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a JobsGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return nil, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no JobsGetter implementations found",
			),
		)
	}

	return getJobs(ctx, implementations)
}
