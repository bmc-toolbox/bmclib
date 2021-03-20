package redfish

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	// Features implemented by ipmitool
	Features = registrar.Features{
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
	}
)

// Conn details for redfish client
type Conn struct {
	Host string
	Port string
	User string
	Pass string
	conn *gofish.APIClient
	Log  logr.Logger
}

// Open a connection to a BMC via redfish
func (c *Conn) Open(ctx context.Context) (err error) {
	config := gofish.ClientConfig{
		Endpoint: "https://" + c.Host,
		Username: c.User,
		Password: c.Pass,
		Insecure: true,
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
