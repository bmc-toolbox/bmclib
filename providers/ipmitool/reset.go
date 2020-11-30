package ipmitool

import (
	"context"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
)

// BmcReset will reset a BMC
func (c *Conn) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	i, err := ipmi.New(c.User, c.Pass, c.Host+":"+c.Port)
	if err != nil {
		return ok, err
	}
	return i.PowerResetBmc(ctx, resetType)
}
