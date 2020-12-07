package ipmitool

import (
	"context"
	"testing"

	"bou.ke/monkey"
	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/bmc-toolbox/bmclib/registry"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestInit(t *testing.T) {
	user := "ADMIN"
	pass := "ADMIN"
	host := "127.0.0.1"
	port := "623"
	ipm := &ipmi.Ipmi{Username: user, Password: pass, Host: host}
	want := &Conn{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		Log:  nil,
		con:  ipm,
	}
	monkey.Patch(ipmi.New, func(username string, password string, host string) (i *ipmi.Ipmi, err error) {
		return ipm, nil
	})
	r := registry.All()
	i, _ := r[0].InitFn(host, port, user, pass, nil)
	n := i.(*Conn)
	diff := cmp.Diff(want, n, cmpopts.IgnoreUnexported(Conn{}))
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestPowerState(t *testing.T) {
	t.Skip("need real ipmi server")
	user := "ADMIN"
	pass := "ADMIN"
	host := "127.0.0.1"
	port := "623"
	i, _ := ipmi.New(user, pass, host+":"+port)
	c := Conn{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		Log:  logging.DefaultLogger(),
		con:  i,
	}
	state, err := c.PowerStateGet(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Fatal()

}

func TestPowerSet1(t *testing.T) {
	t.Skip("need real ipmi server")
	user := "ADMIN"
	pass := "ADMIN"
	host := "127.0.0.1"
	port := "623"
	i, _ := ipmi.New(user, pass, host+":"+port)
	c := Conn{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		Log:  logging.DefaultLogger(),
		con:  i,
	}
	state, err := c.PowerSet(context.Background(), "soft")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Fatal()

}

func TestBootDeviceSet2(t *testing.T) {
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
