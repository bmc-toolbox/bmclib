package sum

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/bmc-toolbox/common/config"
	"github.com/go-logr/logr"
)

// Conn for Sum connection details
type Exec struct {
	SumPath  string
	Log      logr.Logger
	Host     string
	Username string
	Password string
}

// Option for setting optional Client values
type Option func(*Exec)

func WithSumPath(sumPath string) Option {
	return func(c *Exec) {
		c.SumPath = sumPath
	}
}

func WithLogger(log logr.Logger) Option {
	return func(c *Exec) {
		c.Log = log
	}
}

func New(host, user, pass string, opts ...Option) (*Exec, error) {
	sum := &Exec{
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

	return sum, nil
}

// Open a connection to a BMC
func (c *Exec) Open(ctx context.Context) (err error) {
	return nil
}

// Close a connection to a BMC
func (c *Exec) Close(ctx context.Context) (err error) {
	return nil
}

func (s *Exec) run(ctx context.Context, command string, additionalArgs ...string) (output string, err error) {
	// TODO(splaspood) use a tmp file here (as sum supports) to read the password
	sumArgs := []string{"-i", s.Host, "-u", s.Username, "-p", s.Password, "-c", command}
	sumArgs = append(sumArgs, additionalArgs...)

	s.Log.V(9).WithValues(
		"sumArgs",
		sumArgs,
	).Info("Calling sum")

	cmd := exec.CommandContext(ctx, s.SumPath, sumArgs...)
	b_out, err := cmd.CombinedOutput()
	if err != nil || ctx.Err() != nil {
		if err != nil {
			return
		}

		if ctx.Err() != nil {
			return output, ctx.Err()
		}
	}

	return string(b_out), err
}

func (s *Exec) GetCurrentBiosCfg(ctx context.Context) (output string, err error) {
	return s.run(ctx, "GetCurrentBiosCfg")
}

func (s *Exec) LoadDefaultBiosCfg(ctx context.Context) (err error) {
	_, err = s.run(ctx, "LoadDefaultBiosCfg")
	return err
}

func (s *Exec) ChangeBiosCfg(ctx context.Context, cfgFile string, reboot bool) (err error) {
	args := []string{"--file", cfgFile}

	if reboot {
		args = append(args, "--reboot")
	}

	_, err = s.run(ctx, "ChangeBiosCfg", args...)

	return err
}

// GetBiosConfiguration return bios configuration
func (s *Exec) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	biosText, err := s.GetCurrentBiosCfg(ctx)
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
func (s *Exec) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	vcm, err := config.NewVendorConfigManager("xml", common.VendorSupermicro, map[string]string{})
	if err != nil {
		return err
	}

	for k, v := range biosConfig {
		switch {
		case k == "boot_mode":
			if err = vcm.BootMode(v); err != nil {
				return err
			}
		case k == "boot_order":
			if err = vcm.BootOrder(v); err != nil {
				return err
			}
		case k == "intel_sgx":
			if err = vcm.IntelSGX(v); err != nil {
				return err
			}
		case k == "secure_boot":
			switch v {
			case "Enabled":
				if err = vcm.SecureBoot(true); err != nil {
					return err
				}
			case "Disabled":
				if err = vcm.SecureBoot(false); err != nil {
					return err
				}
			}
		case k == "tpm":
			switch v {
			case "Enabled":
				if err = vcm.TPM(true); err != nil {
					return err
				}
			case "Disabled":
				if err = vcm.TPM(false); err != nil {
					return err
				}
			}
		case k == "smt":
			switch v {
			case "Enabled":
				if err = vcm.SMT(true); err != nil {
					return err
				}
			case "Disabled":
				if err = vcm.SMT(false); err != nil {
					return err
				}
			}
		case k == "sr_iov":
			switch v {
			case "Enabled":
				if err = vcm.SRIOV(true); err != nil {
					return err
				}
			case "Disabled":
				if err = vcm.SRIOV(false); err != nil {
					return err
				}
			}
		case strings.HasPrefix(k, "raw:"):
			// k = raw:Menu1,SubMenu1,SubMenuMenu1,SettingName
			pathStr := strings.TrimPrefix(k, "raw:")
			path := strings.Split(pathStr, ",")
			name := path[len(path)-1]
			path = path[:len(path)-1]

			vcm.Raw(name, v, path)
		}
	}

	xmlData, err := vcm.Marshal()
	if err != nil {
		return err
	}

	return s.SetBiosConfigurationFromFile(ctx, xmlData)
}

func (s *Exec) SetBiosConfigurationFromFile(ctx context.Context, cfg string) (err error) {
	// Open tmp file to hold cfg
	inputConfigTmpFile, err := os.CreateTemp("", "bmclib")
	if err != nil {
		return err
	}

	defer os.Remove(inputConfigTmpFile.Name())

	_, err = inputConfigTmpFile.WriteString(cfg)
	if err != nil {
		return err
	}

	err = inputConfigTmpFile.Close()
	if err != nil {
		return err
	}

	return s.ChangeBiosCfg(ctx, inputConfigTmpFile.Name(), false)
}

// ResetBiosConfiguration reset bios configuration
func (s *Exec) ResetBiosConfiguration(ctx context.Context) (err error) {
	return s.LoadDefaultBiosCfg(ctx)
}
