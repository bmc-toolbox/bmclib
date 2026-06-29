package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.SystemEventLog = (*Conn)(nil)

// GetSystemEventLog returns the System Event Log entries as rows of
// [id, created, description, message], aggregated from the BMC log services.
//
// Implements bmc.SystemEventLog.
func (c *Conn) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	return c.redfishwrapper.GetSystemEventLog(ctx)
}

// GetSystemEventLogRaw returns the raw JSON of the SEL log entries.
//
// Implements bmc.SystemEventLog.
func (c *Conn) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	return c.redfishwrapper.GetSystemEventLogRaw(ctx)
}

// ClearSystemEventLog clears the BMC log services via the LogService.ClearLog
// action.
//
// Implements bmc.SystemEventLog.
func (c *Conn) ClearSystemEventLog(ctx context.Context) (err error) {
	return c.redfishwrapper.ClearSystemEventLog(ctx)
}
