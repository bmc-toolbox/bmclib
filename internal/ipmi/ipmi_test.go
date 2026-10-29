package ipmi

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBootFlagsRawCmd(t *testing.T) {
	tests := map[string]struct {
		persistent bool
		efi        bool
		want       []string
	}{
		"once legacy":    {want: []string{"raw", "0x00", "0x08", "0x05", "0x80", "0x20", "0x00", "0x00", "0x00"}},
		"once efi":       {efi: true, want: []string{"raw", "0x00", "0x08", "0x05", "0xa0", "0x20", "0x00", "0x00", "0x00"}},
		"persistent":     {persistent: true, want: []string{"raw", "0x00", "0x08", "0x05", "0xc0", "0x20", "0x00", "0x00", "0x00"}},
		"persistent efi": {persistent: true, efi: true, want: []string{"raw", "0x00", "0x08", "0x05", "0xe0", "0x20", "0x00", "0x00", "0x00"}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := bootFlagsRawCmd(bootDeviceSelectorRemoteCDROM, tc.persistent, tc.efi)
			assert.Equal(t, tc.want, got)
		})
	}
}

// fakeIpmitool writes an executable that records its arguments, one per line, to argsFile.
func fakeIpmitool(t *testing.T, exitCode int) (bin, argsFile string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake ipmitool requires a POSIX shell")
	}
	dir := t.TempDir()
	bin = filepath.Join(dir, "ipmitool")
	argsFile = filepath.Join(dir, "args")
	script := "#!/bin/sh\nfor a in \"$@\"; do echo \"$a\"; done > " + argsFile + "\nexit " + strconv.Itoa(exitCode) + "\n"
	require.NoError(t, os.WriteFile(bin, []byte(script), 0o755))
	return bin, argsFile
}

func TestBootDeviceSetUSBCDROM(t *testing.T) {
	bin, argsFile := fakeIpmitool(t, 0)
	i, err := New("user", "pass", "127.0.0.1", WithIpmitoolPath(bin))
	require.NoError(t, err)

	ok, err := i.BootDeviceSet(context.Background(), "usb_cdrom", false, true)
	require.NoError(t, err)
	assert.True(t, ok)

	b, err := os.ReadFile(argsFile)
	require.NoError(t, err)
	args := strings.Fields(string(b))
	want := []string{"raw", "0x00", "0x08", "0x05", "0xa0", "0x20", "0x00", "0x00", "0x00"}
	require.GreaterOrEqual(t, len(args), len(want))
	assert.Equal(t, want, args[len(args)-len(want):])
}

func TestBootDeviceSetUSBCDROMError(t *testing.T) {
	bin, _ := fakeIpmitool(t, 1)
	i, err := New("user", "pass", "127.0.0.1", WithIpmitoolPath(bin))
	require.NoError(t, err)

	ok, err := i.BootDeviceSet(context.Background(), "usb_cdrom", false, true)
	assert.Error(t, err)
	assert.False(t, ok)
}
