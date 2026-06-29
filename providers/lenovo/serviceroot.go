package lenovo

import (
	"context"

	"github.com/stmcginnis/gofish"
)

// ServiceRoot returns the XCC Redfish service root ("/redfish/v1/") as a
// *gofish.Service.
//
// The gofish service root already models the service identity (Vendor,
// Product, RedfishVersion, ...) and resolves the linked resource collections
// (Systems, Managers, UpdateService, ...), so the provider does not model the
// service root itself nor hard-code its URIs.
func (c *Conn) ServiceRoot(_ context.Context) (*gofish.Service, error) {
	return c.redfishwrapper.ServiceRoot()
}
