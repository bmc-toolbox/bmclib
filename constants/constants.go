// Package constants defines shared constants used across bmclib.
package constants

type (
	// OperationApplyTime is the Redfish operation apply time parameter.
	OperationApplyTime string

	// The FirmwareInstallStep identifies each phase of a firmware install process.
	FirmwareInstallStep string

	// TaskState identifies the state of a firmware install or BMC task.
	TaskState string
)

const (
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
	// Intel is the constant to identify Intel hardware
	Intel = "Intel"

	// Redfish firmware apply at constants

	// Immediate sets the firmware to be installed immediately after upload
	Immediate OperationApplyTime = "Immediate"
	// OnReset sets the firmware to be install on device power cycle/reset
	OnReset OperationApplyTime = "OnReset"
	// OnStartUpdateRequest sets the firmware install to begin after the start request has been sent.
	OnStartUpdateRequest OperationApplyTime = "OnStartUpdateRequest"

	// TODO: rename FirmwareInstall* task status names to FirmwareTaskState and declare a type.

	// Firmware install states returned by bmclib provider FirmwareInstallStatus implementations
	//
	// The redfish from the redfish spec are exposed as a smaller set of bmclib states for callers
	// https://www.dmtf.org/sites/default/files/standards/documents/DSP2046_2020.3.pdf

	// FirmwareInstallInitializing indicates the device is performing init actions to install the update
	// this covers the redfish states - 'starting', 'downloading'
	// no action is required from the callers part in this state
	FirmwareInstallInitializing = "initializing"
	// Initializing is the TaskState equivalent of FirmwareInstallInitializing.
	Initializing TaskState = "initializing"

	// FirmwareInstallQueued indicates the device has queued the update, but has not started the update task yet
	// this covers the redfish states - 'pending', 'new'
	// no action is required from the callers part in this state
	FirmwareInstallQueued = "queued"
	// Queued is the TaskState equivalent of FirmwareInstallQueued.
	Queued TaskState = "queued"

	// FirmwareInstallRunning indicates the device is installing the update
	// this covers the redfish states - 'running', 'stopping', 'cancelling'
	// no action is required from the callers part in this state
	FirmwareInstallRunning = "running"
	// Running is the TaskState equivalent of FirmwareInstallRunning.
	Running TaskState = "running"

	// FirmwareInstallComplete indicates the device completed the firmware install
	// this covers the redfish state - 'complete'
	FirmwareInstallComplete = "complete"
	// Complete is the TaskState equivalent of FirmwareInstallComplete.
	Complete TaskState = "complete"

	// FirmwareInstallFailed indicates the firmware install failed
	// this covers the redfish states - 'interrupted', 'killed', 'exception', 'cancelled', 'suspended'
	FirmwareInstallFailed = "failed"
	// Failed is the TaskState equivalent of FirmwareInstallFailed.
	Failed TaskState = "failed"

	// FirmwareInstallPowerCycleHost indicates the firmware install requires a host power cycle
	FirmwareInstallPowerCycleHost = "powercycle-host"
	// PowerCycleHost is the TaskState equivalent of FirmwareInstallPowerCycleHost.
	PowerCycleHost TaskState = "powercycle-host"

	// FirmwareInstallUnknown indicates the firmware install state could not be determined.
	FirmwareInstallUnknown = "unknown"
	// Unknown is the TaskState equivalent of FirmwareInstallUnknown.
	Unknown TaskState = "unknown"

	// FirmwareInstallStepUploadInitiateInstall identifies the step to upload _and_ initialize the firmware install.
	// as part of the same call.
	FirmwareInstallStepUploadInitiateInstall FirmwareInstallStep = "upload-initiate-install"

	// FirmwareInstallStepInstallStatus identifies the step to verify the status of the firmware install.
	FirmwareInstallStepInstallStatus FirmwareInstallStep = "install-status"

	// FirmwareInstallStepUpload identifies the upload step in the firmware install process.
	FirmwareInstallStepUpload FirmwareInstallStep = "upload"

	// FirmwareInstallStepUploadStatus identifies the step to verify the upload status as part of the firmware install status.
	FirmwareInstallStepUploadStatus FirmwareInstallStep = "upload-status"

	// FirmwareInstallStepInstallUploaded identifies the step to install firmware uploaded in FirmwareInstallStepUpload.
	FirmwareInstallStepInstallUploaded FirmwareInstallStep = "install-uploaded"

	// FirmwareInstallStepPowerOffHost indicates the host requires to be powered off.
	FirmwareInstallStepPowerOffHost FirmwareInstallStep = "power-off-host"

	// FirmwareInstallStepResetBMCPostInstall indicates the BMC requires a reset after the install.
	FirmwareInstallStepResetBMCPostInstall FirmwareInstallStep = "reset-bmc-post-install"

	// FirmwareInstallStepResetBMCOnInstallFailure indicates the BMC requires a reset if an install fails.
	FirmwareInstallStepResetBMCOnInstallFailure FirmwareInstallStep = "reset-bmc-on-install-failure"

	// POSTStateBootINIT identifies the boot init/PXE device BIOS/UEFI POST state.
	POSTStateBootINIT = "boot-init/pxe"
	// POSTStateUEFI identifies the UEFI BIOS/UEFI POST state.
	POSTStateUEFI = "uefi"
	// POSTStateOS identifies the grub/OS BIOS/UEFI POST state.
	POSTStateOS = "grub/os"
	// POSTCodeUnknown identifies an unknown BIOS/UEFI POST state.
	POSTCodeUnknown = "unknown"
)

// ListSupportedVendors  returns a list of supported vendors
func ListSupportedVendors() []string {
	return []string{HP, Dell, Supermicro}
}
