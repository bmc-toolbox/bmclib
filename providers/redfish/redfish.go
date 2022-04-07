package redfish

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish"
	rf "github.com/stmcginnis/gofish/redfish"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "gofish"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "redfish"
)

var (
	// Features implemented by gofish
	Features = registrar.Features{
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
		providers.FeatureUserCreate,
		providers.FeatureUserUpdate,
		providers.FeatureUserDelete,
		providers.FeatureInventoryRead,
		providers.FeatureFirmwareInstall,
		providers.FeatureFirmwareInstallStatus,
	}
)

// Conn details for redfish client
type Conn struct {
	Host                 string
	Port                 string
	User                 string
	Pass                 string
	conn                 *gofish.APIClient
	Log                  logr.Logger
	httpClient           *http.Client
	httpClientSetupFuncs []func(*http.Client)
}

// Option is a function applied to a *Conn
type Option func(*Conn)

// WithHTTPClient returns an option that sets an HTTP client for the connecion
func WithHTTPClient(cli *http.Client) Option {
	return func(c *Conn) {
		c.httpClient = cli
	}
}

// WithSecureTLS returns an option that enables secure TLS with an optional cert pool.
func WithSecureTLS(rootCAs *x509.CertPool) Option {
	return func(c *Conn) {
		c.httpClientSetupFuncs = append(c.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

// New returns a redfish *Conn
func New(host, port, user, pass string, log logr.Logger, opts ...Option) *Conn {
	conn := &Conn{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		Log:  log,
	}
	for _, opt := range opts {
		opt(conn)
	}
	return conn
}

// Open a connection to a BMC via redfish
func (c *Conn) Open(ctx context.Context) (err error) {
	if !strings.HasPrefix(c.Host, "https://") && !strings.HasPrefix(c.Host, "http://") {
		c.Host = "https://" + c.Host
	}

	config := gofish.ClientConfig{
		Endpoint:   "https://" + c.Host,
		Username:   c.User,
		Password:   c.Pass,
		Insecure:   true,
		HTTPClient: c.httpClient,
	}

	if config.HTTPClient == nil {
		config.HTTPClient, err = httpclient.Build(c.httpClientSetupFuncs...)
		if err != nil {
			return err
		}
	} else {
		for _, setupFunc := range c.httpClientSetupFuncs {
			setupFunc(config.HTTPClient)
		}
	}

	debug := os.Getenv("DEBUG_BMCLIB")
	if debug == "true" {
		config.DumpWriter = os.Stdout
	}

	c.conn, err = gofish.ConnectContext(ctx, config)
	if err != nil {
		return err
	}
	return nil
}

// Close a connection to a BMC via redfish
func (c *Conn) Close(ctx context.Context) error {
	c.conn.Logout()
	return nil
}

func (c *Conn) Name() string {
	return ProviderName
}

// Compatible tests whether a BMC is compatible with the gofish provider
func (c *Conn) Compatible(ctx context.Context) bool {
	err := c.Open(ctx)
	if err != nil {
		c.Log.V(0).Error(err, "error checking compatibility: open connection failed")
		return false
	}
	defer c.Close(ctx)
	_, err = c.PowerStateGet(ctx)
	if err != nil {
		c.Log.V(0).Error(err, "error checking compatibility: power state get failed")
	}
	return err == nil
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.status(ctx)
}

// PowerSet sets the power state of a BMC via redfish
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	switch strings.ToLower(state) {
	case "on":
		return c.on(ctx)
	case "off":
		return c.hardoff(ctx)
	case "soft":
		return c.off(ctx)
	case "reset":
		return c.reset(ctx)
	case "cycle":
		return c.cycle(ctx)
	default:
		return false, errors.New("unknown power action")
	}
}

func (c *Conn) on(ctx context.Context) (ok bool, err error) {
	service := c.conn.Service
	ss, err := service.Systems()
	if err != nil {
		return false, err
	}
	for _, system := range ss {
		if system.PowerState == rf.OnPowerState {
			break
		}
		err = system.Reset(rf.OnResetType)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (c *Conn) off(ctx context.Context) (ok bool, err error) {
	service := c.conn.Service
	ss, err := service.Systems()
	if err != nil {
		return false, err
	}
	for _, system := range ss {
		if system.PowerState == rf.OffPowerState {
			break
		}
		err = system.Reset(rf.GracefulShutdownResetType)
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

func (c *Conn) status(ctx context.Context) (result string, err error) {
	service := c.conn.Service
	ss, err := service.Systems()
	if err != nil {
		return "", err
	}
	for _, system := range ss {
		return string(system.PowerState), nil
	}
	return "", errors.New("unable to retrieve status")
}

func (c *Conn) reset(ctx context.Context) (ok bool, err error) {
	service := c.conn.Service
	ss, err := service.Systems()
	if err != nil {
		return false, err
	}
	for _, system := range ss {
		err = system.Reset(rf.PowerCycleResetType)
		if err != nil {
			c.Log.V(1).Info("warning", "msg", err.Error())
			_, _ = c.off(ctx)
			for wait := 1; wait < 10; wait++ {
				status, _ := c.status(ctx)
				if status == "off" {
					break
				}
				time.Sleep(1 * time.Second)
			}
			_, errMsg := c.on(ctx)
			return true, errMsg
		}
	}
	return true, nil
}

func (r *Conn) hardoff(ctx context.Context) (ok bool, err error) {
	service := r.conn.Service
	ss, err := service.Systems()
	if err != nil {
		return false, err
	}
	for _, system := range ss {
		if system.PowerState == rf.OffPowerState {
			break
		}
		err = system.Reset(rf.ForceOffResetType)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (r *Conn) cycle(ctx context.Context) (ok bool, err error) {
	service := r.conn.Service
	ss, err := service.Systems()
	if err != nil {
		return false, err
	}
	res, err := r.status(ctx)
	if err != nil {
		return false, fmt.Errorf("power cycle failed: unable to get current state")
	}
	if strings.ToLower(res) == "off" {
		return false, fmt.Errorf("power cycle failed: Command not supported in present state: %v", res)
	}

	for _, system := range ss {
		err = system.Reset(rf.ForceRestartResetType)
		if err != nil {
			return false, errors.WithMessage(err, "power cycle failed")
		}
	}
	return true, nil
}
