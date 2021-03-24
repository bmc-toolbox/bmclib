package bmclib

import (
	"context"
	"testing"
	"time"

	"github.com/bmc-toolbox/bmclib/bmc"
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
	cl.Registry.Drivers = cl.Registry.FilterForCompatible(ctx)
	var metadata bmc.Metadata
	err := cl.Open(ctx, &metadata)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close(ctx)
	t.Logf("%+v", metadata)

	cl.Registry.Drivers = cl.Registry.PreferDriver("other")
	state, err := cl.GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)

	cl.Registry.Drivers = cl.Registry.PreferProtocol("ipmi")
	if err != nil {
		t.Log(err)
	}

	// if you pass in a the metadata as a pointer to any function
	// it will be updated with details about the call. name of the provider
	// that successfully execute and providers attempted.
	state, err = cl.GetPowerState(ctx, &metadata)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Logf("%+v", metadata)

	users, err := cl.ReadUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users)

	t.Fatal()
}
