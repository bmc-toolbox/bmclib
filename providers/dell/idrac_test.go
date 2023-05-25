package dell

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

const (
	fixturesDir = "./fixtures"
)

func Test_Screenshot(t *testing.T) {
	// byte slice instead of a real image
	img := []byte(`foobar`)

	// endpoint to handler funcs
	type handlerFuncMap map[string]func(http.ResponseWriter, *http.Request)

	testcases := []struct {
		name           string
		imgbytes       []byte
		handlerFuncMap handlerFuncMap
	}{
		{
			"happy path",
			[]byte(`foobar`),
			handlerFuncMap{
				// service root
				"/redfish/v1/": func(w http.ResponseWriter, r *http.Request) {
					// expect either GET or Delete methods
					if r.Method != http.MethodGet && r.Method != http.MethodDelete {
						w.WriteHeader(http.StatusNotFound)
					}

					fixture := fixturesDir + "/serviceroot.json"
					fh, err := os.Open(fixture)
					if err != nil {
						log.Fatal(err)
					}

					defer fh.Close()

					b, err := io.ReadAll(fh)
					if err != nil {
						log.Fatal(err)
					}

					_, _ = w.Write(b)
				},

				"/redfish/v1/Systems": func(w http.ResponseWriter, r *http.Request) {
					// expect either GET or Delete methods
					if r.Method != http.MethodGet && r.Method != http.MethodDelete {
						w.WriteHeader(http.StatusNotFound)
					}

					fixture := fixturesDir + "/systems.json"
					fh, err := os.Open(fixture)
					if err != nil {
						log.Fatal(err)
					}

					defer fh.Close()

					b, err := io.ReadAll(fh)
					if err != nil {
						log.Fatal(err)
					}

					_, _ = w.Write(b)
				},

				"/redfish/v1/Systems/System.Embedded.1": func(w http.ResponseWriter, r *http.Request) {
					// expect either GET or Delete methods
					if r.Method != http.MethodGet && r.Method != http.MethodDelete {
						w.WriteHeader(http.StatusNotFound)
					}

					fixture := fixturesDir + "/systems_embedded.1.json"
					fh, err := os.Open(fixture)
					if err != nil {
						log.Fatal(err)
					}

					defer fh.Close()

					b, err := io.ReadAll(fh)
					if err != nil {
						log.Fatal(err)
					}

					_, _ = w.Write(b)
				},

				// screenshot endpoint
				redfishV1Prefix + screenshotEndpoint: func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, r.Method, http.MethodPost)

					assert.Equal(t, r.Header.Get("Content-Type"), "application/json")

					b, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatal(err)
					}
					assert.Equal(t, []byte(`{"FileType":"ServerScreenShot"}`), b)

					encoded := base64.RawStdEncoding.EncodeToString(img)
					respFmtStr := `{"@Message.ExtendedInfo":[{"Message":"Successfully Completed Request","MessageArgs":[],"MessageArgs@odata.count":0,"MessageId":"Base.1.8.Success","RelatedProperties":[],"RelatedProperties@odata.count":0,"Resolution":"None","Severity":"OK"},{"Message":"The Export Server Screen Shot operation successfully exported the server screen shot file.","MessageArgs":[],"MessageArgs@odata.count":0,"MessageId":"IDRAC.2.5.LC080","RelatedProperties":[],"RelatedProperties@odata.count":0,"Resolution":"Download the encoded Base64 format server screen shot file, decode the Base64 file and then save it as a *.png file.","Severity":"Informational"}],"ServerScreenshotFile":"%s"}`

					_, _ = w.Write([]byte(fmt.Sprintf(respFmtStr, encoded)))
				},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()

			for endpoint, handler := range tc.handlerFuncMap {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			//os.Setenv("DEBUG_BMCLIB", "true")
			client := New(parsedURL.Hostname(), "", "", logr.Discard(), WithPort(parsedURL.Port()), WithUseBasicAuth(true))

			err = client.Open(context.TODO())
			if err != nil {
				t.Fatal(err)
			}

			img, fileType, err := client.Screenshot(context.TODO())
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.imgbytes, img)
			assert.Equal(t, "png", fileType)
		})
	}
}
