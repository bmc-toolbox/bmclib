package redfishwrapper

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

func TestBootDeviceUSBCDROMMapping(t *testing.T) {
	target, err := bootDeviceStringToTarget(string(bmc.BootDeviceTypeUSBCDROM))
	require.NoError(t, err)
	assert.Equal(t, "UsbCd", string(target))

	device, err := bootTargetToBootDeviceType(target)
	require.NoError(t, err)
	assert.Equal(t, bmc.BootDeviceTypeUSBCDROM, device)
}

func TestSystemBootDeviceSetUSBCDROM(t *testing.T) {
	var patched []map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "smc_1.14.0_serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "smc_1.14.0_systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(mustReadFile(t, "smc_1.14.0_systems_1.json"))
		case http.MethodPatch:
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			body := map[string]any{}
			require.NoError(t, json.Unmarshal(b, &body))
			patched = append(patched, body)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	ctx := context.Background()
	require.NoError(t, client.Open(ctx))
	t.Cleanup(func() { _ = client.Close(ctx) })

	ok, err := client.SystemBootDeviceSet(ctx, string(bmc.BootDeviceTypeUSBCDROM), false, true)
	require.NoError(t, err)
	assert.True(t, ok)

	require.Len(t, patched, 1)
	boot, isMap := patched[0]["Boot"].(map[string]any)
	require.True(t, isMap, "PATCH body has no Boot object: %v", patched[0])
	assert.Equal(t, "UsbCd", boot["BootSourceOverrideTarget"])
	assert.Equal(t, "Once", boot["BootSourceOverrideEnabled"])
	assert.Equal(t, "UEFI", boot["BootSourceOverrideMode"])
}
