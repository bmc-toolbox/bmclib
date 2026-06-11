// Package lenovo implements a bmclib provider for Lenovo servers managed by the
// Lenovo XClarity Controller (XCC).
//
// XCC implements the DMTF Redfish standard (Redfish Specification 1.15.0,
// Schema Bundle 2021.4) with a number of Lenovo OEM extensions. This provider
// is built on top of the shared gofish-backed [redfishwrapper.Client] — the
// same foundation used by the generic redfish and the vendor supermicro
// providers — and layers XCC-specific behaviour (the firmware push protocol,
// OEM payload fields and registries, vendor compatibility gating) on top in
// dedicated files.
//
// # XCC conventions
//
// The following XCC behaviours are relevant throughout the provider and are
// documented here once:
//
//   - Authentication: XCC supports both HTTP Basic authentication and Redfish
//     session login (the X-Auth-Token header). Only the service root
//     "/redfish/v1/" is reachable without authentication. Set
//     [WithUseBasicAuth] to use Basic auth instead of a session.
//
//   - Session cap: XCC limits the number of concurrent open sessions to 16.
//     The provider therefore always releases its session on [Conn.Close]; a
//     failed operation must still be followed by Close so the session is not
//     leaked.
//
//   - OEM registries: XCC publishes the Lenovo "ExtendedError",
//     "LenovoExtendedWarning", "LenovoFirmwareUpdateRegistry" and
//     "BiosAttributeRegistry" registries under "/redfish/v1/Registries". The
//     error mapping in errors.go lifts the ExtendedError message id and
//     resolution text into the returned error when present.
//
//   - Task monitor quirk: a GET on the XCC TaskMonitor URI deletes a finished
//     task. Code that waits on long-running operations (e.g. firmware install)
//     must poll the Task resource, never the TaskMonitor URI.
//
// This file holds the provider scaffold: identity, configuration, the
// connection lifecycle and the vendor compatibility check. Data-plane
// capabilities (power, boot, BIOS, inventory, firmware, virtual media, ...) are
// implemented in their own files and registered by appending to [Features].
package lenovo

