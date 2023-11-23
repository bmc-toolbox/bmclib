package supermicro

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/common"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
)

type x11 struct {
	*serviceClient
	model string
	log   logr.Logger
}

func newX11Client(client *serviceClient, logger logr.Logger) bmcQueryor {
	return &x11{
		serviceClient: client,
		log:           logger,
	}
}

func (c *x11) deviceModel() string {
	return c.model
}

func (c *x11) queryDeviceModel(ctx context.Context) (string, error) {
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

	c.model = common.FormatProductName(partNum)

	return c.model, nil
}

func (c *x11) fruInfo(ctx context.Context) (*FruInfo, error) {
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}

	payload := "op=FRU_INFO.XML&r=(0,0)&_="

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBufferString(payload), headers, 0)
	if err != nil {
		return nil, errors.Wrap(ErrQueryFRUInfo, err.Error())
	}

	if status != 200 {
		return nil, unexpectedResponseErr([]byte(payload), body, status)
	}

	if !bytes.Contains(body, []byte(`<IPMI>`)) {
		return nil, unexpectedResponseErr([]byte(payload), body, status)
	}

	data := &IPMI{}
	if err := xml.Unmarshal(body, data); err != nil {
		return nil, errors.Wrap(ErrQueryFRUInfo, err.Error())
	}

	return data.FruInfo, nil
}

func (c *x11) supportsInstall(component string) error {
	errComponentNotSupported := fmt.Errorf("component %s on device %s not supported", component, c.model)

	supported := []string{common.SlugBIOS, common.SlugBMC}
	if !slices.Contains(supported, strings.ToUpper(component)) {
		return errComponentNotSupported
	}

	return nil
}

func (c *x11) firmwareInstallSteps(component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.supportsInstall(component); err != nil {
		return nil, err
	}

	return []constants.FirmwareInstallStep{
		constants.FirmwareInstallStepUpload,
		constants.FirmwareInstallStepInstallUploaded,
		constants.FirmwareInstallStepInstallStatus,
	}, nil
}

func (c *x11) firmwareUpload(ctx context.Context, component string, file *os.File) (string, error) {
	component = strings.ToUpper(component)

	switch component {
	case common.SlugBIOS:
		return "", c.firmwareUploadBIOS(ctx, file)
	case common.SlugBMC:
		return "", c.firmwareUploadBMC(ctx, file)
	}

	return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)
}

func (c *x11) firmwareInstallUploaded(ctx context.Context, component, _ string) (string, error) {
	component = strings.ToUpper(component)

	switch component {
	case common.SlugBIOS:
		return "", c.firmwareInstallUploadedBIOS(ctx)
	case common.SlugBMC:
		return "", c.initiateBMCFirmwareInstall(ctx)
	}

	return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallUploaded, "component unsupported: "+component)
}

func (c *x11) firmwareTaskStatus(ctx context.Context, component, _ string) (state constants.TaskState, status string, err error) {
	component = strings.ToUpper(component)

	switch component {
	case common.SlugBIOS:
		return c.statusBIOSFirmwareInstall(ctx)
	case common.SlugBMC:
		return c.statusBMCFirmwareInstall(ctx)
	}

	return "", "", errors.Wrap(bmclibErrs.ErrFirmwareTaskStatus, "component unsupported: "+component)
}

func (c *x11) inventory(ctx context.Context) (*common.Device, error) {
	if err := c.redfishSession(ctx); err != nil {
		return nil, err
	}

	return c.serviceClient.redfish.Inventory(ctx, false)
}

// power-off - immediate - op=POWER_INFO.XML&r=(1,0)&_=
// power-on - op=POWER_INFO.XML&r=(1,1)&_=
// power-off - acpi/orderly - op=POWER_INFO.XML&r=(1,5)&_=
// reset server - cold powercycle - op=POWER_INFO.XML&r=(1,3)&_=
// power cycle - op=POWER_INFO.XML&r=(1,2)&_=

// powerCycle using SMC XML API
//
// This method is only here for the case when firmware updates are being applied using this provider.
func (c *x11) powerCycle(ctx context.Context) (bool, error) {
	payload := []byte(`op=POWER_INFO.XML&r=(1,3)&_=`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.serviceClient.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return false, err
	}

	if status != http.StatusOK {
		return false, unexpectedResponseErr(payload, body, status)
	}

	return true, nil
}
