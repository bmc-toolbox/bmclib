package supermicro

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	brrs "github.com/bmc-toolbox/bmclib/v2/errors"
	rfw "github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/redfish"
	"golang.org/x/exp/slices"
)

var (
	errUploadTaskIDEmpty = errors.New("firmware upload request returned empty firmware upload verify TaskID")
)

func (c *x12) componentSupported(component string) error {
	errComponentNotSupported := fmt.Errorf("component %s on device %s not supported", component, c.model)

	supported := []string{common.SlugBIOS, common.SlugBMC}
	if !slices.Contains(supported, strings.ToUpper(component)) {
		return errComponentNotSupported
	}

	return nil
}

func (c *x12) firmwareInstallSteps(component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.componentSupported(component); err != nil {
		return nil, err
	}

	return []constants.FirmwareInstallStep{
		constants.FirmwareInstallStepUpload,
		constants.FirmwareInstallStepUploadStatus,
		constants.FirmwareInstallStepInstallUploaded,
		constants.FirmwareInstallStepInstallStatus,
	}, nil
}

// upload firmware
func (c *x12) firmwareUpload(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	if err = c.componentSupported(component); err != nil {
		return "", err
	}

	err = c.firmwareTaskActive(ctx, component)
	if err != nil {
		return "", err
	}

	targetID, err := c.redfishOdataID(ctx, component)
	if err != nil {
		return "", err
	}

	params, err := c.redfishParameters(component, targetID)
	if err != nil {
		return "", err
	}

	taskID, err = c.redfish.FirmwareUpload(ctx, file, params)
	if err != nil {
		if strings.Contains(err.Error(), "OemFirmwareAlreadyInUpdateMode") {
			return "", errors.Wrap(brrs.ErrBMCColdResetRequired, "BMC currently in update mode, either continue the update OR if no update is currently running - reset the BMC")
		}

		return "", errors.Wrap(err, "error in firmware upload")
	}

	if taskID == "" {
		return "", errUploadTaskIDEmpty
	}

	return taskID, nil
}

// returns an error when a bmc firmware install is active
func (c *x12) firmwareTaskActive(ctx context.Context, component string) error {
	tasks, err := c.redfish.Tasks(ctx)
	if err != nil {
		return errors.Wrap(err, "error querying redfish tasks")
	}

	for _, t := range tasks {
		t := t

		if stateFinalized(t.TaskState) {
			continue
		}

		if err := noTasksRunning(component, t); err != nil {
			return err
		}
	}

	return nil
}

// noTasksRunning returns an error if a firmware related task was found active
func noTasksRunning(component string, t *redfish.Task) error {
	errTaskActive := errors.New("A firmware task was found active for component: " + component)

	const (
		// The redfish task name when the BMC is verifies the uploaded BMC firmware.
		verifyBMCFirmware = "BMC Verify"
		// The redfish task name when the BMC is installing the uploaded BMC firmware.
		updateBMCFirmware = "BMC Update"
		// The redfish task name when the BMC is verifies the uploaded BIOS firmware.
		verifyBIOSFirmware = "BIOS Verify"
		// The redfish task name when the BMC is installing the uploaded BIOS firmware.
		updateBIOSFirmware = "BIOS Update"
	)

	var verifyTaskName, updateTaskName string

	switch strings.ToUpper(component) {
	case common.SlugBMC:
		verifyTaskName = verifyBMCFirmware
		updateTaskName = updateBMCFirmware
	case common.SlugBIOS:
		verifyTaskName = verifyBIOSFirmware
		updateTaskName = updateBIOSFirmware
	}

	taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", t.ID, t.TaskState, t.TaskStatus)

	switch t.Name {
	case verifyTaskName:
		return errors.Wrap(errTaskActive, taskInfo)
	case updateTaskName:
		return errors.Wrap(errTaskActive, taskInfo)
	default:
		return nil
	}
}

func stateFinalized(s redfish.TaskState) bool {
	finalized := []redfish.TaskState{
		redfish.CompletedTaskState,
		redfish.CancelledTaskState,
		redfish.InterruptedTaskState,
		redfish.ExceptionTaskState,
	}

	return slices.Contains(finalized, s)
}

