package ilo

import (
	"time"

	"github.com/bmc-toolbox/bmclib/sshmock"

	mrand "math/rand"

	"fmt"
	"testing"
)

// Test server based on:
// http://grokbase.com/t/gg/golang-nuts/165yek1eje/go-nuts-creating-an-ssh-server-instance-for-tests

func init() {
	mrand.Seed(time.Now().Unix())
}

func sshServerAddress(min, max int) string {
	return fmt.Sprintf("127.0.0.1:%d", mrand.Intn(max-min)+min)
}

var (
	sshServer  *sshmock.Server
	sshAnswers = map[string][]byte{
		"power reset":    []byte(`Server resetting .......`),
		"reset /map1":    []byte(`Resetting iLO`),
		"power on":       []byte(`Server powering on .......`),
		"power off hard": []byte(`Forcing server power off .......`),
		"power":          []byte(`power: server power is currently: On`),
	}
)

func setupSSH() (bmc *Ilo, err error) {
	sshServer, err = sshmock.New(sshAnswers, true)
	if err != nil {
		return bmc, err
	}
	address := sshServer.Address()

	bmc, err = setup()
	if err != nil {
		return bmc, err
	}
	bmc.ip = address

	return bmc, err
}

func tearDownSSH() {
	tearDown()
	sshServer.Close()
}

func TestIloPowerCycle(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerCycle()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycle %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestIloPowerCycleBmc(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerCycleBmc()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycleBmc %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestIloPowerOn(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerOn()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOn %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestIloPowerOff(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.PowerOff()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOff %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}

func TestIloIsOn(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	defer tearDownSSH()

	answer, err := bmc.IsOn()
	if err != nil {
		t.Fatalf("Found errors calling bmc.IsOn %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}
}
