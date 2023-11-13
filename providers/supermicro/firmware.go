package supermicro

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
)

var (
	// Its likely the X11 code works on all X11's
	// for now, we list only the ones its been tested on.
	//
	// board part numbers
	//
	supportedModels = []string{
		"X11SCM-F",
		"X11DPH-T",
		"X11SCH-F",
		"X11DGQ",
		"X11DPG-SN",
		"X11DPT-B",
		"X11SSE-F",
		"X12STH-SYS",
	}

	errUploadTaskIDExpected = errors.New("expected an firmware upload taskID")
)

// bmc client interface implementations methods
func (c *Client) FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.serviceClient.supportsFirmwareInstall(ctx, c.bmc.deviceModel()); err != nil {
		return nil, err
	}

	return c.bmc.firmwareInstallSteps(component)
}

func (c *Client) FirmwareUpload(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	if err := c.serviceClient.supportsFirmwareInstall(ctx, c.bmc.deviceModel()); err != nil {
		return "", err
	}

	//	// expect atleast 5 minutes left in the deadline to proceed with the upload
	d, _ := ctx.Deadline()
	if time.Until(d) < 5*time.Minute {
		return "", errors.New("remaining context deadline insufficient to perform update: " + time.Until(d).String())
	}

	return c.bmc.firmwareUpload(ctx, component, file)
}

func (c *Client) FirmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error) {
	if err := c.serviceClient.supportsFirmwareInstall(ctx, c.bmc.deviceModel()); err != nil {
		return "", err
	}

	// x11's don't return a upload Task ID, since the upload mechanism is not redfish
	if !strings.HasPrefix(c.bmc.deviceModel(), "x11") && uploadTaskID == "" {
		return "", errUploadTaskIDExpected
	}

	return c.bmc.firmwareInstallUploaded(ctx, component, uploadTaskID)
}

// FirmwareTaskStatus returns the status of a firmware related task queued on the BMC.
func (c *Client) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state, status string, err error) {
	if err := c.serviceClient.supportsFirmwareInstall(ctx, c.bmc.deviceModel()); err != nil {
		return "", "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, err.Error())
	}

	component = strings.ToUpper(component)
	return c.bmc.firmwareTaskStatus(ctx, component, taskID)
}
