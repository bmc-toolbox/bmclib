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
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"

	log := logging.DefaultLogger()
	cl := NewClient(host, port, user, pass, WithLogger(log))
	cl.DiscoverCompatible(ctx)
	err := cl.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close(ctx)

	cl.Registry = cl.Registry.PreferProtocol("redfish")
	state, err := cl.GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)

	cl.Registry = cl.Registry.PreferProtocol("ipmi")
	if err != nil {
		t.Log(err)
	}
	state, err = cl.GetPowerState(ctx)
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
