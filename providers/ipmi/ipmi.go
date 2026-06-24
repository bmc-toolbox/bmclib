package ipmi

import (
	"context"
	"errors"
	"strconv"
	"strings"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal/goipmi"
	"github.com/bmc-toolbox/bmclib/v2/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "ipmi"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "ipmi"
)

var (
	// Features implemented by the ipmi provider
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

// Conn for IPMI connection details
type Conn struct {
	ipmi      *goipmi.Ipmi
	log      logr.Logger
}

type Config struct {
	CipherSuite string
	Log         logr.Logger
	Port        string
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

func New(host, user, pass string, opts ...Option) (*Conn, error) {
	defaultConfig := &Config{
		Port: "623",
		Log:  logr.Discard(),
	}

	for _, opt := range opts {
		opt(defaultConfig)
	}

	// Convert port string to int
	port := 623
	if portInt, err := strconv.Atoi(defaultConfig.Port); err == nil {
		port = portInt
	}

	iopts := []goipmi.Option{
		goipmi.WithCipherSuite(defaultConfig.CipherSuite),
		goipmi.WithLogger(defaultConfig.Log),
	}
	ipt, err := goipmi.New(user, pass, host, port, iopts...)
	if err != nil {
		return nil, err
	}

	return &Conn{ipmi: ipt, log: defaultConfig.Log}, nil
}

// Open a connection to a BMC
func (c *Conn) Open(ctx context.Context) (err error) {
	_, err = c.ipmi.PowerState(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Close a connection to a BMC
func (c *Conn) Close(ctx context.Context) (err error) {
	return nil
}

// Compatible tests whether a BMC is compatible with the ipmi provider
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

	_, err = c.ipmi.PowerState(ctx)
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
	return c.ipmi.BootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
}

// BmcReset will reset a BMC
func (c *Conn) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	return c.ipmi.PowerResetBmc(ctx, resetType)
}

// DeactivateSOL will deactivate active SOL sessions
func (c *Conn) DeactivateSOL(ctx context.Context) (err error) {
	return c.ipmi.DeactivateSOL(ctx)
}

// UserRead list all users
func (c *Conn) UserRead(ctx context.Context) (users []map[string]string, err error) {
	return c.ipmi.ReadUsers(ctx)
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.ipmi.PowerState(ctx)
}

// PowerSet sets the power state of a BMC machine
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	switch strings.ToLower(state) {
	case "on":
		on, errOn := c.ipmi.IsOn(ctx)
		if errOn != nil || !on {
			ok, err = c.ipmi.PowerOn(ctx)
		} else {
			ok = true
		}
	case "off":
		ok, err = c.ipmi.PowerOff(ctx)
	case "soft":
		ok, err = c.ipmi.PowerSoft(ctx)
	case "reset":
		ok, err = c.ipmi.PowerReset(ctx)
	case "cycle":
		ok, err = c.ipmi.PowerCycle(ctx)
	default:
		err = errors.New("requested state type unknown")
	}

	return ok, err
}

func (c *Conn) ClearSystemEventLog(ctx context.Context) (err error) {
	return c.ipmi.ClearSystemEventLog(ctx)
}

func (c *Conn) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	return c.ipmi.GetSystemEventLog(ctx)
}

func (c *Conn) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	return c.ipmi.GetSystemEventLogRaw(ctx)
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Conn) SendNMI(ctx context.Context) error {
	return c.ipmi.SendPowerDiag(ctx)
}
