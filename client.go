// Package bmclib client.go is intended to be the main the public API.
// Its purpose is to make interacting with bmclib as friendly as possible.
package bmclib

import (
	"context"

	"github.com/bmc-toolbox/bmclib/bmc"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bmc-toolbox/bmclib/registry"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"

	// register providers here
	_ "github.com/bmc-toolbox/bmclib/providers/ipmitool"
)

// Client for BMC interactions
type Client struct {
	Auth     Auth
	Registry registry.Collection
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

// WithRegistry sets the Registry
func WithRegistry(registry registry.Collection) Option {
	return func(args *Client) { args.Registry = registry }
}

// NewClient returns a new Client struct
func NewClient(host, user, pass string, opts ...Option) *Client {
	var (
		defaultClient = &Client{
			Registry: registry.All(),
		}
	)
	for _, opt := range opts {
		opt(defaultClient)
	}

	defaultClient.Auth.Host = host
	defaultClient.Auth.User = user
	defaultClient.Auth.Pass = pass

	return defaultClient
}

// DiscoverProviders probes a BMC to discover what providers are compatible
func (c *Client) DiscoverProviders(ctx context.Context, log logr.Logger) (err error) {
	// try discovering and registering a vendor specific provider
	vendor, scanErr := discover.ScanAndConnect(c.Auth.Host, c.Auth.User, c.Auth.Pass, discover.WithContext(ctx), discover.WithLogger(log))
	if scanErr != nil {
		log.V(1).Info("no vendor specific controller discovered", "error", scanErr.Error())
		err = multierror.Append(err, scanErr)
	} else {
		registry.Register("vendor", "vendor", func(host, user, pass string) (interface{}, error) {
			return vendor, nil
		}, []registry.Feature{})
		c.Registry = registry.All()
	}

	return err
}

// getProviders returns a slice of interfaces for all registered implementations
func (c *Client) getProviders() []interface{} {
	var results []interface{}
	for _, reg := range c.Registry {
		i, _ := reg.InitFn(c.Auth.Host, c.Auth.User, c.Auth.Pass)
		results = append(results, i)
	}
	return results
}

// GetPowerState pass through to library function
func (c *Client) GetPowerState(ctx context.Context, log logr.Logger) (state string, err error) {
	return bmc.GetPowerStateFromInterfaces(ctx, log, c.getProviders())
}

// SetPowerState pass through to library function
func (c *Client) SetPowerState(ctx context.Context, log logr.Logger, state string) (ok bool, err error) {
	return bmc.SetPowerStateFromInterfaces(ctx, log, state, c.getProviders())
}

// CreateUser pass through to library function
func (c *Client) CreateUser(ctx context.Context, log logr.Logger, user, pass, role string) (ok bool, err error) {
	return bmc.CreateUserFromInterfaces(ctx, log, user, pass, role, c.getProviders())
}

// UpdateUser pass through to library function
func (c *Client) UpdateUser(ctx context.Context, log logr.Logger, user, pass, role string) (ok bool, err error) {
	return bmc.UpdateUserFromInterfaces(ctx, log, user, pass, role, c.getProviders())
}

// DeleteUser pass through to library function
func (c *Client) DeleteUser(ctx context.Context, log logr.Logger, user string) (ok bool, err error) {
	return bmc.DeleteUserFromInterfaces(ctx, log, user, c.getProviders())
}

// ReadUsers pass through to library function
func (c *Client) ReadUsers(ctx context.Context, log logr.Logger) (users []map[string]string, err error) {
	return bmc.ReadUsersFromInterfaces(ctx, log, c.getProviders())
}

// SetBootDevice pass through to library function
func (c *Client) SetBootDevice(ctx context.Context, log logr.Logger, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	return bmc.SetBootDeviceFromInterfaces(ctx, log, bootDevice, setPersistent, efiBoot, c.getProviders())
}

// ResetBMC pass through to library function
func (c *Client) ResetBMC(ctx context.Context, log logr.Logger, resetType string) (ok bool, err error) {
	return bmc.ResetBMCFromInterfaces(ctx, log, resetType, c.getProviders())
}
