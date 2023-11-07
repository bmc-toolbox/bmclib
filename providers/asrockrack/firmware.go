package asrockrack

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal"
	"github.com/bmc-toolbox/common"
)

const (
	versionStrError    = -1
	versionStrMatch    = 0
	versionStrMismatch = 1
	versionStrEmpty    = 2
)

// FirmwareInstall uploads and initiates firmware update for the component
func (a *ASRockRack) FirmwareInstall(ctx context.Context, component, applyAt string, forceInstall bool, reader io.Reader) (jobID string, err error) {
	var size int64
	if file, ok := reader.(*os.File); ok {
		finfo, err := file.Stat()
		if err != nil {
			a.log.V(2).Error(err, "unable to determine file size")
		}

		size = finfo.Size()
	}

	component = strings.ToUpper(component)
	switch component {
	case common.SlugBIOS:
		err = a.firmwareInstallBIOS(ctx, reader, size)
	case common.SlugBMC:
		err = a.firmwareInstallBMC(ctx, reader, size)
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)
	}

	if err != nil {
		err = errors.Wrap(bmclibErrs.ErrFirmwareInstall, err.Error())
	}

	return jobID, err
}

// FirmwareInstallStatus returns the status of the firmware install process, a bool value indicating if the component requires a reset
func (a *ASRockRack) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (status string, err error) {
	component = strings.ToUpper(component)
	switch component {
	case common.SlugBIOS, common.SlugBMC:
		return a.firmwareUpdateStatus(ctx, component, installVersion)
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "component unsupported: "+component)
	}
}

// firmwareInstallBMC uploads and installs firmware for the BMC component
func (a *ASRockRack) firmwareInstallBMC(ctx context.Context, reader io.Reader, fileSize int64) error {
	var err error

	// 0. take the model so that we use a different endpoint on E3C256D4ID-NL
	device := common.NewDevice()
	device.Metadata = map[string]string{}
	err = a.fruAttributes(ctx, &device)
	if err != nil {
		return errors.Wrap(err, "failed to get model in step 0/4")
	}

	// 1. set the device to flash mode - prepares the flash
	// Beware: this locks some capabilities, e.g. the access to fruAttributes
	a.log.V(2).WithValues("step", "1/4").Info("set device to flash mode, takes a minute...")
	err = a.setFlashMode(ctx)
	if err != nil {
		return errors.Wrap(err, "failed in step 1/4 - set device to flash mode")
	}

	// 2. upload firmware image file
	fwEndpoint := "api/maintenance/firmware"
	// E3C256D4ID-NL calls a different endpoint for firmware upload
	if strings.EqualFold(device.Model, "E3C256D4ID-NL") {
		fwEndpoint = "api/maintenance/firmware/firmware"
	}
	a.log.V(2).WithValues("step", "2/4").Info("upload BMC firmware image to " + fwEndpoint)
	err = a.uploadFirmware(ctx, fwEndpoint, reader, fileSize)
	if err != nil {
		return errors.Wrap(err, "failed in step 2/4 - upload BMC firmware image")
	}

	// 3. BMC to verify the uploaded file
	err = a.verifyUploadedFirmware(ctx)
	a.log.V(2).WithValues("step", "3/4").Info("verify uploaded BMC firmware")
	if err != nil {
		return errors.Wrap(err, "failed in step 3/4 - verify uploaded BMC firmware")
	}

	// 4. Run the upgrade - preserving current config
	a.log.V(2).WithValues("step", "4/4").Info("proceed with BMC firmware install, preserve current configuration")
	err = a.upgradeBMC(ctx)
	if err != nil {
		return errors.Wrap(err, "failed in step 4/4 - proceed with BMC firmware install")
	}

	return nil
}

// firmwareInstallBIOS uploads and installs firmware for the BIOS component
func (a *ASRockRack) firmwareInstallBIOS(ctx context.Context, reader io.Reader, fileSize int64) error {
	var err error

	// 1. upload firmware image file
	a.log.V(2).WithValues("step", "1/3").Info("upload BIOS firmware image")
	err = a.uploadFirmware(ctx, "api/asrr/maintenance/BIOS/firmware", reader, fileSize)
	if err != nil {
		return errors.Wrap(err, "failed in step 1/3 - upload BIOS firmware image")
	}

	// 2. set update parameters to preserve configurations
	a.log.V(2).WithValues("step", "2/3").Info("set BIOS preserve flash configuration")
	err = a.biosUpgradeConfiguration(ctx)
	if err != nil {
		return errors.Wrap(err, "failed in step 2/3 - set flash configuration")
	}

	// 3. run upgrade
	a.log.V(2).WithValues("step", "3/3").Info("proceed with BIOS firmware install")
	err = a.upgradeBIOS(ctx)
	if err != nil {
		return errors.Wrap(err, "failed in step 3/3 - proceed with BIOS firmware install")
	}

	return nil
}

// firmwareUpdateBIOSStatus returns the BIOS firmware install status
func (a *ASRockRack) firmwareUpdateStatus(ctx context.Context, component string, installVersion string) (status string, err error) {
	var endpoint string
	component = strings.ToUpper(component)
	switch component {
	case common.SlugBIOS:
		endpoint = "api/asrr/maintenance/BIOS/flash-progress"
	case common.SlugBMC:
		endpoint = "api/maintenance/firmware/flash-progress"
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "component unsupported: "+component)
	}

	// 1. query the flash progress endpoint
	//
	// once an update completes/fails this endpoint will return 500
	progress, err := a.flashProgress(ctx, endpoint)
	if err != nil {
		a.log.V(3).Error(err, "bmc query for install progress returned error: ")
	}

	if progress != nil {
		switch progress.State {
		case 0:
			return constants.FirmwareInstallRunning, nil
		case 1: // "Flashing To be done"
			return constants.FirmwareInstallQueued, nil
		case 2:
			return constants.FirmwareInstallComplete, nil
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
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, err.Error())
	}

	switch installStatus {
	case versionStrMatch:
		return constants.FirmwareInstallComplete, nil
	case versionStrEmpty:
		return constants.FirmwareInstallUnknown, nil
	case versionStrMismatch:
		return constants.FirmwareInstallRunning, nil
	}

	return constants.FirmwareInstallUnknown, nil
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
