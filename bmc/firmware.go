package bmc

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bconsts "github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// FirmwareInstaller defines an interface to upload and initiate a firmware install
type FirmwareInstaller interface {
	// FirmwareInstall uploads firmware update payload to the BMC returning the task ID
	//
	// parameters:
	// component - the component slug for the component update being installed.
	// operationsApplyTime - one of the OperationApplyTime constants
	// forceInstall - purge the install task queued/scheduled firmware install BMC task (if any).
	// reader - the io.reader to the firmware update file.
	//
	// return values:
	// taskID - A taskID is returned if the update process on the BMC returns an identifier for the update process.
	FirmwareInstall(ctx context.Context, component string, operationApplyTime string, forceInstall bool, reader io.Reader) (taskID string, err error)
}

// firmwareInstallerProvider is an internal struct to correlate an implementation/provider and its name
type firmwareInstallerProvider struct {
	name string
	FirmwareInstaller
}

// firmwareInstall uploads and initiates firmware update for the component
func firmwareInstall(ctx context.Context, component, operationApplyTime string, forceInstall bool, reader io.Reader, generic []firmwareInstallerProvider) (taskID string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.FirmwareInstaller == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return taskID, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			taskID, vErr := elem.FirmwareInstall(ctx, component, operationApplyTime, forceInstall, reader)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return taskID, metadataLocal, nil
		}
	}

	return taskID, metadataLocal, multierror.Append(err, errors.New("failure in FirmwareInstall"))
}

// FirmwareInstallFromInterfaces identifies implementations of the FirmwareInstaller interface and passes the found implementations to the firmwareInstall() wrapper
func FirmwareInstallFromInterfaces(ctx context.Context, component, operationApplyTime string, forceInstall bool, reader io.Reader, generic []interface{}) (taskID string, metadata Metadata, err error) {
	implementations := make([]firmwareInstallerProvider, 0)
	for _, elem := range generic {
		temp := firmwareInstallerProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FirmwareInstaller:
			temp.FirmwareInstaller = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a FirmwareInstaller implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return taskID, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no FirmwareInstaller implementations found"),
			),
		)
	}

	return firmwareInstall(ctx, component, operationApplyTime, forceInstall, reader, implementations)
}

// Note: this interface is to be deprecated in favour of a more generic FirmwareTaskVerifier.
//
// FirmwareInstallVerifier defines an interface to check firmware install status
type FirmwareInstallVerifier interface {
	// FirmwareInstallStatus returns the status of the firmware install process.
	//
	// parameters:
	// installVersion (required) - the version this method should check is installed.
	// component (optional) - the component slug for the component update being installed.
	// taskID (optional) - the task identifier.
	//
	// return values:
	// status - returns one of the FirmwareInstall statuses (see devices/constants.go).
	FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (status string, err error)
}

// firmwareInstallVerifierProvider is an internal struct to correlate an implementation/provider and its name
type firmwareInstallVerifierProvider struct {
	name string
	FirmwareInstallVerifier
}

// firmwareInstallStatus returns the status of the firmware install process
func firmwareInstallStatus(ctx context.Context, installVersion, component, taskID string, generic []firmwareInstallVerifierProvider) (status string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.FirmwareInstallVerifier == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return status, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			status, vErr := elem.FirmwareInstallStatus(ctx, installVersion, component, taskID)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return status, metadataLocal, nil
		}
	}

	return status, metadataLocal, multierror.Append(err, errors.New("failure in FirmwareInstallStatus"))
}

// FirmwareInstallStatusFromInterfaces identifies implementations of the FirmwareInstallVerifier interface and passes the found implementations to the firmwareInstallStatus() wrapper.
func FirmwareInstallStatusFromInterfaces(ctx context.Context, installVersion, component, taskID string, generic []interface{}) (status string, metadata Metadata, err error) {
	implementations := make([]firmwareInstallVerifierProvider, 0)
	for _, elem := range generic {
		temp := firmwareInstallVerifierProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FirmwareInstallVerifier:
			temp.FirmwareInstallVerifier = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a FirmwareInstallVerifier implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return taskID, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no FirmwareInstallVerifier implementations found"),
			),
		)
	}

	return firmwareInstallStatus(ctx, installVersion, component, taskID, implementations)
}

