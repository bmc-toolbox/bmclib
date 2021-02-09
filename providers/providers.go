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
	// FeatureBmcVersionRead means an implementation that returns the BMC firmware version
	FeatureBmcVersionRead registrar.Feature = "bmcversionread"
	// FeatureBiosVersionRead means an implementation that returns the BIOS firmware version
	FeatureBiosVersionRead registrar.Feature = "biosversionread"
	// FeatureBmcFirmwareUpdate means an implementation that updates the BMC firmware
	FeatureBmcFirmwareUpdate registrar.Feature = "bmcfirwareupdate"
	// FeatureBiosFirmwareUpdate means an implementation that updates the BIOS firmware
	FeatureBiosFirmwareUpdate registrar.Feature = "biosfirwareupdate"
)
