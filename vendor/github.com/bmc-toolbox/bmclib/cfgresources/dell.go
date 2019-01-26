package cfgresources

// Dell holds configuration parameters for Dell assets.
type Dell struct {
	Idrac9BiosSettings *Idrac9BiosSettings `yaml:"idrac9bios"`
}

// Idrac9BiosSettings holds configuration parameters for Idrac9 bios.
type Idrac9BiosSettings struct {
	PxeDev1EnDis string `json:"PxeDev1EnDis,omitempty" yaml:"PxeDev1EnDis,omitempty" validate:"oneof=Enabled Disabled"`
	PxeDev2EnDis string `json:"PxeDev2EnDis,omitempty" yaml:"PxeDev2EnDis,omitempty" validate:"oneof=Enabled Disabled"`
	PxeDev3EnDis string `json:"PxeDev3EnDis,omitempty" yaml:"PxeDev3EnDis,omitempty" validate:"oneof=Enabled Disabled"`
	PxeDev4EnDis string `json:"PxeDev4EnDis,omitempty" yaml:"PxeDev4EnDis,omitempty" validate:"oneof=Enabled Disabled"`
}
