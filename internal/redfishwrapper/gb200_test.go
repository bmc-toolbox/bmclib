package redfishwrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// The tests in this file pin how the client behaves today against a Supermicro
// GB200 NVL compute tray (BMC firmware 70.00.03.19, HGX firmware GB200Nvl-25.06-A).
//
// The tray lists two systems (System_0, then HGX_Baseboard_0), two managers
// (BMC_0, HGX_BMC_0) and 46 chassis. Several tests document behavior that is
// wrong for this hardware; those say so, and a fix should change their assertions.

// gb200FixtureRoot holds the Redfish mockup of the tray.
var gb200FixtureRoot = "fixtures/smc_gb200"

const gb200SELPath = "/redfish/v1/Systems/System_0/LogServices/SEL"

type mockupRequest struct {
	Method string
	Path   string
}

// mockupServer serves a Redfish-Mockup-Creator tree, where the resource at
// /redfish/v1/X lives in <root>/redfish/v1/X/index.json, and records every request.
// Requests other than GET succeed with 204 so tests can assert on what was sent.
type mockupServer struct {
	root     string
	mu       sync.Mutex
	requests []mockupRequest
}

func (m *mockupServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := path.Clean(r.URL.Path)

	m.mu.Lock()
	m.requests = append(m.requests, mockupRequest{Method: r.Method, Path: p})
	m.mu.Unlock()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := os.ReadFile(filepath.Join(m.root, filepath.FromSlash(p), "index.json"))
	if err != nil {
		notFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

func (m *mockupServer) Requests() []mockupRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]mockupRequest(nil), m.requests...)
}

// requested reports whether any request was made for p or a resource under it.
func (m *mockupServer) requested(p string) bool {
	for _, r := range m.Requests() {
		if r.Path == p || strings.HasPrefix(r.Path, p+"/") {
			return true
		}
	}

	return false
}

// writes returns the requests that were not GETs.
func (m *mockupServer) writes() []mockupRequest {
	var out []mockupRequest
	for _, r := range m.Requests() {
		if r.Method != http.MethodGet {
			out = append(out, r)
		}
	}

	return out
}

// newGB200Client opens a client against the GB200 mockup.
func newGB200Client(t *testing.T, opts ...Option) (*Client, *mockupServer) {
	t.Helper()

	mockup := &mockupServer{root: gb200FixtureRoot}
	server := httptest.NewTLSServer(mockup)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	opts = append([]Option{WithBasicAuthEnabled(true)}, opts...)
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", opts...)

	ctx := context.Background()
	require.NoError(t, client.Open(ctx))
	t.Cleanup(func() { _ = client.Close(ctx) })

	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		for _, r := range mockup.Requests() {
			t.Logf("mockup request: %s %s", r.Method, r.Path)
		}
	})

	return client, mockup
}

// With two systems and two managers, the client needs a system name to pick one.
func TestGB200SystemSelection(t *testing.T) {
	ctx := context.Background()

	t.Run("without a system name", func(t *testing.T) {
		client, _ := newGB200Client(t)

		_, err := client.SystemPowerStatus(ctx)
		require.ErrorContains(t, err, "no matching redfish system found")

		_, err = client.Manager(ctx)
		require.ErrorContains(t, err, "no matching redfish manager found")
	})

	t.Run("with System_0", func(t *testing.T) {
		client, _ := newGB200Client(t, WithSystemName("System_0"))

		state, err := client.SystemPowerStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, "On", state)

		manager, err := client.Manager(ctx)
		require.NoError(t, err)
		assert.Equal(t, "BMC_0", manager.ID)
	})
}

// Bug: ClearSystemEventLog walks chassis log services only. The first one it meets
// is the GPU XID log, which has no ClearLog action, so the call fails before any
// write. The SEL under System_0, which does have a ClearLog action, is never touched.
func TestGB200ClearSystemEventLog(t *testing.T) {
	client, mockup := newGB200Client(t, WithSystemName("System_0"))

	err := client.ClearSystemEventLog(context.Background())
	require.ErrorContains(t, err, "no target provided")

	assert.True(t, mockup.requested("/redfish/v1/Chassis/HGX_GPU_0/LogServices/XID"))
	assert.Empty(t, mockup.writes())
	assert.False(t, mockup.requested(gb200SELPath))
}

// Bug: GetSystemEventLog reads every log service of every manager. It returns the
// 5 BMC_0 EventLog entries mixed with 5 BMC dump entries, and never reads the
// 5 entries of the SEL under System_0.
func TestGB200GetSystemEventLog(t *testing.T) {
	client, mockup := newGB200Client(t, WithSystemName("System_0"))

	entries, err := client.GetSystemEventLog(context.Background())
	require.NoError(t, err)
	assert.Len(t, entries, 10)

	assert.True(t, mockup.requested("/redfish/v1/Managers/BMC_0/LogServices/Dump/Entries"))
	assert.True(t, mockup.requested("/redfish/v1/Managers/HGX_BMC_0/LogServices/Dump/Entries"))
	assert.False(t, mockup.requested(gb200SELPath))
}

// Bug: Inventory accepts only chassis, system and manager IDs from a hardcoded
// allowlist, and none of the tray's IDs are on it. Without failOnError, it returns
// a device with no vendor, CPUs, GPUs or BMC, and no error. The tray's CPUs and
// GPUs live under HGX_Baseboard_0, which inventory would not read either.
func TestGB200Inventory(t *testing.T) {
	ctx := context.Background()

	t.Run("failOnError false", func(t *testing.T) {
		client, _ := newGB200Client(t, WithSystemName("System_0"))

		device, err := client.Inventory(ctx, false)
		require.NoError(t, err)

		assert.Empty(t, device.Vendor)
		assert.Empty(t, device.Model)
		assert.Empty(t, device.CPUs)
		assert.Empty(t, device.GPUs)
		require.NotNil(t, device.BMC)
		assert.Empty(t, device.BMC.ID)
	})

	t.Run("failOnError true", func(t *testing.T) {
		client, _ := newGB200Client(t, WithSystemName("System_0"))

		_, err := client.Inventory(ctx, true)
		require.ErrorIs(t, err, bmclibErrs.ErrRedfishChassisOdataID)
	})
}

// Bug: GetChassisTemperatures returns the readings of the first chassis with
// ThermalMetrics. On this tray that is BMC_0, so the result is the BMC board
// temperature alone, with no error.
func TestGB200ChassisTemperatures(t *testing.T) {
	client, _ := newGB200Client(t, WithSystemName("System_0"))

	temps, err := client.GetChassisTemperatures(context.Background())
	require.NoError(t, err)
	require.Len(t, temps, 1)

	assert.Equal(t, "BMC_0_DCSCM_Temp_0", temps[0].Name)
	require.NotNil(t, temps[0].ReadingCelsius)
	assert.InDelta(t, 23.5, *temps[0].ReadingCelsius, 0.001)
}
