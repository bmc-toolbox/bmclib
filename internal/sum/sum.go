package sum

// SUM is Supermicro Update Manager
// https://www.supermicro.com/en/solutions/management-software/supermicro-update-manager

import (
	"context"
	"os"
	"os/exec"
	"strings"

	ex "github.com/bmc-toolbox/bmclib/v2/internal/executor"

	"github.com/bmc-toolbox/common"
	"github.com/bmc-toolbox/common/config"
	"github.com/go-logr/logr"
)

// Sum is a sum command executor object
type Sum struct {
	Executor ex.Executor
	SumPath  string
	Log      logr.Logger
	Host     string
	Username string
	Password string
}

// Option for setting optional Client values
type Option func(*Sum)

func WithSumPath(sumPath string) Option {
	return func(c *Sum) {
		c.SumPath = sumPath
	}
}

func WithLogger(log logr.Logger) Option {
	return func(c *Sum) {
		c.Log = log
	}
}

func New(host, user, pass string, opts ...Option) (*Sum, error) {
	sum := &Sum{
		Host:     host,
		Username: user,
		Password: pass,
		Log:      logr.Discard(),
	}

	for _, opt := range opts {
		opt(sum)
	}

	var err error

	if sum.SumPath == "" {
		sum.SumPath, err = exec.LookPath("sum")
		if err != nil {
			return nil, err
		}
	} else {
		if _, err = os.Stat(sum.SumPath); err != nil {
			return nil, err
		}
	}

	e := ex.NewExecutor(sum.SumPath)
	e.SetEnv([]string{"LC_ALL=C.UTF-8"})
	sum.Executor = e

	return sum, nil
}

// Open a connection to a BMC
func (c *Sum) Open(ctx context.Context) (err error) {
	return nil
}

// Close a connection to a BMC
func (c *Sum) Close(ctx context.Context) (err error) {
	return nil
}

func (c *Sum) run(ctx context.Context, command string, additionalArgs ...string) (output string, err error) {
	// TODO(splaspood) use a tmp file here (as sum supports) to read the password
	sumArgs := make([]string, 0, 8+len(additionalArgs))
	sumArgs = append(sumArgs, "-i", c.Host, "-u", c.Username, "-p", c.Password, "-c", command)
	sumArgs = append(sumArgs, additionalArgs...)

	c.Log.V(9).WithValues(
		"sumArgs",
		sumArgs,
	).Info("Calling sum")

	c.Executor.SetArgs(sumArgs)

	result, err := c.Executor.ExecWithContext(ctx)
	if err != nil {
		return string(result.Stderr), err
	}

	return string(result.Stdout), err
}

func (c *Sum) GetCurrentBiosCfg(ctx context.Context) (output string, err error) {
	return c.run(ctx, "GetCurrentBiosCfg")
}

func (c *Sum) LoadDefaultBiosCfg(ctx context.Context) (err error) {
	_, err = c.run(ctx, "LoadDefaultBiosCfg")
	return err
}

func (c *Sum) ChangeBiosCfg(ctx context.Context, cfgFile string, reboot bool) (err error) {
	args := []string{"--file", cfgFile}

	if reboot {
		args = append(args, "--reboot")
	}

	_, err = c.run(ctx, "ChangeBiosCfg", args...)

	return err
}

// GetBiosConfiguration return bios configuration
func (c *Sum) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	biosText, err := c.GetCurrentBiosCfg(ctx)
	if err != nil {
		return nil, err
	}

	// We need to call vcm here to take the XML returned by SUM and convert it into a simple map
	vcm, err := config.NewVendorConfigManager("xml", common.VendorSupermicro, map[string]string{})
	if err != nil {
		return nil, err
	}

	err = vcm.Unmarshal(biosText)
	if err != nil {
		return nil, err
	}

	biosConfig, err = vcm.StandardConfig()
	if err != nil {
		return nil, err
	}

	return biosConfig, nil
}

// SetBiosConfiguration set bios configuration
func (c *Sum) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	vcm, err := config.NewVendorConfigManager("xml", common.VendorSupermicro, map[string]string{})
	if err != nil {
		return err
	}

	for k, v := range biosConfig {
		if err := applyBiosSetting(vcm, k, v); err != nil {
			return err
		}
	}

	xmlData, err := vcm.Marshal()
	if err != nil {
		return err
	}

	return c.SetBiosConfigurationFromFile(ctx, xmlData)
}

// applyBiosSetting applies a single bios key/value to the vendor config manager.
func applyBiosSetting(vcm config.VendorConfigManager, k, v string) error {
	boolFor := func(v string) (bool, bool) {
		switch v {
		case "Enabled":
			return true, true
		case "Disabled":
			return false, true
		default:
			return false, false
		}
	}

	switch {
	case k == "boot_mode":
		return vcm.BootMode(v)
	case k == "boot_order":
		return vcm.BootOrder(v)
	case k == "intel_sgx":
		return vcm.IntelSGX(v)
	case k == "secure_boot":
		if b, ok := boolFor(v); ok {
			return vcm.SecureBoot(b)
		}
	case k == "tpm":
		if b, ok := boolFor(v); ok {
			return vcm.TPM(b)
		}
	case k == "smt":
		if b, ok := boolFor(v); ok {
			return vcm.SMT(b)
		}
	case k == "sr_iov":
		if b, ok := boolFor(v); ok {
			return vcm.SRIOV(b)
		}
	case strings.HasPrefix(k, "raw:"):
		// k = raw:Menu1,SubMenu1,SubMenuMenu1,SettingName
		pathStr := strings.TrimPrefix(k, "raw:")
		path := strings.Split(pathStr, ",")
		name := path[len(path)-1]
		path = path[:len(path)-1]

		vcm.Raw(name, v, path)
	}

	return nil
}

func (c *Sum) SetBiosConfigurationFromFile(ctx context.Context, cfg string) (err error) {
	// Open tmp file to hold cfg
	inputConfigTmpFile, err := os.CreateTemp("", "bmclib")
	if err != nil {
		return err
	}

	defer func() { _ = os.Remove(inputConfigTmpFile.Name()) }()

	_, err = inputConfigTmpFile.WriteString(cfg)
	if err != nil {
		return err
	}

	err = inputConfigTmpFile.Close()
	if err != nil {
		return err
	}

	return c.ChangeBiosCfg(ctx, inputConfigTmpFile.Name(), true)
}

// ResetBiosConfiguration reset bios configuration
func (c *Sum) ResetBiosConfiguration(ctx context.Context) (err error) {
	return c.LoadDefaultBiosCfg(ctx)
}
