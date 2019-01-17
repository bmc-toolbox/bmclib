package m1000e

import (
	"bufio"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
)

// PowerCycle reboots the chassis
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

// PowerOn power on the chassis
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

// PowerOff power off the chassis
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

// SetIpmiOverLan Enable/Disable IPMI over lan parameter per blade in chassis
func (m *M1000e) SetIpmiOverLan(position int, enable bool) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	var state int
	if enable {
		state = 1
	} else {
		state = 0
	}

	cmd := fmt.Sprintf("config -g cfgServerInfo -o cfgServerIPMIOverLanEnable -i %d %d", position, state)
	output, err := m.sshClient.Run(cmd)
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)

}

// SetDynamicPower Enable/Disable Dynamic Power - Dynamic Power Supply Engagement (DPSE) in Dell jargon.
// Dynamic Power Supply Engagement (DPSE) mode is disabled by default.
// DPSE saves power by optimizing the power efficiency of the PSUs supplying power to the chassis.
// This also increases the PSU life, and reduces heat generation.
func (m *M1000e) SetDynamicPower(enable bool) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	var state int
	if enable {
		state = 1
	} else {
		state = 0
	}

	cmd := fmt.Sprintf("config -g cfgChassisPower -o cfgChassisDynamicPSUEngagementEnable %d", state)
	output, err := m.sshClient.Run(cmd)
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)

}

// SetFlexAddressState Disable/Enable FlexAddress disables flex Addresses for blades
// FlexAddress is a virtual addressing scheme
func (m *M1000e) SetFlexAddressState(position int, enable bool) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	isOn, err := m.IsOnBlade(position)
	if err != nil {
		return false, fmt.Errorf("failed to validate blade %d power status is off, %v", position, err)
	}

	if isOn {
		return false, fmt.Errorf("blade in position %d is currently powered on, it must be powered off before this action", position)
	}

	var cmd string
	if enable {
		cmd = fmt.Sprintf("racadm setflexaddr -i %d 1", position)
	} else {
		cmd = fmt.Sprintf("racadm setflexaddr -i %d 0", position)
	}

	output, err := m.sshClient.Run(cmd)
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "successful") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// GetFirmwareVersion returns the chassis firmware version
func (m *M1000e) GetFirmwareVersion() (version string, err error) {
	err = m.sshLogin()
	if err != nil {
		return version, err
	}

	output, err := m.sshClient.Run("getversion")
	if err != nil {
		return version, fmt.Errorf(output)
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "cmc-1") {
			version = strings.Fields(line)[1]
		}
	}

	return version, err
}

// UpdateFirmware updates the chassis firmware
func (m *M1000e) UpdateFirmware(source, file string) (status bool, err error) {
	err = m.sshLogin()
	if err != nil {
		return status, err
	}

	u, err := url.Parse(source)
	if err != nil {
		return status, err
	}

	password, ok := u.User.Password()
	if !ok {
		password = "anonymous"
	}

	cmd := fmt.Sprintf("fwupdate -f %s %s %s -d %s -m cmc-active -m cmc-standby", u.Host, u.User.Username(), password, u.Path)
	output, err := m.sshClient.Run(cmd)
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "Firmware update has been initiated") {
		return true, err
	}

	return status, err
}

// UpdateFirmwareBmcBlade updates the blade BMC firmware
func (m *M1000e) UpdateFirmwareBmcBlade(position int, host, filepath string) (status bool, err error) {
	// iDRAC 7 or later is not supported by fwupdate on the M1000e
	return status, errors.ErrNotImplemented
}

// AddBladeBmcAdmin adds BMC Admin user accounts through the chassis.
func (m *M1000e) AddBladeBmcAdmin(username string, password string) (err error) {
	return errors.ErrNotImplemented
}

// RemoveBladeBmcUser removes BMC Admin user accounts through the chassis.
func (m *M1000e) RemoveBladeBmcUser(username string) (err error) {
	return errors.ErrNotImplemented
}

// ModBladeBmcUser modifies a BMC Admin user password through the chassis.
func (m *M1000e) ModBladeBmcUser(username string, password string) (err error) {
	return errors.ErrNotImplemented
}
