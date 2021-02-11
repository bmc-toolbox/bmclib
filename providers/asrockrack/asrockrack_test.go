package asrockrack

import (
	"context"
	"os"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

//func TestBmcInterface(t *testing.T) {

//c, err := New(context.TODO(), ip, username, password, logrusr.NewLogger(testLog))
//if err != nil {
//return c, err
//}
//_ = devices.Bmc(c)
//_ = devices.Configure(c)
//_ = devices.Firmware(c)
//}

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

	defer os.Remove(upgradeFile)
	err = aClient.FirmwareUpdateBMC(context.TODO(), upgradeFile)
	if err != nil {
		t.Errorf(err.Error())
	}
}
