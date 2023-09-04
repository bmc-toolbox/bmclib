package redfish

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	gofishcommon "github.com/stmcginnis/gofish/common"
	gofishrf "github.com/stmcginnis/gofish/redfish"
)

func (c *Conn) activeTask(ctx context.Context) (*gofishrf.Task, error) {
	tasks, err := c.redfishwrapper.Tasks(ctx)
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		fmt.Println(t.TaskState)
		fmt.Println(t.TaskStatus)
		fmt.Println("xxx")
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
		//task, err = c.redfishwrapper.Tasks(ctx)
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
			bmclibErrs.ErrNotImplemented,
			"purgeFirmwareInstallTask() for vendor: "+vendor,
		)
	}

	return err
}

func (c *Conn) openbmcGetTask(jsonstr []byte) (task *gofishrf.Task, err error) {

	type TaskStatus struct {
		TaskState  string
		TaskStatus string
	}

	var status TaskStatus

	err = json.Unmarshal(jsonstr, &status)
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
