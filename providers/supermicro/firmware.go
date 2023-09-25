package supermicro

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
)

// FirmwareInstall uploads and initiates firmware update for the component
func (c *Client) FirmwareInstall(ctx context.Context, component, applyAt string, forceInstall bool, reader io.Reader) (jobID string, err error) {
	if err := c.deviceSupported(ctx); err != nil {
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
	}

	var size int64
	if file, ok := reader.(*os.File); ok {
		finfo, err := file.Stat()
		if err != nil {
			c.log.V(2).Error(err, "unable to determine file size")
		}

		size = finfo.Size()
	}

	// expect atleast 10 minutes left in the deadline to proceed with the update
	d, _ := ctx.Deadline()
	if time.Until(d) < 10*time.Minute {
		return "", errors.New("remaining context deadline insufficient to perform update: " + time.Until(d).String())
	}

	component = strings.ToUpper(component)

	switch component {
	case common.SlugBIOS:
		err = c.firmwareInstallBIOS(ctx, reader, size)
	case common.SlugBMC:
		err = c.firmwareInstallBMC(ctx, reader, size)
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)
	}

	if err != nil {
		err = errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
	}

	return jobID, err
}

// FirmwareInstallStatus returns the status of the firmware install process
func (c *Client) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (string, error) {
	component = strings.ToUpper(component)

	switch component {
	case common.SlugBMC:
		return c.statusBMCFirmwareInstall(ctx)
	case common.SlugBIOS:
		return c.statusBIOSFirmwareInstall(ctx)
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "component unsupported: "+component)
	}
}

func (c *Client) deviceSupported(ctx context.Context) error {
	errBoardPartNumUnknown := errors.New("baseboard part number unknown")
	errBoardUnsupported := errors.New("feature not supported/implemented for device")

	// Its likely this works on all X11's
	// for now, we list only the ones its been tested on.
	//
	// board part numbers
	//
	supported := []string{
		"X11SCM-F",
		"X11DPH-T",
		"X11SCH-F",
		"X11DGQ",
		"X11DPG-SN",
		"X11DPT-B",
		"X11SSE-F",
	}

	data, err := c.fruInfo(ctx)
	if err != nil {
		return err
	}

	if data.Board == nil || strings.TrimSpace(data.Board.PartNum) == "" {
		return errors.Wrap(errBoardPartNumUnknown, "baseboard part number empty")
	}

	c.model = strings.TrimSpace(data.Board.PartNum)

	for _, b := range supported {
		if strings.EqualFold(b, strings.TrimSpace(data.Board.PartNum)) {
			return nil
		}
	}

	return errors.Wrap(errBoardUnsupported, data.Board.PartNum)
}
