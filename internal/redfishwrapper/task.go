package redfishwrapper

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

var errUnexpectedTaskState = errors.New("unexpected task state")

// Task returns the redfish task matching taskID, or ErrTaskNotFound if none matches.
func (c *Client) Task(ctx context.Context, taskID string) (*schemas.Task, error) {
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

// TaskStatus returns the converted task state and a human readable status string for the given taskID.
func (c *Client) TaskStatus(ctx context.Context, taskID string) (constants.TaskState, string, error) {
	task, err := c.Task(ctx, taskID)
	if err != nil {
		return "", "", errors.Wrap(err, "error querying redfish for taskID: "+taskID)
	}

	taskInfo := fmt.Sprintf(
		"id: %s, state: %s, status: %s",
		task.ID,
		task.TaskState,
		task.TaskStatus,
	)

	// task message include information that help debug a cause of failure
	if msgs := c.taskMessagesAsString(task.Messages); msgs != "" {
		taskInfo += ", messages: " + msgs
	}

	s := c.ConvertTaskState(string(task.TaskState))
	return s, taskInfo, nil
}

func (c *Client) taskMessagesAsString(messages []schemas.Message) string {
	if len(messages) == 0 {
		return ""
	}

	var found []string
	for i := range messages {
		m := &messages[i]
		if m.Message == "" {
			continue
		}

		found = append(found, m.Message)
	}

	return strings.Join(found, ",")
}

// ConvertTaskState maps a redfish task state string to a bmclib TaskState constant.
func (c *Client) ConvertTaskState(state string) constants.TaskState {
	switch strings.ToLower(state) {
	case "starting", "downloading", "downloaded", "scheduling":
		return constants.Initializing
	case "running", "stopping", "cancelling": //nolint:misspell // Redfish API uses British spelling
		return constants.Running
	case "pending", "new":
		return constants.Queued
	case "scheduled":
		return constants.PowerCycleHost
	case "interrupted", "killed", "exception", "cancelled", "suspended", "failed": //nolint:misspell // Redfish API uses British spelling
		return constants.Failed
	case "completed":
		return constants.Complete
	default:
		return constants.Unknown
	}
}

// TaskStateActive reports whether the given task state indicates the task is still in progress.
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

// GetTasks returns the members of the BMC task collection (TaskService/Tasks). It returns
// ErrTaskServiceNotFound when the service root has no TaskService.
func (c *Client) GetTasks(ctx context.Context) ([]bmc.Task, error) {
	service, err := c.client.Service.Tasks()
	if err != nil {
		return nil, fmt.Errorf("TaskService: %w", err)
	}
	if service == nil {
		return nil, bmclibErrs.ErrTaskServiceNotFound
	}

	tasks, err := service.Tasks()
	if err != nil {
		return nil, fmt.Errorf("TaskService tasks: %w", err)
	}

	out := make([]bmc.Task, 0, len(tasks))
	for _, task := range tasks {
		if task == nil {
			continue
		}
		out = append(out, bmc.Task{
			ID:              task.ID,
			Name:            task.Name,
			TaskState:       string(task.TaskState),
			TaskStatus:      string(task.TaskStatus),
			State:           c.ConvertTaskState(string(task.TaskState)),
			PercentComplete: percentComplete(task.PercentComplete),
			StartTime:       task.StartTime,
			EndTime:         task.EndTime,
			Messages:        messageTexts(task.Messages),
		})
	}

	return out, nil
}

// percentComplete converts a Redfish PercentComplete value, nil when not reported.
func percentComplete(v *uint) *int {
	if v == nil {
		return nil
	}
	p := int(*v)
	return &p
}

// messageTexts returns the non-empty message texts of a Redfish message array, nil when none.
func messageTexts(messages []schemas.Message) []string {
	var out []string
	for i := range messages {
		if messages[i].Message == "" {
			continue
		}
		out = append(out, messages[i].Message)
	}
	return out
}
