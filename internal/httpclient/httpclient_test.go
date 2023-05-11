package httpclient

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func CertPoolFromCert(cert *x509.Certificate) *x509.CertPool {
	certPool := x509.NewCertPool()
	certPool.AddCert(cert)
	return certPool
}

func TestBuildWithOptions(t *testing.T) {
	cases := []struct {
		name         string
		secureClient bool
		withCertPool func(cert *x509.Certificate) *x509.CertPool
		wantErr      bool
	}{
		{
			"Default not secure, no error",
			false,
			func(_ *x509.Certificate) *x509.CertPool { return nil },
			false,
		},
		{
			"Default secure, want an error",
			true,
			func(_ *x509.Certificate) *x509.CertPool { return nil },
			true,
		},
		{
			"Default secure, no error",
			true,
			CertPoolFromCert,
			false,
		},
	}

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"hello": "client"}`)
	}))
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts := []func(*http.Client){}
			if tc.secureClient {
				opts = append(opts, SecureTLSOption(tc.withCertPool(server.Certificate())))
			}
			client := Build(opts...)
			req, _ := http.NewRequest(http.MethodGet, server.URL, nil)

			_, err := client.Do(req)
			if tc.wantErr {
				if err == nil {
					t.Fatal("Missing expected error")
				}

				// Different versions of Go return different error messages so we just
				// check that its a *url.Error{}
				if _, ok := err.(*url.Error); !ok {
					t.Fatalf("Missing expected error: got %T: '%s'", err, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Got unexpected error %s", err)
			}
		})
	}
}
