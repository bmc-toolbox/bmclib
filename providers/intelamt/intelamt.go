package intelamt

import (
	"context"
	"errors"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/iamt"
	"github.com/jacobweinstock/registrar"
	"github.com/metal-toolbox/bmclib/providers"
)

const (
	// ProviderName for the provider AMT implementation
	ProviderName = "IntelAMT"
	// ProviderProtocol for the provider AMT implementation
	ProviderProtocol = "AMT"
)

var (
	// Features implemented by the AMT provider
	Features = registrar.Features{
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
		providers.FeatureBootDeviceSet,
	}
)

// iamtClient interface allows us to mock the client for testing
type iamtClient interface {
	Close(context.Context) error
	IsPoweredOn(context.Context) (bool, error)
	Open(context.Context) error
	PowerCycle(context.Context) error
	PowerOff(context.Context) error
	PowerOn(context.Context) error
	SetPXE(context.Context) error
}

// Conn is a connection to a BMC via Intel AMT
type Conn struct {
	client iamtClient
}

// Option for setting optional Client values
type Option func(*Config)

func WithPort(port uint32) Option {
	return func(c *Config) {
		c.Port = port
	}
}

func WithHostScheme(hostScheme string) Option {
	return func(c *Config) {
		c.HostScheme = hostScheme
	}
}

func WithLogger(logger logr.Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

type Config struct {
	// HostScheme should be either "http" or "https".
	HostScheme string
	// Port is the port number to connect to.
	Port   uint32
	Logger logr.Logger
}

// New creates a new AMT connection
func New(host string, user string, pass string, opts ...Option) *Conn {
	defaultClient := &Config{
		HostScheme: "http",
		Port:       16992,
		Logger:     logr.Discard(),
	}
	for _, opt := range opts {
		opt(defaultClient)
	}

	iopts := []iamt.Option{
		iamt.WithLogger(defaultClient.Logger),
		iamt.WithPort(defaultClient.Port),
		iamt.WithScheme(defaultClient.HostScheme),
	}
	return &Conn{
		client: iamt.NewClient(host, user, pass, iopts...),
	}
}

// Name of the provider
func (c *Conn) Name() string {
	return ProviderName
}

// Open a connection to the BMC via Intel AMT.
func (c *Conn) Open(ctx context.Context) (err error) {
	return c.client.Open(ctx)
}

// Close a connection to a BMC
func (c *Conn) Close() (err error) {
	return c.client.Close(context.Background())
}

// Compatible tests whether a BMC is compatible with the ipmitool provider
func (c *Conn) Compatible(ctx context.Context) bool {
	if err := c.client.Open(ctx); err != nil {
		return false
	}

	if _, err := c.client.IsPoweredOn(ctx); err != nil {
		return false
	}

	return true
}

// BootDeviceSet sets the next boot device with options
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	if strings.ToLower(bootDevice) != "pxe" {
		return false, errors.New("only pxe boot device is supported for AMT provider")
	}
	if err := c.client.SetPXE(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	on, err := c.client.IsPoweredOn(ctx)
	if err != nil {
		return "", err
	}
	if on {
		return "on", nil
	}

	return "off", nil
}

// PowerSet sets the power state of a BMC machine
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	on, _ := c.client.IsPoweredOn(ctx)

	switch strings.ToLower(state) {
	case "on":
		if on {
			return true, nil
		}
		if err := c.client.PowerOn(ctx); err != nil {
			return false, err
		}
		ok = true
	case "off":
		if !on {
			return true, nil
		}
		if err := c.client.PowerOff(ctx); err != nil {
			return false, err
		}
		ok = true
	case "cycle":
		if err := c.client.PowerCycle(ctx); err != nil {
			return false, err
		}
		ok = true
	default:
		err = errors.New("requested state type unknown")
	}

	return ok, err
}
