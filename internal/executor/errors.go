package executor

import (
	"errors"
	"fmt"
)

var (
	// ErrNoCommandOutput is returned when an executed command produced no output.
	ErrNoCommandOutput = errors.New("command returned no output")
	// ErrVersionStrExpectedSemver is returned when a version string is not valid semver.
	ErrVersionStrExpectedSemver = errors.New("expected version string to follow semver format")
	// ErrFakeExecutorInvalidArgs is returned when the fake executor receives an unexpected number of args.
	ErrFakeExecutorInvalidArgs = errors.New("invalid number of args passed to fake executor")
	// ErrRepositoryBaseURL is returned when the update repository base URL is undefined.
	ErrRepositoryBaseURL = errors.New("repository base URL undefined, ensure UpdateOptions.BaseURL OR UPDATE_BASE_URL env var is set")
	// ErrNoUpdatesApplicable is returned when no updates apply to the device.
	ErrNoUpdatesApplicable = errors.New("no updates applicable")
	// ErrDmiDecodeRun is returned when running dmidecode fails.
	ErrDmiDecodeRun = errors.New("error running dmidecode")
	// ErrComponentListExpected is returned when a list of components to update was expected.
	ErrComponentListExpected = errors.New("expected a list of components to apply updates")
	// ErrDeviceInventory is returned when collecting the device inventory fails.
	ErrDeviceInventory = errors.New("failed to collect device inventory")
	// ErrUnsupportedDiskVendor is returned when a disk vendor is not supported.
	ErrUnsupportedDiskVendor = errors.New("unsupported disk vendor")
	// ErrNoUpdateHandlerForComponent is returned when a component slug has no update handler.
	ErrNoUpdateHandlerForComponent = errors.New("component slug has no update handler declared")
	// ErrBinNotExecutable is returned when the binary does not have the executable bit set.
	ErrBinNotExecutable = errors.New("bin has no executable bit set")
	// ErrBinLstat is returned when lstat on the binary fails.
	ErrBinLstat = errors.New("failed to run lstat on bin")
	// ErrBinLookupPath is returned when the binary path cannot be looked up.
	ErrBinLookupPath = errors.New("failed to lookup bin path")
)

// ExecError is returned when the command exits with an error or a non zero exit status
type ExecError struct {
	Cmd      string
	Stderr   string
	Stdout   string
	ExitCode int
}

// Error implements the error interface
func (u *ExecError) Error() string {
	return fmt.Sprintf("cmd %s exited with error: %s\n\t exitCode: %d\n\t stdout: %s", u.Cmd, u.Stderr, u.ExitCode, u.Stdout)
}

func newExecError(cmd string, r *Result) *ExecError {
	return &ExecError{
		Cmd:      cmd,
		Stderr:   string(r.Stderr),
		Stdout:   string(r.Stdout),
		ExitCode: r.ExitCode,
	}
}
