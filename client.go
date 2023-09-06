// Package bmclib client.go is intended to be the main public API.
// Its purpose is to make interacting with bmclib as friendly as possible.
package bmclib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/providers/asrockrack"
	"github.com/bmc-toolbox/bmclib/v2/providers/dell"
	"github.com/bmc-toolbox/bmclib/v2/providers/intelamt"
	"github.com/bmc-toolbox/bmclib/v2/providers/ipmitool"
	"github.com/bmc-toolbox/bmclib/v2/providers/redfish"
	"github.com/bmc-toolbox/bmclib/v2/providers/rpc"
	"github.com/bmc-toolbox/bmclib/v2/providers/supermicro"
	"github.com/bmc-toolbox/common"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

const (
	// default connection timeout
	defaultConnectTimeout = 30 * time.Second
)

// Client for BMC interactions
type Client struct {
	Auth     Auth
	Logger   logr.Logger
	Registry *registrar.Registry

	httpClient             *http.Client
	httpClientSetupFuncs   []func(*http.Client)
	mdLock                 *sync.Mutex
	metadata               *bmc.Metadata
	perProviderTimeout     func(context.Context) time.Duration
	oneTimeRegistry        *registrar.Registry
	oneTimeRegistryEnabled bool
	providerConfig         providerConfig
}

// Auth details for connecting to a BMC
type Auth struct {
	Host string
	User string
	Pass string
}

// providerConfig contains per provider specific configuration.
type providerConfig struct {
	ipmitool   ipmitool.Config
	asrock     asrockrack.Config
	gofish     redfish.Config
	intelamt   intelamt.Config
	dell       dell.Config
	supermicro supermicro.Config
	rpc        rpc.Config
}

// NewClient returns a new Client struct
func NewClient(host, user, pass string, opts ...Option) *Client {
	defaultClient := &Client{
		Logger:                 logr.Discard(),
		Registry:               registrar.NewRegistry(),
		oneTimeRegistryEnabled: false,
		oneTimeRegistry:        registrar.NewRegistry(),
		httpClient:             httpclient.Build(),
		providerConfig: providerConfig{
			ipmitool: ipmitool.Config{
				Port: "623",
			},
			asrock: asrockrack.Config{
				Port: "443",
			},
			gofish: redfish.Config{
				Port:                  "443",
				VersionsNotCompatible: []string{},
			},
			intelamt: intelamt.Config{
				HostScheme: "http",
				Port:       16992,
			},
			dell: dell.Config{
				Port:                  "443",
				VersionsNotCompatible: []string{},
			},
			supermicro: supermicro.Config{
				Port: "443",
			},
			rpc: rpc.Config{},
		},
	}

	for _, opt := range opts {
		opt(defaultClient)
	}
	for _, setupFunc := range defaultClient.httpClientSetupFuncs {
		setupFunc(defaultClient.httpClient)
	}

	defaultClient.Registry.Logger = defaultClient.Logger
	defaultClient.Auth.Host = host
	defaultClient.Auth.User = user
	defaultClient.Auth.Pass = pass
	// len of 0 means that no Registry, with any registered providers, was passed in.
	if len(defaultClient.Registry.Drivers) == 0 {
		defaultClient.registerProviders()
	}
	defaultClient.mdLock = &sync.Mutex{}
	if defaultClient.perProviderTimeout == nil {
		defaultClient.perProviderTimeout = defaultClient.defaultTimeout
	}

	return defaultClient
}

func (c *Client) defaultTimeout(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return defaultConnectTimeout
	}

	l := len(c.registry().Drivers)
	if l == 0 {
		return time.Until(deadline)
	}
	return time.Until(deadline) / time.Duration(l)
}

func (c *Client) registerRPCProvider() error {
	driverRPC := rpc.New(c.providerConfig.rpc.ConsumerURL, c.Auth.Host, c.providerConfig.rpc.Opts.HMAC.Secrets)
	c.providerConfig.rpc.Logger = c.Logger
	if err := mergo.Merge(driverRPC, c.providerConfig.rpc, mergo.WithOverride, mergo.WithTransformers(&rpc.Config{})); err != nil {
		return fmt.Errorf("failed to merge user specified rpc config with the config defaults, rpc provider not available: %w", err)
	}
	c.Registry.Register(rpc.ProviderName, rpc.ProviderProtocol, rpc.Features, nil, driverRPC)

	return nil
}

