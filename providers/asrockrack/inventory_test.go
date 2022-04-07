package asrockrack

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetInventory(t *testing.T) {
	device, err := aClient.GetInventory(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, device)
	assert.Equal(t, "ASRockRack", device.Vendor)
	assert.Equal(t, "E3C246D4I-NL", device.Model)

	assert.Equal(t, "L2.07B", device.BIOS.Firmware.Installed)
	assert.Equal(t, "0.01.00", device.BMC.Firmware.Installed)
	assert.Equal(t, "000000ca", device.CPUs[0].Firmware.Installed)
	assert.Equal(t, "Intel(R) Xeon(R) E-2278G CPU @ 3.40GHz", device.CPUs[0].Model)
	assert.Equal(t, 2, len(device.Memory))
	assert.Equal(t, 2, len(device.Drives))
	assert.Equal(t, "OK", device.Status.Health)
}
