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
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	assert.Equal(t, "l5L29IP7", aClient.loginSession.CSRFToken)
}

func Test_Close(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	err = aClient.httpsLogout(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}
}

func Test_FirwmwareUpdateBMC(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
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
	err = aClient.firmwareInstallBMC(context.TODO(), fh, 0)
	if err != nil {
		t.Errorf(err.Error())
	}
}