func (c *Client) registerProviders() {
	// register the rpc provider
	// without the consumer URL there is no way to send RPC requests.
	if c.providerConfig.rpc.ConsumerURL != "" {
		// when the rpc provider is to be used, we won't register any other providers.
		err := c.registerRPCProvider()
		if err == nil {
			c.Logger.Info("note: with the rpc provider registered, no other providers will be registered and available")
			return
		}
		c.Logger.Info("failed to register rpc provider, falling back to registering all other providers", "error", err.Error())
	}
	// register ipmitool provider
	ipmiOpts := []ipmitool.Option{
		ipmitool.WithLogger(c.Logger),
		ipmitool.WithPort(c.providerConfig.ipmitool.Port),
		ipmitool.WithCipherSuite(c.providerConfig.ipmitool.CipherSuite),
		ipmitool.WithIpmitoolPath(c.providerConfig.ipmitool.IpmitoolPath),
	}
	if driverIpmitool, err := ipmitool.New(c.Auth.Host, c.Auth.User, c.Auth.Pass, ipmiOpts...); err == nil {
		c.Registry.Register(ipmitool.ProviderName, ipmitool.ProviderProtocol, ipmitool.Features, nil, driverIpmitool)
	} else {
		c.Logger.Info("ipmitool provider not available", "error", err.Error())
	}

	// register ASRR vendorapi provider
	asrHttpClient := *c.httpClient
	asrHttpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	driverAsrockrack := asrockrack.NewWithOptions(c.Auth.Host+":"+c.providerConfig.asrock.Port, c.Auth.User, c.Auth.Pass, c.Logger, asrockrack.WithHTTPClient(&asrHttpClient))
	c.Registry.Register(asrockrack.ProviderName, asrockrack.ProviderProtocol, asrockrack.Features, nil, driverAsrockrack)

	// register gofish provider
	gfHttpClient := *c.httpClient
	gfHttpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	gofishOpts := []redfish.Option{
		redfish.WithHttpClient(&gfHttpClient),
		redfish.WithVersionsNotCompatible(c.providerConfig.gofish.VersionsNotCompatible),
		redfish.WithUseBasicAuth(c.providerConfig.gofish.UseBasicAuth),
		redfish.WithPort(c.providerConfig.gofish.Port),
	}
	driverGoFish := redfish.New(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Logger, gofishOpts...)
	c.Registry.Register(redfish.ProviderName, redfish.ProviderProtocol, redfish.Features, nil, driverGoFish)

	// register Intel AMT provider
	iamtOpts := []intelamt.Option{
		intelamt.WithLogger(c.Logger),
		intelamt.WithHostScheme(c.providerConfig.intelamt.HostScheme),
		intelamt.WithPort(c.providerConfig.intelamt.Port),
	}
	driverAMT := intelamt.New(c.Auth.Host, c.Auth.User, c.Auth.Pass, iamtOpts...)
	c.Registry.Register(intelamt.ProviderName, intelamt.ProviderProtocol, intelamt.Features, nil, driverAMT)

	// register Dell gofish provider
	dellGofishHttpClient := *c.httpClient
	//dellGofishHttpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	dellGofishOpts := []dell.Option{
		dell.WithHttpClient(&dellGofishHttpClient),
		dell.WithVersionsNotCompatible(c.providerConfig.dell.VersionsNotCompatible),
		dell.WithUseBasicAuth(c.providerConfig.dell.UseBasicAuth),
		dell.WithPort(c.providerConfig.dell.Port),
	}
	driverGoFishDell := dell.New(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Logger, dellGofishOpts...)
	c.Registry.Register(dell.ProviderName, redfish.ProviderProtocol, dell.Features, nil, driverGoFishDell)

	// register supermicro vendorapi provider
	smcHttpClient := *c.httpClient
	smcHttpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	driverSupermicro := supermicro.NewClient(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Logger, supermicro.WithHttpClient(&smcHttpClient), supermicro.WithPort(c.providerConfig.supermicro.Port))
	c.Registry.Register(supermicro.ProviderName, supermicro.ProviderProtocol, supermicro.Features, nil, driverSupermicro)
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

// registry will return the oneTimeRegistry if the oneTimeRegistryEnabled is true.
func (c *Client) registry() *registrar.Registry {
	if c.oneTimeRegistryEnabled {
		c.oneTimeRegistryEnabled = false
		return c.oneTimeRegistry
	}

	return c.Registry
}

// Open calls the OpenConnectionFromInterfaces library function
// Any providers/drivers that do not successfully connect are removed
// from the client.Registry.Drivers. If client.Registry.Drivers ends up
// being empty then we error.
func (c *Client) Open(ctx context.Context) error {
	ifs, metadata, err := bmc.OpenConnectionFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	defer c.setMetadata(metadata)
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

	return nil
}

