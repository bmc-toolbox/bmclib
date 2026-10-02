package redfishwrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stmcginnis/gofish/schemas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
)

func TestWithVersionsNotCompatible(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name     string
		versions []string
	}{
		{
			"no versions",
			[]string{},
		},
		{
			"with versions",
			[]string{"1.2.3", "4.5.6"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithVersionsNotCompatible(tt.versions))
			assert.Equal(t, tt.versions, client.versionsNotCompatible)
		})
	}
}

func TestWithBasicAuthEnabled(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			"disabled",
			false,
		},
		{
			"enabled",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithBasicAuthEnabled(tt.enabled))
			assert.Equal(t, tt.enabled, client.basicAuth)
		})
	}
}

func TestWithEtagMatchDisabled(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name     string
		disabled bool
	}{
		{
			"disabled",
			true,
		},
		{
			"enabled",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithEtagMatchDisabled(tt.disabled))
			assert.Equal(t, tt.disabled, client.disableEtagMatch)
		})
	}
}

const (
	fixturesDir = "./fixtures"
)

func TestManagerOdataID(t *testing.T) {
	tests := map[string]struct {
		hfunc  map[string]func(http.ResponseWriter, *http.Request)
		expect string
		err    error
	}{
		"happy case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				// service root
				"/redfish/v1/":           endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Systems":    endpointFunc(t, "systems.json"),
				"/redfish/v1/Managers":   endpointFunc(t, "managers.json"),
				"/redfish/v1/Managers/1": endpointFunc(t, "managers_1.json"),
			},
			expect: "/redfish/v1/Managers/1",
			err:    nil,
		},
		"failure case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/": endpointFunc(t, "/serviceroot_no_manager.json"),
			},
			expect: "",
			err:    ErrManagerID,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			handleFunc := tc.hfunc
			for endpoint, handler := range handleFunc {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			// os.Setenv("DEBUG_BMCLIB", "true")
			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "")

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			got, err := client.ManagerOdataID(ctx)
			if err != nil {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.expect, got)

			client.Close(context.Background())
		})
	}
}

func TestSystemsBIOSOdataID(t *testing.T) {
	tests := map[string]struct {
		hfunc  map[string]func(http.ResponseWriter, *http.Request)
		expect string
		err    error
	}{
		"happy case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				// service root
				"/redfish/v1/":               endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Systems":        endpointFunc(t, "systems.json"),
				"/redfish/v1/Systems/1":      endpointFunc(t, "systems_1.json"),
				"/redfish/v1/Systems/1/Bios": endpointFunc(t, "systems_bios.json"),
			},
			expect: "/redfish/v1/Systems/1/Bios",
			err:    nil,
		},
		"failure case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/": endpointFunc(t, "serviceroot.json"),
			},
			expect: "",
			err:    ErrBIOSID,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			handleFunc := tc.hfunc
			for endpoint, handler := range handleFunc {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "")

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			got, err := client.SystemsBIOSOdataID(ctx)
			if err != nil {
				assert.Equal(t, tc.err, err)
			}

			assert.Equal(t, tc.expect, got)

			client.Close(context.Background())
		})
	}
}

func TestRedfishVersionMeetsOrExceeds(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		version string
		exp     bool
	}{
		{
			"empty string",
			"",
			false,
		},
		{
			"short string",
			"1.2",
			false,
		},
		{
			"bogus component",
			"1.asdf.2",
			false,
		},
		{
			"major too low",
			"0.3.4",
			false,
		},
		{
			"minor too low",
			"1.1.3",
			false,
		},
		{
			"patch too low",
			"1.2.2",
			false,
		},
		{
			"meets",
			"1.2.3",
			true,
		},
		{
			"exceeds",
			"1.2.4",
			true,
		},
	}

	for _, tc := range testCases {
		got := redfishVersionMeetsOrExceeds(tc.version, 1, 2, 3)
		assert.Equal(t, tc.exp, got, "testcase %s", tc.name)
	}
}

