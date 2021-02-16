package discover

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/bmc-toolbox/bmclib/providers/asrockrack"
)

var (
	_bmc_endpoints = map[string]map[string][]byte{
		"asrockrack": {"/": []byte(`ASRockRack`)},
	}
)

func mockBMC() *httptest.Server {
	handler := http.NewServeMux()

	for _, endpointResponse := range _bmc_endpoints {
		for endpoint, response := range endpointResponse {
			handler.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(response)
			})
		}
	}
	return httptest.NewTLSServer(handler)
}

func TestProbesv2(t *testing.T) {
	testt := []struct {
		name     string
		wantHint string
		wantType interface{}
	}{
		{
			name:     "asrockrack",
			wantHint: ProbeASRockRack,
			wantType: (*asrockrack.ASRockRack)(nil),
		},
	}

	server := mockBMC()
	mockURL, _ := url.Parse(server.URL)
	logger := logging.DefaultLogger()

	for _, tt := range testt {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// set Options that will be passed to each probe
			opts := &Options{
				Host:    mockURL.Host,
				Context: context.TODO(),
				Logger:  logger,
				Hint:    tt.name,
			}

			// setup probe
			probe, err := NewProbev2(opts)
			if err != nil {
				t.Fatalf("error calling NewProbev2: %v", err)
			}

			var probeResponse interface{}
			switch tt.name {
			case "asrockrack":
				probeResponse, err = probe.asRockRack(context.TODO())
			}

			if err != nil {
				t.Fatalf("error running %s probe: %v", tt.name, err)
			}

			if reflect.TypeOf(tt.wantType) != reflect.TypeOf(probeResponse) {
				t.Errorf("Want %T, got %T", tt.wantType, probeResponse)
			}
		})
	}
}
