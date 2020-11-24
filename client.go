package bmclib

import (
	"context"

	"github.com/bmc-toolbox/bmclib/bmc"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/bmc-toolbox/bmclib/registry"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"

	// for registering provider
	_ "github.com/bmc-toolbox/bmclib/providers/ipmitool"
)

// Client for BMC interactions
type Client struct {
	Auth     Auth
	Logger   logr.Logger
	Registry registry.RegistryCollection
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
func WithRegistry(registry registry.RegistryCollection) Option {
	return func(args *Client) { args.Registry = registry }
}

// NewClient returns a new Client struct
func NewClient(host, user, pass string, opts ...Option) *Client {
	var (
		defaultLogger = logging.DefaultLogger()
		defaultClient = &Client{
			Logger: defaultLogger,
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

// SetDefaultRegistry updates the registry to the default implementations
func (c *Client) SetDefaultRegistry(ctx context.Context, regs registry.RegistryCollection) (err error) {
	// try discovering and registering a vendor specifc provider
	vendor, scanErr := discover.ScanAndConnect(c.Auth.Host, c.Auth.User, c.Auth.Pass, discover.WithContext(ctx), discover.WithLogger(c.Logger))
	if scanErr != nil {
		c.Logger.V(1).Info("no vendor specific controller discovered", "error", scanErr.Error())
		err = multierror.Append(err, scanErr)
	} else {
		registry.Register("vendor", "vendor", func(host, user, pass string) (interface{}, error) {
			return vendor, nil
		}, []string{"power", "userRead"})
	}

	/*
		for _, reg := range regs {
			i, _ := reg.InitFn(c.Auth.Host, c.Auth.User, c.Auth.Pass)
			switch it := i.(type) {
			case bmc.PowerStateSetter:
				reg.Functionality.Power = it
			default:
			}
		}
	*/
	c.Registry = registry.All()

	return nil
}

/*
// GetPowerState pass through to library function
func (c *Client) GetPowerState(ctx context.Context) (state string, err error) {
	var results []bmc.PowerStateSetter
	for _, elem := range registry.All() {
		results = append(results, elem.Functionality.Power)
	}
	return bmc.GetPowerState(ctx, results)
}
*/

// GetPowerState pass through to library function
func (c *Client) GetPowerState(ctx context.Context) (state string, err error) {
	return bmc.GetPowerStateFromInterfaces(ctx, registry.GetProviders(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Registry))
}

// SetPowerState pass through to library function
func (c *Client) SetPowerState(ctx context.Context, state string) (ok bool, err error) {
	return bmc.SetPowerStateFromInterfaces(ctx, state, registry.GetProviders(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Registry))
}

// CreateUser pass through to library function
func (c *Client) CreateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	return bmc.CreateUserFromInterfaces(ctx, user, pass, role, registry.GetProviders(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Registry))
}

// UpdateUser pass through to library function
func (c *Client) UpdateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	return bmc.UpdateUserFromInterfaces(ctx, user, pass, role, registry.GetProviders(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Registry))
}

// DeleteUser pass through to library function
func (c *Client) DeleteUser(ctx context.Context, user string) (ok bool, err error) {
	return bmc.DeleteUserFromInterfaces(ctx, user, registry.GetProviders(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Registry))
}

// ReadUsers pass through to library function
func (c *Client) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	return bmc.ReadUsersFromInterfaces(ctx, registry.GetProviders(c.Auth.Host, c.Auth.User, c.Auth.Pass, c.Registry))
	/*var results []bmc.UserReader
	for _, elem := range registry.All() {
		results = append(results, elem.Functionality.UserRead)
	}
	return bmc.ReadUsers(ctx, results)*/
}