import (
	"context"
	"crypto/x509"
	"net/http"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/bmclib/v2/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

const (
	// ProviderName is the registered name of this provider.
	ProviderName = "lenovo"
	// ProviderProtocol is the transport/protocol this provider speaks.
	ProviderProtocol = "redfish"

	// vendorLenovo is the lower-cased token matched against the device
	// manufacturer/model during the compatibility check.
	//
	// Note: github.com/bmc-toolbox/common does not define a VendorLenovo
	// constant, so the vendor is detected with a case-insensitive substring
	// match rather than a typed comparison.
	vendorLenovo = "lenovo"
)

// Features is the set of bmclib features this provider implements.
//
// Each capability change appends the feature flags it implements so the
// registry advertises only what is actually wired up.
var Features = registrar.Features{
	// power-boot-bios
	providers.FeaturePowerState,
	providers.FeaturePowerSet,
	providers.FeatureBootDeviceSet,
	providers.FeatureBootProgress,
	providers.FeatureGetBiosConfiguration,
	providers.FeatureSetBiosConfiguration,
	providers.FeatureResetBiosConfiguration,
	providers.FeatureSecureBoot,
	providers.FeatureThermalRead,
	providers.FeaturePowerRead,
	providers.FeaturePowerCap,
	// inventory-storage
	providers.FeatureInventoryRead,
	providers.FeatureVolumeRead,
	providers.FeatureVolumeManagement,
	// firmware-tasks
	providers.FeatureFirmwareInstall,
	providers.FeatureFirmwareInstallStatus,
	providers.FeatureFirmwareUpload,
	providers.FeatureFirmwareInstallUploaded,
	providers.FeatureFirmwareUploadInitiateInstall,
	providers.FeatureFirmwareInstallSteps,
	providers.FeatureFirmwareTaskStatus,
	// virtual-media
	providers.FeatureVirtualMedia,
	providers.FeatureUnmountFloppyImage,
	// users-accounts
	providers.FeatureUserRead,
	providers.FeatureUserCreate,
	providers.FeatureUserUpdate,
	providers.FeatureUserDelete,
	// logs-sel
	providers.FeatureGetSystemEventLog,
	providers.FeatureGetSystemEventLogRaw,
	providers.FeatureClearSystemEventLog,
	// bmc-management
	providers.FeatureBmcReset,
	providers.FeatureLicenseManagement,
	providers.FeatureSecureKeyLifecycle,
	// network-serial
	providers.FeatureNetworkInterfaceRead,
	providers.FeatureNetworkInterfaceSet,
	providers.FeatureNetworkProtocolRead,
	providers.FeatureNetworkProtocolSet,
	providers.FeatureSerialRead,
	providers.FeatureSerialSet,
	// events-telemetry
	providers.FeatureEventSubscription,
	providers.FeatureTelemetry,
	// jobs-certs-snmp
	providers.FeatureJobManagement,
	providers.FeatureCertificateManagement,
	providers.FeatureSNMP,
}

// Conn is a connection to a Lenovo XCC BMC.
//
// It wraps a [redfishwrapper.Client]; capability methods delegate to the
// wrapper for standard Redfish behaviour and drop to raw requests only for XCC
// OEM operations.
type Conn struct {
	redfishwrapper *redfishwrapper.Client
	// failInventoryOnError, when true, makes Inventory return on the first
	// sub-resource error instead of collecting best-effort. Defaults to false.
	failInventoryOnError bool
	Log                  logr.Logger
}

// Config holds the optional, user-tunable settings for a [Conn].
type Config struct {
	// HTTPClient is the HTTP client used for all requests. When nil a default
	// client is built.
	HTTPClient *http.Client
	// Port is the TCP port the XCC Redfish service listens on. Defaults to "443".
	Port string
	// VersionsNotCompatible is a list of Redfish version strings to treat as
	// incompatible. With this set, Registry.FilterForCompatible will skip
	// devices reporting one of these versions.
	VersionsNotCompatible []string
	// RootCAs, when set, enables verified TLS against the given cert pool.
	RootCAs *x509.CertPool
	// UseBasicAuth selects HTTP Basic authentication instead of Redfish session
	// login. See the package documentation on the XCC session cap.
	UseBasicAuth bool
	// DisableEtagMatch disables the If-Match Etag header on POST/PATCH requests
	// to System entity endpoints.
	DisableEtagMatch bool
	// SystemName selects a specific ComputerSystem by name on multi-system
	// devices. Empty means the first/only system.
	SystemName string
	// FailInventoryOnError, when true, makes Inventory return on the first
	// sub-resource error instead of collecting best-effort. Defaults to false.
	FailInventoryOnError bool
}

// Option mutates a [Config]. Pass options to [New].
type Option func(*Config)

// WithHTTPClient sets the HTTP client used for all requests.
func WithHTTPClient(c *http.Client) Option {
	return func(cfg *Config) { cfg.HTTPClient = c }
}

// WithPort sets the XCC Redfish service port (default "443").
func WithPort(port string) Option {
	return func(cfg *Config) { cfg.Port = port }
}

// WithVersionsNotCompatible sets the Redfish versions to treat as incompatible.
//
// The version string must match the value returned by the device, e.g.
// curl -k "https://<xcc>/redfish/v1" | jq .RedfishVersion
func WithVersionsNotCompatible(versions []string) Option {
	return func(cfg *Config) { cfg.VersionsNotCompatible = versions }
}

// WithRootCAs enables verified TLS using the provided cert pool.
func WithRootCAs(pool *x509.CertPool) Option {
	return func(cfg *Config) { cfg.RootCAs = pool }
}

// WithUseBasicAuth selects HTTP Basic authentication instead of Redfish session
// login.
func WithUseBasicAuth(use bool) Option {
	return func(cfg *Config) { cfg.UseBasicAuth = use }
}

// WithSystemName selects a specific ComputerSystem by name.
func WithSystemName(name string) Option {
	return func(cfg *Config) { cfg.SystemName = name }
}

// WithEtagMatchDisabled disables the If-Match Etag header on POST/PATCH requests
// to System entity endpoints.
func WithEtagMatchDisabled(d bool) Option {
	return func(cfg *Config) { cfg.DisableEtagMatch = d }
}

// WithFailInventoryOnError makes Inventory return on the first sub-resource
// error instead of collecting best-effort.
func WithFailInventoryOnError(fail bool) Option {
	return func(cfg *Config) { cfg.FailInventoryOnError = fail }
}

// newConfig builds a [Config] with defaults applied, then applies the given
// options. Defaults: HTTP client from httpclient.Build, port "443", empty
// incompatible-versions list.
func newConfig(opts ...Option) *Config {
	cfg := &Config{
		HTTPClient:            httpclient.Build(),
		Port:                  "443",
		VersionsNotCompatible: []string{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// New returns a [Conn] for the given XCC host. The connection is not opened
// until [Conn.Open] is called.
func New(host, user, pass string, log logr.Logger, opts ...Option) *Conn {
	cfg := newConfig(opts...)

	rfOpts := []redfishwrapper.Option{
		redfishwrapper.WithHTTPClient(cfg.HTTPClient),
		redfishwrapper.WithVersionsNotCompatible(cfg.VersionsNotCompatible),
		redfishwrapper.WithEtagMatchDisabled(cfg.DisableEtagMatch),
		redfishwrapper.WithBasicAuthEnabled(cfg.UseBasicAuth),
		redfishwrapper.WithSystemName(cfg.SystemName),
	}

	if cfg.RootCAs != nil {
		rfOpts = append(rfOpts, redfishwrapper.WithSecureTLS(cfg.RootCAs))
	}

	return &Conn{
		Log:                  log,
		failInventoryOnError: cfg.FailInventoryOnError,
		redfishwrapper:       redfishwrapper.NewClient(host, cfg.Port, user, pass, rfOpts...),
	}
}

// Name returns the provider name ("lenovo").
func (c *Conn) Name() string {
	return ProviderName
}

// Open establishes the connection: it opens a Redfish session (or sets up Basic
// auth when [WithUseBasicAuth] was given).
func (c *Conn) Open(ctx context.Context) error {
	return c.redfishwrapper.Open(ctx)
}

// Close releases the connection.
//
// The session is always released so the XCC 16-session cap is respected, even
// when a preceding operation failed. Under Basic auth there is no session and
// Close is effectively a no-op.
func (c *Conn) Close(ctx context.Context) error {
	return c.redfishwrapper.Close(ctx)
}

// Compatible reports whether this provider can manage the device.
//
// The device is compatible only when a session can be established, the Redfish
// version is not excluded via [WithVersionsNotCompatible], and the device
// identifies as Lenovo. It never panics on an unreachable host; it returns
// false and logs at V(2).
func (c *Conn) Compatible(ctx context.Context) bool {
	if err := c.Open(ctx); err != nil {
		c.Log.V(2).WithValues("provider", c.Name()).
			Info(bmclibErrs.ErrCompatibilityCheck.Error(), "error", err.Error())

		return false
	}
	defer c.Close(ctx)

	if !c.redfishwrapper.VersionCompatible() {
		c.Log.V(2).WithValues("provider", c.Name()).
			Info(bmclibErrs.ErrCompatibilityCheck.Error(), "reason", "incompatible redfish version")

		return false
	}

	ok, err := c.deviceIsLenovo(ctx)
	if err != nil {
		c.Log.V(2).WithValues("provider", c.Name()).
			Info(bmclibErrs.ErrCompatibilityCheck.Error(), "error", err.Error())

		return false
	}

	return ok
}

// deviceIsLenovo returns true when the device manufacturer or model identifies
// as Lenovo.
//
// common.VendorLenovo does not exist, so this uses a case-insensitive substring
// match against the manufacturer and model strings reported by the
// ComputerSystem.
func (c *Conn) deviceIsLenovo(ctx context.Context) (bool, error) {
	vendor, model, err := c.redfishwrapper.DeviceVendorModel(ctx)
	if err != nil {
		return false, err
	}

	if strings.Contains(strings.ToLower(vendor), vendorLenovo) ||
		strings.Contains(strings.ToLower(model), vendorLenovo) {
		return true, nil
	}

	return false, nil
}
