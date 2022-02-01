package ipmitool

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
