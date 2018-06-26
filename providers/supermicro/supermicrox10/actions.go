package supermicrox10

import (
	"github.com/bmc-toolbox/bmclib/internal/ipmi"
)

// PowerCycle reboots the machine via bmc
func (s *SupermicroX10) PowerCycle() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PowerCycle()
	return status, err
}

// PowerCycleBmc reboots the bmc we are connected to
func (s *SupermicroX10) PowerCycleBmc() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PowerCycleBmc()
	return status, err
}

// PowerOn power on the machine via bmc
func (s *SupermicroX10) PowerOn() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PowerOn()
	return status, err
}

// PowerOff power off the machine via bmc
func (s *SupermicroX10) PowerOff() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PowerOff()
	return status, err
}

// PxeOnce makes the machine to boot via pxe once
func (s *SupermicroX10) PxeOnce() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PxeOnceMbr()
	if err != nil {
		return false, err
	}
	return i.PowerCycle()
}

// IsOn tells if a machine is currently powered on
func (s *SupermicroX10) IsOn() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.IsOn()
	return status, err
}
