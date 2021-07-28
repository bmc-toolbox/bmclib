package ibmc

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
)

// The ibmc model is part of the dummy vendor,
// for lack of a better name and since most annoying devices begin with "i"
// ibmc only exists to make testing easier.

const (
	// BMCType defines the bmc model that is supported by this package
	BMCType = "ibmc"
)

// Ibmc holds the status and properties of a connection to an iDrac device
type Ibmc struct {
	ip       string
	username string
	password string
}

// New returns a new ibmc ready to be used
func New(ip string, username string, password string) (i *Ibmc, err error) {
	return &Ibmc{ip: ip, username: username, password: password}, err
}

// ApplyCfg implements the Bmc interface
func (i *Ibmc) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return nil
}

// BiosVersion implements the Bmc interface
func (i *Ibmc) BiosVersion() (string, error) {
	return "", nil
}

// HardwareType implements the Bmc interface
func (i *Ibmc) HardwareType() string {
	return ""
}

// Version implements the Bmc interface
func (i *Ibmc) Version() (string, error) {
	return "", nil
}

// CPU implements the Bmc interface
func (i *Ibmc) CPU() (string, int, int, int, error) {
	return "", 0, 0, 0, nil
}

// CheckCredentials implements the Bmc interface
func (i *Ibmc) CheckCredentials() error {
	return nil
}

// Disks implements the Bmc interface
func (i *Ibmc) Disks() ([]*devices.Disk, error) {
	return make([]*devices.Disk, 0), nil
}

// IsBlade implements the Bmc interface
func (i *Ibmc) IsBlade() (bool, error) {
	return false, nil
}

// License implements the Bmc interface
func (i *Ibmc) License() (string, string, error) {
	return "", "", nil
}

// Close implements the Bmc interface
func (i *Ibmc) Close(ctx context.Context) error {
	return nil
}

// Memory implements the Bmc interface
func (i *Ibmc) Memory() (int, error) {
	return 0, nil
}

// Model implements the Bmc interface
func (i *Ibmc) Model() (string, error) {
	return "", nil
}

// Name implements the Bmc interface
func (i *Ibmc) Name() (string, error) {
	return "", nil
}

// Nics implements the Bmc interface
func (i *Ibmc) Nics() ([]*devices.Nic, error) {
	return make([]*devices.Nic, 0), nil
}

// PowerKw implements the Bmc interface
func (i *Ibmc) PowerKw() (float64, error) {
	return 0.0, nil
}

// PowerState implements the Bmc interface
func (i *Ibmc) PowerState() (string, error) {
	return "", nil
}

// PowerCycleBmc implements the Bmc interface
func (i *Ibmc) PowerCycleBmc() (status bool, err error) {
	return false, nil
}

// PowerCycle implements the Bmc interface
func (i *Ibmc) PowerCycle() (status bool, err error) {
	return false, nil
}

// PowerOff implements the Bmc interface
func (i *Ibmc) PowerOff() (status bool, err error) {
	return false, nil
}

// PowerOn implements the Bmc interface
func (i *Ibmc) PowerOn() (status bool, err error) {
	return false, nil
}

// PxeOnce implements the Bmc interface
func (i *Ibmc) PxeOnce() (status bool, err error) {
	return false, nil
}

// Serial implements the Bmc interface
func (i *Ibmc) Serial() (string, error) {
	return "", nil
}

// ChassisSerial implements the Bmc interface
func (i *Ibmc) ChassisSerial() (string, error) {
	return "", nil
}

// Status implements the Bmc interface
func (i *Ibmc) Status() (string, error) {
	return "", nil
}

// TempC implements the Bmc interface
func (i *Ibmc) TempC() (int, error) {
	return 0, nil
}

// Vendor implements the Bmc interface
func (i *Ibmc) Vendor() string {
	return ""
}

// Screenshot implements the Bmc interface
func (i *Ibmc) Screenshot() ([]byte, string, error) {
	return []byte{}, "", nil
}

// ServerSnapshot implements the Bmc interface
func (i *Ibmc) ServerSnapshot() (interface{}, error) {
	return nil, nil
}

// UpdateCredentials implements the Bmc interface
func (i *Ibmc) UpdateCredentials(string, string) {
}

// Slot implements the Bmc interface
func (i *Ibmc) Slot() (int, error) {
	return -1, nil
}

// UpdateFirmware implements the Bmc inteface
func (i *Ibmc) UpdateFirmware(string, string) (bool, string, error) {
	return false, "Not yet implemented", fmt.Errorf("not yet implemented")
}

// IsOn implements the Bmc interface
func (i *Ibmc) IsOn() (status bool, err error) {
	return false, nil
}

// BiosVersion returns the BIOS version from the BMC, implements the Firmware interface
func (i *Ibmc) GetBIOSVersion(ctx context.Context) (string, error) {
	return "", errors.ErrNotImplemented
}

// BMCVersion returns the BMC version, implements the Firmware interface
func (i *Ibmc) GetBMCVersion(ctx context.Context) (string, error) {
	return "", errors.ErrNotImplemented
}

// Updates the BMC firmware, implements the Firmware interface
func (i *Ibmc) FirmwareUpdateBMC(ctx context.Context, filePath string) error {
	return errors.ErrNotImplemented
}
