package c7000

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal"
)

// PowerCycle reboots the chassis
func (c *C7000) PowerCycle() (bool, error) {
	output, err := c.sshClient.Run("RESTART OA ACTIVE")
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Restarting Onboard Administrator") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerOn power on the chassis
func (c *C7000) PowerOn() (bool, error) {
	return false, errors.ErrFeatureUnavailable
}

// PowerOff power off the chassis
func (c *C7000) PowerOff() (bool, error) {
	return false, errors.ErrFeatureUnavailable
}

// IsOn tells if a machine is currently powered on
func (c *C7000) IsOn() (bool, error) {
	if c.sshClient != nil { // TODO: run "help"?
		return true, nil
	}
	return false, nil
}

// FindBladePosition receives a serial and find the position of the blade using it
func (c *C7000) FindBladePosition(serial string) (int, error) {
	output, err := c.sshClient.Run("SHOW SERVER NAMES")
	if err != nil {
		return -1, err
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.Replace(line, "Server-", "", -1)
		data := strings.FieldsFunc(line, internal.IsntLetterOrNumber)
		for _, field := range data {
			if strings.EqualFold(serial, field) {
				return strconv.Atoi(data[0])
			}
		}
	}

	return -1, fmt.Errorf("unable to find the blade in this chassis")
}

// PowerCycleBlade reboots the machine via bmc
func (c *C7000) PowerCycleBlade(position int) (bool, error) {
	output, err := c.sshClient.Run(fmt.Sprintf("REBOOT SERVER %d FORCE", position))
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "currently powered off") {
		return c.PowerOnBlade(position)
	}

	if strings.Contains(output, "Forcing reboot of Blade") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// ReseatBlade reboots the machine via bmc
func (c *C7000) ReseatBlade(position int) (bool, error) {
	output, err := c.sshClient.Run(fmt.Sprintf("RESET SERVER %d", position))
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Successfully") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerOnBlade power on the machine via bmc
func (c *C7000) PowerOnBlade(position int) (bool, error) {
	output, err := c.sshClient.Run(fmt.Sprintf("POWERON SERVER %d", position))
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Powering on") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerOffBlade power off the machine via bmc
func (c *C7000) PowerOffBlade(position int) (bool, error) {
	output, err := c.sshClient.Run(fmt.Sprintf("POWEROFF SERVER %d FORCE", position))
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}
	if strings.Contains(output, "powering down.") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// IsOnBlade tells if a machine is currently powered on
func (c *C7000) IsOnBlade(position int) (bool, error) {
	output, err := c.sshClient.Run(fmt.Sprintf("SHOW SERVER STATUS %d", position))
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Power: On") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PowerCycleBmcBlade reboots the bmc we are connected to
func (c *C7000) PowerCycleBmcBlade(position int) (bool, error) {
	output, err := c.sshClient.Run(fmt.Sprintf("RESET ILO %d", position))
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Successfully") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// PxeOnceBlade makes the machine to boot via pxe once
func (c *C7000) PxeOnceBlade(position int) (bool, error) {
	status, err := c.PowerCycleBlade(position)
	if err != nil {
		return status, err
	}

	output, err := c.sshClient.Run(fmt.Sprintf("SET SERVER BOOT ONCE PXE %d", position))
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "boot order changed to PXE") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// SetDynamicPower configure the dynamic power behavior
func (c *C7000) SetDynamicPower(enable bool) (bool, error) {
	var state string
	if enable {
		state = "ON"
	} else {
		state = "OFF"
	}

	cmd := fmt.Sprintf("SET POWER SAVINGS %s", state)
	output, err := c.sshClient.Run(cmd)
	if err != nil {
		return false, fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Dynamic Power: Disabled") {
		return true, nil
	}

	return false, fmt.Errorf(output)
}

// UpdateFirmware updates the chassis firmware
func (c *C7000) UpdateFirmware(source, file string) (bool, string, error) {
	cmd := fmt.Sprintf("update image %s/%s", source, file)
	output, err := c.sshClient.Run(cmd)
	if err != nil {
		return false, "", fmt.Errorf("output: %q: %w", output, err)
	}

	if strings.Contains(output, "Flashing Active Onboard Administrator") {
		return true, output, nil
	}

	return false, output, fmt.Errorf(output)
}

func (c *C7000) CheckFirmwareVersion() (version string, err error) {
	return "", fmt.Errorf("not supported yet")
}

// ModBladeBmcUser modfies BMC Admin user account password through the chassis,
// this method will attempt to modify a user account on all BMCs in a chassis.
func (c *C7000) ModBladeBmcUser(username string, password string) error {
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

	output, err := c.sshClient.Run(ribcl)
	if err != nil {
		return fmt.Errorf("output: %q: %w", output, err)
	}

	//since there are multiple blades and this command
	//could fail on any of the blades because they are un responsive
	//we only validate the command actually ran and not if it succeeded on each blade.
	if !strings.Contains(output, "END RIBCL RESULTS") {
		return fmt.Errorf(output)
	}

	return nil
}

// AddBladeBmcAdmin configures BMC Admin user accounts through the chassis.
// this method will attempt to add the user to all BMCs in a chassis.
func (c *C7000) AddBladeBmcAdmin(username string, password string) error {
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

	output, err := c.sshClient.Run(ribcl)
	if err != nil {
		return fmt.Errorf("output: %q: %w", output, err)
	}

	//since there are multiple blades and this command
	//could fail on any of the blades because they are un responsive
	//we only validate the command actually ran and not if it succeeded on each blade.
	if !strings.Contains(output, "END RIBCL RESULTS") {
		return fmt.Errorf(output)
	}

	return nil
}

// RemoveBladeBmcUser removes the user account from all BMCs through the chassis.
func (c *C7000) RemoveBladeBmcUser(username string) error {
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

	output, err := c.sshClient.Run(ribcl)
	if err != nil {
		return fmt.Errorf("output: %q: %w", output, err)
	}

	//since there are multiple blades and this command
	//could fail on any of the blades because they are un responsive
	//we only validate the command actually ran and not if it succeeded on each blade.
	if !strings.Contains(output, "END RIBCL RESULTS") {
		return fmt.Errorf(output)
	}

	return nil
}

// SetFlexAddressState Enable/Disable FlexAddress disables flex Addresses for blades
// FlexAddress is a virtual addressing scheme
func (c *C7000) SetFlexAddressState(_ int, _ bool) (bool, error) {
	return false, errors.ErrNotImplemented
}

// SetIpmiOverLan Enable/Disable IPMI over lan parameter per blade in chassis
func (c *C7000) SetIpmiOverLan(_ int, _ bool) (bool, error) {
	return false, errors.ErrNotImplemented
}
