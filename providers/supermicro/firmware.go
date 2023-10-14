package supermicro

import (
	"context"
	"io"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/common"
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

func (c *Client) FirmwareInstallActions(ctx context.Context, component string) ([]constants.FirmwareAction, error) {
	if err := c.firmwareInstallSupported(ctx); err != nil {
		return nil, err
	}

	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		return c.x12().firmwareInstallActions(component)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		//return c.x11().statusBMCFirmwareInstall(ctx)
	}

	return nil, errors.Wrap(errUnexpectedModel, c.model)
}

func (c *Client) FirmwareUpload(ctx context.Context, component string, operationApplyTime constants.OperationApplyTime, reader io.Reader) (taskID string, err error) {
	if err := c.firmwareInstallSupported(ctx); err != nil {
		return "", err
	}

	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		return c.x12().firmwareUpload(ctx, component, operationApplyTime, reader)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		//return c.x11().statusBMCFirmwareInstall(ctx)
	}

	return "", errors.Wrap(errUnexpectedModel, c.model)
}

func (c *Client) FirmwareInstallWithOptions(ctx context.Context, component string, reader io.Reader, opts *bmc.FirmwareInstallOptions) (installTaskID string, err error) {
	if err := c.firmwareInstallSupported(ctx); err != nil {
		return "", err
	}

	if opts.UploadTaskID == "" {
		return "", errUploadTaskIDExpected
	}

	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		return c.x12().firmwareInstall(ctx, component, opts.UploadTaskID)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		//return c.x11().statusBMCFirmwareInstall(ctx)
	}

	return "", errors.Wrap(errUnexpectedModel, c.model)

}

// FirmwareInstallStatus returns the status of the firmware install process
func (c *Client) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (status string, err error) {
	if err := c.firmwareInstallSupported(ctx); err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
	}

	component = strings.ToUpper(component)

	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		return c.x12().firmwareInstallStatus(ctx, component, installVersion, taskID)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		//return c.x11().statusBMCFirmwareInstall(ctx)
	}

	switch component {
	case common.SlugBMC:
		return c.statusFirmwareInstallBMC(ctx, installVersion, component, taskID)
	case common.SlugBIOS:
		return c.statusFirmwareInstallBIOS(ctx, installVersion, component, taskID)
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "component unsupported: "+component)
	}
}

// FirmwareInstall uploads and initiates firmware update for the component
//func (c *Client) FirmwareInstall(ctx context.Context, component string, applyAt constants.OperationApplyTime, forceInstall bool, reader io.Reader) (jobID string, err error) {
//	if err := c.deviceSupported(ctx); err != nil {
//		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
//	}
//
//	var size int64
//	if file, ok := reader.(*os.File); ok {
//		finfo, err := file.Stat()
//		if err != nil {
//			c.log.V(2).Error(err, "unable to determine file size")
//		}
//
//		size = finfo.Size()
//	}
//
//	// expect atleast 10 minutes left in the deadline to proceed with the update
//	d, _ := ctx.Deadline()
//	if time.Until(d) < 10*time.Minute {
//		return "", errors.New("remaining context deadline insufficient to perform update: " + time.Until(d).String())
//	}
//
//	component = strings.ToUpper(component)
//
//	switch component {
//	case common.SlugBIOS:
//		err = c.installFirmwareBIOS(ctx, reader, size)
//	case common.SlugBMC:
//		//return c.installFirmwareBMC(ctx, reader, size)
//	default:
//		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)
//	}
//
//	if err != nil {
//		err = errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
//	}
//
//	return jobID, err
//}

// installFirmwareBMC invoke the device model specific BIOS firmware install method
func (c *Client) installFirmwareBIOS(ctx context.Context, reader io.Reader, fileSize int64) error {
	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		//return c.x12().firmwareInstallBIOS(ctx, reader, fileSize)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		return c.x11().firmwareInstallBIOS(ctx, reader, fileSize)
	}

	return errors.Wrap(errUnexpectedModel, c.model)
}

func (c *Client) statusFirmwareInstallBMC(ctx context.Context, installVersion, component, taskID string) (string, error) {
	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
	//	return c.x12().statusFirmwareInstallBMC(ctx)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		return c.x11().statusBMCFirmwareInstall(ctx)
	}

	return "", errors.Wrap(errUnexpectedModel, c.model)
}

func (c *Client) statusFirmwareInstallBIOS(ctx context.Context, installVersion, component, taskID string) (string, error) {
	switch {
	case strings.HasPrefix(strings.ToLower(c.model), "x12"):
		//return c.x12().firmwareInstallBIOS(ctx, reader, fileSize)
	case strings.HasPrefix(strings.ToLower(c.model), "x11"):
		return c.x11().statusBIOSFirmwareInstall(ctx)
	}

	return "", errors.Wrap(errUnexpectedModel, c.model)
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
