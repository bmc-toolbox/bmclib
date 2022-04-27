package asrockrack

import (
	"context"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

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

	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	fwInfo, err := aClient.firmwareInfo(context.TODO())
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, expected, fwInfo)
}

func Test_inventoryInfo(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	inventory, err := aClient.inventoryInfo(context.TODO())
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, 6, len(inventory))
	assert.Equal(t, "CPU", inventory[0].DeviceType)
	assert.Equal(t, "Storage device", inventory[5].DeviceType)
}
func Test_fruInfo(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	frus, err := aClient.fruInfo(context.TODO())
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, 3, len(frus))
}

func Test_sensors(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	sensors, err := aClient.sensors(context.TODO())
	if err != nil {
		t.Fatal(err.Error())
	}

	assert.Equal(t, 27, len(sensors))
}

func Test_biosPOSTCode(t *testing.T) {
	expected := biosPOSTCode{
		PostStatus: 1,
		PostData:   160,
	}

	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	info, err := aClient.postCodeInfo(context.TODO())
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, expected, info)
}

func Test_chassisStatus(t *testing.T) {
	expected := chassisStatus{
		PowerStatus: 1,
		LEDStatus:   0,
	}

	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf(err.Error())
	}

	info, err := aClient.chassisStatusInfo(context.TODO())
	if err != nil {
		t.Error(err.Error())
	}

	assert.Equal(t, expected, info)
}
