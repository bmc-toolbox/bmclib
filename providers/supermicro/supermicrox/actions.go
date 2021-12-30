package supermicrox

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
)

// PowerCycle reboots the machine via bmc
func (s *SupermicroX) PowerCycle() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PowerCycle(context.Background())
	return status, err
}

// PowerCycleBmc reboots the bmc we are connected to
func (s *SupermicroX) PowerCycleBmc() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PowerCycleBmc(context.Background())
	return status, err
}

// PowerOn power on the machine via bmc
func (s *SupermicroX) PowerOn() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PowerOn(context.Background())
	return status, err
}

// PowerOff power off the machine via bmc
func (s *SupermicroX) PowerOff() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.PowerOff(context.Background())
	return status, err
}

// PxeOnce makes the machine to boot via pxe once
func (s *SupermicroX) PxeOnce() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	_, err = i.PxeOnceEfi(context.Background())
	if err != nil {
		return false, err
	}
	return i.ForceRestart(context.Background())
}

// IsOn tells if a machine is currently powered on
func (s *SupermicroX) IsOn() (status bool, err error) {
	i, err := ipmi.New(s.username, s.password, s.ip)
	if err != nil {
		return status, err
	}
	status, err = i.IsOn(context.Background())
	return status, err
}

// UpdateFirmware updates the bmc firmware
func (s *SupermicroX) UpdateFirmware(source, file string) (status bool, output string, err error) {
	return false, "Not yet implemented", fmt.Errorf("not yet implemented")
}

func (s *SupermicroX) CheckFirmwareVersion() (version string, err error) {
	return "Not yet implemented", fmt.Errorf("not yet implemented")
}
