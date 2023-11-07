package supermicro

import (
	"context"
	"io"
	"strings"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
)

func (c *x11) firmwareUpload(ctx context.Context, component string, reader io.Reader) (string, error) {
	component = strings.ToUpper(component)

	switch component {
	case common.SlugBIOS:
		return "", c.firmwareUploadBIOS(ctx, reader)
	case common.SlugBMC:
		return "", c.firmwareUploadBMC(ctx, reader)
	}

	return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)
}

func (c *x11) firmwareInstallUploaded(ctx context.Context, component string) error {
	component = strings.ToUpper(component)

	switch component {
	case common.SlugBIOS:
		return c.firmwareInstallUploadedBIOS(ctx)
	case common.SlugBMC:
		return c.initiateBMCFirmwareInstall(ctx)
	}

	return errors.Wrap(bmclibErrs.ErrFirmwareInstallUploaded, "component unsupported: "+component)
}

func (c *x11) firmwareTaskStatus(ctx context.Context, component string) (state, status string, err error) {
	component = strings.ToUpper(component)

	switch component {
	case common.SlugBIOS:
		return c.statusBIOSFirmwareInstall(ctx)
	case common.SlugBMC:
		return c.statusBMCFirmwareInstall(ctx)
	}

	return "", "", errors.Wrap(bmclibErrs.ErrFirmwareTaskStatus, "component unsupported: "+component)
}
