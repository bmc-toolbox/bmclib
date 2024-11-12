package supermicro

import (
	"bytes"
	"context"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"github.com/metal-toolbox/bmclib/constants"
	"github.com/metal-toolbox/bmclib/internal/httpclient"
	"github.com/stretchr/testify/assert"
)

func TestX11SetBMCFirmwareInstallMode(t *testing.T) {
	testcases := []struct {
		name          string
		errorContains string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"BMC fw install lock acquired",
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

func TestX11UploadBMCFirmware(t *testing.T) {
	testcases := []struct {
		name           string
		errorContains  string
		endpoint       string
		fwFilename     string
		fwFileContents string
		handler        func(http.ResponseWriter, *http.Request)
	}{
		{
			"upload works",
			"",
			"/cgi/oem_firmware_upload.cgi",
			"blob.bin",
			"dummy fw image",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				// validate content type
				mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				assert.Nil(t, err)

				assert.Equal(t, "multipart/form-data", mediaType)

				// read form parts from boundary
				reader := multipart.NewReader(bytes.NewReader(b), params["boundary"])

				// validate firmware image part
				part, err := reader.NextPart()
				assert.Nil(t, err)

				assert.Equal(t, `form-data; name="fw_image"; filename="blob.bin"`, part.Header.Get("Content-Disposition"))

				// validate csrf-token part
				part, err = reader.NextPart()
				assert.Nil(t, err)

				assert.Equal(t, `form-data; name="CSRF_TOKEN"`, part.Header.Get("Content-Disposition"))
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

			// create tmp firmware file
			var fwReader *os.File
			if tc.fwFilename != "" {
				tmpdir := t.TempDir()
				binPath := filepath.Join(tmpdir, tc.fwFilename)
				err := os.WriteFile(binPath, []byte(tc.fwFileContents), 0600)
				if err != nil {
					t.Fatal(err)
				}

				fwReader, err = os.Open(binPath)
				if err != nil {
					t.Fatalf("%s -> %s", err.Error(), binPath)
				}

				defer os.Remove(binPath)
			}

			serviceClient, err := newBmcServiceClient(parsedURL.Hostname(), parsedURL.Port(), "foo", "bar", httpclient.Build())
			assert.Nil(t, err)
			serviceClient.csrfToken = "foobar"
			client := &x11{serviceClient: serviceClient, log: logr.Discard()}

			if err := client.uploadBMCFirmware(context.Background(), fwReader); err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}

				assert.Nil(t, err)
			}
		})
	}
}

func TestX11VerifyBMCFirmwareVersion(t *testing.T) {
	testcases := []struct {
		name          string
		errorContains string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"verify successful",
			"",
			"/cgi/ipmi.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, []byte(`op=UPLOAD_FW_VERSION.XML&r=(0,0)&_=`), b)

				resp := []byte(`<?xml version="1.0"?>  <IPMI>  <FW_VERSION NEW="017409" OLD="017409"/>  </IPMI>`)
				_, err = w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			"unexpected response",
			"unexpected response",
			"/cgi/ipmi.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				resp := []byte(`bad bmc does not comply`)
				_, err := w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			"unexpected status code",
			"Unexpected status code: 403",
			"/cgi/ipmi.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(http.StatusForbidden)
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

			if err := client.verifyBMCFirmwareVersion(context.Background()); err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}

				assert.Nil(t, err)
			}
		})
	}
}

func TestX11InitiateBMCFirmwareInstall(t *testing.T) {
	testcases := []struct {
		name          string
		errorContains string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"install intiated successfully",
			"",
			"/cgi/op.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, []byte(`op=main_fwupdate&preserve_config=1&preserve_sdr=1&preserve_ssl=1`), b)

				resp := []byte(`Upgrade progress.. 1%`)
				_, err = w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			"unexpected response",
			"unexpected response",
			"/cgi/op.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				resp := []byte(`bad bmc does not comply`)
				_, err := w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			"unexpected status code",
			"Unexpected status code: 403",
			"/cgi/op.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				w.WriteHeader(http.StatusForbidden)
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

			if err := client.initiateBMCFirmwareInstall(context.Background()); err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}

				assert.Nil(t, err)
			}
		})
	}
}

func TestX11StatusBMCFirmwareInstall(t *testing.T) {
	testcases := []struct {
		name          string
		expectState   constants.TaskState
		expectStatus  string
		errorContains string
		endpoint      string
		handler       func(http.ResponseWriter, *http.Request)
	}{
		{
			"state complete 0",
			constants.Complete,
			"0%",
			"",
			"/cgi/upgrade_process.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, []byte(`fwtype=0&_`), b)

				resp := []byte(`  <?xml version="1.0"?>
				<IPMI>
				  <percent>0</percent>
				</IPMI>`)
				_, err = w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			"state complete 100",
			constants.Complete,
			"100%",
			"",
			"/cgi/upgrade_process.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, []byte(`fwtype=0&_`), b)

				resp := []byte(`  <?xml version="1.0"?>
				<IPMI>
				  <percent>100</percent>
				</IPMI>`)
				_, err = w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			"state initializing",
			constants.Initializing,
			"1%",
			"",
			"/cgi/upgrade_process.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, []byte(`fwtype=0&_`), b)

				resp := []byte(`  <?xml version="1.0"?>
				<IPMI>
				  <percent>1</percent>
				</IPMI>`)
				_, err = w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			"status running",
			constants.Running,
			"95%",
			"",
			"/cgi/upgrade_process.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, []byte(`fwtype=0&_`), b)

				resp := []byte(`  <?xml version="1.0"?>
				<IPMI>
				  <percent>95</percent>
				</IPMI>`)
				_, err = w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			"status unknown",
			constants.Unknown,
			"",
			"session expired",
			"/cgi/upgrade_process.cgi",
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				b, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, []byte(`fwtype=0&_`), b)

				resp := []byte(`<html> <head>uh what</head> </html>`)
				_, err = w.Write(resp)
				if err != nil {
					t.Fatal(err)
				}
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

			gotState, gotStatus, err := client.statusBMCFirmwareInstall(context.Background())
			if err != nil {
				if tc.errorContains != "" {
					assert.ErrorContains(t, err, tc.errorContains)

					return
				}
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectState, gotState)
			assert.Equal(t, tc.expectStatus, gotStatus)
		})
	}
}