// FirmwareInstallerWithOpts defines an interface to install firmware that was previously uploaded with FirmwareUpload
type FirmwareInstallerUploaded interface {
	// FirmwareInstallUploaded uploads firmware update payload to the BMC returning the firmware install task ID
	//
	// parameters:
	// component - the component slug for the component update being installed.
	// uploadTaskID - the taskID for the firmware upload verify task (returned by FirmwareUpload)
	//
	// return values:
	// installTaskID - A installTaskID is returned if the update process on the BMC returns an identifier for the firmware install process.
	FirmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (taskID string, err error)
}

// firmwareInstallerProvider is an internal struct to correlate an implementation/provider and its name
type firmwareInstallerWithOptionsProvider struct {
	name string
	FirmwareInstallerUploaded
}

// firmwareInstallUploaded uploads and initiates firmware update for the component
func firmwareInstallUploaded(ctx context.Context, component, uploadTaskID string, generic []firmwareInstallerWithOptionsProvider) (installTaskID string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.FirmwareInstallerUploaded == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return installTaskID, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			var vErr error
			installTaskID, vErr = elem.FirmwareInstallUploaded(ctx, component, uploadTaskID)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return installTaskID, metadataLocal, nil
		}
	}

	return installTaskID, metadataLocal, multierror.Append(err, errors.New("failure in FirmwareInstallUploaded"))
}

// FirmwareInstallerUploadedFromInterfaces identifies implementations of the FirmwareInstallUploaded interface and passes the found implementations to the firmwareInstallUploaded() wrapper
func FirmwareInstallerUploadedFromInterfaces(ctx context.Context, component, uploadTaskID string, generic []interface{}) (installTaskID string, metadata Metadata, err error) {
	implementations := make([]firmwareInstallerWithOptionsProvider, 0)
	for _, elem := range generic {
		temp := firmwareInstallerWithOptionsProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FirmwareInstallerUploaded:
			temp.FirmwareInstallerUploaded = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a FirmwareInstallerUploaded implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return installTaskID, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no FirmwareInstallerUploaded implementations found"),
			),
		)
	}

	return firmwareInstallUploaded(ctx, component, uploadTaskID, implementations)
}

type FirmwareInstallStepsGetter interface {
	FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error)
}

// firmwareInstallStepsGetterProvider is an internal struct to correlate an implementation/provider and its name
type firmwareInstallStepsGetterProvider struct {
	name string
	FirmwareInstallStepsGetter
}

// FirmwareInstallStepsFromInterfaces identifies implementations of the FirmwareInstallStepsGetter interface and passes the found implementations to the firmwareInstallSteps() wrapper.
func FirmwareInstallStepsFromInterfaces(ctx context.Context, component string, generic []interface{}) (steps []constants.FirmwareInstallStep, metadata Metadata, err error) {
	implementations := make([]firmwareInstallStepsGetterProvider, 0)
	for _, elem := range generic {
		temp := firmwareInstallStepsGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FirmwareInstallStepsGetter:
			temp.FirmwareInstallStepsGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a FirmwareInstallStepsGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return steps, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no FirmwareInstallStepsGetter implementations found"),
			),
		)
	}

	return firmwareInstallSteps(ctx, component, implementations)
}

func firmwareInstallSteps(ctx context.Context, component string, generic []firmwareInstallStepsGetterProvider) (steps []constants.FirmwareInstallStep, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.FirmwareInstallStepsGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return steps, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			steps, vErr := elem.FirmwareInstallSteps(ctx, component)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return steps, metadataLocal, nil
		}
	}

	return steps, metadataLocal, multierror.Append(err, errors.New("failure in FirmwareInstallSteps"))
}

