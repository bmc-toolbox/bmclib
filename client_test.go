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
	cl.Registry.Drivers = cl.Registry.FilterForCompatible(ctx)
	err := cl.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close(ctx)

	cl.Registry.Drivers = cl.Registry.PreferProtocol("redfish")
	state, err := cl.GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)

	cl.Registry.Drivers = cl.Registry.PreferProtocol("ipmi")
	if err != nil {
		t.Log(err)
	}

	var providerName string
	// if you pass in a string pointer to any function
	// it will be updated with the name of the provider
	// that successfully execute
	state, err = cl.GetPowerState(ctx, &providerName)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Log(providerName)

	users, err := cl.ReadUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users)

	t.Fatal()
}
