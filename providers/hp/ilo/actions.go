package ilo

import (
	"fmt"
	"strings"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
)

// PowerCycle reboots the machine via bmc
func (i *Ilo) PowerCycle() (status bool, err error) {
	output, err := i.sshClient.Run("power reset")
	if err != nil {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "Server power off") {
		return i.PowerOn()
	}

	if strings.Contains(output, "Server resetting") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerCycleBmc reboots the bmc we are connected to
func (i *Ilo) PowerCycleBmc() (status bool, err error) {
	output, err := i.sshClient.Run("reset /map1")
	if err != nil && !strings.Contains(output, "Resetting iLO") {
		return false, fmt.Errorf(output)
	}

	if strings.Contains(output, "Resetting iLO") {
		return true, nil
	}

	return status, fmt.Errorf(output)
}

// PowerOn power on the machine via bmc
func (i *Ilo) PowerOn() (status bool, err error) {
	output, err := i.sshClient.Run("power on")
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "Server powering on") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PowerOff power off the machine via bmc
func (i *Ilo) PowerOff() (status bool, err error) {
	output, err := i.sshClient.Run("power off hard")
	if err != nil {
		return false, fmt.Errorf(output)
	}
	if strings.Contains(output, "Forcing server") {
		return true, err
	}

	return status, fmt.Errorf(output)
}

// PxeOnce makes the machine to boot via pxe once
func (i *Ilo) PxeOnce() (status bool, err error) {
	im, err := ipmi.New(i.username, i.password, i.ip)
	if err != nil {
		return status, err
	}
	// Just to be clear. I didn't choose to do this
	// HP is not reliable to boot a machine and pxe (HAHAHAHA)
	// so this is the only usable solution.
	status, err = im.PxeOnceEfi()
	if err != nil {
		return false, err
	}
	status, err = im.PowerCycle()
	if err != nil {
		return false, err
	}
	im.PowerOnForce()
	return status, err
}

// IsOn tells if a machine is currently powered on
func (i *Ilo) IsOn() (status bool, err error) {
	output, err := i.sshClient.Run("power")
	if err != nil {
		return false, fmt.Errorf("%v: %v", err, output)
	}
	if strings.Contains(output, "currently: On") {
		return true, err
	}
	return status, fmt.Errorf(output)
}
