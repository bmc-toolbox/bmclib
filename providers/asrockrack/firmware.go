package asrockrack

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/constants"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/metal-toolbox/bmclib/internal"
)

const (
	versionStrError    = -1
	versionStrMatch    = 0
	versionStrMismatch = 1
	versionStrEmpty    = 2
)

// bmc client interface implementations methods
func (a *ASRockRack) FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error) {
	if err := a.supported(ctx); err != nil {
		return nil, bmclibErrs.NewErrUnsupportedHardware(err.Error())
	}

	switch strings.ToUpper(component) {
	case common.SlugBMC:
		return []constants.FirmwareInstallStep{
			constants.FirmwareInstallStepUpload,
			constants.FirmwareInstallStepInstallUploaded,
			constants.FirmwareInstallStepInstallStatus,
			constants.FirmwareInstallStepResetBMCPostInstall,
			constants.FirmwareInstallStepResetBMCOnInstallFailure,
		}, nil
	}

	return nil, errors.Wrap(bmclibErrs.ErrFirmwareUpload, "component unsupported: "+component)
}

func (a *ASRockRack) FirmwareUpload(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	switch strings.ToUpper(component) {
	case common.SlugBIOS:
		return "", a.firmwareUploadBIOS(ctx, file)
	case common.SlugBMC:
		return "", a.firmwareUploadBMC(ctx, file)
	}

	return "", errors.Wrap(bmclibErrs.ErrFirmwareUpload, "component unsupported: "+component)

}

func (a *ASRockRack) firmwareUploadBMC(ctx context.Context, file *os.File) error {
	//	// expect atleast 5 minutes left in the deadline to proceed with the upload
	d, _ := ctx.Deadline()
	if time.Until(d) < 5*time.Minute {
		return errors.New("remaining context deadline insufficient to perform update: " + time.Until(d).String())
	}

	// Beware: this locks some capabilities, e.g. the access to fruAttributes
	a.log.V(2).WithValues("step", "1/4").Info("set device to flash mode, takes a minute...")
	err := a.setFlashMode(ctx)
	if err != nil {
		return errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"failed in step 1/3 - set device to flash mode: "+err.Error(),
		)
	}

	var fwEndpoint string
	switch a.deviceModel {
	// E3C256D4ID-NL calls a different endpoint for firmware upload
	case "E3C256D4ID-NL":
		fwEndpoint = "api/maintenance/firmware/firmware"
	default:
		fwEndpoint = "api/maintenance/firmware"
	}

	a.log.V(2).WithValues("step", "2/4").Info("upload BMC firmware image to " + fwEndpoint)
	err = a.uploadFirmware(ctx, fwEndpoint, file)
	if err != nil {
		return errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"failed in step 2/3 - upload BMC firmware image: "+err.Error(),
		)
	}

	a.log.V(2).WithValues("step", "3/4").Info("verify uploaded BMC firmware")
	err = a.verifyUploadedFirmware(ctx)
	if err != nil {
		return errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"failed in step 3/3 - verify uploaded BMC firmware: "+err.Error(),
		)
	}

	return nil
}

func (a *ASRockRack) firmwareUploadBIOS(ctx context.Context, file *os.File) error {
	a.log.V(2).WithValues("step", "1/3").Info("upload BIOS firmware image")
	err := a.uploadFirmware(ctx, "api/asrr/maintenance/BIOS/firmware", file)
	if err != nil {
		return errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"failed in step 1/3 - upload BIOS firmware image: "+err.Error(),
		)
	}

	a.log.V(2).WithValues("step", "2/3").Info("set BIOS preserve flash configuration")
	err = a.biosUpgradeConfiguration(ctx)
	if err != nil {
		return errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"failed in step 2/3 - set flash configuration: "+err.Error(),
		)
	}

	// 3. run upgrade
	a.log.V(2).WithValues("step", "3/3").Info("proceed with BIOS firmware install")
	err = a.upgradeBIOS(ctx)
	if err != nil {
		return errors.Wrap(
			bmclibErrs.ErrFirmwareUpload,
			"failed in step 3/3 - proceed with BIOS firmware install: "+err.Error(),
		)
	}

	return nil
}

func (a *ASRockRack) FirmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error) {
	switch strings.ToUpper(component) {
	case common.SlugBIOS:
		return "", a.firmwareInstallUploadedBIOS(ctx)
	case common.SlugBMC:
		return "", a.firmwareInstallUploadedBMC(ctx)
	}

	return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)
}

