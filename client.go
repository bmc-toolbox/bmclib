// Package bmclib client.go is intended to be the main public API.
// Its purpose is to make interacting with bmclib as friendly as possible.
package bmclib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/bmc"
	"github.com/metal-toolbox/bmclib/constants"
	"github.com/metal-toolbox/bmclib/internal/httpclient"
	"github.com/metal-toolbox/bmclib/providers/asrockrack"
	"github.com/metal-toolbox/bmclib/providers/dell"
	"github.com/metal-toolbox/bmclib/providers/intelamt"
	"github.com/metal-toolbox/bmclib/providers/ipmitool"
	"github.com/metal-toolbox/bmclib/providers/openbmc"
	"github.com/metal-toolbox/bmclib/providers/redfish"
	"github.com/metal-toolbox/bmclib/providers/rpc"
	"github.com/metal-toolbox/bmclib/providers/supermicro"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

const (
	// default connection timeout
	defaultConnectTimeout = 30 * time.Second
	pkgName               = "github.com/metal-toolbox/bmclib"
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
	traceprovider          oteltrace.TracerProvider
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
	rpc        rpc.Provider
	openbmc    openbmc.Config
}

// NewClient returns a new Client struct
func NewClient(host, user, pass string, opts ...Option) *Client {
	defaultClient := &Client{
		Logger:                 logr.Discard(),
		Registry:               registrar.NewRegistry(),
		oneTimeRegistryEnabled: false,
		oneTimeRegistry:        registrar.NewRegistry(),
		httpClient:             httpclient.Build(),
		traceprovider:          tracenoop.NewTracerProvider(),
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
			rpc: rpc.Provider{},
			openbmc: openbmc.Config{
				Port: "443",
			},
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
	httpClient := *c.httpClient
	httpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	c.providerConfig.rpc.HTTPClient = &httpClient
	if err := mergo.Merge(driverRPC, c.providerConfig.rpc, mergo.WithOverride, mergo.WithTransformers(&rpc.Provider{})); err != nil {
		return fmt.Errorf("failed to merge user specified rpc config with the config defaults, rpc provider not available: %w", err)
	}
	c.Registry.Register(rpc.ProviderName, rpc.ProviderProtocol, rpc.Features, nil, driverRPC)

	return nil
}

// register ipmitool provider
func (c *Client) registerIPMIProvider() error {
	ipmiOpts := []ipmitool.Option{
		ipmitool.WithLogger(c.Logger),
		ipmitool.WithPort(c.providerConfig.ipmitool.Port),
		ipmitool.WithCipherSuite(c.providerConfig.ipmitool.CipherSuite),
		ipmitool.WithIpmitoolPath(c.providerConfig.ipmitool.IpmitoolPath),
	}

	driverIpmitool, err := ipmitool.New(c.Auth.Host, c.Auth.User, c.Auth.Pass, ipmiOpts...)
	if err != nil {
		return err
	}

	c.Registry.Register(ipmitool.ProviderName, ipmitool.ProviderProtocol, ipmitool.Features, nil, driverIpmitool)

	return nil
}

// register ASRR vendorapi provider
func (c *Client) registerASRRProvider() {
	asrHttpClient := *c.httpClient
	asrHttpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	driverAsrockrack := asrockrack.NewWithOptions(c.Auth.Host+":"+c.providerConfig.asrock.Port, c.Auth.User, c.Auth.Pass, c.Logger, asrockrack.WithHTTPClient(&asrHttpClient))
	c.Registry.Register(asrockrack.ProviderName, asrockrack.ProviderProtocol, asrockrack.Features, nil, driverAsrockrack)
}

// register gofish provider
func (c *Client) registerGofishProvider() {
	gfHttpClient := *c.httpClient
	gfHttpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	gofishOpts := []redfish.Option{
		redfish.WithHttpClient(&gfHttpClient),
		redfish.WithVersionsNotCompatible(c.providerConfig.gofish.VersionsNotCompatible),
		redfish.WithUseBasicAuth(c.providerConfig.gofish.UseBasicAuth),
		redfish.WithPort(c.providerConfig.gofish.Port),
		redfish.WithEtagMatchDisabled(c.providerConfig.gofish.DisableEtagMatch),
		redfish.WithSystemName(c.providerConfig.gofish.SystemName),
	}

	driverGoFish := redfish.New(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Logger, gofishOpts...)
	c.Registry.Register(redfish.ProviderName, redfish.ProviderProtocol, redfish.Features, nil, driverGoFish)
}

// register Intel AMT provider
func (c *Client) registerIntelAMTProvider() {

	iamtOpts := []intelamt.Option{
		intelamt.WithLogger(c.Logger),
		intelamt.WithHostScheme(c.providerConfig.intelamt.HostScheme),
		intelamt.WithPort(c.providerConfig.intelamt.Port),
	}
	driverAMT := intelamt.New(c.Auth.Host, c.Auth.User, c.Auth.Pass, iamtOpts...)
	c.Registry.Register(intelamt.ProviderName, intelamt.ProviderProtocol, intelamt.Features, nil, driverAMT)
}

// register Dell gofish provider
func (c *Client) registerDellProvider() {
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
}

// register supermicro vendorapi provider
func (c *Client) registerSupermicroProvider() {
	smcHttpClient := *c.httpClient
	smcHttpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	driverSupermicro := supermicro.NewClient(
		c.Auth.Host,
		c.Auth.User,
		c.Auth.Pass,
		c.Logger,
		supermicro.WithHttpClient(&smcHttpClient),
		supermicro.WithPort(c.providerConfig.supermicro.Port),
	)

	c.Registry.Register(supermicro.ProviderName, supermicro.ProviderProtocol, supermicro.Features, nil, driverSupermicro)
}

func (c *Client) registerOpenBMCProvider() {
	httpClient := *c.httpClient
	httpClient.Transport = c.httpClient.Transport.(*http.Transport).Clone()
	driver := openbmc.New(
		c.Auth.Host,
		c.Auth.User,
		c.Auth.Pass,
		c.Logger,
		openbmc.WithHttpClient(&httpClient),
		openbmc.WithPort(c.providerConfig.openbmc.Port),
	)

	c.Registry.Register(openbmc.ProviderName, openbmc.ProviderProtocol, openbmc.Features, nil, driver)
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

	if err := c.registerIPMIProvider(); err != nil {
		c.Logger.Info("ipmitool provider not available", "error", err.Error())
	}

	c.registerASRRProvider()
	c.registerGofishProvider()
	c.registerIntelAMTProvider()
	c.registerDellProvider()
	c.registerSupermicroProvider()
	c.registerOpenBMCProvider()
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

func (c *Client) RegisterSpanAttributes(m bmc.Metadata, span oteltrace.Span) {
	span.SetAttributes(attribute.String("host", c.Auth.Host))

	span.SetAttributes(attribute.String("successful-provider", m.SuccessfulProvider))

	span.SetAttributes(
		attribute.String("successful-open-conns", strings.Join(m.SuccessfulOpenConns, ",")),
	)

	span.SetAttributes(
		attribute.String("successful-close-conns", strings.Join(m.SuccessfulCloseConns, ",")),
	)

	span.SetAttributes(
		attribute.String("attempted-providers", strings.Join(m.ProvidersAttempted, ",")),
	)

	for p, e := range m.FailedProviderDetail {
		span.SetAttributes(
			attribute.String("provider-errs-"+p, e),
		)
	}
}

// Open calls the OpenConnectionFromInterfaces library function
// Any providers/drivers that do not successfully connect are removed
// from the client.Registry.Drivers. If client.Registry.Drivers ends up
// being empty then we error.
func (c *Client) Open(ctx context.Context) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Open")
	defer span.End()

	ifs, metadata, err := bmc.OpenConnectionFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	metadata.RegisterSpanAttributes(c.Auth.Host, span)
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

	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Close")
	defer span.End()

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
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

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
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GetPowerState")
	defer span.End()

	state, metadata, err := bmc.GetPowerStateFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return state, err
}

// SetPowerState pass through to library function
func (c *Client) SetPowerState(ctx context.Context, state string) (ok bool, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetPowerState")
	defer span.End()

	ok, metadata, err := bmc.SetPowerStateFromInterfaces(ctx, c.perProviderTimeout(ctx), state, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ok, err
}

// CreateUser pass through to library function
func (c *Client) CreateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "CreateUser")
	defer span.End()

	ok, metadata, err := bmc.CreateUserFromInterfaces(ctx, c.perProviderTimeout(ctx), user, pass, role, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ok, err
}

// UpdateUser pass through to library function
func (c *Client) UpdateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "UpdateUser")
	defer span.End()

	ok, metadata, err := bmc.UpdateUserFromInterfaces(ctx, c.perProviderTimeout(ctx), user, pass, role, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ok, err
}

