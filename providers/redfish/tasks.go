package redfish

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	gofishcommon "github.com/stmcginnis/gofish/common"
	gofishrf "github.com/stmcginnis/gofish/redfish"
)

func (c *Conn) activeTask(ctx context.Context) (*gofishrf.Task, error) {
	resp, err := c.redfishwrapper.Get("/redfish/v1/TaskService/Tasks")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		err = errors.Wrap(
			bmclibErrs.ErrFirmwareInstallStatus,
			"HTTP Error: "+fmt.Sprint(resp.StatusCode),
		)

		return nil, err
	}

	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	type TaskId struct {
		OdataID    string `json:"@odata.id"`
		TaskState  string
		TaskStatus string
	}

	type Tasks struct {
		Members []TaskId
	}

	var status Tasks

	err = json.Unmarshal(data, &status)
	if err != nil {
		fmt.Println(err)
	}

	// For each task, check if it's running
	// It's usually the latest that is running, so it would be faster to
	// start by the end, but an easy way to do this is only available in go 1.21
	// for _, t := range slices.Reverse(status.Members) { // when go 1.21
	for _, t := range status.Members {
		re := regexp.MustCompile("/redfish/v1/TaskService/Tasks/([0-9]+)")
		taskmatch := re.FindSubmatch([]byte(t.OdataID))
		if len(taskmatch) < 1 {
			continue
		}

		tasknum := string(taskmatch[1])

		task, err := c.GetTask(tasknum)
		if err != nil {
			continue
		}

		if task.TaskState == "Running" {
			return task, nil
		}
	}

	return nil, nil
}

// GetFirmwareInstallTaskQueued returns the redfish task object for a queued update task
func (c *Conn) GetFirmwareInstallTaskQueued(ctx context.Context, component string) (*gofishrf.Task, error) {
	vendor, _, err := c.DeviceVendorModel(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine device vendor, model attributes")
	}

	var task *gofishrf.Task

	// check an update task for the component is currently scheduled
	switch {
	case strings.Contains(vendor, constants.Dell):
		task, err = c.getDellFirmwareInstallTaskScheduled(component)
	default:
		task, err = c.activeTask(ctx)
	}

	if err != nil {
		return nil, err
	}

	return task, nil
}

// purgeQueuedFirmwareInstallTask removes any existing queued firmware install task for the given component slug
func (c *Conn) purgeQueuedFirmwareInstallTask(ctx context.Context, component string) error {
	vendor, _, err := c.DeviceVendorModel(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to determine device vendor, model attributes")
	}

	// check an update task for the component is currently scheduled
	switch {
	case strings.Contains(vendor, constants.Dell):
		err = c.dellPurgeScheduledFirmwareInstallJob(component)
	default:
		err = errors.Wrap(
			bmclibErrs.ErrFirmwareInstall,
			"Update is already running",
		)
	}

	return err
}

// GetTask returns the current Task fir the given TaskID
func (c *Conn) GetTask(taskID string) (task *gofishrf.Task, err error) {

	resp, err := c.redfishwrapper.Get("/redfish/v1/TaskService/Tasks/" + taskID)
	if err != nil {
		if strings.HasPrefix(err.Error(), "404") {
			return nil, errors.Wrap(bmclibErrs.ErrTaskNotFound, "task with ID not found: "+taskID)
		}
		return nil, err
	}
	if resp.StatusCode != 200 {
		err = errors.Wrap(
			bmclibErrs.ErrFirmwareInstallStatus,
			"HTTP Error: "+fmt.Sprint(resp.StatusCode),
		)

		return nil, err
	}

	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	type TaskStatus struct {
		TaskState  string
		TaskStatus string
	}

	var status TaskStatus

	err = json.Unmarshal(data, &status)
	if err != nil {
		fmt.Println(err)
	} else {
		task = &gofishrf.Task{
			TaskState:  gofishrf.TaskState(status.TaskState),
			TaskStatus: gofishcommon.Health(status.TaskStatus),
		}
	}

	return task, err
}
