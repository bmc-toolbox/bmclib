package ipmi

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// Ipmi holds the date for an ipmi connection
type Ipmi struct {
	Username string
	Password string
	Host     string
	ipmitool string
}

// New returns a new ipmi instance
func New(username string, password string, host string) (ipmi *Ipmi, err error) {
	ipmi = &Ipmi{
		Username: username,
		Password: password,
		Host:     host,
	}

	ipmi.ipmitool, err = ipmi.findBin("ipmitool")
	if err != nil {
		return nil, err
	}

	return ipmi, err
}

func (i *Ipmi) run(ctx context.Context, command []string) (output string, err error) {
	ipmiArgs := []string{"-I", "lanplus", "-U", i.Username, "-E", "-N", "5", "-H", i.Host}
	ipmiArgs = append(ipmiArgs, command...)
	cmd := exec.CommandContext(ctx, i.ipmitool, ipmiArgs...)
	cmd.Env = []string{fmt.Sprintf("IPMITOOL_PASSWORD=%s", i.Password)}
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), ctx.Err()
	}
	return string(out), err
}

func (i *Ipmi) findBin(binary string) (binaryPath string, err error) {
	path, err := exec.LookPath(binary)
	if err == nil {
		return path, nil
	}
	locations := []string{"/bin", "/sbin", "/usr/bin", "/usr/sbin", "/usr/local/sbin"}

	for _, path := range locations {
		lookup := path + "/" + binary
		fileInfo, err := os.Stat(path + "/" + binary)

		if err != nil {
			continue
		}

		if !fileInfo.IsDir() {
			return lookup, nil
		}
	}

	return binaryPath, fmt.Errorf("Unable to find binary: %v", binary)
}

// PowerCycle reboots the machine via bmc
func (i *Ipmi) PowerCycle() (status bool, err error) {
	output, err := i.run(context.Background(), []string{"chassis", "power", "reset"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: Reset") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerCycleBmc reboots the bmc we are connected to
func (i *Ipmi) PowerCycleBmc() (status bool, err error) {
	output, err := i.run(context.Background(), []string{"mc", "reset", "cold"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Sent cold reset command to MC") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerOn power on the machine via bmc
func (i *Ipmi) PowerOn() (status bool, err error) {
	ctx := context.Background()
	s, err := i.IsOn(ctx)
	if err != nil {
		return false, err
	}

	if s {
		return true, nil
	}

	output, err := i.run(ctx, []string{"chassis", "power", "on"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: Up/On") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerOnForce power on the machine via bmc even when the machine is already on (Thanks HP!)
func (i *Ipmi) PowerOnForce() (status bool, err error) {
	output, err := i.run(context.Background(), []string{"chassis", "power", "on"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: Up/On") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerOff power off the machine via bmc
func (i *Ipmi) PowerOff() (status bool, err error) {
	ctx := context.Background()
	s, err := i.IsOn(ctx)
	if err != nil {
		return false, err
	}

	if !s {
		return true, nil
	}

	output, err := i.run(ctx, []string{"chassis", "power", "off"})
	if strings.Contains(output, "Chassis Power Control: Down/Off") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PxeOnceEfi makes the machine to boot via pxe once using EFI
func (i *Ipmi) PxeOnceEfi() (status bool, err error) {
	output, err := i.run(context.Background(), []string{"chassis", "bootdev", "pxe", "options=efiboot"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// BootDevice sets a next boot device with options
func (i *Ipmi) BootDevice(ctx context.Context, device string, options []string) (bool, error) {
	output, err := i.run(context.Background(), []string{"chassis", "bootdev", device, fmt.Sprintf("options=%v", strings.Join(options, ","))})
	if err != nil {
		return false, errors.Wrapf(err, "output: %v", output)
	}

	if !strings.Contains(output, fmt.Sprintf("Set Boot Device to %v", device)) {
		return false, errors.New(fmt.Sprintf("unexpected output: %v", output))
	}
	return true, nil
}

// PxeOnceMbr makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnceMbr() (status bool, err error) {
	output, err := i.run(context.Background(), []string{"chassis", "bootdev", "pxe"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PxeOnce makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnce() (status bool, err error) {
	return i.PxeOnceMbr()
}

// IsOn tells if a machine is currently powered on
func (i *Ipmi) IsOn(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "status"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.Contains(output, "Chassis Power is on") {
		return true, err
	}
	return false, err
}

// CreateUser via ipmitool
func (i *Ipmi) CreateUser(ctx context.Context) (status bool, err error) {
	return false, errors.New("not implemented")
}

// Info about BMC
func (i *Ipmi) Info(ctx context.Context) error {
	output, err := i.run(ctx, []string{"bmc", "info"})
	if err != nil {
		return errors.Wrap(err, "error getting bmc info")
	}

	if !strings.Contains(output, "IPMI Version              : 2.0") {
		return errors.New(fmt.Sprintf("unexpected output: %v", output))
	}
	return nil
}

// ReadUsers list all BMC users
func (i *Ipmi) ReadUsers(ctx context.Context) (users []map[string]string, err error) {
	output, err := i.run(ctx, []string{"user", "list"})
	if err != nil {
		return users, errors.Wrap(err, "error getting user list")
	}
	header := map[int]string{}
	firstLine := true
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.Fields(scanner.Text())
		if firstLine {

			firstLine = false
			for x := 0; x < 5; x++ {
				header[x] = line[x]
			}
			continue
		}
		entry := map[string]string{}
		if line[1] != "true" {
			for x := 0; x < 5; x++ {
				entry[header[x]] = line[x]

			}
			users = append(users, entry)
		}
	}

	return users, err
}
