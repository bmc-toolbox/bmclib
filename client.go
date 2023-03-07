// Package bmclib client.go is intended to be the main public API.
// Its purpose is to make interacting with bmclib as friendly as possible.
package bmclib

import (
	"context"
	"crypto/x509"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/bmclib/v2/providers/asrockrack"
	"github.com/bmc-toolbox/bmclib/v2/providers/intelamt"
	"github.com/bmc-toolbox/bmclib/v2/providers/ipmitool"
	"github.com/bmc-toolbox/bmclib/v2/providers/redfish"
	"github.com/bmc-toolbox/common"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

// Client for BMC interactions
type Client struct {
	Auth           Auth
	connectTimeout time.Duration
	Logger         logr.Logger
	Registry       *registrar.Registry
	metadata       *bmc.Metadata
	mdLock         *sync.Mutex

	redfishVersionsNotCompatible []string
	httpClient                   *http.Client
	httpClientSetupFuncs         []func(*http.Client)
}

var (
	// default connection open timeout
	defaultConnectTimeout = 30 * time.Second
)

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

// WithSecureTLS enforces trusted TLS connections, with an optional CA certificate pool.
// Using this option with an nil pool uses the system CAs.
func WithSecureTLS(rootCAs *x509.CertPool) Option {
	return func(args *Client) {
		args.httpClientSetupFuncs = append(args.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

// WithHTTPClient sets an http client
func WithHTTPClient(c *http.Client) Option {
	return func(args *Client) {
		args.httpClient = c
	}
}

// WithConnectTimeout sets the timeout when connecting to a BMC to create a new session.
// When not defined the default connection timeout applies.
func WithConnectTimeout(t time.Duration) Option {
	return func(args *Client) {
		args.connectTimeout = t
	}
}

// WithRedfishVersionsNotCompatible sets the list of incompatible redfish versions.
//
// With this option set, The bmclib.Registry.FilterForCompatible(ctx) method will not proceed on
// devices with the given redfish version(s).
func WithRedfishVersionsNotCompatible(versions []string) Option {
	return func(args *Client) {
		args.redfishVersionsNotCompatible = append(args.redfishVersionsNotCompatible, versions...)
	}
}

// NewClient returns a new Client struct
func NewClient(host, port, user, pass string, opts ...Option) *Client {
	var defaultClient = &Client{
		Logger:                       logr.Discard(),
		Registry:                     registrar.NewRegistry(),
		redfishVersionsNotCompatible: []string{},
		connectTimeout:               defaultConnectTimeout,
	}

	for _, opt := range opts {
		opt(defaultClient)
	}
	if defaultClient.httpClient == nil {
		defaultClient.httpClient, _ = httpclient.Build(defaultClient.httpClientSetupFuncs...)
	} else {
		for _, setupFunc := range defaultClient.httpClientSetupFuncs {
			setupFunc(defaultClient.httpClient)
		}
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
	driverAsrockrack, _ := asrockrack.NewWithOptions(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Logger, asrockrack.WithHTTPClient(c.httpClient))
	c.Registry.Register(asrockrack.ProviderName, asrockrack.ProviderProtocol, asrockrack.Features, nil, driverAsrockrack)

	// register gofish provider
	driverGoFish := redfish.New(c.Auth.Host, c.Auth.Port, c.Auth.User, c.Auth.Pass, c.Logger, redfishwrapper.WithHTTPClient(c.httpClient), redfishwrapper.WithVersionsNotCompatible(c.redfishVersionsNotCompatible))
	c.Registry.Register(redfish.ProviderName, redfish.ProviderProtocol, redfish.Features, nil, driverGoFish)

	// register AMT provider
	driverAMT := intelamt.New(c.Logger, c.Auth.Host, c.Auth.Port, c.Auth.User, c.Auth.Pass)
	c.Registry.Register(intelamt.ProviderName, intelamt.ProviderProtocol, intelamt.Features, nil, driverAMT)

	// register dell idrac9 provider
	// driverIdrac9 := idrac9.NewConn(c.Auth.Host, c.Auth.Port, c.Auth.User, c.Auth.Pass, c.Logger, idrac9.WithHTTPClientConnOption(c.httpClient))
	// c.Registry.Register(idrac9.ProviderName, idrac9.ProviderProtocol, idrac9.Features, nil, driverIdrac9)

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
	ifs, metadata, err := bmc.OpenConnectionFromInterfaces(ctx, c.connectTimeout, c.Registry.GetDriverInterfaces())
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

// SetVirtualMedia controls the virtual media simulated by the BMC as being connected to the
// server. Specifically, the method ejects any currently attached virtual media, and then if
// mediaURL isn't empty, attaches a virtual media device of type kind whose contents are
// streamed from the indicated URL.
func (c *Client) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (ok bool, err error) {
	ok, metadata, err := bmc.SetVirtualMediaFromInterfaces(ctx, kind, mediaURL, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// ResetBMC pass through to library function
func (c *Client) ResetBMC(ctx context.Context, resetType string) (ok bool, err error) {
	ok, metadata, err := bmc.ResetBMCFromInterfaces(ctx, resetType, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// Inventory pass through library function to collect hardware and firmware inventory
func (c *Client) Inventory(ctx context.Context) (device *common.Device, err error) {
	device, metadata, err := bmc.GetInventoryFromInterfaces(ctx, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return device, err
}

func (c *Client) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	biosConfig, metadata, err := bmc.GetBiosConfigurationInterfaces(ctx, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return biosConfig, err
}

// FirmwareInstall pass through library function to upload firmware and install firmware
func (c *Client) FirmwareInstall(ctx context.Context, component, applyAt string, forceInstall bool, reader io.Reader) (taskID string, err error) {
	taskID, metadata, err := bmc.FirmwareInstallFromInterfaces(ctx, component, applyAt, forceInstall, reader, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return taskID, err
}

// FirmwareInstallStatus pass through library function to check firmware install status
func (c *Client) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (status string, err error) {
	status, metadata, err := bmc.FirmwareInstallStatusFromInterfaces(ctx, installVersion, component, taskID, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return status, err

}

// PostCodeGetter pass through library function to return the BIOS/UEFI POST code
func (c *Client) PostCode(ctx context.Context) (status string, code int, err error) {
	status, code, metadata, err := bmc.GetPostCodeInterfaces(ctx, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	return status, code, err
}
