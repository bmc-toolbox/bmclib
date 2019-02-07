package c7000

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
)

// PowerCycle reboots the chassis
func (c *C7000) PowerCycle() (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run("RESTART OA ACTIVE")
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "Restarting Onboard Administrator") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOn power on the chassis
func (c *C7000) PowerOn() (status bool, err error) {
	return status, fmt.Errorf("Unsupported action")
}

// PowerOff power off the chassis
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

	return position, fmt.Errorf("unable to find the blade in this chassis")
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

// SetDynamicPower configure the dynamic power behaviour
func (c *C7000) SetDynamicPower(enable bool) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	var state string
	if enable {
		state = "ON"
	} else {
		state = "OFF"
	}

	cmd := fmt.Sprintf("SET POWER SAVINGS %s", state)
	output, err := c.sshClient.Run(cmd)
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "Dynamic Power: Disabled") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// GetFirmwareVersion returns the chassis firmware version
func (c *C7000) GetFirmwareVersion() (version string, err error) {
	err = c.sshLogin()
	if err != nil {
		return version, err
	}

	output, err := c.sshClient.Run("SHOW OA INFO")
	if err != nil {
		return version, fmt.Errorf(output)
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Firmware Ver") {
			version = strings.Fields(line)[3]
		}
	}

	return version, err
}

// UpdateFirmware updates the chassis firmware
func (c *C7000) UpdateFirmware(source, file string) (status bool, err error) {
	err = c.sshLogin()
	if err != nil {
		return status, err
	}

	cmd := fmt.Sprintf("update image %s/%s", source, file)
	output, err := c.sshClient.Run(cmd)
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "Restarting Onboard Administrator") {
		return true, err
	}

	return status, err
}

// ModBladeBmcUser modfies BMC Admin user account password through the chassis,
// this method will attempt to modify a user account on all BMCs in a chassis.
func (c *C7000) ModBladeBmcUser(username string, password string) (err error) {

	ribcl := `HPONCFG all  << end_marker
<RIBCL VERSION="2.0">
<LOGIN USER_LOGIN="__USERNAME__" PASSWORD="__PASSWORD__">
 <USER_INFO MODE="write">
  <MOD_USER USER_LOGIN="__USERNAME__">
     <PASSWORD value="__PASSWORD__"/>
</MOD_USER>
 </USER_INFO>
 </LOGIN>
</RIBCL>
end_marker`

	ribcl = strings.Replace(ribcl, "__USERNAME__", username, -1)
	ribcl = strings.Replace(ribcl, "__PASSWORD__", password, -1)

	err = c.sshLogin()
	if err != nil {
		return err
	}

	output, err := c.sshClient.Run(ribcl)
	if err != nil {
		return fmt.Errorf(output)
	}

	//since there are multiple blades and this command
	//could fail on any of the blades because they are un responsive
	//we only validate the command actually ran and not if it succeeded on each blade.
	if !strings.Contains(output, "END RIBCL RESULTS") {
		return fmt.Errorf(output)
	}

	return err
}

// AddBladeBmcAdmin configures BMC Admin user accounts through the chassis.
// this method will attempt to add the user to all BMCs in a chassis.
func (c *C7000) AddBladeBmcAdmin(username string, password string) (err error) {

	ribcl := `HPONCFG all  << end_marker
<RIBCL VERSION="2.0">
<LOGIN USER_LOGIN="__USERNAME__" PASSWORD="__PASSWORD__">
 <USER_INFO MODE="write">
   <ADD_USER
     USER_NAME="__USERNAME__"
     USER_LOGIN="__USERNAME__"
     PASSWORD="__PASSWORD__">
     <ADMIN_PRIV value ="Yes"/>
     <REMOTE_CONS_PRIV value ="Yes"/>
     <RESET_SERVER_PRIV value ="Yes"/>
     <VIRTUAL_MEDIA_PRIV value ="Yes"/>
     <CONFIG_ILO_PRIV value="Yes"/>
   </ADD_USER>
 </USER_INFO>
 </LOGIN>
</RIBCL>
end_marker`

	ribcl = strings.Replace(ribcl, "__USERNAME__", username, -1)
	ribcl = strings.Replace(ribcl, "__PASSWORD__", password, -1)

	err = c.sshLogin()
	if err != nil {
		return err
	}

	output, err := c.sshClient.Run(ribcl)
	if err != nil {
		return fmt.Errorf(output)
	}

	//since there are multiple blades and this command
	//could fail on any of the blades because they are un responsive
	//we only validate the command actually ran and not if it succeeded on each blade.
	if !strings.Contains(output, "END RIBCL RESULTS") {
		return fmt.Errorf(output)
	}

	return err
}

// RemoveBladeBmcUser removes the user account from all BMCs through the chassis.
func (c *C7000) RemoveBladeBmcUser(username string) (err error) {

	ribcl := `HPONCFG all  << end_marker
<RIBCL VERSION="2.0">
<LOGIN USER_LOGIN="__USERNAME__" PASSWORD="">
 <USER_INFO MODE="write">
   <DELETE_USER USER_LOGIN="__USERNAME__" />
 </USER_INFO>
 </LOGIN>
</RIBCL>
end_marker`

	ribcl = strings.Replace(ribcl, "__USERNAME__", username, -1)

	err = c.sshLogin()
	if err != nil {
		return err
	}

	output, err := c.sshClient.Run(ribcl)
	if err != nil {
		return fmt.Errorf(output)
	}

	//since there are multiple blades and this command
	//could fail on any of the blades because they are un responsive
	//we only validate the command actually ran and not if it succeeded on each blade.
	if !strings.Contains(output, "END RIBCL RESULTS") {
		return fmt.Errorf(output)
	}

	return err
}

// SetFlexAddressState Enable/Disable FlexAddress disables flex Addresses for blades
// FlexAddress is a virtual addressing scheme
func (c *C7000) SetFlexAddressState(position int, enable bool) (status bool, err error) {
	return status, errors.ErrNotImplemented
}

// SetIpmiOverLan Enable/Disable IPMI over lan parameter per blade in chassis
func (c *C7000) SetIpmiOverLan(position int, enable bool) (status bool, err error) {
	return status, errors.ErrNotImplemented
}
