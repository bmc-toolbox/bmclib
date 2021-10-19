package asrockrack

import (
	"context"
	"os"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func Test_Compatible(t *testing.T) {
	b := aClient.Compatible()
	if !b {
		t.Errorf("expected true, got false")
	}
}

func Test_httpLogin(t *testing.T) {
	err := aClient.httpsLogin()
	if err != nil {
		t.Errorf(err.Error())
	}

	assert.Equal(t, "l5L29IP7", aClient.loginSession.CSRFToken)
}

func Test_Close(t *testing.T) {
	err := aClient.httpsLogin()
	if err != nil {
		t.Errorf(err.Error())
	}

	err = aClient.httpsLogout()
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test_FirmwareInfo(t *testing.T) {
	expected := firmwareInfo{
		BMCVersion:       "0.01.00",
		BIOSVersion:      "L2.07B",
		MEVersion:        "5.1.3.78",
		MicrocodeVersion: "000000ca",
		CPLDVersion:      "N/A",
		CMVersion:        "0.13.01",
		BPBVersion:       "0.0.002.0",
		NodeID:           "2",
	}

	err := aClient.httpsLogin()
	if err != nil {
		t.Errorf(err.Error())
	}

	fwInfo, err := aClient.firmwareInfo()
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, expected, fwInfo)
}

func Test_FirwmwareUpdateBMC(t *testing.T) {
	err := aClient.httpsLogin()
	if err != nil {
		t.Errorf(err.Error())
	}

	upgradeFile := "/tmp/dummy-E3C246D4I-NL_L0.01.00.ima"
	_, err = os.Create(upgradeFile)
	if err != nil {
		t.Errorf(err.Error())
	}

	fh, err := os.Open(upgradeFile)
	if err != nil {
		t.Errorf(err.Error())
	}

	defer fh.Close()
	err = aClient.FirmwareUpdateBMC(context.TODO(), fh, 0)
	if err != nil {
		t.Errorf(err.Error())
	}
}
