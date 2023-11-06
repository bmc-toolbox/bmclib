package redfishwrapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	gofishrf "github.com/stmcginnis/gofish/redfish"
)

func (c *Client) Task(ctx context.Context, taskID string) (*gofishrf.Task, error) {
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

func (c *Client) TaskStatus(ctx context.Context, taskID string) (state, status string, err error) {
	task, err := c.Task(ctx, taskID)
	if err != nil {
		return "", "", errors.Wrap(err, "error querying redfish for taskID: "+taskID)
	}

	taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", task.ID, task.TaskState, task.TaskStatus)

	state = strings.ToLower(string(task.TaskState))

	switch state {
	case "starting", "downloading", "downloaded":
		return constants.FirmwareInstallInitializing, taskInfo, nil
	case "running", "stopping", "cancelling", "scheduling":
		return constants.FirmwareInstallRunning, taskInfo, nil
	case "pending", "new":
		return constants.FirmwareInstallQueued, taskInfo, nil
	case "scheduled":
		return constants.FirmwareInstallPowerCycleHost, taskInfo, nil
	case "interrupted", "killed", "exception", "cancelled", "suspended", "failed":
		return constants.FirmwareInstallFailed, taskInfo, nil
	case "completed":
		return constants.FirmwareInstallComplete, taskInfo, nil
	default:
		return constants.FirmwareInstallUnknown, taskInfo, nil
	}
}
