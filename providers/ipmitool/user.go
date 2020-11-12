package ipmitool

import (
	"context"

	"github.com/bmc-toolbox/bmclib/internal/ipmi"
)

// UserRead list all users
func (c *Conn) UserRead(ctx context.Context) (users []map[string]string, err error) {
	i, err := ipmi.New(c.User, c.Pass, c.Host)
	if err != nil {
		return users, err
	}
	return i.ReadUsers(ctx)
}
