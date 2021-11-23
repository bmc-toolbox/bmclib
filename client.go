// Package bmclib client.go is intended to be the main the public API.
// Its purpose is to make interacting with bmclib as friendly as possible.
package bmclib

import (
	"context"
	"io"
	"sync"

	"github.com/bmc-toolbox/bmclib/bmc"
	"github.com/bmc-toolbox/bmclib/providers/asrockrack"
	"github.com/bmc-toolbox/bmclib/providers/dell/idrac9"
	"github.com/bmc-toolbox/bmclib/providers/goipmi"
	"github.com/bmc-toolbox/bmclib/providers/ipmitool"
	"github.com/bmc-toolbox/bmclib/providers/redfish"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

// Client for BMC interactions
type Client struct {
	Auth     Auth
	Logger   logr.Logger
	Registry *registrar.Registry
	metadata *bmc.Metadata
	mdLock   *sync.Mutex
}

// Auth details for connecting to a BMC
type Auth struct {
	Host string
	Port string
	User string
	Pass string
}

// Option for setting optional Client values
type Option func(*Client)

// WithLogger sets the logger
func WithLogger(logger logr.Logger) Option {
	return func(args *Client) { args.Logger = logger }
}

// WithRegistry sets the Registry
func WithRegistry(registry *registrar.Registry) Option {
	return func(args *Client) { args.Registry = registry }
}

// NewClient returns a new Client struct
func NewClient(host, port, user, pass string, opts ...Option) *Client {
	var defaultClient = &Client{
		Logger:   logr.Discard(),
		Registry: registrar.NewRegistry(),
	}

	for _, opt := range opts {
		opt(defaultClient)
	}

	defaultClient.Registry.Logger = defaultClient.Logger
	defaultClient.Auth.Host = host
	defaultClient.Auth.Port = port
	defaultClient.Auth.User = user
	defaultClient.Auth.Pass = pass
	// len of 0 means that no Registry, with any registered providers was passed in.
	if len(defaultClient.Registry.Drivers) == 0 {
		defaultClient.registerProviders()
	}
	defaultClient.mdLock = &sync.Mutex{}
	return defaultClient
}

func (c *Client) registerProviders() {
	// register ipmitool provider
	driverIpmitool := &ipmitool.Conn{Host: c.Auth.Host, Port: c.Auth.Port, User: c.Auth.User, Pass: c.Auth.Pass, Log: c.Logger}
	c.Registry.Register(ipmitool.ProviderName, ipmitool.ProviderProtocol, ipmitool.Features, nil, driverIpmitool)

	// register ASRR vendorapi provider
	driverAsrockrack, _ := asrockrack.New(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Logger)
	c.Registry.Register(asrockrack.ProviderName, asrockrack.ProviderProtocol, asrockrack.Features, nil, driverAsrockrack)

	// register goipmi provider
	driverGoIpmi := &goipmi.Conn{Host: c.Auth.Host, Port: c.Auth.Port, User: c.Auth.User, Pass: c.Auth.Pass, Log: c.Logger}
	c.Registry.Register(goipmi.ProviderName, goipmi.ProviderProtocol, goipmi.Features, nil, driverGoIpmi)

	// register gofish provider
	driverGoFish := &redfish.Conn{Host: c.Auth.Host, Port: c.Auth.Port, User: c.Auth.User, Pass: c.Auth.Pass, Log: c.Logger}
	c.Registry.Register(redfish.ProviderName, redfish.ProviderProtocol, redfish.Features, nil, driverGoFish)

	// register dell idrac9 provider
	driverIdrac9 := &idrac9.Conn{Host: c.Auth.Host, Port: c.Auth.Port, User: c.Auth.User, Pass: c.Auth.Pass, Log: c.Logger}
	c.Registry.Register(idrac9.ProviderName, idrac9.ProviderProtocol, idrac9.Features, nil, driverIdrac9)
	/*
		// dummy used for testing
		driverDummy := &dummy.Conn{FailOpen: true}
		c.Registry.Register(dummy.ProviderName, dummy.ProviderProtocol, dummy.Features, nil, driverDummy)
	*/
}

// GetMetadata returns the metadata that is populated after each BMC function/method call
func (c *Client) GetMetadata() bmc.Metadata {
	if c.metadata != nil {
		return *c.metadata
	}
	return bmc.Metadata{}
}

// setMetadata wraps setting metadata with a mutex for cases where users are
// making calls to multiple *Client.X functions/methods across goroutines
func (c *Client) setMetadata(metadata bmc.Metadata) {
	// a mutex is created with the NewClient func, in the case
	// where a user doesn't call NewClient we handle by checking if
	// the mutex is nil
	if c.mdLock != nil {
		c.mdLock.Lock()
		defer c.mdLock.Unlock()
	}
	c.metadata = &metadata
}

// Open calls the OpenConnectionFromInterfaces library function
// Any providers/drivers that do not successfully connect are removed
// from the client.Registry.Drivers. If client.Registry.Drivers ends up
// being empty then we error.
func (c *Client) Open(ctx context.Context) error {
	ifs, metadata, err := bmc.OpenConnectionFromInterfaces(ctx, c.Registry.GetDriverInterfaces())
	if err != nil {
		return err
	}
	var reg registrar.Drivers
	for _, elem := range c.Registry.Drivers {
		for _, em := range ifs {
			if em == elem.DriverInterface {
				elem.DriverInterface = em
				reg = append(reg, elem)
			}
		}
	}
	c.Registry.Drivers = reg
	c.setMetadata(metadata)
	return nil
}

// Close pass through to library function
func (c *Client) Close(ctx context.Context) (err error) {
	metadata, err := bmc.CloseConnectionFromInterfaces(ctx, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return err
}

// GetPowerState pass through to library function
func (c *Client) GetPowerState(ctx context.Context) (state string, err error) {
	state, metadata, err := bmc.GetPowerStateFromInterfaces(ctx, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return state, err
}

// SetPowerState pass through to library function
func (c *Client) SetPowerState(ctx context.Context, state string) (ok bool, err error) {
	ok, metadata, err := bmc.SetPowerStateFromInterfaces(ctx, state, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// CreateUser pass through to library function
func (c *Client) CreateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	ok, metadata, err := bmc.CreateUserFromInterfaces(ctx, user, pass, role, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// UpdateUser pass through to library function
func (c *Client) UpdateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	ok, metadata, err := bmc.UpdateUserFromInterfaces(ctx, user, pass, role, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// DeleteUser pass through to library function
func (c *Client) DeleteUser(ctx context.Context, user string) (ok bool, err error) {
	ok, metadata, err := bmc.DeleteUserFromInterfaces(ctx, user, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// ReadUsers pass through to library function
func (c *Client) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	users, metadata, err := bmc.ReadUsersFromInterfaces(ctx, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return users, err
}

// SetBootDevice pass through to library function
func (c *Client) SetBootDevice(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	ok, metadata, err := bmc.SetBootDeviceFromInterfaces(ctx, bootDevice, setPersistent, efiBoot, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// ResetBMC pass through to library function
func (c *Client) ResetBMC(ctx context.Context, resetType string) (ok bool, err error) {
	ok, metadata, err := bmc.ResetBMCFromInterfaces(ctx, resetType, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// GetBMCVersion pass through library function
func (c *Client) GetBMCVersion(ctx context.Context) (version string, err error) {
	return bmc.GetBMCVersionFromInterfaces(ctx, c.Registry.GetDriverInterfaces())
}

// UpdateBMCFirmware pass through library function
func (c *Client) UpdateBMCFirmware(ctx context.Context, fileReader io.Reader, fileSize int64) (err error) {
	return bmc.UpdateBMCFirmwareFromInterfaces(ctx, fileReader, fileSize, c.Registry.GetDriverInterfaces())
}

// GetBIOSVersion pass through library function
func (c *Client) GetBIOSVersion(ctx context.Context) (version string, err error) {
	return bmc.GetBIOSVersionFromInterfaces(ctx, c.Registry.GetDriverInterfaces())
}

// UpdateBIOSFirmware pass through library function
func (c *Client) UpdateBIOSFirmware(ctx context.Context, fileReader io.Reader, fileSize int64) (err error) {
	return bmc.UpdateBIOSFirmwareFromInterfaces(ctx, fileReader, fileSize, c.Registry.GetDriverInterfaces())
}
