package supermicro

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/constants"
	brrs "github.com/metal-toolbox/bmclib/errors"
	rfw "github.com/metal-toolbox/bmclib/internal/redfishwrapper"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/redfish"
	"golang.org/x/exp/slices"
)

type x13 struct {
	*serviceClient
	model string
	log   logr.Logger
}

func newX13Client(client *serviceClient, logger logr.Logger) bmcQueryor {
	return &x13{
		serviceClient: client,
		log:           logger,
	}
}

func (c *x13) deviceModel() string {
	return c.model
}

func (c *x13) queryDeviceModel(ctx context.Context) (string, error) {
	if err := c.redfishSession(ctx); err != nil {
		return "", err
	}

	_, model, err := c.redfish.DeviceVendorModel(ctx)
	if err != nil {
		return "", err
	}

	if model == "" {
		return "", errors.Wrap(ErrModelUnknown, "empty value")
	}

	c.model = common.FormatProductName(model)

	return c.model, nil
}

func (c *x13) supportsInstall(component string) error {
	errComponentNotSupported := fmt.Errorf("component %s on device %s not supported", component, c.model)

	supported := []string{common.SlugBIOS, common.SlugBMC}
	if !slices.Contains(supported, strings.ToUpper(component)) {
		return errComponentNotSupported
	}

	return nil
}

func (c *x13) firmwareInstallSteps(component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.supportsInstall(component); err != nil {
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
func (c *x13) firmwareUpload(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	if err = c.supportsInstall(component); err != nil {
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
		return "", ErrUploadTaskIDEmpty
	}

	return taskID, nil
}

// returns an error when a bmc firmware install is active
func (c *x13) firmwareTaskActive(ctx context.Context, component string) error {
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

// redfish OEM fw install parameters
func (c *x13) biosFwInstallParams() (map[string]bool, error) {
	switch c.model {
	case "x13dem":
		return map[string]bool{
			"PreserveSMBIOS":     true,
			"PreserveOA":         true,
			"PreserveSETUPCONF":  true,
			"PreserveSETUPPWD":   true,
			"PreserveSECBOOTKEY": true,
			"PreserveBOOTCONF":   true,
			"BackupBIOS":         false,
		}, nil
	default:
		// ideally we never get in this position, since theres model number validation in parent callers.
		return nil, errors.New("unsupported model for X13 BIOS fw install: " + c.model)
	}
}

// redfish OEM fw install parameters
func (c *x13) bmcFwInstallParams() map[string]bool {
	return map[string]bool{
		"PreserveCfg": true,
		"PreserveSdr": true,
		"PreserveSsl": true,
		"BackupBMC":   false,
	}
}

func (c *x13) redfishParameters(component, targetODataID string) (*rfw.RedfishUpdateServiceParameters, error) {
	errUnsupported := errors.New("redfish parameters for x13 hardware component not supported: " + component)

	oem := OEM{}

	biosInstallParams, err := c.biosFwInstallParams()
	if err != nil {
		return nil, err
	}

	switch strings.ToUpper(component) {
	case common.SlugBIOS:
		oem.Supermicro.BIOS = biosInstallParams
	case common.SlugBMC:
		oem.Supermicro.BMC = c.bmcFwInstallParams()
	default:
		return nil, errUnsupported
	}

	b, err := json.Marshal(oem)
	if err != nil {
		return nil, errors.Wrap(err, "error preparing redfish parameters")
	}

	return &rfw.RedfishUpdateServiceParameters{
		// NOTE:
		// X13s support the OnReset Apply time for BIOS updates if we want to implement that in the future.
		OperationApplyTime: constants.OnStartUpdateRequest,
		Targets:            []string{targetODataID},
		Oem:                b,
	}, nil
}

func (c *x13) redfishOdataID(ctx context.Context, component string) (string, error) {
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

func (c *x13) firmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error) {
	if err = c.supportsInstall(component); err != nil {
		return "", err
	}

	task, err := c.redfish.Task(ctx, uploadTaskID)
	if err != nil {
		e := fmt.Sprintf("error querying redfish tasks for firmware upload taskID: %s, err: %s", uploadTaskID, err.Error())
		return "", errors.Wrap(brrs.ErrFirmwareVerifyTask, e)
	}

	taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", task.ID, task.TaskState, task.TaskStatus)

	if task.TaskState != redfish.CompletedTaskState && task.TaskState != redfish.PendingTaskState {
		return "", errors.Wrap(brrs.ErrFirmwareVerifyTask, taskInfo)
	}

	if task.TaskStatus != "OK" {
		return "", errors.Wrap(brrs.ErrFirmwareVerifyTask, taskInfo)
	}

	err = c.redfish.StartUpdateForUploadedFirmwareNoTaskID(ctx)
	if err != nil {
		return "", err
	}

	// X13s dont create a new task id when going from upload to install, so we pass through the same one
	return uploadTaskID, nil
}

func (c *x13) firmwareTaskStatus(ctx context.Context, component, taskID string) (state constants.TaskState, status string, err error) {
	if err = c.supportsInstall(component); err != nil {
		return "", "", errors.Wrap(brrs.ErrFirmwareTaskStatus, err.Error())
	}

	return c.redfish.TaskStatus(ctx, taskID)
}

func (c *x13) getBootProgress() (*redfish.BootProgress, error) {
	bps, err := c.redfish.GetBootProgress()
	if err != nil {
		return nil, err
	}
	return bps[0], nil
}

// this is some syntactic sugar to avoid having to code potentially provider- or model-specific knowledge into a caller
func (c *x13) bootComplete() (bool, error) {
	bp, err := c.getBootProgress()
	if err != nil {
		return false, err
	}

	// we determined this by experiment on X12STH-SYS with redfish 1.14.0
	return bp.LastState == redfish.SystemHardwareInitializationCompleteBootProgressTypes, nil
}