func TestGetBootProgress(t *testing.T) {
	tests := map[string]struct {
		hfunc  map[string]func(http.ResponseWriter, *http.Request)
		expect []*schemas.BootProgress
		err    error
	}{
		"happy case": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				// service root
				"/redfish/v1/":          endpointFunc(t, "smc_1.14.0_serviceroot.json"),
				"/redfish/v1/Systems":   endpointFunc(t, "smc_1.14.0_systems.json"),
				"/redfish/v1/Systems/1": endpointFunc(t, "smc_1.14.0_systems_1.json"),
			},
			expect: []*schemas.BootProgress{
				{
					LastState: schemas.SystemHardwareInitializationCompleteBootProgressTypes,
				},
			},
			err: nil,
		},
		"insufficient redfish version": {
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/": endpointFunc(t, "smc_1.9.0_serviceroot.json"),
			},
			expect: nil,
			err:    bmclibErrs.ErrRedfishVersionIncompatible,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			handleFunc := tc.hfunc
			for endpoint, handler := range handleFunc {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "")

			err = client.Open(context.TODO())
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close(context.TODO())

			got, err := client.GetBootProgress()
			if err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}

			assert.ElementsMatch(t, tc.expect, got)
		})
	}
}

func TestOpenRestoresHTTPClientTimeout(t *testing.T) {
	// Open bounds the connect by the ctx deadline via the HTTP client's Timeout,
	// because gofish ignores per-call contexts. That deadline must not outlive
	// Open: the same client serves every later call on the connection, and a
	// connect deadline is typically much shorter than a slow-but-valid operation.
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "systems.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "managers.json"))

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	assert.NoError(t, err)

	httpClient := httpclient.Build()
	configured := httpClient.Timeout
	assert.NotZero(t, configured, "test needs a non-zero configured timeout to be meaningful")

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithHTTPClient(httpClient))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	assert.NoError(t, client.Open(ctx))
	defer client.Close(ctx)

	assert.Equal(t, configured, httpClient.Timeout)
}

func TestOpenBoundsConnectByContext(t *testing.T) {
	// Complements TestOpenRestoresHTTPClientTimeout: that test proves the configured
	// timeout isn't left mutated after Open returns, but a no-op fix (never touching
	// HTTPClient.Timeout at all) would pass it just as trivially. This test proves the
	// other half - that Open actually still bounds the connect by the shorter ctx
	// deadline, not by the client's much longer configured timeout - by giving the
	// server a response delay in between the two and asserting Open fails around the
	// short ctx deadline rather than hanging until the long configured one.
	const ctxDeadline = 100 * time.Millisecond
	const serverDelay = 500 * time.Millisecond
	const configuredTimeout = 10 * time.Second

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(serverDelay)
		endpointFunc(t, "serviceroot.json")(w, r)
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	assert.NoError(t, err)

	httpClient := httpclient.Build()
	httpClient.Timeout = configuredTimeout

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithHTTPClient(httpClient))

	ctx, cancel := context.WithTimeout(context.Background(), ctxDeadline)
	defer cancel()

	start := time.Now()
	err = client.Open(ctx)
	elapsed := time.Since(start)

	assert.Error(t, err, "expected the connect to fail once it outlives the ctx deadline")
	assert.Less(t, elapsed, serverDelay, "expected Open to fail before the server even responds, bounded by ctx rather than by the much longer configured timeout")
}

func TestJobs(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Members": [{"@odata.id": "/redfish/v1/JobService/Jobs/JID_1"}], "Members@odata.count": 1}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_1", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Id": "JID_1", "@odata.id": "/redfish/v1/JobService/Jobs/JID_1", "Name": "ConfigBIOS:BIOS.Setup.1-1", "JobState": "Scheduled"}`))
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(ctx))
	defer client.Close(ctx)

	jobs, err := client.Jobs(ctx)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, "JID_1", jobs[0].ID)
	assert.Equal(t, "ConfigBIOS:BIOS.Setup.1-1", jobs[0].Name)
	assert.Equal(t, "Scheduled", string(jobs[0].JobState))
}

func TestJobsWithoutJobService(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(ctx))
	defer client.Close(ctx)

	_, err = client.Jobs(ctx)
	assert.Error(t, err)
}

