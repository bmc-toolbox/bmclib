package ipmitool

import (
	"context"
	"errors"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/metal-toolbox/bmclib/internal/ipmi"
	"github.com/metal-toolbox/bmclib/providers"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "ipmitool"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "ipmi"
)

var (
	// Features implemented by ipmitool
	Features = registrar.Features{
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
		providers.FeatureUserRead,
		providers.FeatureBmcReset,
		providers.FeatureBootDeviceSet,
		providers.FeatureClearSystemEventLog,
		providers.FeatureGetSystemEventLog,
		providers.FeatureGetSystemEventLogRaw,
		providers.FeatureDeactivateSOL,
	}
)

// Conn for Ipmitool connection details
type Conn struct {
	ipmitool *ipmi.Ipmi
	log      logr.Logger
}

type Config struct {
	CipherSuite  string
	IpmitoolPath string
	Log          logr.Logger
	Port         string
}

// Option for setting optional Client values
type Option func(*Config)

func WithLogger(log logr.Logger) Option {
	return func(c *Config) {
		c.Log = log
	}
}

func WithPort(port string) Option {
	return func(c *Config) {
		c.Port = port
	}
}

func WithCipherSuite(cipherSuite string) Option {
	return func(c *Config) {
		c.CipherSuite = cipherSuite
	}
}

func WithIpmitoolPath(ipmitoolPath string) Option {
	return func(c *Config) {
		c.IpmitoolPath = ipmitoolPath
	}
}

func New(host, user, pass string, opts ...Option) (*Conn, error) {
	defaultConfig := &Config{
		Port: "623",
		Log:  logr.Discard(),
	}

	for _, opt := range opts {
		opt(defaultConfig)
	}

	iopts := []ipmi.Option{
		ipmi.WithIpmitoolPath(defaultConfig.IpmitoolPath),
		ipmi.WithCipherSuite(defaultConfig.CipherSuite),
		ipmi.WithLogger(defaultConfig.Log),
	}
	ipt, err := ipmi.New(user, pass, host+":"+defaultConfig.Port, iopts...)
	if err != nil {
		return nil, err
	}

	return &Conn{ipmitool: ipt, log: defaultConfig.Log}, nil
}

// Open a connection to a BMC
func (c *Conn) Open(ctx context.Context) (err error) {
	_, err = c.ipmitool.PowerState(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Close a connection to a BMC
func (c *Conn) Close(ctx context.Context) (err error) {
	return nil
}

// Compatible tests whether a BMC is compatible with the ipmitool provider
func (c *Conn) Compatible(ctx context.Context) bool {
	err := c.Open(ctx)
	if err != nil {
		c.log.V(2).WithValues(
			"provider",
			c.Name(),
		).Info("warn", bmclibErrs.ErrCompatibilityCheck.Error(), err.Error())

		return false
	}
	defer c.Close(ctx)

	_, err = c.ipmitool.PowerState(ctx)
	if err != nil {
		c.log.V(2).WithValues(
			"provider",
			c.Name(),
		).Info("warn", bmclibErrs.ErrCompatibilityCheck.Error(), err.Error())
	}

	return err == nil
}

func (c *Conn) Name() string {
	return ProviderName
}

// BootDeviceSet sets the next boot device with options
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	return c.ipmitool.BootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
}

// BmcReset will reset a BMC
func (c *Conn) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	return c.ipmitool.PowerResetBmc(ctx, resetType)
}

// DeactivateSOL will deactivate active SOL sessions
func (c *Conn) DeactivateSOL(ctx context.Context) (err error) {
	return c.ipmitool.DeactivateSOL(ctx)
}

// UserRead list all users
func (c *Conn) UserRead(ctx context.Context) (users []map[string]string, err error) {
	return c.ipmitool.ReadUsers(ctx)
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.ipmitool.PowerState(ctx)
}

// PowerSet sets the power state of a BMC machine
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	switch strings.ToLower(state) {
	case "on":
		on, errOn := c.ipmitool.IsOn(ctx)
		if errOn != nil || !on {
			ok, err = c.ipmitool.PowerOn(ctx)
		} else {
			ok = true
		}
	case "off":
		ok, err = c.ipmitool.PowerOff(ctx)
	case "soft":
		ok, err = c.ipmitool.PowerSoft(ctx)
	case "reset":
		ok, err = c.ipmitool.PowerReset(ctx)
	case "cycle":
		ok, err = c.ipmitool.PowerCycle(ctx)
	default:
		err = errors.New("requested state type unknown")
	}

	return ok, err
}

func (c *Conn) ClearSystemEventLog(ctx context.Context) (err error) {
	return c.ipmitool.ClearSystemEventLog(ctx)
}

func (c *Conn) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	return c.ipmitool.GetSystemEventLog(ctx)
}

func (c *Conn) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	return c.ipmitool.GetSystemEventLogRaw(ctx)
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Conn) SendNMI(ctx context.Context) error {
	return c.ipmitool.SendPowerDiag(ctx)
}
