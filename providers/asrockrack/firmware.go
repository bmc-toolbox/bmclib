package asrockrack

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/bmc-toolbox/bmclib/devices"
	bmclibErrs "github.com/bmc-toolbox/bmclib/errors"
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
			a.log.V(2).Info("warn", "unable to determine file size: "+err.Error())
		}

		size = finfo.Size()
	}

	switch component {
	case devices.SlugBIOS:
		err = a.firmwareInstallBIOS(ctx, reader, size)
	case devices.SlugBMC:
		err = a.firmwareInstallBMC(ctx, reader, size)
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)

	}

	return jobID, err
}

// FirmwareInstallStatus returns the status of the firmware install process, a bool value indicating if the component requires a reset
func (a *ASRockRack) FirmwareInstallStatus(ctx context.Context, component, installVersion, taskID string) (status string, err error) {
	switch component {
	case devices.SlugBIOS, devices.SlugBMC:
		return a.firmwareUpdateStatus(ctx, component, installVersion)
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "component unsupported: "+component)
	}
}

// firmwareInstallBMC uploads and installs firmware for the BMC component
func (a *ASRockRack) firmwareInstallBMC(ctx context.Context, reader io.Reader, fileSize int64) error {
	var err error

	// 1. set the device to flash mode - prepares the flash
	a.log.V(2).Info("info", "action", "set device to flash mode, takes a minute...", "step", "1/4")
	err = a.setFlashMode(ctx)
	if err != nil {
		return fmt.Errorf("failed in step 1/4 - set device to flash mode: " + err.Error())
	}

	// 2. upload firmware image file
	a.log.V(2).Info("info", "action", "upload BMC firmware image", "step", "2/4")
	err = a.uploadFirmware(ctx, "api/maintenance/firmware", reader, fileSize)
	if err != nil {
		return fmt.Errorf("failed in step 2/4 - upload BMC firmware image: " + err.Error())
	}

	// 3. BMC to verify the uploaded file
	err = a.verifyUploadedFirmware(ctx)
	a.log.V(2).Info("info", "action", "BMC verify uploaded firmware", "step", "3/4")
	if err != nil {
		return fmt.Errorf("failed in step 3/4 - BMC verify uploaded firmware: " + err.Error())
	}

	// 4. Run the upgrade - preserving current config
	a.log.V(2).Info("info", "action", "proceed with upgrade, preserve current configuration", "step", "4/4")
	err = a.upgradeBMC(ctx)
	if err != nil {
		return fmt.Errorf("failed in step 4/4 - proceed with upgrade: " + err.Error())
	}

	return nil
}

// firmwareInstallBIOS uploads and installs firmware for the BIOS component
func (a *ASRockRack) firmwareInstallBIOS(ctx context.Context, reader io.Reader, fileSize int64) error {
	var err error

	// 1. upload firmware image file
	a.log.V(2).Info("info", "action", "upload BIOS firmware image", "step", "1/3")
	err = a.uploadFirmware(ctx, "api/asrr/maintenance/BIOS/firmware", reader, fileSize)
	if err != nil {
		return fmt.Errorf("failed in step 1/3 - upload firmware image: " + err.Error())
	}

	// 2. set update parameters to preserve configurations
	a.log.V(2).Info("info", "action", "set flash configuration", "step", "2/3")
	err = a.biosUpgradeConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("failed in step 2/3 - set flash configuration: " + err.Error())
	}

	// 3. run upgrade
	a.log.V(2).Info("info", "action", "proceed with upgrade", "step", "3/3")
	err = a.upgradeBIOS(ctx)
	if err != nil {
		return fmt.Errorf("failed in step 3/3 - proceed with upgrade: " + err.Error())
	}

	return nil
}

// firmwareUpdateBIOSStatus returns the BIOS firmware install status
func (a *ASRockRack) firmwareUpdateStatus(ctx context.Context, component string, installVersion string) (status string, err error) {
	// TODO: purge debug logging
	os.Setenv("BMCLIB_LOG_LEVEL", "trace")
	defer os.Unsetenv("BMCLIB_LOG_LEVEL")

	var endpoint string
	switch component {
	case devices.SlugBIOS:
		endpoint = "api/asrr/maintenance/BIOS/flash-progress"
	case devices.SlugBMC:
		endpoint = "api/maintenance/firmware/flash-progress"
	default:
		return "", errors.Wrap(bmclibErrs.ErrFirmwareInstallStatus, "component unsupported: "+component)
	}

	// 1. query the flash progress endpoint
	//
	// once an update completes/fails this endpoint will return 500
	progress, err := a.flashProgress(ctx, endpoint)
	if err != nil {
		a.log.V(3).Info("warn", "bmc query for install progress returned error: "+err.Error())
	}

	if progress != nil {
		switch progress.State {
		case 0:
			return devices.FirmwareInstallRunning, nil
		case 2:
			return devices.FirmwareInstallComplete, nil
		default:
			a.log.V(3).Info("warn", "bmc returned unknown flash progress state: "+strconv.Itoa(progress.State))
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
		return devices.FirmwareInstallComplete, nil
	case versionStrEmpty:
		return devices.FirmwareInstallUnknown, nil
	case versionStrMismatch:
		return devices.FirmwareInstallRunning, nil
	}

	return devices.FirmwareInstallUnknown, nil
}

// versionInstalled returns int values on the status of the firmware version install
//
// - 0 indicates the given version parameter matches the version installed
// - 1 indicates the given version parameter does not match the version installed
// - 2 the version parameter returned from the BMC is empty (which means the BMC needs a reset)
func (a *ASRockRack) versionInstalled(ctx context.Context, component, version string) (status int, err error) {
	fwInfo, err := a.firmwareInfo(ctx)
	if err != nil {
		err = errors.Wrap(err, "error querying for firmware info: ")
		a.log.V(3).Info("warn", err.Error())
		return versionStrError, err
	}

	var installed string

	switch component {
	case devices.SlugBIOS:
		installed = fwInfo.BIOSVersion
	case devices.SlugBMC:
		installed = fwInfo.BMCVersion
	default:
		return versionStrError, errors.Wrap(bmclibErrs.ErrFirmwareInstall, "component unsupported: "+component)
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
