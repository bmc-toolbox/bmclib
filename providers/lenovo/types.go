package lenovo

// derefFloat64 returns the value of p, or 0 if p is nil.
func derefFloat64(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

// derefFloat32 returns the value of p as a float64, or 0 if p is nil.
func derefFloat32(p *float32) float64 {
	if p == nil {
		return 0
	}
	return float64(*p)
}

// derefInt returns the value of p, or 0 if p is nil.
func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// odataID is the common Redfish reference link shape: {"@odata.id": "/redfish/..."}.
type odataID struct {
	ODataID string `json:"@odata.id"`
}

// ServiceRoot is a partial model of the XCC Redfish service root
// ("/redfish/v1/") holding the fields and links the provider needs to discover
// the rest of the service. It is intentionally not exhaustive — only the links
// consumed by this provider's capabilities are modelled.
type ServiceRoot struct {
	ID             string  `json:"Id"`
	Name           string  `json:"Name"`
	Vendor         string  `json:"Vendor"`
	RedfishVersion string  `json:"RedfishVersion"`
	UUID           string  `json:"UUID"`
	Product        string  `json:"Product"`
	Systems        odataID `json:"Systems"`
	Managers       odataID `json:"Managers"`
	Chassis        odataID `json:"Chassis"`
	UpdateService  odataID `json:"UpdateService"`
	AccountService odataID `json:"AccountService"`
	// SessionService is the link to the session service; the Sessions
	// collection itself is nested under Links.
	SessionService     odataID `json:"SessionService"`
	Tasks              odataID `json:"Tasks"`
	EventService       odataID `json:"EventService"`
	TelemetryService   odataID `json:"TelemetryService"`
	JobService         odataID `json:"JobService"`
	CertificateService odataID `json:"CertificateService"`
	Registries         odataID `json:"Registries"`
	Links              struct {
		Sessions odataID `json:"Sessions"`
	} `json:"Links"`
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
