package racadm

import (
	"context"
	"os"

	ex "github.com/metal-toolbox/bmclib/internal/executor"

	"github.com/go-logr/logr"
)

type Racadm struct {
	Executor   ex.Executor
	RacadmPath string
	Log        logr.Logger
	Host       string
	Username   string
	Password   string
}

type Option func(*Racadm)

func WithRacadmPath(racadmPath string) Option {
	return func(c *Racadm) {
		c.RacadmPath = racadmPath
	}
}

func WithLogger(log logr.Logger) Option {
	return func(c *Racadm) {
		c.Log = log
	}
}

func New(host, user, pass string, opts ...Option) *Racadm {
	racadm := &Racadm{
		Host:     host,
		Username: user,
		Password: pass,
		Log:      logr.Discard(),
	}

	for _, opt := range opts {
		opt(racadm)
	}

	e := ex.NewExecutor(racadm.RacadmPath)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	racadm.Executor = e

	return racadm
}

// Open a connection to a BMC
func (c *Racadm) Open(ctx context.Context) (err error) {
	return nil
}

// Close a connection to a BMC
func (c *Racadm) Close(ctx context.Context) (err error) {
	return nil
}

func (s *Racadm) run(ctx context.Context, command string, additionalArgs ...string) (output string, err error) {
	racadmArgs := []string{"-r", s.Host, "-u", s.Username, "-p", s.Password, "--nocertwarm", command}
	racadmArgs = append(racadmArgs, additionalArgs...)

	s.Log.V(9).WithValues(
		"racadmArgs",
		racadmArgs,
	).Info("Calling racadm")

	s.Executor.SetArgs(racadmArgs)

	result, err := s.Executor.ExecWithContext(ctx)
	if err != nil {
		return string(result.Stderr), err
	}

	return string(result.Stdout), err
}

func (s *Racadm) ChangeBiosCfg(ctx context.Context, cfgFile string) (err error) {
	args := []string{"-t", "xml", "-f", cfgFile}

	_, err = s.run(ctx, "set", args...)

	return err
}

func (s *Racadm) SetBiosConfigurationFromFile(ctx context.Context, cfg string) (err error) {
	// Open tmp file to hold cfg
	inputConfigTmpFile, err := os.CreateTemp("", "bmclib")
	if err != nil {
		return err
	}

	defer os.Remove(inputConfigTmpFile.Name())

	_, err = inputConfigTmpFile.WriteString(cfg)
	if err != nil {
		return err
	}

	err = inputConfigTmpFile.Close()
	if err != nil {
		return err
	}

	return s.ChangeBiosCfg(ctx, inputConfigTmpFile.Name())
}