// firmwareInstallUploadedBIOS uploads and installs firmware for the BMC component
func (a *ASRockRack) firmwareInstallUploadedBIOS(ctx context.Context) error {
	// 4. Run the upgrade - preserving current config
	a.log.V(2).WithValues("step", "install").Info("proceed with BIOS firmware install, preserve current configuration")
	err := a.upgradeBIOS(ctx)
	if err != nil {
		return errors.Wrap(
			bmclibErrs.ErrFirmwareInstallUploaded,
			"failed in step 4/4 - proceed with BMC firmware install: "+err.Error(),
		)
	}

	return nil
}

// firmwareInstallUploadedBMC uploads and installs firmware for the BMC component
func (a *ASRockRack) firmwareInstallUploadedBMC(ctx context.Context) error {
	// 4. Run the upgrade - preserving current config
	a.log.V(2).WithValues("step", "install").Info("proceed with BMC firmware install, preserve current configuration")
	err := a.upgradeBMC(ctx)
	if err != nil {
		return errors.Wrap(
			bmclibErrs.ErrFirmwareInstallUploaded,
			"failed in step 4/4 - proceed with BMC firmware install"+err.Error(),
		)
	}

	return nil
}

// FirmwareTaskStatus returns the status of a firmware related task queued on the BMC.
func (a *ASRockRack) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state constants.TaskState, status string, err error) {
	component = strings.ToUpper(component)
	switch component {
	case common.SlugBIOS, common.SlugBMC:
		return a.firmwareUpdateStatus(ctx, component, installVersion)
	default:
		return "", "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "component unsupported: "+component)
	}
}

// firmwareUpdateBIOSStatus returns the BIOS firmware install status
func (a *ASRockRack) firmwareUpdateStatus(ctx context.Context, component string, installVersion string) (state constants.TaskState, status string, err error) {
	var endpoint string
	component = strings.ToUpper(component)
	switch component {
	case common.SlugBIOS:
		endpoint = "api/asrr/maintenance/BIOS/flash-progress"
	case common.SlugBMC:
		endpoint = "api/maintenance/firmware/flash-progress"
	default:
		return "", "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "component unsupported: "+component)
	}

	// 1. query the flash progress endpoint
	//
	// once an update completes/fails this endpoint will return 500
	progress, err := a.flashProgress(ctx, endpoint)
	if err != nil {
		a.log.V(3).Error(err, "bmc query for install progress returned error: ")
	}

	if progress != nil {
		status = fmt.Sprintf("action: %s, progress: %s", progress.Action, progress.Progress)

		switch progress.State {
		case 0:
			return constants.Running, status, nil
		case 1: // "Flashing To be done"
			return constants.Queued, status, nil
		case 2:
			return constants.Complete, status, nil
		default:
			a.log.V(3).WithValues("state", progress.State).Info("warn", "bmc returned unknown flash progress state")
		}
	}

	// 2. query the firmware info endpoint to determine the update status
	//
	// at this point the flash-progress endpoint isn't returning useful information
	var installStatus int

	installStatus, err = a.versionInstalled(ctx, component, installVersion)
	if err != nil {
		return "", "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, err.Error())
	}

	switch installStatus {
	case versionStrMatch:
		if progress == nil {
			// TODO: we should pass the force parameter to firmwareUpdateStatus,
			// so that we can know if we expect a version change or not
			a.log.V(3).Info("Nil progress + no version change -> unknown")
			return constants.Unknown, status, nil
		}

		return constants.Complete, status, nil
	case versionStrEmpty:
		return constants.Unknown, status, nil
	case versionStrMismatch:
		return constants.Running, status, nil
	}

	return constants.Unknown, status, nil
}

// versionInstalled returns int values on the status of the firmware version install
//
// - 0 indicates the given version parameter matches the version installed
// - 1 indicates the given version parameter does not match the version installed
// - 2 the version parameter returned from the BMC is empty (which means the BMC needs a reset)
func (a *ASRockRack) versionInstalled(ctx context.Context, component, version string) (status int, err error) {
	component = strings.ToUpper(component)
	if !internal.StringInSlice(component, []string{common.SlugBIOS, common.SlugBMC}) {
		return versionStrError, errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)
	}

	fwInfo, err := a.firmwareInfo(ctx)
	if err != nil {
		err = errors.Wrap(err, "error querying for firmware info: ")
		a.log.V(3).Info("warn", err.Error())
		return versionStrError, err
	}

	var installed string

	switch component {
	case common.SlugBIOS:
		installed = fwInfo.BIOSVersion
	case common.SlugBMC:
		installed = fwInfo.BMCVersion
	}

	// version match
	if strings.EqualFold(installed, version) {
		return versionStrMatch, nil
	}

	// fwinfo returned an empty string for firmware revision
	// this indicates the BMC is out of sync with the firmware versions installed
	if strings.TrimSpace(installed) == "" {
		return versionStrEmpty, nil
	}

	return 1, nil
}
