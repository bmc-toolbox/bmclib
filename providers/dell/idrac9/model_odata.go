package idrac9

import "github.com/bmc-toolbox/bmclib/cfgresources"

type Jobs map[string]string

type Odata struct {
	Attributes   *BiosSettings       `json:"Attributes,omitempty"`
	Members      []map[string]string `json:"Members,omitempty"`
	MembersCount int                 `json:"Members@odata.count,omitempty"`
	JobType      string              `json:"JobType,omitempty"`
	JobState     string              `json:"JobState,omitempty"`
}

//All supported bios settings can be queried from through redfish/v1/Systems/System.Embedded.1/Bios
//NOTE: all fields int this struct are expected to be of type string, for details see diffBiosSettings()
// This type aliasing tightly couples config resources, maybe theres another aproach here.
type BiosSettings = cfgresources.Idrac9BiosSettings

//Post Jobs to be done
type TargetSettingsUri struct {
	TargetSettingsUri string `json:"TargetSettingsURI"` //e.g /redfish/v1/Systems/System.Embedded.1/Bios/Settings
}
