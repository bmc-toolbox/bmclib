package ipmi

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
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

func (i *Ipmi) run(command []string) (output string, err error) {
	ipmiArgs := []string{"-I", "lanplus", "-U", i.Username, "-E", "-H", i.Host}
	ipmiArgs = append(ipmiArgs, command...)
	cmd := exec.Command(i.ipmitool, ipmiArgs...)
	cmd.Env = []string{fmt.Sprintf("IPMITOOL_PASSWORD=%s", i.Password)}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (i *Ipmi) findBin(binary string) (binaryPath string, err error) {
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
	output, err := i.run([]string{"chassis", "power", "reset"})
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
	output, err := i.run([]string{"mc", "reset", "cold"})
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
	s, err := i.IsOn()
	if err != nil {
		return false, err
	}

	if s == true {
		return false, fmt.Errorf("server is already on")
	}

	output, err := i.run([]string{"chassis", "power", "on"})
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
	output, err := i.run([]string{"chassis", "power", "on"})
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
	s, err := i.IsOn()
	if err != nil {
		return false, err
	}

	if s == false {
		return false, fmt.Errorf("server is already off")
	}

	output, err := i.run([]string{"chassis", "power", "off"})
	if strings.Contains(output, "Chassis Power Control: Down/Off") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PxeOnceEfi makes the machine to boot via pxe once using EFI
func (i *Ipmi) PxeOnceEfi() (status bool, err error) {
	output, err := i.run([]string{"chassis", "bootdev", "pxe", "options=efiboot"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.Contains(output, "Set Boot Device to pxe") {
		return true, err
	}
	return false, fmt.Errorf("%v: %v", err, output)
}

// PxeOnceMbr makes the machine to boot via pxe once using MBR
func (i *Ipmi) PxeOnceMbr() (status bool, err error) {
	output, err := i.run([]string{"chassis", "bootdev", "pxe"})
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
func (i *Ipmi) IsOn() (status bool, err error) {
	output, err := i.run([]string{"chassis", "power", "status"})
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}

	if strings.Contains(output, "Chassis Power is on") {
		return true, err
	}
	return false, err
}
