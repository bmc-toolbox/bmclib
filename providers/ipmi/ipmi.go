// Package ipmi implements a BMC client using native IPMI over LAN.
package ipmi

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal/goipmi"
	"github.com/bmc-toolbox/bmclib/v2/providers"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "ipmi"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "ipmi"
)

// Features implemented by the ipmi provider
var Features = registrar.Features{
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

// Conn for IPMI connection details
type Conn struct {
	ipmi   *goipmi.Ipmi
	log    logr.Logger
	openMu sync.Mutex
	open   bool
}

// Config holds the optional configuration for an ipmi connection.
type Config struct {
	CipherSuite string
	Log         logr.Logger
	Port        string
}

// Option for setting optional Client values
type Option func(*Config)

// WithLogger sets the logger used by the provider.
func WithLogger(log logr.Logger) Option {
	return func(c *Config) {
		c.Log = log
	}
}

// WithPort sets the port used to connect to the BMC.
func WithPort(port string) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithCipherSuite sets the IPMI cipher suite used for the connection.
func WithCipherSuite(cipherSuite string) Option {
	return func(c *Config) {
		c.CipherSuite = cipherSuite
	}
}

// New returns a new ipmi connection.
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
	c.openMu.Lock()
	defer c.openMu.Unlock()
	if c.open {
		return nil
	}
	if err := c.ipmi.Open(ctx); err != nil {
		return err
	}
	c.open = true
	return nil
}

// Close a connection to a BMC
func (c *Conn) Close(ctx context.Context) (err error) {
	c.openMu.Lock()
	defer c.openMu.Unlock()
	if !c.open {
		return nil
	}
	// Clear the open state before closing so a failed Close still allows
	// callers to retry Open; the underlying session is torn down regardless.
	c.open = false
	return c.ipmi.Close(ctx)
}

// Compatible tests whether a BMC is compatible with the ipmi provider
func (c *Conn) Compatible(ctx context.Context) bool {
	// Use an isolated, throwaway connection so the compatibility check never
	// affects the session state of the caller's connection.
	clone, err := c.ipmi.Clone()
	if err != nil {
		c.log.V(2).WithValues("provider", c.Name()).
			Info("warn", bmclibErrs.ErrCompatibilityCheck.Error(), err.Error())
		return false
	}
	probe := &Conn{ipmi: clone, log: c.log}
	if err := probe.Open(ctx); err != nil {
		c.log.V(2).WithValues("provider", c.Name()).
			Info("warn", bmclibErrs.ErrCompatibilityCheck.Error(), err.Error())
		return false
	}
	defer func() { _ = probe.Close(ctx) }()

	if _, err := probe.ipmi.PowerState(ctx); err != nil {
		c.log.V(2).WithValues("provider", c.Name()).
			Info("warn", bmclibErrs.ErrCompatibilityCheck.Error(), err.Error())
		return false
	}
	return true
}

// Name returns the name of this provider.
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

// ClearSystemEventLog clears the BMC System Event Log (SEL).
func (c *Conn) ClearSystemEventLog(ctx context.Context) (err error) {
	return c.ipmi.ClearSystemEventLog(ctx)
}

// GetSystemEventLog returns the BMC System Event Log (SEL) entries.
func (c *Conn) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	return c.ipmi.GetSystemEventLog(ctx)
}

// GetSystemEventLogRaw returns the raw BMC System Event Log (SEL).
func (c *Conn) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	return c.ipmi.GetSystemEventLogRaw(ctx)
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Conn) SendNMI(ctx context.Context) error {
	return c.ipmi.SendPowerDiag(ctx)
}
