package redfishwrapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/common"
	redfish "github.com/stmcginnis/gofish/redfish"
)

var (
	errUnexpectedTaskState = errors.New("unexpected task state")
)

func (c *Client) Task(ctx context.Context, taskID string) (*redfish.Task, error) {
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

func (c *Client) TaskStatus(ctx context.Context, taskID string) (constants.TaskState, string, error) {
	task, err := c.Task(ctx, taskID)
	if err != nil {
		return "", "", errors.Wrap(err, "error querying redfish for taskID: "+taskID)
	}
	taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", task.ID, task.TaskState, task.TaskStatus)

	state := strings.ToLower(string(task.TaskState))
	return c.ConvertTaskState(state), taskInfo, nil
}

func (c *Client) ConvertTaskState(state string) constants.TaskState {
	switch state {
	case "starting", "downloading", "downloaded":
		return constants.Initializing
	case "running", "stopping", "cancelling", "scheduling":
		return constants.Running
	case "pending", "new":
		return constants.Queued
	case "scheduled":
		return constants.PowerCycleHost
	case "interrupted", "killed", "exception", "cancelled", "suspended", "failed":
		return constants.Failed
	case "completed":
		return constants.Complete
	default:
		return constants.Unknown
	}
}

func (c *Client) TaskStateActive(state constants.TaskState) (bool, error) {
	switch state {
	case constants.Initializing, constants.Running, constants.Queued:
		return true, nil
	case constants.Complete, constants.Failed:
		return false, nil
	default:
		return false, errors.Wrap(errUnexpectedTaskState, string(state))
	}
}
