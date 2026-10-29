package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.TasksGetter = (*Conn)(nil)
	_ bmc.JobsGetter  = (*Conn)(nil)
)

// GetTasks returns the BMC task collection.
//
// Implements bmc.TasksGetter.
func (c *Conn) GetTasks(ctx context.Context) (tasks []bmc.Task, err error) {
	return c.redfishwrapper.GetTasks(ctx)
}

// GetJobs returns the BMC job collection.
//
// Implements bmc.JobsGetter.
func (c *Conn) GetJobs(ctx context.Context) (jobs []bmc.Job, err error) {
	return c.redfishwrapper.GetJobs(ctx)
}
