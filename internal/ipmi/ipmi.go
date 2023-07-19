package ipmi

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// Ipmi holds the date for an ipmi connection
type Ipmi struct {
	Username    string
	Password    string
	Host        string
	ipmitool    string
	cipherSuite string
	log         logr.Logger
}

// Option for setting optional Ipmi values
type Option func(*Ipmi)

func WithIpmitoolPath(path string) Option {
	return func(i *Ipmi) {
		i.ipmitool = path
	}
}

func WithCipherSuite(cipherSuite string) Option {
	return func(i *Ipmi) {
		i.cipherSuite = cipherSuite
	}
}

func WithLogger(log logr.Logger) Option {
	return func(i *Ipmi) {
		i.log = log
	}
}

// New returns a new ipmi instance
func New(username string, password string, host string, opts ...Option) (ipmi *Ipmi, err error) {
	ipmi = &Ipmi{
		Username: username,
		Password: password,
		Host:     host,
		log:      logr.Discard(),
	}

	for _, opt := range opts {
		opt(ipmi)
	}

	if ipmi.ipmitool == "" {
		ipmi.ipmitool, err = exec.LookPath("ipmitool")
		if err != nil {
			return nil, err
		}
	} else {
		if _, err = os.Stat(ipmi.ipmitool); err != nil {
			return nil, err
		}
	}

	return ipmi, err
}

func (i *Ipmi) run(ctx context.Context, command []string, stdin ...byte) (output string, err error) {
	var out []byte
	var ipmiCiphers = []string{"3", "17"}
	ipmiArgs := []string{"-I", "lanplus", "-U", i.Username, "-E", "-N", "5"}

	if strings.Contains(i.Host, ":") {
		host, port, err := net.SplitHostPort(i.Host)
		if err == nil {
			ipmiArgs = append(ipmiArgs, "-H", host, "-p", port)
		}
	} else {
		ipmiArgs = append(ipmiArgs, "-H", i.Host)
	}

	if i.cipherSuite != "" {
		ipmiCiphers = []string{i.cipherSuite}
	}
	for _, cipherString := range ipmiCiphers {
		ipmiCmd := append(ipmiArgs, "-C", cipherString)
		i.log.V(3).Info("ipmitool options", "opts", formatOptions(ipmiCmd))
		ipmiCmd = append(ipmiCmd, command...)
		cmd := exec.CommandContext(ctx, i.ipmitool, ipmiCmd...)
		if stdin != nil {
			stdinPipe, err := cmd.StdinPipe()
			if err != nil {
				return "", err
			}
			stdinPipe.Write(stdin)
		}
		cmd.Env = []string{fmt.Sprintf("IPMITOOL_PASSWORD=%s", i.Password)}
		out, err = cmd.CombinedOutput()
		if err == nil || ctx.Err() != nil {
			break
		}
	}

	if ctx.Err() == context.DeadlineExceeded {
		return string(out), ctx.Err()
	}

	return string(out), errors.Wrap(err, strings.TrimSpace(string(out)))
}

type cmdOpt struct {
	Opt string `json:"opt"`
	Val string `json:"val"`
}

func formatOptions(opts []string) []cmdOpt {
	result := []cmdOpt{}
	for _, opt := range opts {
		if strings.HasPrefix(opt, "-") {
			o := cmdOpt{Opt: opt}
			if opt == "-E" {
				o.Val = "-E"
			}
			result = append(result, o)
		} else {
			result[len(result)-1].Val = opt
		}
	}

	return result
}

