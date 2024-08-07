package executor

import (
	"errors"
	"fmt"
)

var (
	ErrNoCommandOutput             = errors.New("command returned no output")
	ErrVersionStrExpectedSemver    = errors.New("expected version string to follow semver format")
	ErrFakeExecutorInvalidArgs     = errors.New("invalid number of args passed to fake executor")
	ErrRepositoryBaseURL           = errors.New("repository base URL undefined, ensure UpdateOptions.BaseURL OR UPDATE_BASE_URL env var is set")
	ErrNoUpdatesApplicable         = errors.New("no updates applicable")
	ErrDmiDecodeRun                = errors.New("error running dmidecode")
	ErrComponentListExpected       = errors.New("expected a list of components to apply updates")
	ErrDeviceInventory             = errors.New("failed to collect device inventory")
	ErrUnsupportedDiskVendor       = errors.New("unsupported disk vendor")
	ErrNoUpdateHandlerForComponent = errors.New("component slug has no update handler declared")
	ErrBinNotExecutable            = errors.New("bin has no executable bit set")
	ErrBinLstat                    = errors.New("failed to run lstat on bin")
	ErrBinLookupPath               = errors.New("failed to lookup bin path")
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
