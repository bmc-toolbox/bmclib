package redfish

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-logr/logr"
)

const (
	fixturesDir = "./fixtures"
)

var (
	mockServer  *httptest.Server
	mockBMCHost *url.URL
	mockClient  *Conn
)

// jsonResponse returns the fixture json response for a request URI
func jsonResponse(endpoint string) []byte {
	jsonResponsesMap := map[string]string{
		"/redfish/v1/Managers":                                            fixturesDir + "/v1/dell/managers.json",
		"/redfish/v1/Managers/iDRAC.Embedded.1":                           fixturesDir + "/v1/dell/manager.idrac.embedded.1.json",
		"/redfish/v1/Managers/iDRAC.Embedded.1/LogServices":               fixturesDir + "/v1/dell/logservices.json",
		"/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel":           fixturesDir + "/v1/dell/logservices.sel.json",
		"/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries":   fixturesDir + "/v1/dell/entries.json",
		"/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries/1": fixturesDir + "/v1/dell/selentries/1.json",
		"/redfish/v1/Managers/iDRAC.Embedded.1/LogServices/Sel/Entries/2": fixturesDir + "/v1/dell/selentries/2.json",

		"/redfish/v1/":                          fixturesDir + "/v1/serviceroot.json",
		"/redfish/v1/UpdateService":             fixturesDir + "/v1/updateservice.json",
		"/redfish/v1/Systems":                   fixturesDir + "/v1/systems.json",
	}

	fh, err := os.Open(jsonResponsesMap[endpoint])
	if err != nil {
		log.Fatalf("%s, failed to open fixture: %s for endpoint: %s", err.Error(), jsonResponsesMap[endpoint], endpoint)
	}

	defer fh.Close()

	b, err := io.ReadAll(fh)
	if err != nil {
		log.Fatalf("%s, failed to read fixture: %s for endpoint: %s", err.Error(), jsonResponsesMap[endpoint], endpoint)
	}

	return b
}

func TestMain(m *testing.M) {
	// setup mock server
	mockServer = func() *httptest.Server {
		handler := http.NewServeMux()
		handler.HandleFunc("/redfish/v1/", serviceRoot)
		handler.HandleFunc("/redfish/v1/SessionService/Sessions", sessionService)
		return httptest.NewTLSServer(handler)
	}()

	mockBMCHost, _ = url.Parse(mockServer.URL)
	mockClient = New(mockBMCHost.Hostname(), "foo", "bar", logr.Discard(), WithPort(mockBMCHost.Port()))
	err := mockClient.Open(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func serviceRoot(w http.ResponseWriter, r *http.Request) {
	// expect either GET or Delete methods
	if r.Method != http.MethodGet && r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusNotFound)
	}

	_, _ = w.Write(jsonResponse(r.RequestURI))
}

func sessionService(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
	}

	_, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("X-Auth-Token", "hunter2")
	w.Header().Set("Location", r.URL.String())
	_, _ = w.Write([]byte(`is cool`))
}
