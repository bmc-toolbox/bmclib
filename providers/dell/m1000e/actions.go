package m1000e

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/internal/sshclient"
)

// PowerCycle reboots the machine via bmc
func (m *M1000e) PowerCycle() (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run("racadm racreset")
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOn power on the machine via bmc
func (m *M1000e) PowerOn() (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run("chassisaction powerup")
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOff power off the machine via bmc
func (m *M1000e) PowerOff() (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run("chassisaction powerdown")
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// IsOn tells if a machine is currently powered on
func (m *M1000e) IsOn() (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run("getsysinfo")
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, " = ON") {
		return true, err
	}

	return status, err
}

// FindBladePosition receives a serial and find the position of the blade using it
func (m *M1000e) FindBladePosition(serial string) (position int, err error) {
	err = m.sshLogin()
	if err != nil {
		return position, err
	}

	output, err := m.sshClient.Run("getsvctag")
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
func (m *M1000e) PowerCycleBlade(position int) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run(fmt.Sprintf("serveraction -m server-%d hardreset", position))
	if err != nil {
		if strings.Contains(output, "is already powered OFF") {
			return m.PowerOnBlade(position)
		}
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// ReseatBlade reboots the machine via bmc
func (m *M1000e) ReseatBlade(position int) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run(fmt.Sprintf("serveraction -m server-%d reseat -f", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOnBlade power on the machine via bmc
func (m *M1000e) PowerOnBlade(position int) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run(fmt.Sprintf("serveraction -m server-%d powerup", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOffBlade power off the machine via bmc
func (m *M1000e) PowerOffBlade(position int) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run(fmt.Sprintf("serveraction -m server-%d powerdown", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// IsOnBlade tells if a machine is currently powered on
func (m *M1000e) IsOnBlade(position int) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run(fmt.Sprintf("serveraction -m server-%d powerstatus", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "ON") {
		return true, err
	}

	return status, err
}

// PowerCycleBmcBlade reboots the bmc we are connected to
func (m *M1000e) PowerCycleBmcBlade(position int) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run(fmt.Sprintf("racreset -m server-%d", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PxeOnceBlade makes the machine to boot via pxe once
func (m *M1000e) PxeOnceBlade(position int) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	status, err = m.PowerCycleBlade(position)
	if err != nil {
		return status, err
	}

	output, err := m.sshClient.Run(fmt.Sprintf("deploy -m server-%d -b PXE -o yes", position))
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}
