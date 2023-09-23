package redfish

import "context"

func (c *Conn) ClearSystemEventLog(ctx context.Context) (err error) {
	return c.redfishwrapper.ClearSystemEventLog(ctx)
}
