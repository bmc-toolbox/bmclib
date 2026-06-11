package bmc

import "context"

// JobServiceInfo describes the Redfish JobService configuration.
type JobServiceInfo struct {
	// ServiceEnabled reports whether the job service is enabled.
	ServiceEnabled bool
}

// JobInfo describes a Redfish job.
type JobInfo struct {
	// ID is the Redfish Job Id.
	ID string
	// Name is the job name.
	Name string
	// JobState is the job state (e.g. "Scheduled", "Running", "Completed").
	JobState string
	// PercentComplete is the job completion percentage.
	PercentComplete int
	// StartTime is the scheduled/actual start time.
	StartTime string
}

// JobManager is implemented by providers that can read jobs and update a job's
// schedule.
type JobManager interface {
	// JobService returns the job-service configuration.
	JobService(ctx context.Context) (JobServiceInfo, error)
	// Jobs lists the jobs.
	Jobs(ctx context.Context) ([]JobInfo, error)
	// Job returns a job by Id.
	Job(ctx context.Context, id string) (JobInfo, error)
	// JobUpdateSchedule PATCHes a job's Schedule with the given properties.
	JobUpdateSchedule(ctx context.Context, id string, schedule map[string]any) error
}
