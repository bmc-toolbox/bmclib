package lenovo

import (
	"context"
	"fmt"
)

// XCC log-service ids. XCC exposes several log services beyond the IPMI SEL;
// these constants name the ones documented in the XCC REST API guide.
const (
	LogServiceActive         = "Active"
	LogServiceAudit          = "AuditLog"
	LogServicePlatform       = "PlatformLog"
	LogServiceMaintenance    = "MaintenanceLog"
	LogServiceServiceAdvisor = "ServiceAdvisorLog"
	LogServiceDiagnostic     = "DiagnosticLog"
)

// EventLog returns the entries of a specific BMC log service, identified by its
// Redfish LogService Id (e.g. the XCC OEM log types: "AuditLog", "PlatformLog",
// "MaintenanceLog", "ServiceAdvisorLog", "DiagnosticLog", "Active").
//
// Entries are returned as rows of [id, created, severity, message]. A descriptive
// error is returned when no log service with the given id is present. This is an
// XCC-specific provider method (the additional log types are not modelled by a
// bmc.Feature interface).
func (c *Conn) EventLog(ctx context.Context, logServiceID string) ([][]string, error) {
	managers, err := c.redfishwrapper.Managers(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range managers {
		logServices, err := m.LogServices()
		if err != nil {
			return nil, err
		}

		for _, ls := range logServices {
			if ls.ID != logServiceID {
				continue
			}

			lentries, err := ls.Entries()
			if err != nil {
				return nil, fmt.Errorf("reading entries of log service %q: %w", logServiceID, err)
			}

			rows := make([][]string, 0, len(lentries))
			for _, e := range lentries {
				rows = append(rows, []string{e.ID, e.Created, string(e.Severity), e.Message})
			}

			return rows, nil
		}
	}

	return nil, fmt.Errorf("log service %q not found on this device", logServiceID)
}
