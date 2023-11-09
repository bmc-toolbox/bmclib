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

	errUnexpectedModel      = errors.New("unexpected device model")
	errUploadTaskIDExpected = errors.New("expected an firmware upload taskID")
)

// bmc client interface implementations methods

func (c *Client) FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.firmwareInstallSupported(ctx); err != nil {
		return nil, err
	}

	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		return c.x12().firmwareInstallSteps(component)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		return c.x11().firmwareInstallSteps(component)
	}

	return nil, errors.Wrap(errUnexpectedModel, c.model)
}

func (c *Client) FirmwareUpload(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	if err := c.firmwareInstallSupported(ctx); err != nil {
		return "", err
	}

	//	// expect atleast 5 minutes left in the deadline to proceed with the upload
	d, _ := ctx.Deadline()
	if time.Until(d) < 5*time.Minute {
		return "", errors.New("remaining context deadline insufficient to perform update: " + time.Until(d).String())
	}

	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		return c.x12().firmwareUpload(ctx, component, file)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		return c.x11().firmwareUpload(ctx, component, file)
	}

	return "", errors.Wrap(errUnexpectedModel, c.model)
}

func (c *Client) FirmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error) {
	if err := c.firmwareInstallSupported(ctx); err != nil {
		return "", err
	}

	if uploadTaskID == "" {
		return "", errUploadTaskIDExpected
	}

	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		return c.x12().firmwareInstallUploaded(ctx, component, uploadTaskID)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		return "", c.x11().firmwareInstallUploaded(ctx, component)
	}

	return "", errors.Wrap(errUnexpectedModel, c.model)

}

// FirmwareTaskStatus returns the status of a firmware related task queued on the BMC.
func (c *Client) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state, status string, err error) {
	if err := c.firmwareInstallSupported(ctx); err != nil {
		return "", "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, err.Error())
	}

	component = strings.ToUpper(component)

	if strings.HasPrefix(strings.ToLower(c.model), "x12") {
		return c.x12().firmwareTaskStatus(ctx, component, taskID)
	} else if strings.HasPrefix(strings.ToLower(c.model), "x11") {
		return c.x11().firmwareTaskStatus(ctx, component)

	}

	return "", "", errors.Wrap(errUnexpectedModel, c.model)
}

func (c *Client) firmwareInstallSupported(ctx context.Context) error {
	errBoardUnsupported := errors.New("firmware install not supported/implemented for device model")

	for _, s := range supportedModels {
		if strings.EqualFold(s, c.model) {
			return nil
		}
	}

	return errBoardUnsupported
}
