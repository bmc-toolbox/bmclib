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
	}
	state, err := i.PowerStateGet(context.Background(), logging.DefaultLogger())
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
	}
	state, err := i.PowerSet(context.Background(), logging.DefaultLogger(), "soft")
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
	}
	state, err := i.BootDeviceSet(context.Background(), logging.DefaultLogger(), "disk", false, false)
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
	}
	state, err := i.BmcReset(context.Background(), logging.DefaultLogger(), "warm")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Fatal()

}
