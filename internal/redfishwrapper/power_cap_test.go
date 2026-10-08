package redfishwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// powerPatch captures the PATCH payloads sent to the Dell Power resource.
type powerPatch struct {
	PowerControl []struct {
		PowerLimit struct {
			LimitInWatts *float64 `json:"LimitInWatts"`
		} `json:"PowerLimit"`
	} `json:"PowerControl"`
}

func newDellPowerClient(t *testing.T, mux *http.ServeMux) *Client {
	t.Helper()

	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Chassis", endpointFunc(t, "dell/chassis_collection.json"))
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1", endpointFunc(t, "dell/chassis.system.embedded.1.json"))

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

// dellPowerHandler serves the Power fixture on GET and records PATCH payloads.
func dellPowerHandler(t *testing.T, patches *[]powerPatch) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(mustReadFile(t, "dell/power.json"))
		case http.MethodPatch:
			var p powerPatch
			require.NoError(t, json.NewDecoder(r.Body).Decode(&p))
			*patches = append(*patches, p)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func TestGetPowerMetrics(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1/Power", endpointFunc(t, "dell/power.json"))
	client := newDellPowerClient(t, mux)

	metrics, err := client.GetPowerMetrics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, float64(234), metrics.ConsumedWatts)
	assert.Equal(t, float64(1400), metrics.CapacityWatts)
	assert.Nil(t, metrics.LimitInWatts, "fixture reports LimitInWatts null, expected no cap")
}

func TestSetPowerCap(t *testing.T) {
	var patches []powerPatch

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1/Power", dellPowerHandler(t, &patches))
	client := newDellPowerClient(t, mux)

	limit := 800.0
	require.NoError(t, client.SetPowerCap(context.Background(), &limit))

	require.Len(t, patches, 1, "expected SetPowerCap to PATCH the Power resource once")
	require.Len(t, patches[0].PowerControl, 1)
	require.NotNil(t, patches[0].PowerControl[0].PowerLimit.LimitInWatts)
	assert.Equal(t, limit, *patches[0].PowerControl[0].PowerLimit.LimitInWatts)
}

func TestSetPowerCapClear(t *testing.T) {
	var patches []powerPatch

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1/Power", dellPowerHandler(t, &patches))
	client := newDellPowerClient(t, mux)

	require.NoError(t, client.SetPowerCap(context.Background(), nil))

	require.Len(t, patches, 1, "expected SetPowerCap(nil) to PATCH the Power resource once")
	require.Len(t, patches[0].PowerControl, 1)
	assert.Nil(t, patches[0].PowerControl[0].PowerLimit.LimitInWatts, "expected LimitInWatts: null to clear the cap")
}

func TestSetPowerCapRejected(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1/Power", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(mustReadFile(t, "dell/power.json"))
		case http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"LimitInWatts out of range"}}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	client := newDellPowerClient(t, mux)

	limit := 1.0
	err := client.SetPowerCap(context.Background(), &limit)
	assert.ErrorIs(t, err, bmclibErrs.ErrPowerCapSet)
	assert.ErrorContains(t, err, "LimitInWatts out of range")
}

func TestPowerControlNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/Chassis/System.Embedded.1/Power", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	client := newDellPowerClient(t, mux)

	_, err := client.GetPowerMetrics(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrPowerControlNotFound)

	err = client.SetPowerCap(context.Background(), nil)
	assert.ErrorIs(t, err, bmclibErrs.ErrPowerControlNotFound)
}
