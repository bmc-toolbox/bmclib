package ipmitool

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/metal-toolbox/bmclib/logging"
)

func TestMain(m *testing.M) {
	var tempDir string
	_, err := exec.LookPath("ipmitool")
	if err != nil {
		tempDir, err = os.MkdirTemp("/tmp", "")
		if err != nil {
			os.Exit(2)
		}
		path := os.Getenv("PATH") + ":" + tempDir
		os.Setenv("PATH", path)
		fmt.Println(os.Getenv("PATH"))
		f := filepath.Join(tempDir, "ipmitool")
		err = os.WriteFile(f, []byte{}, 0755)
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
	c, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
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
	c, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
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
	host := "127.0.0.1"
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"
	i, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
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
	host := "127.0.0.1"
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"
	i, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
	}
	state, err := i.BmcReset(context.Background(), "warm")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Fatal()
}

func TestDeactivateSOL(t *testing.T) {
	t.Skip("need real ipmi server")
	host := "127.0.0.1"
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"
	i, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
	}
	err = i.DeactivateSOL(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(err != nil)
	t.Fatal()
}

func TestSystemEventLogClear(t *testing.T) {
	t.Skip("need real ipmi server")
	host := "127.0.0.1"
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"
	i, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
	}
	err = i.ClearSystemEventLog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("System Event Log cleared")
	t.Fatal()
}

func TestSystemEventLogGet(t *testing.T) {
	t.Skip("need real ipmi server")
	host := "127.0.0.1"
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"
	i, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := i.GetSystemEventLog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(entries)
	t.Fatal()
}

func TestSystemEventLogGetRaw(t *testing.T) {
	t.Skip("need real ipmi server")
	host := "127.0.0.1"
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"
	i, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
	}
	eventlog, err := i.GetSystemEventLogRaw(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(eventlog)
	t.Fatal()
}

func TestSendNMI(t *testing.T) {
	t.Skip("need real ipmi server")
	host := "127.0.0.1"
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"
	i, err := New(host, user, pass, WithPort(port), WithLogger(logging.DefaultLogger()))
	if err != nil {
		t.Fatal(err)
	}
	err = i.SendNMI(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log("NMI sent")
	t.Fatal()
}
