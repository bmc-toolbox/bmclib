package lenovo

// odataID is the common Redfish reference link shape: {"@odata.id": "/redfish/..."}.
type odataID struct {
	ODataID string `json:"@odata.id"`
}

// redfishError is the standard Redfish error envelope returned in the body of a
// failed request. XCC populates @Message.ExtendedInfo with entries that may
// reference the Lenovo "ExtendedError" registry.
//
// See the package documentation and errors.go for how this is mapped into a
// bmclib error.
type redfishError struct {
	Error struct {
		Code         string                     `json:"code"`
		Message      string                     `json:"message"`
		ExtendedInfo []redfishErrorExtendedInfo `json:"@Message.ExtendedInfo"`
	} `json:"error"`
}

// redfishErrorExtendedInfo is a single entry of the Redfish extended error
// information, carrying the registry message id, the human readable message and
// (for XCC OEM messages) a suggested resolution.
type redfishErrorExtendedInfo struct {
	MessageID  string `json:"MessageId"`
	Message    string `json:"Message"`
	Resolution string `json:"Resolution"`
	Severity   string `json:"Severity"`
}
