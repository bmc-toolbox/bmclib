// Package compal implements a bmclib provider for Compal servers, such as the
// SGX30-2, whose BMC lists the host and an HGX baseboard as two Redfish systems.
package compal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/bmclib/v2/providers"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

const (
	// ProviderName is the registered name of this provider.
	ProviderName = "compal"
	// ProviderProtocol is the transport/protocol this provider speaks.
	ProviderProtocol = "redfish"
)

// Features is the set of bmclib features this provider implements.
var Features = registrar.Features{
	providers.FeaturePowerState,
	providers.FeaturePowerSet,
	providers.FeatureBootDeviceSet,
}

// Conn is a connection to a Compal BMC.
type Conn struct {
	redfishwrapper *redfishwrapper.Client
	Log            logr.Logger
}

// Config holds the optional settings for a Conn.
type Config struct {
	HTTPClient            *http.Client
	Port                  string
	VersionsNotCompatible []string
	UseBasicAuth          bool
}

// Option mutates a Config.
type Option func(*Config)

// WithHTTPClient sets the HTTP client used for all requests.
func WithHTTPClient(c *http.Client) Option {
	return func(cfg *Config) { cfg.HTTPClient = c }
}

// WithPort sets the Redfish service port (default "443").
func WithPort(port string) Option {
	return func(cfg *Config) { cfg.Port = port }
}

// WithVersionsNotCompatible sets the Redfish versions to treat as incompatible.
func WithVersionsNotCompatible(versions []string) Option {
	return func(cfg *Config) { cfg.VersionsNotCompatible = versions }
}

// WithUseBasicAuth selects HTTP Basic authentication instead of a Redfish session.
func WithUseBasicAuth(use bool) Option {
	return func(cfg *Config) { cfg.UseBasicAuth = use }
}

// New returns a Conn for the given BMC host. Open connects to it.
func New(host, user, pass string, log logr.Logger, opts ...Option) *Conn {
	cfg := &Config{HTTPClient: httpclient.Build(), Port: "443"}
	for _, opt := range opts {
		opt(cfg)
	}

	return &Conn{
		Log: log,
		redfishwrapper: redfishwrapper.NewClient(host, cfg.Port, user, pass,
			redfishwrapper.WithHTTPClient(cfg.HTTPClient),
			redfishwrapper.WithVersionsNotCompatible(cfg.VersionsNotCompatible),
			redfishwrapper.WithBasicAuthEnabled(cfg.UseBasicAuth),
			redfishwrapper.WithSystemSelector(hostSystem),
		),
	}
}

// hostSystem returns the only system with a Reset action; the HGX baseboard has none.
func hostSystem(systems []*schemas.ComputerSystem) (*schemas.ComputerSystem, error) {
	resettable := make([]*schemas.ComputerSystem, 0, 1)
	listed := make([]string, 0, len(systems))

	for _, s := range systems {
		if s == nil {
			continue
		}

		listed = append(listed, fmt.Sprintf("%s (%s)", s.ID, s.Name))

		var raw struct {
			Actions struct {
				Reset *struct {
					Target string `json:"target"`
				} `json:"#ComputerSystem.Reset"`
			} `json:"Actions"`
		}
		if err := json.Unmarshal(s.RawData, &raw); err != nil {
			return nil, fmt.Errorf("read reset action of system %q: %w", s.ID, err)
		}

		if raw.Actions.Reset == nil {
			continue
		}

		if raw.Actions.Reset.Target == "" {
			return nil, fmt.Errorf("reset action of system %q has no target", s.ID)
		}

		resettable = append(resettable, s)
	}

	if len(resettable) != 1 {
		return nil, fmt.Errorf("want one system with a reset action, found %d among: %s", len(resettable), strings.Join(listed, ", "))
	}

	host := resettable[0]
	if !strings.Contains(strings.ToLower(host.Manufacturer), common.VendorCompal) {
		return nil, fmt.Errorf("system %q is made by %q, not Compal", host.ID, host.Manufacturer)
	}

	return host, nil
}

// Name returns the provider name.
func (c *Conn) Name() string {
	return ProviderName
}

// Open connects to the BMC.
func (c *Conn) Open(ctx context.Context) error {
	return c.redfishwrapper.Open(ctx)
}

// Close closes the connection.
func (c *Conn) Close(ctx context.Context) error {
	return c.redfishwrapper.Close(ctx)
}

// Compatible reports whether the BMC is a Compal BMC with one resettable host system.
func (c *Conn) Compatible(ctx context.Context) bool {
	if err := c.Open(ctx); err != nil {
		c.Log.V(2).WithValues("provider", c.Name()).Info(bmclibErrs.ErrCompatibilityCheck.Error(), "error", err.Error())
		return false
	}
	defer func() { _ = c.Close(ctx) }()

	if !c.redfishwrapper.VersionCompatible() {
		c.Log.V(2).WithValues("provider", c.Name()).Info(bmclibErrs.ErrCompatibilityCheck.Error(), "reason", "incompatible redfish version")
		return false
	}

	if _, err := c.redfishwrapper.System(); err != nil {
		c.Log.V(2).WithValues("provider", c.Name()).Info(bmclibErrs.ErrCompatibilityCheck.Error(), "error", err.Error())
		return false
	}

	return true
}

// PowerStateGet returns the power state of the host system.
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.redfishwrapper.SystemPowerStatus(ctx)
}

// PowerSet sets the power state of the host system.
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	return c.redfishwrapper.PowerSet(ctx, state)
}

// BootDeviceSet sets the next boot device of the host system.
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	return c.redfishwrapper.SystemBootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
}