// DeleteUser pass through to library function
func (c *Client) DeleteUser(ctx context.Context, user string) (ok bool, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "DeleteUser")
	defer span.End()

	ok, metadata, err := bmc.DeleteUserFromInterfaces(ctx, c.perProviderTimeout(ctx), user, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ok, err
}

// ReadUsers pass through to library function
func (c *Client) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "ReadUsers")
	defer span.End()

	users, metadata, err := bmc.ReadUsersFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return users, err
}

// GetBootDeviceOverride pass through to library function
func (c *Client) GetBootDeviceOverride(ctx context.Context) (override bmc.BootDeviceOverride, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GetBootDeviceOverride")
	defer span.End()

	override, metadata, err := bmc.GetBootDeviceOverrideFromInterface(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)

	return override, err
}

// SetBootDevice pass through to library function
func (c *Client) SetBootDevice(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetBootDevice")
	defer span.End()

	ok, metadata, err := bmc.SetBootDeviceFromInterfaces(ctx, c.perProviderTimeout(ctx), bootDevice, setPersistent, efiBoot, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ok, err
}

// SetVirtualMedia controls the virtual media simulated by the BMC as being connected to the
// server. Specifically, the method ejects any currently attached virtual media, and then if
// mediaURL isn't empty, attaches a virtual media device of type kind whose contents are
// streamed from the indicated URL.
func (c *Client) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (ok bool, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetVirtualMedia")
	defer span.End()

	ok, metadata, err := bmc.SetVirtualMediaFromInterfaces(ctx, kind, mediaURL, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ok, err
}

// ResetBMC pass through to library function
func (c *Client) ResetBMC(ctx context.Context, resetType string) (ok bool, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "ResetBMC")
	defer span.End()

	ok, metadata, err := bmc.ResetBMCFromInterfaces(ctx, c.perProviderTimeout(ctx), resetType, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return ok, err
}

// DeactivateSOL pass through library function to deactivate active SOL sessions
func (c *Client) DeactivateSOL(ctx context.Context) (err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "DeactivateSOL")
	defer span.End()
	metadata, err := bmc.DeactivateSOLFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return err
}

// Inventory pass through library function to collect hardware and firmware inventory
func (c *Client) Inventory(ctx context.Context) (device *common.Device, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Inventory")
	defer span.End()

	device, metadata, err := bmc.GetInventoryFromInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return device, err
}

func (c *Client) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GetBiosConfiguration")
	defer span.End()

	biosConfig, metadata, err := bmc.GetBiosConfigurationInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return biosConfig, err
}

