package ipmitool

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/bmclib/logging"
)

func TestPowerState(t *testing.T) {
	t.Skip("need real ipmi server")
	i := Conn{
		Host: "127.0.0.1",
		Port: "623",
		User: "ADMIN",
		Pass: "ADMIN",
		Log:  logging.DefaultLogger(),
	}
	state, err := i.PowerStateGet(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Fatal()

}

func TestPowerSet(t *testing.T) {
	t.Skip("need real ipmi server")
	i := Conn{
		Host: "127.0.0.1",
		Port: "623",
		User: "ADMIN",
		Pass: "ADMIN",
		Log:  logging.DefaultLogger(),
	}
	state, err := i.PowerSet(context.Background(), "soft")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Fatal()

}

func TestBootDeviceSet(t *testing.T) {
	t.Skip("need real ipmi server")
	i := Conn{
		Host: "127.0.0.1",
		Port: "623",
		User: "ADMIN",
		Pass: "ADMIN",
		Log:  logging.DefaultLogger(),
	}
	state, err := i.BootDeviceSet(context.Background(), "disk", false, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Fatal()

}

func TestBMCReset(t *testing.T) {
	t.Skip("need real ipmi server")
	i := Conn{
		Host: "127.0.0.1",
		Port: "623",
		User: "ADMIN",
		Pass: "ADMIN",
		Log:  logging.DefaultLogger(),
	}
	state, err := i.BmcReset(context.Background(), "warm")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Fatal()

}