type FirmwareUploader interface {
	FirmwareUpload(ctx context.Context, component string, file *os.File) (uploadVerifyTaskID string, err error)
}

// firmwareUploaderProvider is an internal struct to correlate an implementation/provider and its name
type firmwareUploaderProvider struct {
	name string
	FirmwareUploader
}

// FirmwareUploaderFromInterfaces identifies implementations of the FirmwareUploader interface and passes the found implementations to the firmwareUpload() wrapper.
func FirmwareUploadFromInterfaces(ctx context.Context, component string, file *os.File, generic []interface{}) (taskID string, metadata Metadata, err error) {
	implementations := make([]firmwareUploaderProvider, 0)
	for _, elem := range generic {
		temp := firmwareUploaderProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FirmwareUploader:
			temp.FirmwareUploader = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a FirmwareUploader implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return taskID, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no FirmwareUploader implementations found"),
			),
		)
	}

	return firmwareUpload(ctx, component, file, implementations)
}

func firmwareUpload(ctx context.Context, component string, file *os.File, generic []firmwareUploaderProvider) (taskID string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.FirmwareUploader == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return taskID, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			taskID, vErr := elem.FirmwareUpload(ctx, component, file)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return taskID, metadataLocal, nil
		}
	}

	return taskID, metadataLocal, multierror.Append(err, errors.New("failure in FirmwareUpload"))
}

// FirmwareTaskVerifier defines an interface to check the status for firmware related tasks queued on the BMC.
// these could be a an firmware upload and verify task or a firmware install task.
//
// This is to replace the FirmwareInstallVerifier interface
type FirmwareTaskVerifier interface {
	// FirmwareTaskStatus returns the status of the firmware upload process.
	//
	// parameters:
	// kind (required) - The FirmwareInstallStep
	// component (optional) - the component slug for the component that the firmware was uploaded for.
	// taskID (required) - the task identifier.
	// installVersion (optional) -  the firmware version being installed as part of the task if applicable.
	//
	// return values:
	// state - returns one of the FirmwareTask statuses (see devices/constants.go).
	// status - returns firmware task progress or other arbitrary task information.
	FirmwareTaskStatus(ctx context.Context, kind bconsts.FirmwareInstallStep, component, taskID, installVersion string) (state string, status string, err error)
}

// firmwareTaskVerifierProvider is an internal struct to correlate an implementation/provider and its name
type firmwareTaskVerifierProvider struct {
	name string
	FirmwareTaskVerifier
}

// firmwareTaskStatus returns the status of the firmware upload process.
func firmwareTaskStatus(ctx context.Context, kind bconsts.FirmwareInstallStep, component, taskID, installVersion string, generic []firmwareTaskVerifierProvider) (state, status string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.FirmwareTaskVerifier == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return state, status, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			state, status, vErr := elem.FirmwareTaskStatus(ctx, kind, component, taskID, installVersion)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return state, status, metadataLocal, nil
		}
	}

	return state, status, metadataLocal, multierror.Append(err, errors.New("failure in FirmwareTaskStatus"))
}

// FirmwareTaskStatusFromInterfaces identifies implementations of the FirmwareTaskVerifier interface and passes the found implementations to the firmwareTaskStatus() wrapper.
func FirmwareTaskStatusFromInterfaces(ctx context.Context, kind bconsts.FirmwareInstallStep, component, taskID, installVersion string, generic []interface{}) (state, status string, metadata Metadata, err error) {
	implementations := make([]firmwareTaskVerifierProvider, 0)
	for _, elem := range generic {
		temp := firmwareTaskVerifierProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FirmwareTaskVerifier:
			temp.FirmwareTaskVerifier = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a FirmwareTaskVerifier implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return state, status, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no FirmwareTaskVerifier implementations found"),
			),
		)
	}

	return firmwareTaskStatus(ctx, kind, component, taskID, installVersion, implementations)
}
