package compal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stmcginnis/gofish/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	selfPath = "/redfish/v1/Systems/Self"
	hgxPath  = "/redfish/v1/Systems/HGX_Baseboard_0"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("fixtures", name))
	require.NoError(t, err)
	return body
}

func fixtureWith(t *testing.T, name, field, value string) []byte {
	t.Helper()
	var doc map[string]any
	require.NoError(t, json.Unmarshal(fixture(t, name), &doc))
	doc[field] = value
	body, err := json.Marshal(doc)
	require.NoError(t, err)
	return body
}

func newTestConn(t *testing.T, overrides map[string][]byte, resetTargets *[]string) *Conn {
	t.Helper()
	bodies := map[string][]byte{
		"/redfish/v1/":                fixture(t, "serviceroot.json"),
		"/redfish/v1/Systems":         fixture(t, "systems.json"),
		selfPath:                      fixture(t, "system_self.json"),
		selfPath + "/ResetActionInfo": fixture(t, "system_self_resetactioninfo.json"),
		hgxPath:                       fixture(t, "system_hgx_baseboard_0.json"),
	}
	for path, body := range overrides {
		bodies[path] = body
	}

	mux := http.NewServeMux()
	for path, body := range bodies {
		mux.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(body) })
	}
	mux.HandleFunc(selfPath+"/Actions/ComputerSystem.Reset", func(w http.ResponseWriter, r *http.Request) {
		var req struct{ ResetType string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		*resetTargets = append(*resetTargets, r.URL.Path+" "+req.ResetType)
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	require.NoError(t, err)

	return New(u.Hostname(), "", "", logr.Discard(), WithPort(u.Port()), WithUseBasicAuth(true))
}

func TestPowerSetTargetsHostSystem(t *testing.T) {
	tests := map[string]map[string][]byte{
		"compal fixtures":                     nil,
		"baseboard has the same name as host": {hgxPath: fixtureWith(t, "system_hgx_baseboard_0.json", "Name", "System")},
	}

	for name, overrides := range tests {
		t.Run(name, func(t *testing.T) {
			var resetTargets []string
			conn := newTestConn(t, overrides, &resetTargets)
			ctx := context.Background()
			require.NoError(t, conn.Open(ctx))
			t.Cleanup(func() { _ = conn.Close(ctx) })

			ok, err := conn.PowerSet(ctx, "soft")
			require.NoError(t, err)
			assert.True(t, ok)
			assert.Equal(t, []string{selfPath + "/Actions/ComputerSystem.Reset GracefulShutdown"}, resetTargets)
		})
	}
}

func TestCompatible(t *testing.T) {
	tests := map[string]struct {
		overrides map[string][]byte
		want      bool
	}{
		"compal host system": {want: true},
		"host system from another vendor": {
			overrides: map[string][]byte{selfPath: fixtureWith(t, "system_self.json", "Manufacturer", "Dell Inc.")},
			want:      false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			conn := newTestConn(t, tc.overrides, nil)
			assert.Equal(t, tc.want, conn.Compatible(context.Background()))
		})
	}
}

func TestHostSystem(t *testing.T) {
	const reset = `{"Actions":{"#ComputerSystem.Reset":{"target":"/redfish/v1/Systems/Self/Actions/ComputerSystem.Reset"}}}`
	system := func(id, name, manufacturer, raw string) *schemas.ComputerSystem {
		return &schemas.ComputerSystem{Entity: schemas.Entity{ID: id, Name: name}, Manufacturer: manufacturer, RawData: []byte(raw)}
	}
	hgx := system("HGX_Baseboard_0", "HGX_Baseboard_0", "NVIDIA", `{}`)

	tests := map[string]struct {
		systems []*schemas.ComputerSystem
		wantID  string
		wantErr string
	}{
		"host is the only resettable system": {
			systems: []*schemas.ComputerSystem{hgx, system("Self", "System", "COMPAL", reset)},
			wantID:  "Self",
		},
		"no resettable system": {
			systems: []*schemas.ComputerSystem{hgx},
			wantErr: "found 0 among: HGX_Baseboard_0 (HGX_Baseboard_0)",
		},
		"two resettable systems": {
			systems: []*schemas.ComputerSystem{system("HGX_Baseboard_0", "HGX_Baseboard_0", "NVIDIA", reset), system("Self", "System", "COMPAL", reset)},
			wantErr: "found 2 among: HGX_Baseboard_0 (HGX_Baseboard_0), Self (System)",
		},
		"reset action without a target": {
			systems: []*schemas.ComputerSystem{hgx, system("Self", "System", "COMPAL", `{"Actions":{"#ComputerSystem.Reset":{}}}`)},
			wantErr: `reset action of system "Self" has no target`,
		},
		"resettable system from another vendor": {
			systems: []*schemas.ComputerSystem{hgx, system("Self", "System", "Dell Inc.", reset)},
			wantErr: `system "Self" is made by "Dell Inc.", not Compal`,
		},
		"unreadable system": {
			systems: []*schemas.ComputerSystem{system("Self", "System", "COMPAL", "{")},
			wantErr: `read reset action of system "Self"`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			host, err := hostSystem(tc.systems)
			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantID, host.ID)
		})
	}
}
