package redfishwrapper

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

func TestGetChassisThermal(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1/Thermal", endpointFunc(t, "dell/thermal.json"))
	client := newDellPowerClient(t, mux)

	thermal, err := client.GetChassisThermal(context.Background())
	require.NoError(t, err)

	require.Len(t, thermal.Temperatures, 2)
	cpu1 := thermal.Temperatures[0]
	assert.Equal(t, "CPU1 Temp", cpu1.Name)
	assert.Equal(t, "CPU", cpu1.PhysicalContext)
	require.NotNil(t, cpu1.ReadingCelsius)
	assert.Equal(t, float64(45), *cpu1.ReadingCelsius)
	assert.Equal(t, "OK", cpu1.Health)
	assert.Equal(t, "Enabled", cpu1.State)

	cpu2 := thermal.Temperatures[1]
	assert.Nil(t, cpu2.ReadingCelsius, "fixture reports ReadingCelsius null for the absent CPU2 sensor")
	assert.Equal(t, "Absent", cpu2.State)

	require.Len(t, thermal.Fans, 2)
	fan := thermal.Fans[0]
	assert.Equal(t, "System Board Fan1A", fan.Name)
	require.NotNil(t, fan.Reading)
	assert.Equal(t, 5880, *fan.Reading)
	assert.Equal(t, "RPM", fan.ReadingUnits)
	assert.Equal(t, "OK", fan.Health)
}

func TestGetChassisThermalFanNameFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1/Thermal", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Thermal",
			"Fans": [{"FanName": "Fan 1", "Reading": 40, "ReadingUnits": "Percent"}]
		}`))
	})
	client := newDellPowerClient(t, mux)

	thermal, err := client.GetChassisThermal(context.Background())
	require.NoError(t, err)
	require.Len(t, thermal.Fans, 1)
	assert.Equal(t, "Fan 1", thermal.Fans[0].Name)
	assert.Equal(t, "Percent", thermal.Fans[0].ReadingUnits)
}

func TestThermalNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1/Thermal", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	client := newDellPowerClient(t, mux)

	_, err := client.GetChassisThermal(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrThermalNotFound)
}
