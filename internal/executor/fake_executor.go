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

func NewFakeExecutor(cmd string) Executor {
	return &FakeExecute{Cmd: cmd, CheckBin: false}
}

// nolint:gocyclo // TODO: break this method up and move into each $util_test.go
// FakeExecute method returns whatever you want it to return
// Set e.Stdout and e.Stderr to data to be returned
func (e *FakeExecute) ExecWithContext(_ context.Context) (*Result, error) {
	// switch e.Cmd {
	// case "ipmicfg":
	// 	if e.Args[0] == "-summary" {
	// 		buf := new(bytes.Buffer)

	// 		_, err := buf.ReadFrom(e.Stdin)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		e.Stdout = buf.Bytes()
	// 	}
	// }

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

func (e *FakeExecute) SetArgs(a []string) {
	e.Args = a
}

func (e *FakeExecute) SetEnv(env []string) {
	e.Env = env
}

func (e *FakeExecute) SetQuiet() {
	e.Quiet = true
}

func (e *FakeExecute) SetVerbose() {
	e.Quiet = false
}

func (e *FakeExecute) SetStdout(b []byte) {
	e.Stdout = b
}

func (e *FakeExecute) SetStderr(b []byte) {
	e.Stderr = b
}

func (e *FakeExecute) SetStdin(r io.Reader) {
	e.Stdin = r
}

func (e *FakeExecute) DisableBinCheck() {
	e.CheckBin = false
}

func (e *FakeExecute) SetExitCode(i int) {
	e.ExitCode = i
}

func (e *FakeExecute) GetCmd() string {
	cmd := []string{e.Cmd}
	cmd = append(cmd, e.Args...)

	return strings.Join(cmd, " ")
}
