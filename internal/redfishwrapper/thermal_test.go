package redfishwrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

const (
	dellThermalSubsystemPath = "/redfish/v1/Chassis/System.Embedded.1/ThermalSubsystem"
	dellThermalPath          = "/redfish/v1/Chassis/System.Embedded.1/Thermal"
)

// notFound answers 404 for a link the chassis advertises but the firmware does not implement.
func notFound(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`{"error": {"code": "Base.1.0.ResourceMissingAtURI", "message": "not found"}}`))
}

// newDellChassisClient serves the Dell service root and chassis collection with the given
// chassis handler and returns an opened client.
func newDellChassisClient(t *testing.T, mux *http.ServeMux, chassis http.HandlerFunc) *Client {
	t.Helper()

	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Chassis", endpointFunc(t, "dell/chassis_collection.json"))
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1", chassis)

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	ctx := context.Background()
	require.NoError(t, client.Open(ctx))
	t.Cleanup(func() { _ = client.Close(ctx) })

	return client
}

// newDellThermalSubsystemClient serves the ThermalSubsystem, Fans and ThermalMetrics fixtures.
func newDellThermalSubsystemClient(t *testing.T, mux *http.ServeMux) *Client {
	t.Helper()

	mux.HandleFunc(dellThermalSubsystemPath, endpointFunc(t, "dell/thermalsubsystem.json"))
	mux.HandleFunc(dellThermalSubsystemPath+"/Fans", endpointFunc(t, "dell/thermalsubsystem_fans.json"))
	mux.HandleFunc(dellThermalSubsystemPath+"/Fans/0", endpointFunc(t, "dell/thermalsubsystem_fan_0.json"))
	mux.HandleFunc(dellThermalSubsystemPath+"/Fans/1", endpointFunc(t, "dell/thermalsubsystem_fan_1.json"))
	mux.HandleFunc(dellThermalSubsystemPath+"/ThermalMetrics", endpointFunc(t, "dell/thermalmetrics.json"))

	return newDellPowerClient(t, mux)
}

func TestGetChassisFanSpeeds(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellThermalPath, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("legacy Thermal must not be read when ThermalSubsystem is served")
		w.WriteHeader(http.StatusInternalServerError)
	})
	client := newDellThermalSubsystemClient(t, mux)

	fans, err := client.GetChassisFanSpeeds(context.Background())
	require.NoError(t, err)
	require.Len(t, fans, 2)

	fan1A := fans[0]
	assert.Equal(t, "System Board Fan1A", fan1A.Name)
	assert.Equal(t, "SystemBoard", fan1A.PhysicalContext)
	require.NotNil(t, fan1A.SpeedRPM)
	assert.Equal(t, float64(5880), *fan1A.SpeedRPM)
	require.NotNil(t, fan1A.SpeedPercent)
	assert.Equal(t, float64(42), *fan1A.SpeedPercent)
	assert.Equal(t, "OK", fan1A.Health)
	assert.Equal(t, "Enabled", fan1A.State)

	fan1B := fans[1]
	assert.Equal(t, "System Board Fan1B", fan1B.Name)
	assert.Nil(t, fan1B.SpeedRPM, "fixture reports no reading for the absent fan")
	assert.Nil(t, fan1B.SpeedPercent)
	assert.Equal(t, "Absent", fan1B.State)
}

func TestGetChassisTemperatures(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellThermalPath, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("legacy Thermal must not be read when ThermalSubsystem is served")
		w.WriteHeader(http.StatusInternalServerError)
	})
	client := newDellThermalSubsystemClient(t, mux)

	temperatures, err := client.GetChassisTemperatures(context.Background())
	require.NoError(t, err)
	require.Len(t, temperatures, 2)

	cpu1 := temperatures[0]
	assert.Equal(t, "CPU1 Temp", cpu1.Name)
	assert.Equal(t, "CPU", cpu1.PhysicalContext)
	require.NotNil(t, cpu1.ReadingCelsius)
	assert.Equal(t, float64(45), *cpu1.ReadingCelsius)
	assert.Empty(t, cpu1.Health, "ThermalMetrics excerpts carry readings only")
	assert.Empty(t, cpu1.State)

	inlet := temperatures[1]
	assert.Equal(t, "Inlet Temp", inlet.Name)
	assert.Equal(t, "Intake", inlet.PhysicalContext)
	require.NotNil(t, inlet.ReadingCelsius)
	assert.Equal(t, float64(23), *inlet.ReadingCelsius)
}