// Close pass through to library function
func (c *Client) Close(ctx context.Context) (err error) {
	// Generally, we always want the close function to run.
	// We don't want a context timeout or cancellation to prevent this.
	// But because the current model is to pass just a single context to all
	// functions, we need to create a new context here allowing closing connections.
	// This is a short term solution, and we should consider a better/more holistic model.
	if err := ctx.Err(); err != nil {
		var done context.CancelFunc
		ctx, done = context.WithTimeout(context.Background(), defaultConnectTimeout)
		defer done()
	}
	metadata, err := bmc.CloseConnectionFromInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return err
}

// FilterForCompatible removes any drivers/providers that are not compatible. It wraps the
// Client.Registry.FilterForCompatible func in order to provide a per provider timeout.
func (c *Client) FilterForCompatible(ctx context.Context) {
	perProviderTimeout, cancel := context.WithTimeout(ctx, c.perProviderTimeout(ctx))
	defer cancel()

	reg := c.registry().FilterForCompatible(perProviderTimeout)
	c.Registry.Drivers = reg
}

// GetPowerState pass through to library function
func (c *Client) GetPowerState(ctx context.Context) (state string, err error) {
	state, metadata, err := bmc.GetPowerStateFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return state, err
}

// SetPowerState pass through to library function
func (c *Client) SetPowerState(ctx context.Context, state string) (ok bool, err error) {
	ok, metadata, err := bmc.SetPowerStateFromInterfaces(ctx, c.perProviderTimeout(ctx), state, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// CreateUser pass through to library function
func (c *Client) CreateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	ok, metadata, err := bmc.CreateUserFromInterfaces(ctx, c.perProviderTimeout(ctx), user, pass, role, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// UpdateUser pass through to library function
func (c *Client) UpdateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	ok, metadata, err := bmc.UpdateUserFromInterfaces(ctx, c.perProviderTimeout(ctx), user, pass, role, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// DeleteUser pass through to library function
func (c *Client) DeleteUser(ctx context.Context, user string) (ok bool, err error) {
	ok, metadata, err := bmc.DeleteUserFromInterfaces(ctx, c.perProviderTimeout(ctx), user, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// ReadUsers pass through to library function
func (c *Client) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	users, metadata, err := bmc.ReadUsersFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return users, err
}

// SetBootDevice pass through to library function
func (c *Client) SetBootDevice(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	ok, metadata, err := bmc.SetBootDeviceFromInterfaces(ctx, c.perProviderTimeout(ctx), bootDevice, setPersistent, efiBoot, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// SetVirtualMedia controls the virtual media simulated by the BMC as being connected to the
// server. Specifically, the method ejects any currently attached virtual media, and then if
// mediaURL isn't empty, attaches a virtual media device of type kind whose contents are
// streamed from the indicated URL.
func (c *Client) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (ok bool, err error) {
	ok, metadata, err := bmc.SetVirtualMediaFromInterfaces(ctx, kind, mediaURL, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// ResetBMC pass through to library function
func (c *Client) ResetBMC(ctx context.Context, resetType string) (ok bool, err error) {
	ok, metadata, err := bmc.ResetBMCFromInterfaces(ctx, c.perProviderTimeout(ctx), resetType, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return ok, err
}

// Inventory pass through library function to collect hardware and firmware inventory
func (c *Client) Inventory(ctx context.Context) (device *common.Device, err error) {
	device, metadata, err := bmc.GetInventoryFromInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return device, err
}

func (c *Client) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	biosConfig, metadata, err := bmc.GetBiosConfigurationInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return biosConfig, err
}

// FirmwareInstall pass through library function to upload firmware and install firmware
func (c *Client) FirmwareInstall(ctx context.Context, component, applyAt string, forceInstall bool, reader io.Reader) (taskID string, err error) {
	taskID, metadata, err := bmc.FirmwareInstallFromInterfaces(ctx, component, applyAt, forceInstall, reader, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return taskID, err
}

// FirmwareInstallStatus pass through library function to check firmware install status
func (c *Client) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (status string, err error) {
	status, metadata, err := bmc.FirmwareInstallStatusFromInterfaces(ctx, installVersion, component, taskID, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return status, err

}

// PostCodeGetter pass through library function to return the BIOS/UEFI POST code
func (c *Client) PostCode(ctx context.Context) (status string, code int, err error) {
	status, code, metadata, err := bmc.GetPostCodeInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return status, code, err
}

func (c *Client) Screenshot(ctx context.Context) (image []byte, fileType string, err error) {
	image, fileType, metadata, err := bmc.ScreenshotFromInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)

	return image, fileType, err
}