func (c *Client) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetBiosConfiguration")
	defer span.End()

	metadata, err := bmc.SetBiosConfigurationInterfaces(ctx, c.registry().GetDriverInterfaces(), biosConfig)
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

func (c *Client) SetBiosConfigurationFromFile(ctx context.Context, cfg string) (err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SetBiosConfigurationFromFile")
	defer span.End()

	metadata, err := bmc.SetBiosConfigurationFromFileInterfaces(ctx, c.registry().GetDriverInterfaces(), cfg)
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

func (c *Client) ResetBiosConfiguration(ctx context.Context) (err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "ResetBiosConfiguration")
	defer span.End()

	metadata, err := bmc.ResetBiosConfigurationInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// FirmwareInstall pass through library function to upload firmware and install firmware
func (c *Client) FirmwareInstall(ctx context.Context, component string, operationApplyTime string, forceInstall bool, reader io.Reader) (taskID string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "FirmwareInstall")
	defer span.End()

	taskID, metadata, err := bmc.FirmwareInstallFromInterfaces(ctx, component, operationApplyTime, forceInstall, reader, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return taskID, err
}

// Note: this interface is to be deprecated in favour of a more generic FirmwareTaskStatus.
//
// FirmwareInstallStatus pass through library function to check firmware install status
func (c *Client) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (status string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "FirmwareInstallStatus")
	defer span.End()

	status, metadata, err := bmc.FirmwareInstallStatusFromInterfaces(ctx, installVersion, component, taskID, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return status, err

}

// PostCodeGetter pass through library function to return the BIOS/UEFI POST code
func (c *Client) PostCode(ctx context.Context) (status string, code int, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "PostCode")
	defer span.End()

	status, code, metadata, err := bmc.GetPostCodeInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return status, code, err
}

