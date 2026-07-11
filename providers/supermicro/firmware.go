package supermicro

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

var (
	// Its likely the X11 code works on all X11's
	// for now, we list only the ones its been tested on.
	//
	// board part numbers
	//
	supportedModels = []string{
		"X11SSL-F",
		"X11SCM-F",
		"X11DPH-T",
		"X11SCH-F",
		"X11DGQ",
		"X11DPG-SN",
		"X11DPT-B",
		"X11SSE-F",
		"X12STH-SYS",
		"X12SPO-NTF",
	}

	errUploadTaskIDExpected = errors.New("expected an firmware upload taskID")
)

// FirmwareInstallSteps returns the ordered steps required to install firmware on the given component.
func (c *Client) FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.serviceClient.supportsFirmwareInstall(c.bmc.deviceModel()); err != nil {
		return nil, err
	}

	return c.bmc.firmwareInstallSteps(component)
}

// FirmwareUpload uploads the firmware image for the given component and returns the upload task ID.
func (c *Client) FirmwareUpload(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	if err := c.serviceClient.supportsFirmwareInstall(c.bmc.deviceModel()); err != nil {
		return "", err
	}

	// expect atleast 5 minutes left in the deadline to proceed with the upload
	d, set := ctx.Deadline()
	if set && time.Until(d) < 5*time.Minute {
		return "", errors.New("remaining context deadline insufficient to perform update: " + time.Until(d).String())
	}

	return c.bmc.firmwareUpload(ctx, component, file)
}

// FirmwareInstallUploaded initiates installation of a previously uploaded firmware image and returns the install task ID.
func (c *Client) FirmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error) {
	if err := c.serviceClient.supportsFirmwareInstall(c.bmc.deviceModel()); err != nil {
		return "", err
	}

	// x11's don't return a upload Task ID, since the upload mechanism is not redfish
	if !strings.HasPrefix(strings.ToLower(c.bmc.deviceModel()), "x11") && uploadTaskID == "" {
		return "", errors.Wrap(errUploadTaskIDExpected, "device model: "+c.bmc.deviceModel())
	}

	return c.bmc.firmwareInstallUploaded(ctx, component, uploadTaskID)
}

// FirmwareTaskStatus returns the status of a firmware related task queued on the BMC.
func (c *Client) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state constants.TaskState, status string, err error) {
	if err := c.serviceClient.supportsFirmwareInstall(c.bmc.deviceModel()); err != nil {
		return "", "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, err.Error())
	}

	component = strings.ToUpper(component)
	return c.bmc.firmwareTaskStatus(ctx, component, taskID)
}
