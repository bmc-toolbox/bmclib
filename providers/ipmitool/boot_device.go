package ipmitool

import (
	"context"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
)

// BootDeviceSet sets the next boot device with options
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	i, err := ipmi.New(c.User, c.Pass, c.Host+":"+c.Port)
	if err != nil {
		return ok, err
	}
	return i.BootDeviceSet(ctx, bootDevice, setPersistent, efiBoot)
}
