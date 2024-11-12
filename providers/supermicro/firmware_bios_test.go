package supermicro

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-logr/logr"
	"github.com/metal-toolbox/bmclib/internal/httpclient"
	"github.com/stretchr/testify/assert"
)

func Test_setComponentUpdateMisc(t *testing.T) {
	testcases := []struct {
		name          string
		stage         string
		errorContains string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"preUpdate",
			"preUpdate",
			"",
			"/cgi/ipmi.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/x-www-form-urlencoded; charset=UTF-8", r.Header.Get("Content-Type"))

				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, `op=COMPONENT_UPDATE_MISC.XML&r=(0,0)&_=`, string(b))

				_, _ = w.Write([]byte(`<?xml version="1.0"?>
				<IPMI>
				  <MISC_INFO RES="-1" SYSOFF="0"/>
				</IPMI>`))
			},
		},
		{
			"postUpdate",
			"postUpdate",
			"",
			"/cgi/ipmi.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/x-www-form-urlencoded; charset=UTF-8", r.Header.Get("Content-Type"))

				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, `op=COMPONENT_UPDATE_MISC.XML&r=(1,0)&_=`, string(b))

				_, _ = w.Write([]byte(`<?xml version="1.0"?>
				<IPMI>
				  <MISC_INFO RES="0" SYSOFF="0"/>
				</IPMI>`))
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(tc.endpoint, tc.handler)

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			serviceClient, err := newBmcServiceClient(parsedURL.Hostname(), parsedURL.Port(), "foo", "bar", httpclient.Build())
			assert.Nil(t, err)

			serviceClient.csrfToken = "foobar"
			client := &x11{serviceClient: serviceClient, log: logr.Discard()}

			if err := client.checkComponentUpdateMisc(context.Background(), tc.stage); err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}

				assert.Nil(t, err)
			}
		})
	}
}

func Test_setBIOSFirmwareInstallMode(t *testing.T) {
	testcases := []struct {
		name          string
		errorContains string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"BIOS fw install lock acquired",
			"",
			"/cgi/ipmi.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/x-www-form-urlencoded; charset=UTF-8", r.Header.Get("Content-Type"))

				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, `op=LOCK_UPLOAD_FW.XML&r=(0,0)&_=`, string(b))

				_, _ = w.Write([]byte(`<?xml version="1.0"?>
				<IPMI>
				  <LOCK_FW_UPLOAD RES="1"/>
				</IPMI>`))
			},
		},
		{
			"lock not acquired",
			"BMC cold reset required",
			"/cgi/ipmi.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/x-www-form-urlencoded; charset=UTF-8", r.Header.Get("Content-Type"))

				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, `op=LOCK_UPLOAD_FW.XML&r=(0,0)&_=`, string(b))

				_, _ = w.Write([]byte(`<?xml version="1.0"?>
				<IPMI>
				  <LOCK_FW_UPLOAD RES="0"/>
				</IPMI>`))
			},
		},
		{
			"error returned",
			"400",
			"/cgi/ipmi.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(400)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.HandleFunc(tc.endpoint, tc.handler)

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			serviceClient, err := newBmcServiceClient(parsedURL.Hostname(), parsedURL.Port(), "foo", "bar", httpclient.Build())
			assert.Nil(t, err)

			serviceClient.csrfToken = "foobar"
			client := &x11{serviceClient: serviceClient, log: logr.Discard()}

			if err := client.setBMCFirmwareInstallMode(context.Background()); err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}

				assert.Nil(t, err)
			}
		})
	}
}
