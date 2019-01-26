package idrac9

import "github.com/bmc-toolbox/bmclib/cfgresources"

// Jobs type is how job payloads are unmarshalled.
type Jobs map[string]string

// Odata struct declares parameters for redfish odata payload.
type Odata struct {
	Attributes   *BiosSettings       `json:"Attributes,omitempty"`
	Members      []map[string]string `json:"Members,omitempty"`
	MembersCount int                 `json:"Members@odata.count,omitempty"`
	JobType      string              `json:"JobType,omitempty"`
	JobState     string              `json:"JobState,omitempty"`
}

// BiosSettings is an alias type of cfgresources.Idrac9BiosSettings
//All supported bios settings can be queried from through redfish/v1/Systems/System.Embedded.1/Bios
//NOTE: all fields int this struct are expected to be of type string, for details see diffBiosSettings()
// This type aliasing tightly couples config resources, maybe theres another aproach here.
type BiosSettings = cfgresources.Idrac9BiosSettings

// TargetSettingsURI struct declares parameters to unmarshal TargetSettingsURI.
//Post Jobs to be done
type TargetSettingsURI struct {
	TargetSettingsURI string `json:"TargetSettingsURI"` //e.g /redfish/v1/Systems/System.Embedded.1/Bios/Settings
}
