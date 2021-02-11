package asrockrack

import (
	"crypto/x509"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
)

// ApplyCfg implements the Bmc interface
func (a *ASRockRack) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return nil
}

// BiosVersion implements the Bmc interface
func (a *ASRockRack) BiosVersion() (string, error) {
	return "", nil
}

// HardwareType implements the Bmc interface
func (a *ASRockRack) HardwareType() string {
	return ""
}

// Version implements the Bmc interface
func (a *ASRockRack) Version() (string, error) {
	return "", nil
}

// CPU implements the Bmc interface
func (a *ASRockRack) CPU() (string, int, int, int, error) {
	return "", 0, 0, 0, nil
}

// Disks implements the Bmc interface
func (a *ASRockRack) Disks() ([]*devices.Disk, error) {
	return make([]*devices.Disk, 0), nil
}

// IsBlade implements the Bmc interface
func (a *ASRockRack) IsBlade() (bool, error) {
	return false, nil
}

// License implements the Bmc interface
func (a *ASRockRack) License() (string, string, error) {
	return "", "", nil
}

// Memory implements the Bmc interface
func (a *ASRockRack) Memory() (int, error) {
	return 0, nil
}

// Model implements the Bmc interface
func (a *ASRockRack) Model() (string, error) {
	return "", nil
}

// Name implements the Bmc interface
func (a *ASRockRack) Name() (string, error) {
	return "", nil
}

// Nics implements the Bmc interface
func (a *ASRockRack) Nics() ([]*devices.Nic, error) {
	return make([]*devices.Nic, 0), nil
}

// PowerKw implements the Bmc interface
func (a *ASRockRack) PowerKw() (float64, error) {
	return 0.0, nil
}

// PowerState implements the Bmc interface
func (a *ASRockRack) PowerState() (string, error) {
	return "", nil
}

// PowerCycleBmc implements the Bmc interface
func (a *ASRockRack) PowerCycleBmc() (status bool, err error) {
	return false, nil
}

// PowerCycle implements the Bmc interface
func (a *ASRockRack) PowerCycle() (status bool, err error) {
	return false, nil
}

// PowerOff implements the Bmc interface
func (a *ASRockRack) PowerOff() (status bool, err error) {
	return false, nil
}

// PowerOn implements the Bmc interface
func (a *ASRockRack) PowerOn() (status bool, err error) {
	return false, nil
}

// PxeOnce implements the Bmc interface
func (a *ASRockRack) PxeOnce() (status bool, err error) {
	return false, nil
}

// Serial implements the Bmc interface
func (a *ASRockRack) Serial() (string, error) {
	return "", nil
}

// ChassisSerial implements the Bmc interface
func (a *ASRockRack) ChassisSerial() (string, error) {
	return "", nil
}

// Status implements the Bmc interface
func (a *ASRockRack) Status() (string, error) {
	return "", nil
}

// TempC implements the Bmc interface
func (a *ASRockRack) TempC() (int, error) {
	return 0, nil
}

// Vendor implements the Bmc interface
func (a *ASRockRack) Vendor() string {
	return ""
}

// Screenshot implements the Bmc interface
func (a *ASRockRack) Screenshot() ([]byte, string, error) {
	return []byte{}, "", nil
}

// ServerSnapshot implements the Bmc interface
func (a *ASRockRack) ServerSnapshot() (interface{}, error) {
	return nil, nil
}

// UpdateCredentials implements the Bmc interface
func (a *ASRockRack) UpdateCredentials(string, string) {
}

// Slot implements the Bmc interface
func (a *ASRockRack) Slot() (int, error) {
	return -1, nil
}

// UpdateFirmware implements the Bmc inteface
func (a *ASRockRack) UpdateFirmware(string, string) (b bool, e error) {
	return b, e
}

// IsOn implements the Bmc interface
func (a *ASRockRack) IsOn() (status bool, err error) {
	return false, nil
}

// Resources returns a slice of supported resources and
// the order they are to be applied in.
func (a *ASRockRack) Resources() []string {
	return []string{
		"user",
		"syslog",
		"ntp",
		"ldap",
		"ldap_group",
		"network",
	}
}

// Power implemented the Configure interface
func (a *ASRockRack) Power(cfg *cfgresources.Power) (err error) {
	return err
}

// User method implements the Configure interface
func (a *ASRockRack) User(cfg []*cfgresources.User) error {
	return nil
}

// Syslog method implements the Configure interface
func (a *ASRockRack) Syslog(cfg *cfgresources.Syslog) error {
	return nil
}

// Ntp method implements the Configure interface
func (a *ASRockRack) Ntp(cfg *cfgresources.Ntp) error {
	return nil
}

// Ldap method implements the Configure interface
func (a *ASRockRack) Ldap(cfg *cfgresources.Ldap) error {
	return nil
}

// LdapGroup method implements the Configure interface
func (a *ASRockRack) LdapGroup(cfgGroup []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) error {
	return nil
}

// Network method implements the Configure interface
func (a *ASRockRack) Network(cfg *cfgresources.Network) (bool, error) {
	return false, nil
}

// SetLicense implements the Configure interface
func (a *ASRockRack) SetLicense(*cfgresources.License) error {
	return nil
}

// Bios method implements the Configure interface
func (a *ASRockRack) Bios(cfg *cfgresources.Bios) error {
	return nil
}

// GenerateCSR generates a CSR request on the BMC.
// GenerateCSR implements the Configure interface.
func (a *ASRockRack) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {
	return []byte{}, nil
}

// UploadHTTPSCert uploads the given CRT cert,
// UploadHTTPSCert implements the Configure interface.
func (a *ASRockRack) UploadHTTPSCert(cert []byte, certFileName string, key []byte, keyFileName string) (bool, error) {
	return false, nil
}

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// The bool value returned indicates if the BMC supports CSR generation.
// CurrentHTTPSCert implements the Configure interface.
func (a *ASRockRack) CurrentHTTPSCert() ([]*x509.Certificate, bool, error) {
	return nil, false, nil
}