// jobServiceMux serves a JobService with one job. If expand is true the collection answers
// $expand with the members inline, and reading the job on its own then fails the test; otherwise
// the collection answers with links only, whatever it was asked for.
func jobServiceMux(t *testing.T, expand bool) *http.ServeMux {
	t.Helper()

	const job = `{"Id": "JID_1", "@odata.id": "/redfish/v1/JobService/Jobs/JID_1", "Name": "ConfigBIOS:BIOS.Setup.1-1", "JobState": "Scheduled"}`

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
		if expand && r.URL.Query().Has("$expand") {
			_, _ = w.Write([]byte(`{"Members": [` + job + `], "Members@odata.count": 1}`))
			return
		}
		if r.URL.Query().Has("$expand") && !expand {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"Members": [{"@odata.id": "/redfish/v1/JobService/Jobs/JID_1"}], "Members@odata.count": 1}`))
	})
	mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_1", func(w http.ResponseWriter, r *http.Request) {
		if expand {
			t.Error("job read on its own although the collection was returned expanded")
		}
		_, _ = w.Write([]byte(job))
	})

	return mux
}

func TestJobsReadsTheExpandedCollectionInOneRequest(t *testing.T) {
	server := httptest.NewTLSServer(jobServiceMux(t, true))
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(ctx))
	defer client.Close(ctx)

	jobs, err := client.Jobs(ctx)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, "JID_1", jobs[0].ID)
	assert.Equal(t, "ConfigBIOS:BIOS.Setup.1-1", jobs[0].Name)
	assert.Equal(t, "Scheduled", string(jobs[0].JobState))
	assert.Equal(t, "/redfish/v1/JobService/Jobs/JID_1", jobs[0].ODataID, "expected the job's own URI, needed to act on it")
}

func TestJobsFallsBackToReadingEachJobWhenExpandIsRefused(t *testing.T) {
	server := httptest.NewTLSServer(jobServiceMux(t, false))
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(ctx))
	defer client.Close(ctx)

	jobs, err := client.Jobs(ctx)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
	assert.Equal(t, "JID_1", jobs[0].ID)
}

// TestJobsFallsBackWhenTheExpandedCollectionIsIncomplete: an expanded answer that is paginated,
// reports more members than it holds, or has a member that cannot be acted on must not be taken
// at face value, since a job missing from it would go unnoticed.
func TestJobsFallsBackWhenTheExpandedCollectionIsIncomplete(t *testing.T) {
	const job = `{"Id": "JID_1", "@odata.id": "/redfish/v1/JobService/Jobs/JID_1", "Name": "ConfigBIOS:BIOS.Setup.1-1", "JobState": "Scheduled"}`

	tests := map[string]string{
		"next link":          `{"Members": [` + job + `], "Members@odata.count": 1, "Members@odata.nextLink": "/redfish/v1/JobService/Jobs?$skip=1"}`,
		"count mismatch":     `{"Members": [` + job + `], "Members@odata.count": 2}`,
		"member without uri": `{"Members": [{"Id": "JID_1", "Name": "ConfigBIOS:BIOS.Setup.1-1", "JobState": "Scheduled"}], "Members@odata.count": 1}`,
	}

	for name, expanded := range tests {
		t.Run(name, func(t *testing.T) {
			var readOnOwn bool

			mux := http.NewServeMux()
			mux.HandleFunc("/redfish/v1/", endpointFunc(t, "/dell/serviceroot.json"))
			mux.HandleFunc("/redfish/v1/JobService", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"Jobs": {"@odata.id": "/redfish/v1/JobService/Jobs"}}`))
			})
			mux.HandleFunc("/redfish/v1/JobService/Jobs", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Has("$expand") {
					_, _ = w.Write([]byte(expanded))
					return
				}
				_, _ = w.Write([]byte(`{"Members": [{"@odata.id": "/redfish/v1/JobService/Jobs/JID_1"}], "Members@odata.count": 1}`))
			})
			mux.HandleFunc("/redfish/v1/JobService/Jobs/JID_1", func(w http.ResponseWriter, r *http.Request) {
				readOnOwn = true
				_, _ = w.Write([]byte(job))
			})

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			ctx := context.Background()
			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
			require.NoError(t, client.Open(ctx))
			defer client.Close(ctx)

			jobs, err := client.Jobs(ctx)
			require.NoError(t, err)
			require.Len(t, jobs, 1)
			assert.Equal(t, "/redfish/v1/JobService/Jobs/JID_1", jobs[0].ODataID)
			assert.True(t, readOnOwn, "expected the jobs to be read one by one instead")
		})
	}
}
