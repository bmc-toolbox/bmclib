package constants

import "strings"

const (
	// Unknown is the constant that defines unknown things
	Unknown = "Unknown"

	// EnvEnableDebug is the const for the environment variable to cause bmclib to dump debugging debugging information.
	// the valid parameter for this environment variable is 'true'
	EnvEnableDebug = "DEBUG_BMCLIB"

	// Vendor constants

	// HP is the constant that defines the vendor HP
	HP = "HP"
	// Dell is the constant that defines the vendor Dell
	Dell = "Dell"
	// Supermicro is the constant that defines the vendor Supermicro
	Supermicro = "Supermicro"
	// Cloudline is the constant that defines the cloudlines
	Cloudline = "Cloudline"
	// Quanta is the contant to identify Quanta hardware
	Quanta = "Quanta"
	// Quanta is the contant to identify Intel hardware
	Intel = "Intel"

	// Redfish firmware apply at constants
	// FirmwareApplyImmediate sets the firmware to be installed immediately after upload
	FirmwareApplyImmediate = "Immediate"
	//FirmwareApplyOnReset sets the firmware to be install on device power cycle/reset
	FirmwareApplyOnReset = "OnReset"

	// Firmware install states returned by bmclib provider FirmwareInstallStatus implementations
	//
	// The redfish from the redfish spec are exposed as a smaller set of bmclib states for callers
	// https://www.dmtf.org/sites/default/files/standards/documents/DSP2046_2020.3.pdf

	// FirmwareInstallInitializing indicates the device is performing init actions to install the update
	// this covers the redfish states - 'starting', 'downloading'
	// no action is required from the callers part in this state
	FirmwareInstallInitializing = "initializing"

	// FirmwareInstallQueued indicates the device has queued the update, but has not started the update task yet
	// this covers the redfish states - 'pending', 'new'
	// no action is required from the callers part in this state
	FirmwareInstallQueued = "queued"

	// FirmwareInstallRunner indicates the device is installing the update
	// this covers the redfish states - 'running', 'stopping', 'cancelling'
	// no action is required from the callers part in this state
	FirmwareInstallRunning = "running"

	// FirmwareInstallComplete indicates the device completed the firmware install
	// this covers the redfish state - 'complete'
	FirmwareInstallComplete = "complete"

	// FirmwareInstallFailed indicates the firmware install failed
	// this covers the redfish states - 'interrupted', 'killed', 'exception', 'cancelled', 'suspended'
	FirmwareInstallFailed = "failed"

	// FirmwareInstallPowerCycleHost indicates the firmware install requires a host power cycle
	FirmwareInstallPowerCyleHost = "powercycle-host"

	// FirmwareInstallPowerCycleBMC indicates the firmware install requires a BMC power cycle
	FirmwareInstallPowerCycleBMC = "powercycle-bmc"

	FirmwareInstallUnknown = "unknown"

	// device BIOS/UEFI POST code bmclib identifiers
	POSTStateBootINIT = "boot-init/pxe"
	POSTStateUEFI     = "uefi"
	POSTStateOS       = "grub/os"
	POSTCodeUnknown   = "unknown"
)

// ListSupportedVendors  returns a list of supported vendors
func ListSupportedVendors() []string {
	return []string{HP, Dell, Supermicro}
}

// VendorFromProductName attempts to identify the vendor from the given productname
func VendorFromProductName(productName string) string {
	n := strings.ToLower(productName)
	switch {
	case strings.Contains(n, "intel"):
		return Intel
	case strings.Contains(n, "dell"):
		return Dell
	case strings.Contains(n, "supermicro"):
		return Supermicro
	case strings.Contains(n, "cloudline"):
		return Cloudline
	case strings.Contains(n, "quanta"):
		return Quanta
	default:
		return productName
	}
}
