package idrac9

import (
	"testing"

	"github.com/bmc-toolbox/bmclib/sshmock"
)

var (
	sshServer  *sshmock.Server
	sshAnswers = map[string][]byte{
		"racadm serveraction hardreset": []byte(`Server power operation successful`),
		"racadm racreset hard": []byte(`RAC reset operation initiated successfully. It may take a few
			minutes for the RAC to come online again.
		   `),
		"racadm serveraction powerup":     []byte(`Server power operation successful`),
		"racadm serveraction powerdown":   []byte(`Server power operation successful`),
		"racadm serveraction powerstatus": []byte(`Server power status: ON`),
		"racadm config -g cfgServerInfo -o cfgServerBootOnce 1": []byte(`Object value modified successfully


			RAC1169: The RACADM "config" command will be deprecated in a
			future version of iDRAC firmware. Run the RACADM 
			"racadm set" command to configure the iDRAC configuration parameters.
			For more information on the set command, run the RACADM command
			"racadm help set".
			
			`),
		"racadm config -g cfgServerInfo -o cfgServerFirstBootDevice PXE": []byte(`Object value modified successfully


			RAC1169: The RACADM "config" command will be deprecated in a
			future version of iDRAC firmware. Run the RACADM 
			"racadm set" command to configure the iDRAC configuration parameters.
			For more information on the set command, run the RACADM command
			"racadm help set".
			
			`),
	}
)

func setupSSH() (bmc *IDrac9, err error) {
	username := "super"
	password := "test"

	sshServer, err = sshmock.New(sshAnswers, true)
	if err != nil {
		return bmc, err
	}
	address := sshServer.Address()

	bmc, err = New(address, username, password)
	if err != nil {
		return bmc, err
	}

	return bmc, err
}

func tearDownSSH() {
	sshServer.Close()
}

func TestIDracPowerCycle(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerCycle()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycle %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracPowerCycleBmc(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerCycleBmc()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerCycleBmc %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracPowerOn(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerOn()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOn %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracPowerOff(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerOff()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerOff %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracPxeOnce(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PxeOnce()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PxeOnce %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}

func TestIDracIsOn(t *testing.T) {
	expectedAnswer := true

	bmc, err := setupSSH()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.IsOn()
	if err != nil {
		t.Fatalf("Found errors calling bmc.IsOn %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDownSSH()
}
