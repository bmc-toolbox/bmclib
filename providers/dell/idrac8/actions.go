package idrac8

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// PowerCycle reboots the machine via bmc
func (i *IDrac8) PowerCycle() (bool, error) {
	output, err := i.sshClient.Run("racadm serveraction hardreset")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "successful") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerCycleBmc reboots the bmc we are connected to
func (i *IDrac8) PowerCycleBmc() (bool, error) {
	output, err := i.sshClient.Run("racadm racreset hard")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "initiated successfully") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerOn power on the machine via bmc
func (i *IDrac8) PowerOn() (bool, error) {
	output, err := i.sshClient.Run("racadm serveraction powerup")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "successful") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerOff power off the machine via bmc
func (i *IDrac8) PowerOff() (bool, error) {
	output, err := i.sshClient.Run("racadm serveraction powerdown")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "successful") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PxeOnce makes the machine to boot via pxe once
func (i *IDrac8) PxeOnce() (bool, error) {
	output, err := i.sshClient.Run("racadm config -g cfgServerInfo -o cfgServerBootOnce 1")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "successful") {
		output, err = i.sshClient.Run("racadm config -g cfgServerInfo -o cfgServerFirstBootDevice PXE")
		if err != nil {
			return false, fmt.Errorf("output: %q: %w", output, err)
		}

		if strings.Contains(output, "successful") {
			return i.PowerCycle()
		}
	}

	return false, fmt.Errorf(output)
}

// IsOn tells if a machine is currently powered on
func (i *IDrac8) IsOn() (bool, error) {
	output, err := i.sshClient.Run("racadm serveraction powerstatus")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Server power status: ON") {
		return true, nil
	}

	if strings.Contains(output, "Server power status: OFF") {
		return false, nil
	}

	return false, fmt.Errorf(output)
}

// UpdateFirmware updates the bmc firmware
func (i *IDrac8) UpdateFirmware(source, file string) (bool, string, error) {
	_, err := url.Parse(source)
	if err != nil {
		return false, "", err
	}

	cmd := fmt.Sprintf("racadm fwupdate -g -u -a %s -d %s", source, file)
	output, err := i.sshClient.Run(cmd)
	if err != nil {
		return false, output, fmt.Errorf("unexpected output: %q, error: %w", output, err)
	}

	if strings.Contains(output, "Firmware update completed successfully") {
		return true, output, nil
	}

	return false, output, fmt.Errorf("unexpected output: %q", output)
}

func (i *IDrac8) CheckFirmwareVersion() (version string, err error) {
	output, err := i.sshClient.Run("racadm getversion -f idrac")
	if err != nil {
		return "", fmt.Errorf("output: %q: %w", output, err)
	}

	re := regexp.MustCompile(`.*iDRAC Version.*= ((\d+\.)+\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("unexpected output: %q", output)
}
