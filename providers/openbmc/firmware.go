package openbmc

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/constants"

	bmcliberrs "github.com/metal-toolbox/bmclib/errors"
	rfw "github.com/metal-toolbox/bmclib/internal/redfishwrapper"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/redfish"
)

// bmc client interface implementations methods
func (c *Conn) FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.deviceSupported(ctx); err != nil {
		return nil, err
	}

	switch strings.ToUpper(component) {
	case common.SlugBIOS:
		return []constants.FirmwareInstallStep{
			constants.FirmwareInstallStepPowerOffHost,
			constants.FirmwareInstallStepUploadInitiateInstall,
			constants.FirmwareInstallStepInstallStatus,
		}, nil
	case common.SlugBMC:
		return []constants.FirmwareInstallStep{
			constants.FirmwareInstallStepUploadInitiateInstall,
			constants.FirmwareInstallStepInstallStatus,
		}, nil
	default:
		return nil, errors.New("component firmware install not supported: " + component)
	}
}

func (c *Conn) FirmwareInstallUploadAndInitiate(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	if err := c.deviceSupported(ctx); err != nil {
		return "", errNotOpenBMCDevice
	}

	//	// expect atleast 5 minutes left in the deadline to proceed with the upload
	d, _ := ctx.Deadline()
	if time.Until(d) < 10*time.Minute {
		return "", errors.New("remaining context deadline insufficient to perform update: " + time.Until(d).String())
	}

	// list current tasks on BMC
	tasks, err := c.redfishwrapper.Tasks(ctx)
	if err != nil {
		return "", errors.Wrap(err, "error listing bmc redfish tasks")
	}

	// validate a new firmware install task can be queued
	if err := c.checkQueueability(component, tasks); err != nil {
		return "", errors.Wrap(bmcliberrs.ErrFirmwareInstall, err.Error())
	}

	params := &rfw.RedfishUpdateServiceParameters{
		Targets:            []string{},
		OperationApplyTime: constants.OnReset,
		Oem:                []byte(`{}`),
	}

	return c.redfishwrapper.FirmwareUpload(ctx, file, params)
}

// returns an error when a bmc firmware install is active
func (c *Conn) checkQueueability(component string, tasks []*redfish.Task) error {
	errTaskActive := errors.New("A firmware job was found active for component: " + component)

	for _, t := range tasks {
		// taskInfo returned in error if any.
		taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", t.ID, t.TaskState, t.TaskStatus)

		// convert redfish task state to bmclib state
		convstate := c.redfishwrapper.ConvertTaskState(string(t.TaskState))
		// check if task is active based on converted state
		active, err := c.redfishwrapper.TaskStateActive(convstate)
		if err != nil {
			return errors.Wrap(err, taskInfo)
		}

		if active {
			return errors.Wrap(errTaskActive, taskInfo)
		}
	}

	return nil
}

// FirmwareTaskStatus returns the status of a firmware related task queued on the BMC.
func (c *Conn) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state constants.TaskState, status string, err error) {
	return c.redfishwrapper.TaskStatus(ctx, taskID)
}
