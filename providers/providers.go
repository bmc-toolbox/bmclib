package providers

import "github.com/jacobweinstock/registrar"

const (
	// FeaturePowerState represents the powerstate functionality
	// an implementation will use these when they have implemented
	// the corresponding interface method.
	FeaturePowerState registrar.Feature = "powerstate"
	// FeaturePowerSet means an implementation can set a BMC power state
	FeaturePowerSet registrar.Feature = "powerset"
	// FeatureUserCreate means an implementation can create BMC users
	FeatureUserCreate registrar.Feature = "usercreate"
	// FeatureUserDelete means an implementation can delete BMC users
	FeatureUserDelete registrar.Feature = "userdelete"
	// FeatureUserUpdate means an implementation can update BMC users
	FeatureUserUpdate registrar.Feature = "userupdate"
	// FeatureUserRead means an implementation can read BMC users
	FeatureUserRead registrar.Feature = "userread"
	// FeatureBmcReset means an implementation can warm or cold reset a BMC
	FeatureBmcReset registrar.Feature = "bmcreset"
	// FeatureBootDeviceSet means an implementation the next boot device
	FeatureBootDeviceSet registrar.Feature = "bootdeviceset"
	// FeaturesVirtualMedia means an implementation can manage virtual media devices
	FeatureVirtualMedia registrar.Feature = "virtualmedia"
	// FeatureMountFloppyImage means an implementation uploads a floppy image for mounting as virtual media.
	//
	// note: This is differs from FeatureVirtualMedia which is limited to accepting a URL to download the image from.
	FeatureMountFloppyImage registrar.Feature = "mountFloppyImage"
	// FeatureUnmountFloppyImage means an implementation removes a floppy image that was previously uploaded.
	FeatureUnmountFloppyImage registrar.Feature = "unmountFloppyImage"
	// FeatureFirmwareInstall means an implementation that initiates the firmware install process
	FeatureFirmwareInstall registrar.Feature = "firmwareinstall"
	// FeatureFirmwareInstallSatus means an implementation that returns the firmware install status
	FeatureFirmwareInstallStatus registrar.Feature = "firmwareinstallstatus"
	// FeatureInventoryRead means an implementation that returns the hardware and firmware inventory
	FeatureInventoryRead registrar.Feature = "inventoryread"
	// FeaturePostCodeRead means an implementation that returns the boot BIOS/UEFI post code status and value
	FeaturePostCodeRead registrar.Feature = "postcoderead"
	// FeatureScreenshot means an implementation that returns a screenshot of the video.
	FeatureScreenshot registrar.Feature = "screenshot"
	// FeatureClearSystemEventLog means an implementation that clears the BMC System Event Log (SEL)
	FeatureClearSystemEventLog registrar.Feature = "clearsystemeventlog"

	// FeatureFirmwareInstallSteps means an implementation returns the steps part of the firmware update process.
	FeatureFirmwareInstallSteps registrar.Feature = "firmwareinstallactions"

	// FeatureFirmwareUpload means an implementation that uploads firmware for installing.
	FeatureFirmwareUpload registrar.Feature = "firmwareupload"

	// 	FeatureFirmwareUploadVerify means an implementation that verifies the uploaded firmware.
	FeatureFirmwareUploadVerify registrar.Feature = "firmwareuploadverify"

	// FeatureFirmwareInstallUpload means an implementation that both uploads and initiates the firmware install process.
	FeatureFirmwareInstallUpload registrar.Feature = "firmwareinstallupload"
)
