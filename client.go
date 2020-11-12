package bmclib

import (
	"context"
	"errors"

	"github.com/bmc-toolbox/bmclib/bmc"
	"github.com/bmc-toolbox/bmclib/discover"
	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/bmc-toolbox/bmclib/providers/ipmitool"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
)

// Client for BMC interactions
type Client struct {
	Auth     Auth
	Logger   logr.Logger
	Registry []interface{}
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
func WithRegistry(registry []interface{}) Option {
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
func (c *Client) SetDefaultRegistry(ctx context.Context) (err error) {
	// try discovering and registering a vendor specifc provider
	vendor, scanErr := discover.ScanAndConnect(c.Auth.Host, c.Auth.User, c.Auth.Pass, discover.WithContext(ctx), discover.WithLogger(c.Logger))
	if scanErr != nil {
		c.Logger.V(1).Info("no vendor specific controller discovered", "error", scanErr.Error())
		err = multierror.Append(err, scanErr)
	} else {
		c.Registry = append(c.Registry, vendor)
	}

	// register generic controllers
	c.Registry = append(c.Registry, &ipmitool.Conn{
		Host: c.Auth.Host,
		User: c.Auth.User,
		Pass: c.Auth.Pass,
		Log:  c.Logger,
	})

	return nil
}

// GetPowerState pass through to library function
func (c *Client) GetPowerState(ctx context.Context) (state string, err error) {
	var powerStateSetters []bmc.PowerStateSetter
	for _, elem := range c.Registry {
		switch p := elem.(type) {
		case bmc.PowerStateSetter:
			powerStateSetters = append(powerStateSetters, p)
		default:
		}
	}
	if len(powerStateSetters) == 0 {
		return state, errors.New("no registered providers found")
	}
	return bmc.GetPowerState(ctx, powerStateSetters)
}

// SetPowerState pass through to library function
func (c *Client) SetPowerState(ctx context.Context, state string) (ok bool, err error) {
	var powerStateSetters []bmc.PowerStateSetter
	for _, elem := range c.Registry {
		switch p := elem.(type) {
		case bmc.PowerStateSetter:
			powerStateSetters = append(powerStateSetters, p)
		default:
		}
	}
	if len(powerStateSetters) == 0 {
		return ok, errors.New("no registered providers found")
	}
	return bmc.SetPowerState(ctx, state, powerStateSetters)
}

// CreateUser pass through to library function
func (c *Client) CreateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	var userCreators []bmc.UserCreator
	for _, elem := range c.Registry {
		switch u := elem.(type) {
		case bmc.UserCreator:
			userCreators = append(userCreators, u)
		default:
		}
	}
	if len(userCreators) == 0 {
		return ok, errors.New("no registered providers found")
	}
	return bmc.CreateUser(ctx, user, pass, role, userCreators)
}

// UpdateUser pass through to library function
func (c *Client) UpdateUser(ctx context.Context, user, pass, role string) (ok bool, err error) {
	var userUpdaters []bmc.UserUpdater
	for _, elem := range c.Registry {
		switch u := elem.(type) {
		case bmc.UserUpdater:
			userUpdaters = append(userUpdaters, u)
		default:
		}
	}
	if len(userUpdaters) == 0 {
		return ok, errors.New("no registered providers found")
	}
	return bmc.UpdateUser(ctx, user, pass, role, userUpdaters)
}

// DeleteUser pass through to library function
func (c *Client) DeleteUser(ctx context.Context, user string) (ok bool, err error) {
	var userDeleters []bmc.UserDeleter
	for _, elem := range c.Registry {
		switch u := elem.(type) {
		case bmc.UserDeleter:
			userDeleters = append(userDeleters, u)
		default:
		}
	}
	if len(userDeleters) == 0 {
		return ok, errors.New("no registered providers found")
	}
	return bmc.DeleteUser(ctx, user, userDeleters)
}

// ReadUsers pass through to library function
func (c *Client) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	var userReaders []bmc.UserReader
	for _, elem := range c.Registry {
		switch u := elem.(type) {
		case bmc.UserReader:
			userReaders = append(userReaders, u)
		default:
		}
	}
	if len(userReaders) == 0 {
		return users, errors.New("no registered providers found")
	}
	return bmc.ReadUsers(ctx, userReaders)
}
