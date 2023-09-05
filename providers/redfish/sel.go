package redfish

import "context"

func (c *Conn) ClearSEL(ctx context.Context) (err error) {
	return c.redfishwrapper.ClearSEL(ctx)
}
