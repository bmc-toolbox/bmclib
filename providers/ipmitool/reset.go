package ipmitool

import (
	"context"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/go-logr/logr"
)

// BmcReset will reset a BMC
func (c *Conn) BmcReset(ctx context.Context, log logr.Logger, resetType string) (ok bool, err error) {
	i, err := ipmi.New(c.User, c.Pass, c.Host+":"+c.Port)
	if err != nil {
		return ok, err
	}
	return i.PowerResetBmc(ctx, resetType)
}
