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

	log := logging.DefaultLogger()
	cl := NewClient(host, user, pass)
	dErr := cl.DiscoverProviders(ctx, log)
	if dErr != nil {
		t.Fatal(dErr)
	}
	state, err := cl.GetPowerState(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)

	users, err := cl.ReadUsers(ctx, log)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users)

	t.Fatal()
}
