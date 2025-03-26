package redfishwrapper

import (
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

func mustReadFile(t *testing.T, filename string) []byte {
	t.Helper()

	fixture := fixturesDir + "/" + filename
	fh, err := os.Open(fixture)
	if err != nil {
		log.Fatal(err)
	}

	defer fh.Close()

	b, err := io.ReadAll(fh)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

var endpointFunc = func(t *testing.T, file string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		t.Helper()
		if file == "404" {
			w.WriteHeader(http.StatusNotFound)
		}

		// expect either GET or Delete methods
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		_, _ = w.Write(mustReadFile(t, file))
	}
}
