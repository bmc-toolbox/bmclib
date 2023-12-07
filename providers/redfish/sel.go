package redfish

import "context"

func (c *Conn) ClearSystemEventLog(ctx context.Context) (err error) {
	return c.redfishwrapper.ClearSystemEventLog(ctx)
}

func (c *Conn) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	return c.redfishwrapper.GetSystemEventLog(ctx)
}

func (c *Conn) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	return c.redfishwrapper.GetSystemEventLogRaw(ctx)
}
