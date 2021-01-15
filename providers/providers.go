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
)