// PowerCycle reboots the machine via bmc
func (i *Ipmi) PowerCycle(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "cycle"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: Cycle") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// ForceRestart does the chassis power cycle even if the chassis is turned off.
// From the RedFish spec (https://www.dmtf.org/sites/default/files/standards/documents/DSP2046_2018.1.pdf):
//
//	Perform an immediate (non-graceful) shutdown, followed by a restart.
func (i *Ipmi) ForceRestart(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "status"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	command := "on"
	reply := "Up/On"
	if strings.HasPrefix(output, "Chassis Power is on") {
		command = "cycle"
		reply = "Cycle"
	} else if !strings.HasPrefix(output, "Chassis Power is off") {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	output, err = i.run(ctx, []string{"chassis", "power", command})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: "+reply) {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerReset reboots the machine via bmc
func (i *Ipmi) PowerReset(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "reset"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if !strings.HasPrefix(output, "Chassis Power Control: Reset") {
		return false, fmt.Errorf("%v: %v", err, output)
	}
	return true, err
}

// PowerCycleBmc reboots the bmc we are connected to
func (i *Ipmi) PowerCycleBmc(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"mc", "reset", "cold"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Sent cold reset command to MC") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerResetBmc reboots the bmc we are connected to
func (i *Ipmi) PowerResetBmc(ctx context.Context, resetType string) (ok bool, err error) {
	output, err := i.run(ctx, []string{"mc", "reset", strings.ToLower(resetType)})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, fmt.Sprintf("Sent %v reset command to MC", strings.ToLower(resetType))) {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerOn power on the machine via bmc
func (i *Ipmi) PowerOn(ctx context.Context) (status bool, err error) {
	s, err := i.IsOn(ctx)
	if err != nil {
		return false, errors.Wrap(err, "error checking power state")
	}
	if s {
		return true, nil
	}

	output, err := i.run(ctx, []string{"chassis", "power", "on"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: Up/On") {
		return true, nil
	}
	return false, fmt.Errorf("stderr/stdout: %v", output)
}

// PowerOnForce power on the machine via bmc even when the machine is already on (Thanks HP!)
func (i *Ipmi) PowerOnForce(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "on"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.HasPrefix(output, "Chassis Power Control: Up/On") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerOff power off the machine via bmc
func (i *Ipmi) PowerOff(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "power", "off"})
	if strings.Contains(output, "Chassis Power Control: Down/Off") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PowerSoft power off the machine via bmc
func (i *Ipmi) PowerSoft(ctx context.Context) (status bool, err error) {
	on, _ := i.IsOn(ctx)
	if !on {
		return true, nil
	}

	output, err := i.run(ctx, []string{"chassis", "power", "soft"})
	if !strings.Contains(output, "Chassis Power Control: Soft") {
		return false, fmt.Errorf("%v: %v", err, output)
	}
	return true, err
}

// PxeOnceEfi makes the machine to boot via pxe once using EFI
func (i *Ipmi) PxeOnceEfi(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "bootdev", "pxe", "options=efiboot"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// BootDeviceSet sets the next boot device with options
func (i *Ipmi) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	var atLeastOneOptionSelected bool
	ipmiCmd := []string{"chassis", "bootdev", strings.ToLower(bootDevice)}
	var opts []string
	if setPersistent {
		opts = append(opts, "persistent")
		atLeastOneOptionSelected = true
	}
	if efiBoot {
		opts = append(opts, "efiboot")
		atLeastOneOptionSelected = true
	}
	if atLeastOneOptionSelected {
		optsJoined := strings.Join(opts, ",")
		optsFull := fmt.Sprintf("options=%v", optsJoined)
		ipmiCmd = append(ipmiCmd, optsFull)
	}

	output, err := i.run(ctx, ipmiCmd)
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.Contains(output, fmt.Sprintf("Set Boot Device to %v", strings.ToLower(bootDevice))) {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PxeOnceMbr makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnceMbr(ctx context.Context) (status bool, err error) {
	output, err := i.run(ctx, []string{"chassis", "bootdev", "pxe"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PxeOnce makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnce(ctx context.Context) (status bool, err error) {
	return i.PxeOnceMbr(ctx)
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

// PowerState returns the current power state of the machine
func (i *Ipmi) PowerState(ctx context.Context) (state string, err error) {
	return i.run(ctx, []string{"chassis", "power", "status"})
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