// Firmware advertising the ThermalSubsystem link without implementing it answers 404;
// the readings then come from the deprecated Thermal resource.
func TestGetChassisThermalLegacyFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellThermalSubsystemPath, notFound)
	mux.HandleFunc(dellThermalPath, endpointFunc(t, "dell/thermal.json"))
	client := newDellPowerClient(t, mux)

	fans, err := client.GetChassisFanSpeeds(context.Background())
	require.NoError(t, err)
	require.Len(t, fans, 2)
	assert.Equal(t, "System Board Fan1A", fans[0].Name)
	require.NotNil(t, fans[0].SpeedRPM)
	assert.Equal(t, float64(5880), *fans[0].SpeedRPM)
	assert.Nil(t, fans[0].SpeedPercent)
	assert.Equal(t, "OK", fans[0].Health)

	temperatures, err := client.GetChassisTemperatures(context.Background())
	require.NoError(t, err)
	require.Len(t, temperatures, 2)
	assert.Equal(t, "CPU1 Temp", temperatures[0].Name)
	require.NotNil(t, temperatures[0].ReadingCelsius)
	assert.Equal(t, float64(45), *temperatures[0].ReadingCelsius)
	assert.Equal(t, "Enabled", temperatures[0].State)
	assert.Nil(t, temperatures[1].ReadingCelsius, "fixture reports ReadingCelsius null for the absent CPU2 sensor")
	assert.Equal(t, "Absent", temperatures[1].State)
}

func TestGetChassisFanSpeedsLegacyPercent(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellThermalSubsystemPath, notFound)
	mux.HandleFunc(dellThermalPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"@odata.id": "/redfish/v1/Chassis/System.Embedded.1/Thermal",
			"Fans": [{"FanName": "Fan 1", "Reading": 40, "ReadingUnits": "Percent"}]
		}`))
	})
	client := newDellPowerClient(t, mux)

	fans, err := client.GetChassisFanSpeeds(context.Background())
	require.NoError(t, err)
	require.Len(t, fans, 1)
	assert.Equal(t, "Fan 1", fans[0].Name, "pre-1.1.0 Fan schema only carries FanName")
	assert.Nil(t, fans[0].SpeedRPM)
	require.NotNil(t, fans[0].SpeedPercent)
	assert.Equal(t, float64(40), *fans[0].SpeedPercent)
}

// A failure other than 404 while collecting is returned, not skipped.
func TestGetChassisThermalErrorSurfaced(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellThermalSubsystemPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc(dellThermalPath, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("legacy Thermal must not be read after a ThermalSubsystem failure")
	})
	client := newDellPowerClient(t, mux)

	_, err := client.GetChassisFanSpeeds(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "chassis System.Embedded.1 ThermalSubsystem")

	_, err = client.GetChassisTemperatures(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "chassis System.Embedded.1 ThermalSubsystem")
}

func TestGetChassisThermalNotFound(t *testing.T) {
	// A chassis without ThermalSubsystem or Thermal links.
	client := newDellChassisClient(t, http.NewServeMux(), func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"@odata.id": "/redfish/v1/Chassis/System.Embedded.1",
			"Id": "System.Embedded.1",
			"Name": "Computer System Chassis"
		}`))
	})

	_, err := client.GetChassisFanSpeeds(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrChassisFansNotFound)

	_, err = client.GetChassisTemperatures(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrChassisTemperaturesNotFound)
}
