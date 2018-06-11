package c7000

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/internal/sshclient"
)

// PowerCycle reboots the machine via bmc
func (c *C7000) PowerCycle() (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run("restart oa active")
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOn power on the machine via bmc
func (c *C7000) PowerOn() (status bool, err error) {
	return status, fmt.Errorf("Unsupported action")
}

// PowerOff power off the machine via bmc
func (c *C7000) PowerOff() (status bool, err error) {
	return status, fmt.Errorf("Unsupported action")
}

// IsOn tells if a machine is currently powered on
func (c *C7000) IsOn() (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	if c.sshClient != nil {
		return true, err
	}
	return status, err
}

// FindBladePosition receives a serial and find the position of the blade using it
func (c *C7000) FindBladePosition(serial string) (position int, err error) {
	err = c.sshLogin()
	if err != nil {
		return position, err
	}

	output, err := c.sshClient.Run("SHOW SERVER NAMES")
	if err != nil {
		return position, err
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.Replace(line, "Server-", "", -1)
		data := strings.FieldsFunc(line, sshclient.IsntLetterOrNumber)
		for _, field := range data {
			if strings.ToLower(serial) == strings.ToLower(field) {
				position, err := strconv.Atoi(data[0])
				return position, err
			}
		}
	}

	return position, fmt.Errorf("Unable to find the blade in this chassis")
}

// PowerCycleBlade reboots the machine via bmc
func (c *C7000) PowerCycleBlade(position int) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run(fmt.Sprintf("REBOOT SERVER %d FORCE", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "currently powered off") {
		return c.PowerOnBlade(position)
	}

	if strings.Contains(output, "Forcing reboot of Blade") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// ReseatBlade reboots the machine via bmc
func (c *C7000) ReseatBlade(position int) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run(fmt.Sprintf("RESET SERVER %d", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "Successfully") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOnBlade power on the machine via bmc
func (c *C7000) PowerOnBlade(position int) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run(fmt.Sprintf("POWERON SERVER %d", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "Powering on") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOffBlade power off the machine via bmc
func (c *C7000) PowerOffBlade(position int) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run(fmt.Sprintf("POWEROFF SERVER %d FORCE", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "powering down.") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// IsOnBlade tells if a machine is currently powered on
func (c *C7000) IsOnBlade(position int) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run(fmt.Sprintf("SHOW SERVER STATUS %d", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "Power: On") {
		return true, err
	}

	return status, err
}

// PowerCycleBmcBlade reboots the bmc we are connected to
func (c *C7000) PowerCycleBmcBlade(position int) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run(fmt.Sprintf("RESET ILO %d", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "Successfully") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PxeOnceBlade makes the machine to boot via pxe once
func (c *C7000) PxeOnceBlade(position int) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	status, err = c.PowerCycleBlade(position)
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run(fmt.Sprintf("SET SERVER BOOT ONCE PXE %d", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "boot order changed to PXE") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// Enable/Disable FlexAddress disables flex Addresses for blades
// FlexAddress is a virtual addressing scheme
func (c *C7000) SetFlexAddressState(position int, enable bool) (status bool, err error) {
	return status, fmt.Errorf("Not implemented.")
}

// Enable/Disable IPMI over lan parameter per blade in chassis
func (c *C7000) SetIpmiOverLan(position int, enable bool) (status bool, err error) {
	return status, fmt.Errorf("Not implemented")
}
