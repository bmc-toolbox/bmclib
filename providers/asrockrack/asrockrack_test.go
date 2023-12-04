package asrockrack

import (
	"context"
	"os"
	"testing"
	"time"

	"gopkg.in/go-playground/assert.v1"
)

func TestHttpLogin(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	assert.Equal(t, "l5L29IP7", aClient.loginSession.CSRFToken)
}

func TestClose(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	err = aClient.httpsLogout(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestFirwmwareUpdateBMC(t *testing.T) {
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
	ctx, cancel := context.WithTimeout(context.TODO(), time.Minute*15)
	defer cancel()

	err = aClient.firmwareUploadBMC(ctx, fh)
	if err != nil {
		t.Errorf(err.Error())
	}
}
