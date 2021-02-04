package ipmitool

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/bmc-toolbox/bmclib/logging"
)

func TestMain(m *testing.M) {
	var tempDir string
	_, err := exec.LookPath("ipmitool")
	if err != nil {
		tempDir, err = ioutil.TempDir("/tmp", "")
		if err != nil {
			os.Exit(2)
		}
		path := os.Getenv("PATH") + ":" + tempDir
		os.Setenv("PATH", path)
		fmt.Println(os.Getenv("PATH"))
		f := filepath.Join(tempDir, "ipmitool")
		err = ioutil.WriteFile(f, []byte{}, 0755)
		if err != nil {
			os.RemoveAll(tempDir)
			os.Exit(3)
		}
	}

	code := m.Run()
	os.RemoveAll(tempDir)
	os.Exit(code)
}

func TestIsCompatible(t *testing.T) {
	testCases := []struct {
		name string
		ok   bool
	}{
		{"true", true},
		{"false", false},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var ipm *ipmi.Ipmi
			monkey.PatchInstanceMethod(reflect.TypeOf(ipm), "PowerState", func(_ *ipmi.Ipmi, _ context.Context) (status string, err error) {
				if !tc.ok {
					err = errors.New("not compatible")
				}
				return "on", err
			})
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			user := "ADMIN"
			pass := "ADMIN"
			host := "127.1.1.1"
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
			ok := c.Compatible(ctx)
			if ok != tc.ok {
				t.Fatalf("got: %v, expected: %v", ok, tc.ok)
			}
		})
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
