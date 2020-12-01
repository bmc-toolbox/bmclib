package bmclib

import (
	"context"
	"testing"
	"time"

	"github.com/bmc-toolbox/bmclib/logging"
)

func TestBMC(t *testing.T) {
	t.Skip("needs ipmitool and real ipmi server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	cl := NewClient(host, user, pass, WithLogger(logging.DefaultLogger()))
	dErr := cl.DiscoverProviders(ctx)
	if dErr != nil {
		t.Fatal(dErr)
	}
	state, err := cl.GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)

	users, err := cl.ReadUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users)

	t.Fatal()
}