func (c *Client) Screenshot(ctx context.Context) (image []byte, fileType string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "Screenshot")
	defer span.End()

	image, fileType, metadata, err := bmc.ScreenshotFromInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return image, fileType, err
}

func (c *Client) ClearSystemEventLog(ctx context.Context) (err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "ClearSystemEventLog")
	defer span.End()

	metadata, err := bmc.ClearSystemEventLogFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

func (c *Client) MountFloppyImage(ctx context.Context, image io.Reader) (err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "MountFloppyImage")
	defer span.End()

	metadata, err := bmc.MountFloppyImageFromInterfaces(ctx, image, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

func (c *Client) UnmountFloppyImage(ctx context.Context) (err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "UnmountFloppyImage")
	defer span.End()

	metadata, err := bmc.UnmountFloppyImageFromInterfaces(ctx, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return err
}

// FirmwareInstallSteps return the order of actions required install firmware for a component.
func (c *Client) FirmwareInstallSteps(ctx context.Context, component string) (actions []constants.FirmwareInstallStep, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "FirmwareInstallSteps")
	defer span.End()

	status, metadata, err := bmc.FirmwareInstallStepsFromInterfaces(ctx, component, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return status, err
}

// FirmwareUpload just uploads the firmware for install, it returns a task ID to verify the upload status.
func (c *Client) FirmwareUpload(ctx context.Context, component string, file *os.File) (uploadVerifyTaskID string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "FirmwareUpload")
	defer span.End()

	uploadVerifyTaskID, metadata, err := bmc.FirmwareUploadFromInterfaces(ctx, component, file, c.Registry.GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return uploadVerifyTaskID, err
}

// FirmwareTaskStatus pass through library function to check firmware task statuses
func (c *Client) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state constants.TaskState, status string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "FirmwareTaskStatus")
	defer span.End()

	state, status, metadata, err := bmc.FirmwareTaskStatusFromInterfaces(ctx, kind, component, taskID, installVersion, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return state, status, err
}

// FirmwareInstallUploaded kicks off firmware install for a firmware uploaded with FirmwareUpload.
func (c *Client) FirmwareInstallUploaded(ctx context.Context, component, uploadVerifyTaskID string) (installTaskID string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "FirmwareInstallUploaded")
	defer span.End()

	installTaskID, metadata, err := bmc.FirmwareInstallerUploadedFromInterfaces(ctx, component, uploadVerifyTaskID, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return installTaskID, err
}

func (c *Client) FirmwareInstallUploadAndInitiate(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "FirmwareInstallUploadAndInitiate")
	defer span.End()

	taskID, metadata, err := bmc.FirmwareInstallUploadAndInitiateFromInterfaces(ctx, component, file, c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	metadata.RegisterSpanAttributes(c.Auth.Host, span)

	return taskID, err
}

// GetSystemEventLog queries for the SEL and returns the entries in an opinionated format.
func (c *Client) GetSystemEventLog(ctx context.Context) (entries bmc.SystemEventLogEntries, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GetSystemEventLog")
	defer span.End()

	entries, metadata, err := bmc.GetSystemEventLogFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return entries, err
}

// GetSystemEventLogRaw queries for the SEL and returns the raw response.
func (c *Client) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "GetSystemEventLogRaw")
	defer span.End()

	eventlog, metadata, err := bmc.GetSystemEventLogRawFromInterfaces(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)
	return eventlog, err
}

// SendNMI tells the BMC to issue an NMI to the device
func (c *Client) SendNMI(ctx context.Context) error {
	ctx, span := c.traceprovider.Tracer(pkgName).Start(ctx, "SendNMI")
	defer span.End()

	metadata, err := bmc.SendNMIFromInterface(ctx, c.perProviderTimeout(ctx), c.registry().GetDriverInterfaces())
	c.setMetadata(metadata)

	return err
}
