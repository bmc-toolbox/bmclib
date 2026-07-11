package redfish

import "context"

// ClearSystemEventLog clears the System Event Log (SEL).
func (c *Conn) ClearSystemEventLog(ctx context.Context) (err error) {
	return c.redfishwrapper.ClearSystemEventLog(ctx)
}

// GetSystemEventLog returns the System Event Log (SEL) entries.
func (c *Conn) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	return c.redfishwrapper.GetSystemEventLog(ctx)
}

// GetSystemEventLogRaw returns the raw System Event Log (SEL) content.
func (c *Conn) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	return c.redfishwrapper.GetSystemEventLogRaw(ctx)
}