// redfish OEM parameter structs
type BIOS struct {
	PreserveME         bool `json:"PreserveME"`
	PreserveNVRAM      bool `json:"PreserveNVRAM"`
	PreserveSMBIOS     bool `json:"PreserveSMBIOS"`
	PreserveOA         bool `json:"PreserveOA"`
	PreserveSETUPCONF  bool `json:"PreserveSETUPCONF"`
	PreserveSETUPPWD   bool `json:"PreserveSETUPPWD"`
	PreserveSECBOOTKEY bool `json:"PreserveSECBOOTKEY"`
	PreserveBOOTCONF   bool `json:"PreserveBOOTCONF"`
}

type BMC struct {
	PreserveCfg bool `json:"PreserveCfg"`
	PreserveSdr bool `json:"PreserveSdr"`
	PreserveSsl bool `json:"PreserveSsl"`
}

type Supermicro struct {
	*BIOS `json:"BIOS,omitempty"`
	*BMC  `json:"BMC,omitempty"`
}

type OEM struct {
	Supermicro `json:"Supermicro"`
}

func (c *x12) redfishParameters(component, targetODataID string) (*rfw.RedfishUpdateServiceParameters, error) {
	errUnsupported := errors.New("redfish parameters for x12 hardware component not supported: " + component)

	oem := OEM{}

	switch strings.ToUpper(component) {
	case common.SlugBIOS:
		oem.Supermicro.BIOS = &BIOS{
			PreserveME:         false,
			PreserveNVRAM:      false,
			PreserveSMBIOS:     true,
			PreserveOA:         true,
			PreserveSETUPCONF:  true,
			PreserveSETUPPWD:   true,
			PreserveSECBOOTKEY: true,
			PreserveBOOTCONF:   true,
		}
	case common.SlugBMC:
		oem.Supermicro.BMC = &BMC{
			PreserveCfg: true,
			PreserveSdr: true,
			PreserveSsl: true,
		}
	default:
		return nil, errUnsupported
	}

	b, err := json.Marshal(oem)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing redfish parameters")
	}

	return &rfw.RedfishUpdateServiceParameters{
		OperationApplyTime: constants.OnStartUpdateRequest,
		Targets:            []string{targetODataID},
		Oem:                b,
	}, nil
}

func (c *x12) redfishOdataID(ctx context.Context, component string) (string, error) {
	errUnsupported := errors.New("unable to return redfish OData ID for unsupported component: " + component)

	switch strings.ToUpper(component) {
	case common.SlugBMC:
		return c.redfish.ManagerOdataID(ctx)
	case common.SlugBIOS:
		// hardcoded since SMCs without the DCMS license will throw license errors
		return "/redfish/v1/Systems/1/Bios", nil
		//return c.redfish.SystemsBIOSOdataID(ctx)
	}

	return "", errUnsupported
}

func (c *x12) firmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error) {
	if err = c.componentSupported(component); err != nil {
		return "", err
	}

	task, err := c.redfish.Task(ctx, uploadTaskID)
	if err != nil {
		e := fmt.Sprintf("error querying redfish tasks for firmware upload taskID: %s, err: %s", uploadTaskID, err.Error())
		return "", errors.Wrap(brrs.ErrFirmwareVerifyTask, e)
	}

	taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", task.ID, task.TaskState, task.TaskStatus)

	if task.TaskState != redfish.CompletedTaskState {
		return "", errors.Wrap(brrs.ErrFirmwareVerifyTask, taskInfo)
	}

	if task.TaskStatus != "OK" {
		return "", errors.Wrap(brrs.ErrFirmwareVerifyTask, taskInfo)
	}

	return c.redfish.StartUpdateForUploadedFirmware(ctx)
}

func (c *x12) firmwareTaskStatus(ctx context.Context, component, taskID string) (state, status string, err error) {
	if err = c.componentSupported(component); err != nil {
		return "", "", errors.Wrap(brrs.ErrFirmwareTaskStatus, err.Error())
	}

	return c.redfish.TaskStatus(ctx, taskID)
}
