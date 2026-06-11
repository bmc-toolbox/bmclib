package lenovo

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/constants"
)

// simpleUpdateURI is the XCC UpdateService.SimpleUpdate action.
const simpleUpdateURI = "/redfish/v1/UpdateService/Actions/UpdateService.SimpleUpdate"

// compile-time assertions that the provider implements the firmware interfaces.
var (
	_ bmc.FirmwareInstaller          = (*Conn)(nil)
	_ bmc.FirmwareUploader           = (*Conn)(nil)
	_ bmc.FirmwareInstallerUploaded  = (*Conn)(nil)
	_ bmc.FirmwareInstallProvider    = (*Conn)(nil)
	_ bmc.FirmwareInstallVerifier    = (*Conn)(nil)
	_ bmc.FirmwareTaskVerifier       = (*Conn)(nil)
	_ bmc.FirmwareInstallStepsGetter = (*Conn)(nil)
)

// FirmwareInstall uploads a firmware image and initiates the install in a single
// step using the XCC push protocol, returning the install task id.
//
// component is the XCC FirmwareInventory target id (e.g. "BMC-Backup", "UEFI")
// or empty to let the XCC auto-detect the target from the image.
// operationApplyTime is one of the bmclib OperationApplyTime values (defaults to
// Immediate). Implements bmc.FirmwareInstaller.
func (c *Conn) FirmwareInstall(ctx context.Context, component, operationApplyTime string, forceInstall bool, reader io.Reader) (taskID string, err error) {
	file, cleanup, err := tempFileFromReader(reader)
	if err != nil {
		return "", err
	}
	defer cleanup()

	return c.pushFirmware(ctx, component, file, operationApplyTimeOrDefault(operationApplyTime))
}

// FirmwareInstallUploadAndInitiate uploads a firmware image and initiates the
// install in a single step, returning the task id.
//
// Implements bmc.FirmwareInstallProvider.
func (c *Conn) FirmwareInstallUploadAndInitiate(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	return c.pushFirmware(ctx, component, file, constants.Immediate)
}

// FirmwareUpload uploads a firmware image, returning the upload/verify task id.
//
// The image is pushed with OperationApplyTime "OnStartUpdateRequest" so the
// install is started separately by FirmwareInstallUploaded. Implements
// bmc.FirmwareUploader.
func (c *Conn) FirmwareUpload(ctx context.Context, component string, file *os.File) (uploadVerifyTaskID string, err error) {
	return c.pushFirmware(ctx, component, file, constants.OnStartUpdateRequest)
}

// FirmwareInstallUploaded starts the install for firmware previously uploaded
// with FirmwareUpload, returning the install task id.
//
// Implements bmc.FirmwareInstallerUploaded.
func (c *Conn) FirmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error) {
	return c.redfishwrapper.StartUpdateForUploadedFirmware(ctx)
}

// FirmwareInstallStatus returns the install status for the given task id.
//
// Implements bmc.FirmwareInstallVerifier.
func (c *Conn) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (status string, err error) {
	state, _, err := c.firmwareTask(ctx, taskID)
	if err != nil {
		return "", err
	}

	return string(state), nil
}

// FirmwareTaskStatus returns the state and status of a firmware upload/install
// task.
//
// The task is read from the Task resource — never the XCC TaskMonitor URI, since
// a GET on TaskMonitor deletes a finished task. When the task reaches a terminal
// state the update service claim is released. Implements bmc.FirmwareTaskVerifier.
func (c *Conn) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state constants.TaskState, status string, err error) {
	return c.firmwareTask(ctx, taskID)
}

// FirmwareInstallSteps returns the ordered steps the provider performs for a
// firmware install: the XCC push uploads and initiates the install in one step,
// followed by polling the install task.
//
// Implements bmc.FirmwareInstallStepsGetter.
func (c *Conn) FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error) {
	return []constants.FirmwareInstallStep{
		constants.FirmwareInstallStepUploadInitiateInstall,
		constants.FirmwareInstallStepInstallStatus,
	}, nil
}

// SimpleUpdate triggers an UpdateService.SimpleUpdate where the XCC pulls the
// image from imageURI itself, returning the created task id.
//
// transferProtocol is an optional Redfish TransferProtocol (e.g. "HTTPS",
// "SFTP"); when empty it is inferred by the XCC from the URI scheme. This is an
// XCC-specific provider method (not part of a bmc.Feature interface).
func (c *Conn) SimpleUpdate(ctx context.Context, imageURI, transferProtocol string) (taskID string, err error) {
	payload := map[string]any{"ImageURI": imageURI}
	if transferProtocol != "" {
		payload["TransferProtocol"] = transferProtocol
	}

	resp, err := c.redfishwrapper.PostWithHeaders(ctx, simpleUpdateURI, payload, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", parseRedfishError(resp)
	}

	if loc := resp.Header.Get("Location"); loc != "" {
		return path.Base(strings.TrimRight(loc, "/")), nil
	}

	return "", nil
}

// firmwareTask polls the Task resource for a firmware task and releases the
// update service claim once the task reaches a terminal state.
func (c *Conn) firmwareTask(ctx context.Context, taskID string) (constants.TaskState, string, error) {
	state, status, err := c.redfishwrapper.TaskStatus(ctx, taskID)
	if err != nil {
		return "", "", err
	}

	if state == constants.Complete || state == constants.Failed {
		// Best-effort release of the busy claim once the update is done.
		_ = c.releaseUpdateService(ctx, true)
	}

	return state, status, nil
}

// tempFileFromReader writes reader to a temporary file and returns it positioned
// at the start, along with a cleanup function that closes and removes it.
func tempFileFromReader(reader io.Reader) (*os.File, func(), error) {
	f, err := os.CreateTemp("", "lenovo-fw-*")
	if err != nil {
		return nil, func() {}, fmt.Errorf("creating temp firmware file: %w", err)
	}

	cleanup := func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}

	if _, err := io.Copy(f, reader); err != nil {
		cleanup()
		return nil, func() {}, fmt.Errorf("buffering firmware image: %w", err)
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		cleanup()
		return nil, func() {}, fmt.Errorf("rewinding firmware image: %w", err)
	}

	return f, cleanup, nil
}
