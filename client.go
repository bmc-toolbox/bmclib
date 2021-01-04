// Package bmclib client.go is intended to be the main the public API.
// Its purpose is to make interacting with bmclib as friendly as possible.
package bmclib

import (
	"context"
	"sync"

	"github.com/bmc-toolbox/bmclib/bmc"
	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/bmc-toolbox/bmclib/registry"
	"github.com/go-logr/logr"

	// register providers here
	_ "github.com/bmc-toolbox/bmclib/providers/ipmitool"
)

// Client for BMC interactions
type Client struct {
	Auth     Auth
	Logger   logr.Logger
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

// WithLogger sets the logger
func WithLogger(logger logr.Logger) Option {
	return func(args *Client) { args.Logger = logger }
}

// WithRegistry sets the Registry
func WithRegistry(registry registry.Collection) Option {
	return func(args *Client) { args.Registry = registry }
}

// NewClient returns a new Client struct
func NewClient(host, port, user, pass string, opts ...Option) *Client {
	var defaultClient = &Client{
		Logger:   logging.DefaultLogger(),
		Registry: registry.All(),
	}
	for _, opt := range opts {
		opt(defaultClient)
	}

	defaultClient.Auth.Host = host
	defaultClient.Auth.Port = port
	defaultClient.Auth.User = user
	defaultClient.Auth.Pass = pass
	defaultClient.setProviders()

	return defaultClient
}

// setProviders updates the Registry with corresponding interfaces for all registered implementations
func (c *Client) setProviders() {
	for _, elem := range c.Registry {
		providerInterface, isCompatFn, err := elem.InitFn(c.Auth.Host, c.Auth.Port, c.Auth.User, c.Auth.Pass, c.Logger)
		if err != nil {
			c.Logger.V(0).Info("provider registration error", "error", err.Error(), "provider", elem.Provider)
			continue
		}
		elem.ProviderInterface = providerInterface
		elem.IsCompatibleFn = isCompatFn
	}
}

// getProviders returns a slice of interfaces for all registered implementations
func (c *Client) getProviders() []interface{} {
	results := make([]interface{}, len(c.Registry))
	for _, elem := range c.Registry {
		results = append(results, elem.ProviderInterface)
	}
	return results
}

// DiscoverCompatible updates the registry with only compatible BMCs
func (c *Client) DiscoverCompatible(ctx context.Context) {
	var wg sync.WaitGroup
	result := make(registry.Collection, 0)
	for _, elem := range c.Registry {
		wg.Add(1)
		go func(isCompat registry.IsCompatibleFn, reg *registry.Registry, wg *sync.WaitGroup) {
			if isCompat(ctx) {
				result = append(result, reg)
			}
			wg.Done()
		}(elem.IsCompatibleFn, elem, &wg)
	}
	wg.Wait()
	c.Registry = result
}

// Open pass through to library function
func (c *Client) Open(ctx context.Context) (err error) {
	return bmc.OpenConnectionFromInterfaces(ctx, c.getProviders())
}

// Close pass through to library function
func (c *Client) Close(ctx context.Context) (err error) {
	return bmc.CloseConnectionFromInterfaces(ctx, c.getProviders())
}

// GetPowerState pass through to library function
func (c *Client) GetPowerState(ctx context.Context) (state string, err error) {
	return bmc.GetPowerStateFromInterfaces(ctx, c.getProviders())
}

// SetPowerState pass through to library function
func (c *Client) SetPowerState(ctx context.Context, state string) (ok bool, err error) {
	return bmc.SetPowerStateFromInterfaces(ctx, state, c.getProviders())
}

// CreateUser pass through to library function
func (c *Client) CreateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	return bmc.CreateUserFromInterfaces(ctx, user, pass, role, c.getProviders())
}

// UpdateUser pass through to library function
func (c *Client) UpdateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	return bmc.UpdateUserFromInterfaces(ctx, user, pass, role, c.getProviders())
}

// DeleteUser pass through to library function
func (c *Client) DeleteUser(ctx context.Context, user string) (ok bool, err error) {
	return bmc.DeleteUserFromInterfaces(ctx, user, c.getProviders())
}

// ReadUsers pass through to library function
func (c *Client) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	return bmc.ReadUsersFromInterfaces(ctx, c.getProviders())
}

// SetBootDevice pass through to library function
func (c *Client) SetBootDevice(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	return bmc.SetBootDeviceFromInterfaces(ctx, bootDevice, setPersistent, efiBoot, c.getProviders())
}

// ResetBMC pass through to library function
func (c *Client) ResetBMC(ctx context.Context, resetType string) (ok bool, err error) {
	return bmc.ResetBMCFromInterfaces(ctx, resetType, c.getProviders())
}
