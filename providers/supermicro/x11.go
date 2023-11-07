package supermicro

import (
	"context"
	"fmt"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

type x11 struct{ *Client }

func (c *x11) deviceModel(ctx context.Context) (string, error) {
	errBoardPartNumUnknown := errors.New("baseboard part number unknown")
	data, err := c.fruInfo(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return "", ErrXMLAPIUnsupported
		}

		return "", err
	}

	partNum := strings.TrimSpace(data.Board.PartNum)

	if data.Board == nil || partNum == "" {
		return "", errors.Wrap(errBoardPartNumUnknown, "baseboard part number empty")
	}

	return partNum, nil
}

func (c *x11) componentSupported(component string) error {
	errComponentNotSupported := fmt.Errorf("component %s on device %s not supported", component, c.model)

	supported := []string{common.SlugBIOS, common.SlugBMC}
	if !slices.Contains(supported, strings.ToUpper(component)) {
		return errComponentNotSupported
	}

	return nil
}

func (c *x11) firmwareInstallSteps(component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.componentSupported(component); err != nil {
		return nil, err
	}

	return []constants.FirmwareInstallStep{
		constants.FirmwareInstallStepUpload,
		constants.FirmwareInstallStepInstallUploaded,
		constants.FirmwareInstallStepInstallStatus,
	}, nil
}
