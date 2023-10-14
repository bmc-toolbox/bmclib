package supermicro

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	brrs "github.com/bmc-toolbox/bmclib/v2/errors"
	rfw "github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/redfish"
)

var (
	errInsufficientCtxTimeout = errors.New("remaining context timeout insufficient to install firmware")

	errBmcInstallRunning        = errors.New("BMC firmware install is currently in progress")
	errBmcVerifyRunning         = errors.New("BMC firmware verify is currently in progress")
	errBmcFirmwareTasksNotFound = errors.New("No BMC verify firmware task was found")
	//errBmcResetRequired         = errors.New("Ex")
	errVerifyTaskIDEmpty = errors.New("firmware upload returned empty firmware verify TaskID")
	errStartUpdate       = errors.New("error starting update")
)

type installMethod string

const (
	// The redfish task name when the BMC is verifies the uploaded firmware.
	taskBMCVerify = "BMC Verify"
	// The redfish task name when the BMC is updating firmware.
	taskBMCUpdate = "BMC Update"

	verifyTaskWaits = 20
	verifyTaskDelay = 2 * time.Second
)

//return errors.Wrap(
//	bmclibErrs.ErrBMCColdResetRequired,
//	"BMC verify firmware task found, please reset the BMC before queueing another install",
//)

func (c *x12) firmwareInstallActions(component string) ([]constants.FirmwareAction, error) {
	errComponentNotSupported := errors.New("firmware install for component on x12 hardware not supported")

	switch component {
	case common.SlugBMC:
		return c.firmwareInstallActionsBMC()
	case common.SlugBIOS:
		return c.firmwareInstallActionsBIOS()
	default:
		return nil, errors.Wrap(errComponentNotSupported, component)
	}
}

func (c *x12) firmwareUpload(ctx context.Context, component string, operationApplyTime constants.OperationApplyTime, reader io.Reader) (taskID string, err error) {
	errComponentNotSupported := errors.New("firmware install on x12 hardware not supported for component: " + component)

	switch component {
	case common.SlugBMC:
		return c.firmwareUploadBMC(ctx, operationApplyTime, reader)
	case common.SlugBIOS:
		//		return c.firmwareUploadBIOS()
	}

	return taskID, errComponentNotSupported
}

func (c *x12) firmwareInstallStatus(ctx context.Context, component, installVersion, taskID string) (status string, err error) {
	errComponentNotSupported := errors.New("firmware install status on x12 hardware not supported for component: " + component)

	switch component {
	case common.SlugBMC:
		// return c.firmwareInstallStatusBMC(ctx, component, installVersion, taskID)
	case common.SlugBIOS:
		//		return c.firmwareUploadBIOS()
	}

	return taskID, errComponentNotSupported
}

func (c *x12) firmwareInstallActionsBIOS() ([]constants.FirmwareAction, error) {
	return []constants.FirmwareAction{
		constants.FirmwareActionUpload,
		constants.FirmwareActionVerifyUpload,
		constants.FirmwareActionInstall,
		constants.FirmwareActionVerifyInstall,
	}, nil
}

func (c *x12) firmwareInstallActionsBMC() ([]constants.FirmwareAction, error) {
	return []constants.FirmwareAction{
		constants.FirmwareActionUpload,
		constants.FirmwareActionVerifyUpload,
		constants.FirmwareActionInstall,
		constants.FirmwareActionVerifyInstall,
	}, nil
}

// BMC

// upload firmware
func (c *x12) firmwareUploadBMC(ctx context.Context, operationApplyTime constants.OperationApplyTime, reader io.Reader) (taskID string, err error) {
	err = c.firmwareInstallActiveBMC(ctx)
	if err != nil {
		return "", err
	}

	oemParameters := []byte(`{"Supermicro":{"BMC":{"PreserveCfg":true,"PreserveSdr":true,"PreserveSsl":true}}}`)

	target, err := c.redfish.ManagerOdataID(ctx)
	if err != nil {
		return "", err
	}

	params := &rfw.RedfishUpdateServiceParameters{
		OperationApplyTime: operationApplyTime,
		Targets:            []string{target},
		Oem:                oemParameters,
	}

	taskID, err = c.redfish.FirmwareUpload(ctx, reader, params)
	if err != nil {
		return "", errors.Wrap(err, "error in firmware upload")
	}

	if taskID == "" {
		return "", errVerifyTaskIDEmpty
	}

	return taskID, nil
}

// returns an error when a bmc firmware install is active
func (c *x12) firmwareInstallActiveBMC(ctx context.Context) error {
	tasks, err := c.redfish.Tasks(ctx)
	if err != nil {
		return errors.Wrap(err, "error querying redfish tasks")
	}

	for _, t := range tasks {
		if t.TaskState == redfish.CompletedTaskState {
			continue
		}

		taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", t.ID, t.TaskState, t.TaskStatus)

		switch t.Name {
		case taskBMCUpdate:
			return errors.Wrap(errBmcInstallRunning, taskInfo)
		case taskBMCVerify:
			return errors.Wrap(errBmcVerifyRunning, taskInfo)
		}
	}

	return nil
}

func (c *x12) firmwareInstall(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error) {
	errComponentNotSupported := errors.New("firmware install on x12 hardware not supported for component: " + component)

	switch component {
	case common.SlugBMC:
		return c.firmwareInstallBMC(ctx, uploadTaskID)
	case common.SlugBIOS:
		//		return c.firmwareUploadBIOS()
	}

	return installTaskID, errComponentNotSupported
}

// firmwareInstallBMC
//
// When a ErrFirmwareVerifyTaskRunning, the caller must retry this action
func (c *x12) firmwareInstallBMC(ctx context.Context, uploadTaskID string) (installTaskID string, err error) {
	task, err := c.redfish.Task(ctx, uploadTaskID)
	if err != nil {
		return "", errors.Wrap(err, "error querying redfish tasks for firmware upload taskID: "+uploadTaskID)
	}

	taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", task.ID, task.TaskState, task.TaskStatus)

	if task.TaskState != redfish.CompletedTaskState {
		return "", errors.Wrap(brrs.ErrFirmwareVerifyTaskRunning, taskInfo)
	}

	if task.TaskStatus != "OK" {
		return "", errors.Wrap(brrs.ErrFirmwareVerifyTaskFailed, taskInfo)
	}

	return c.redfish.StartUpdateForUploadedFirmware(ctx)
}
