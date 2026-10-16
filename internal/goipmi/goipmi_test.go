package goipmi

import (
	"testing"

	"github.com/bougou/go-ipmi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClone(t *testing.T) {
	orig, err := New("user", "pass", "host", 623, WithCipherSuite("17"))
	require.NoError(t, err)

	clone, err := orig.Clone()
	require.NoError(t, err)

	// The clone must carry a fresh, independent client so operations on it
	// never touch the original's session.
	assert.NotSame(t, orig, clone)
	assert.NotSame(t, orig.client, clone.client)

	// All connection parameters must be preserved.
	assert.Equal(t, orig.Username, clone.Username)
	assert.Equal(t, orig.Password, clone.Password)
	assert.Equal(t, orig.Host, clone.Host)
	assert.Equal(t, orig.Port, clone.Port)
	assert.Equal(t, orig.cipherSuite, clone.cipherSuite)
}

func TestBootDeviceSelector(t *testing.T) {
	tests := map[string]struct {
		device  string
		want    ipmi.BootDeviceSelector
		wantErr bool
	}{
		"none":             {device: "none", want: ipmi.BootDeviceSelectorNoOverride},
		"pxe":              {device: "pxe", want: ipmi.BootDeviceSelectorForcePXE},
		"disk":             {device: "disk", want: ipmi.BootDeviceSelectorForceHardDrive},
		"hd alias":         {device: "hd", want: ipmi.BootDeviceSelectorForceHardDrive},
		"safe":             {device: "safe", want: ipmi.BootDeviceSelectorForceHardDriveSafe},
		"diag":             {device: "diag", want: ipmi.BootDeviceSelectorForceDiagnosticPartition},
		"cdrom":            {device: "cdrom", want: ipmi.BootDeviceSelectorForceCDROM},
		"bios":             {device: "bios", want: ipmi.BootDeviceSelectorForceBIOSSetup},
		"floppy":           {device: "floppy", want: ipmi.BootDeviceSelectorForceFloppy},
		"usb_cdrom":        {device: "usb_cdrom", want: ipmi.BootDeviceSelectorForceRemoteCDROM},
		"case-insensitive": {device: "USB_CDROM", want: ipmi.BootDeviceSelectorForceRemoteCDROM},
		"unsupported":      {device: "remote_drive", wantErr: true},
		"empty":            {device: "", wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := bootDeviceSelector(tc.device)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
