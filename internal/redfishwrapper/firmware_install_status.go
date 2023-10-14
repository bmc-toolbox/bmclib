package redfishwrapper

import (
	"context"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	gofishrf "github.com/stmcginnis/gofish/redfish"
)

// FirmwareInstallStatus returns the status of the firmware install task queued
func (c *Client) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (state string, err error) {
	var task *gofishrf.Task
	task, err = c.Task(ctx, taskID)
	if err != nil {
		return state, err
	}

	state = strings.ToLower(string(task.TaskState))

	// so much for standards...
	switch state {
	case "starting", "downloading", "downloaded":
		return constants.FirmwareInstallInitializing, nil
	case "running", "stopping", "cancelling", "scheduling":
		return constants.FirmwareInstallRunning, nil
	case "pending", "new":
		return constants.FirmwareInstallQueued, nil
	case "scheduled":
		return constants.FirmwareInstallPowerCyleHost, nil
	case "interrupted", "killed", "exception", "cancelled", "suspended", "failed":
		return constants.FirmwareInstallFailed, nil
	case "completed":
		return constants.FirmwareInstallComplete, nil
	default:
		return constants.FirmwareInstallUnknown + ": " + state, nil
	}

}

func (c *Client) Task(ctx context.Context, taskID string) (*gofishrf.Task, error) {
	c.client.Service.Tasks()
	tasks, err := c.Tasks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error querying redfish tasks")
	}

	for _, t := range tasks {
		if t.ID != taskID {
			continue
		}

		return t, nil
	}

	return nil, bmclibErrs.ErrTaskNotFound
}
