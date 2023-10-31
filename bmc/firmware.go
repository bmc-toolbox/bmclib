package bmc

import (
	"context"
	"fmt"
	"io"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// FirmwareInstaller defines an interface to install firmware updates
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

type FirmwareInstallOptions struct {
	// The firmware upload task ID if any.
	UploadTaskID string
	// operationsApplyTime - one of the OperationApplyTime constants
	OperationApplyTime constants.OperationApplyTime
}

// FirmwareInstallerWithOpts defines an interface to install firmware updates with the given install parameters
type FirmwareInstallerWithOptions interface {
	// FirmwareInstallWithOptions uploads firmware update payload to the BMC returning the task ID
	//
	// parameters:
	// component - the component slug for the component update being installed.
	// reader - the io.reader to the firmware update file.
	//
	// return values:
	// taskID - A taskID is returned if the update process on the BMC returns an identifier for the update process.
	FirmwareInstallWithOptions(ctx context.Context, component string, reader io.Reader, opts *FirmwareInstallOptions) (taskID string, err error)
}

// firmwareInstallerProvider is an internal struct to correlate an implementation/provider and its name
type firmwareInstallerWithOptionsProvider struct {
	name string
	FirmwareInstallerWithOptions
}

// firmwareInstallWithOptions uploads and initiates firmware update for the component
func firmwareInstallWithOptions(ctx context.Context, component string, reader io.Reader, opts *FirmwareInstallOptions, generic []firmwareInstallerWithOptionsProvider) (taskID string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.FirmwareInstallerWithOptions == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return taskID, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			taskID, vErr := elem.FirmwareInstallWithOptions(ctx, component, reader, opts)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return taskID, metadataLocal, nil
		}
	}

	return taskID, metadataLocal, multierror.Append(err, errors.New("failure in FirmwareInstallWithOptions"))
}

// FirmwareInstallWithOptionsFromInterfaces identifies implementations of the FirmwareInstallerWithOptions interface and passes the found implementations to the firmwareInstallWithOptions() wrapper
func FirmwareInstallWithOptionsFromInterfaces(ctx context.Context, component string, reader io.Reader, opts *FirmwareInstallOptions, generic []interface{}) (taskID string, metadata Metadata, err error) {
	implementations := make([]firmwareInstallerWithOptionsProvider, 0)
	for _, elem := range generic {
		temp := firmwareInstallerWithOptionsProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FirmwareInstallerWithOptions:
			temp.FirmwareInstallerWithOptions = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a FirmwareInstallerWithOptions implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return taskID, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no FirmwareInstallerWithOptions implementations found"),
			),
		)
	}

	return firmwareInstallWithOptions(ctx, component, reader, opts, implementations)
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
	FirmwareUpload(ctx context.Context, component string, reader io.Reader) (uploadVerifyTaskID string, err error)
}

// firmwareUploaderProvider is an internal struct to correlate an implementation/provider and its name
type firmwareUploaderProvider struct {
	name string
	FirmwareUploader
}

// FirmwareUploaderFromInterfaces identifies implementations of the FirmwareUploader interface and passes the found implementations to the firmwareUpload() wrapper.
func FirmwareUploadFromInterfaces(ctx context.Context, component string, reader io.Reader, generic []interface{}) (taskID string, metadata Metadata, err error) {
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

	return firmwareUpload(ctx, component, reader, implementations)
}

func firmwareUpload(ctx context.Context, component string, reader io.Reader, generic []firmwareUploaderProvider) (taskID string, metadata Metadata, err error) {
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
			taskID, vErr := elem.FirmwareUpload(ctx, component, reader)
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
