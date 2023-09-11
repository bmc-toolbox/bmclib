package goipmi

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/providers"
	"github.com/bougou/go-ipmi"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

const (
	// ProviderName for the provider pure go ipmi (go-ipmi) implementation.
	ProviderName = "goipmi"
	// ProviderProtocol for the provider pure go ipmi (go-ipmi) implementation.
	ProviderProtocol = "ipmi"
)

var (
	// Features implemented by goipmi.
	Features = registrar.Features{
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
		providers.FeatureBootDeviceSet,
		providers.FeatureUserRead,
	}
)

// Config for goipmi provider.
type Config struct {
	CipherSuite int
	Log         logr.Logger
	Port        int

	client *ipmi.Client
}

// Option for setting optional Client values.
type Option func(*Config)

func New(host string, port int, user, pass string, opts ...Option) (*Config, error) {
	cl, err := ipmi.NewClient(host, port, user, pass)
	if err != nil {
		return nil, err
	}
	c := &Config{
		CipherSuite: 3,
		Log:         logr.Discard(),
		client:      cl,
	}
	for _, opt := range opts {
		opt(c)
	}
	c.client.WithInterface(ipmi.InterfaceLanplus)
	c.client.WithCipherSuiteID(toCipherSuiteID(c.CipherSuite))

	return c, nil
}

func WithCipherSuite(cipherSuite int) Option {
	return func(c *Config) {
		c.CipherSuite = cipherSuite
	}
}

func (c *Config) Name() string {
	return ProviderName
}

func (c *Config) Open(ctx context.Context) error {
	c.client.WithTimeout(getTimeout(ctx))
	return c.client.Connect()
}

func getTimeout(ctx context.Context) time.Duration {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 30 * time.Second
	}

	return time.Until(deadline)
}

func (c *Config) Close(_ context.Context) error {
	return c.client.Close()
}

func (c *Config) PowerStateGet(_ context.Context) (string, error) {
	r, err := c.client.GetChassisStatus()
	if err != nil {
		return "", err
	}
	state := "off"
	if r.PowerIsOn {
		state = "on"
	}

	return state, nil
}

func (c *Config) PowerSet(_ context.Context, state string) (bool, error) {
	var action ipmi.ChassisControl
	switch strings.ToLower(state) {
	case "on":
		action = ipmi.ChassisControlPowerUp
	case "off":
		action = ipmi.ChassisControlPowerDown
	case "soft":
		action = ipmi.ChassisControlSoftShutdown
	case "reset":
		action = ipmi.ChassisControlHardwareRest
	case "cycle":
		action = ipmi.ChassisControlPowerCycle
	default:
		return false, fmt.Errorf("unknown or unimplemented state request: %v", state)
	}

	// ipmi.ChassisControlResponse is an empty struct.
	// No methods return any actual response. So we ignore it.
	_, err := c.client.ChassisControl(action)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *Config) BootDeviceSet(_ context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	d := ipmi.BootDeviceSelectorNoOverride
	switch strings.ToLower(bootDevice) {
	case "pxe":
		d = ipmi.BootDeviceSelectorForcePXE
	case "disk":
		d = ipmi.BootDeviceSelectorForceHardDrive
	case "safe":
		d = ipmi.BootDeviceSelectorForceHardDriveSafe
	case "diag":
		d = ipmi.BootDeviceSelectorForceDiagnosticPartition
	case "cdrom":
		d = ipmi.BootDeviceSelectorForceCDROM
	case "bios":
		d = ipmi.BootDeviceSelectorForceBIOSSetup
	case "floppy":
		d = ipmi.BootDeviceSelectorForceFloppy
	}
	bt := ipmi.BIOSBootTypeLegacy
	if efiBoot {
		bt = ipmi.BIOSBootTypeEFI
	}

	if err := c.client.SetBootDevice(d, bt, setPersistent); err != nil {
		return false, err
	}

	return true, nil
}

func (c *Config) UserRead(_ context.Context) (users []map[string]string, err error) {
	u, err := c.client.ListUser(0)
	if err != nil {
		return nil, err
	}

	for _, v := range u {
		if v.Name == "" {
			continue
		}
		users = append(users, map[string]string{
			"id":               fmt.Sprintf("%v", v.ID),
			"name":             v.Name,
			"callin":           fmt.Sprintf("%v", v.Callin),
			"linkAuth":         fmt.Sprintf("%v", v.LinkAuthEnabled),
			"ipmiMsg":          fmt.Sprintf("%v", v.IPMIMessagingEnabled),
			"channelPrivLimit": fmt.Sprintf("%v", v.MaxPrivLevel),
		})
	}

	return users, nil
}

func toCipherSuiteID(c int) ipmi.CipherSuiteID {
	switch c {
	case 0:
		return ipmi.CipherSuiteID0
	case 1:
		return ipmi.CipherSuiteID1
	case 2:
		return ipmi.CipherSuiteID2
	case 3:
		return ipmi.CipherSuiteID3
	case 4:
		return ipmi.CipherSuiteID4
	case 5:
		return ipmi.CipherSuiteID5
	case 6:
		return ipmi.CipherSuiteID6
	case 7:
		return ipmi.CipherSuiteID7
	case 8:
		return ipmi.CipherSuiteID8
	case 9:
		return ipmi.CipherSuiteID9
	case 10:
		return ipmi.CipherSuiteID10
	case 11:
		return ipmi.CipherSuiteID11
	case 12:
		return ipmi.CipherSuiteID12
	case 13:
		return ipmi.CipherSuiteID13
	case 14:
		return ipmi.CipherSuiteID14
	case 15:
		return ipmi.CipherSuiteID15
	case 16:
		return ipmi.CipherSuiteID16
	case 17:
		return ipmi.CipherSuiteID17
	case 18:
		return ipmi.CipherSuiteID18
	case 19:
		return ipmi.CipherSuiteID19
	default:
		return ipmi.CipherSuiteID3
	}

}
