package executor

import (
	"context"
	"io"
	"strings"
)

// FakeExecute implements the utils.Executor interface
// to enable testing
type FakeExecute struct {
	Cmd      string
	Args     []string
	Env      []string
	CheckBin bool
	Stdin    io.Reader
	Stdout   []byte // Set this for the dummy data to be returned
	Stderr   []byte // Set this for the dummy data to be returned
	Quiet    bool
	ExitCode int
}

// NewFakeExecutor returns a fake Executor for use in tests.
func NewFakeExecutor(cmd string) Executor {
	return &FakeExecute{Cmd: cmd, CheckBin: false}
}

// ExecWithContext returns the preconfigured Stdout, Stderr and ExitCode as a Result.
func (e *FakeExecute) ExecWithContext(_ context.Context) (*Result, error) {
	return &Result{Stdout: e.Stdout, Stderr: e.Stderr, ExitCode: 0}, nil
}

// CheckExecutable implements the Executor interface
func (e *FakeExecute) CheckExecutable() error {
	return nil
}

// CmdPath returns the absolute path to the executable
// this means the caller should not have disabled CheckBin.
func (e *FakeExecute) CmdPath() string {
	return e.Cmd
}

// SetArgs sets the command arguments.
func (e *FakeExecute) SetArgs(a []string) {
	e.Args = a
}

// SetEnv sets the command environment variables.
func (e *FakeExecute) SetEnv(env []string) {
	e.Env = env
}

// SetQuiet enables quiet mode, suppressing command output.
func (e *FakeExecute) SetQuiet() {
	e.Quiet = true
}

// SetVerbose disables quiet mode, allowing command output.
func (e *FakeExecute) SetVerbose() {
	e.Quiet = false
}

// SetStdout sets the data to be returned as stdout.
func (e *FakeExecute) SetStdout(b []byte) {
	e.Stdout = b
}

// SetStderr sets the data to be returned as stderr.
func (e *FakeExecute) SetStderr(b []byte) {
	e.Stderr = b
}

// SetStdin sets the reader used as the command's standard input.
func (e *FakeExecute) SetStdin(r io.Reader) {
	e.Stdin = r
}

// DisableBinCheck disables checking that the command binary is executable.
func (e *FakeExecute) DisableBinCheck() {
	e.CheckBin = false
}

// SetExitCode sets the exit code to be returned.
func (e *FakeExecute) SetExitCode(i int) {
	e.ExitCode = i
}

// GetCmd returns the command with its arguments joined as a single string.
func (e *FakeExecute) GetCmd() string {
	cmd := make([]string, 0, 1+len(e.Args))
	cmd = append(cmd, e.Cmd)
	cmd = append(cmd, e.Args...)

	return strings.Join(cmd, " ")
}
