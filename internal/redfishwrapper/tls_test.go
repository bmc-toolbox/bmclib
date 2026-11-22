package redfishwrapper

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
)

// The parent TestOpenSecureTLSSystemRoots passes its servers' URLs to its child process
// in these variables. envSystemRootsServer also tells the test it is running as the child.
const (
	envSystemRootsServer          = "BMCLIB_TEST_SYSTEM_ROOTS_SERVER"
	envSystemRootsUntrustedServer = "BMCLIB_TEST_SYSTEM_ROOTS_UNTRUSTED_SERVER"
)

// redfishServer starts a TLS Redfish server on cfg, or on httptest's default config when
// cfg is nil. The returned counter counts the requests that reached the server; a request
// arrives only after a completed TLS handshake, so it shows whether a client accepted the
// server's certificate.
func redfishServer(t *testing.T, cfg *tls.Config) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "systems.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/1", endpointFunc(t, "managers_1.json"))

	requests := &atomic.Int32{}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		mux.ServeHTTP(w, r)
	}))
	server.TLS = cfg
	server.StartTLS()
	t.Cleanup(server.Close)

	return server, requests
}

// selfSignedCert returns a certificate for the loopback addresses httptest listens on that
// no test trusts. It is valid in every way except its issuer, so a client can reject it
// only by checking the CA.
func selfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "untrusted"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

func TestOpenTLSVerification(t *testing.T) {
	// httptest's certificate is not in the system roots.
	server, requests := redfishServer(t, nil)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	serverCA := x509.NewCertPool()
	serverCA.AddCert(server.Certificate())

	tests := map[string]struct {
		opts    []Option
		wantErr bool
	}{
		"no HTTP client skips verification": {},
		"default HTTP client skips verification": {
			opts: []Option{WithHTTPClient(httpclient.Build())},
		},
		"WithSecureTLS with system roots rejects an untrusted certificate": {
			opts:    []Option{WithSecureTLS(nil)},
			wantErr: true,
		},
		// bmclib.NewClient with bmclib.WithSecureTLS hands providers a client built this way.
		"secured HTTP client rejects an untrusted certificate": {
			opts:    []Option{WithHTTPClient(httpclient.Build(httpclient.SecureTLSOption(nil)))},
			wantErr: true,
		},
		"WithSecureTLS with the server's CA accepts its certificate": {
			opts: []Option{WithSecureTLS(serverCA)},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			requests.Store(0)
			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", tc.opts...)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := client.Open(ctx)
			if tc.wantErr {
				assert.ErrorContains(t, err, "certificate")
				assert.Zero(t, requests.Load(), "a request completed against an untrusted certificate")
				return
			}
			require.NoError(t, err)
			defer client.Close(ctx)
			assert.NotZero(t, requests.Load())
		})
	}
}

func TestCallsAfterOpenVerifyTLS(t *testing.T) {
	// Every call must verify the certificate, not only Open. After Open, the server
	// switches to a certificate from another CA, as an attacker on the BMC network would.
	// bmclib's transport disables keep-alives, so the next call makes a new handshake.
	untrusted := selfSignedCert(t)
	var swapped atomic.Bool
	server, requests := redfishServer(t, &tls.Config{
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			if swapped.Load() {
				return &tls.Config{Certificates: []tls.Certificate{untrusted}}, nil
			}
			return nil, nil // httptest's certificate
		},
	})

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	serverCA := x509.NewCertPool()
	serverCA.AddCert(server.Certificate())
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithSecureTLS(serverCA))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, client.Open(ctx))

	// Before the switch, the call reaches the server, so it does go over the network.
	before := requests.Load()
	_, err = client.ManagerOdataID(ctx)
	require.NoError(t, err)
	require.Greater(t, requests.Load(), before)

	swapped.Store(true)
	before = requests.Load()
	_, err = client.ManagerOdataID(ctx)
	assert.ErrorContains(t, err, "certificate")
	assert.Equal(t, before, requests.Load(), "a request completed against an untrusted certificate")
}

func TestOpenSecureTLSSystemRoots(t *testing.T) {
	// WithSecureTLS(nil) means the system roots: a caller trusts a BMC's CA by adding it
	// to them, for example by mounting it into a container's trust store. Go loads the
	// system roots once per process, so the check runs in a child process whose
	// SSL_CERT_FILE names one server's CA. The child must connect to that server and be
	// refused by a second server whose certificate comes from another CA.
	if serverURL := os.Getenv(envSystemRootsServer); serverURL != "" {
		openWithSystemRoots(t, serverURL, os.Getenv(envSystemRootsUntrustedServer))
		return
	}
	if runtime.GOOS != "linux" {
		t.Skip("SSL_CERT_FILE sets Go's system roots on Linux; macOS and Windows use the platform verifier")
	}

	server, requests := redfishServer(t, nil)
	untrusted, untrustedRequests := redfishServer(t, &tls.Config{Certificates: []tls.Certificate{selfSignedCert(t)}})

	caFile := filepath.Join(t.TempDir(), "ca.pem")
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	require.NoError(t, os.WriteFile(caFile, caPEM, 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestOpenSecureTLSSystemRoots$", "-test.count=1")
	cmd.Env = append(os.Environ(),
		"SSL_CERT_FILE="+caFile,
		envSystemRootsServer+"="+server.URL,
		envSystemRootsUntrustedServer+"="+untrusted.URL,
	)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "child process output:\n%s", out)

	// The child exits 0 when -test.run matches nothing; the server must have seen it.
	assert.NotZero(t, requests.Load(), "child process never reached the server:\n%s", out)
	assert.Zero(t, untrustedRequests.Load(), "a request completed against an untrusted certificate:\n%s", out)
}

// openWithSystemRoots runs in the child process of TestOpenSecureTLSSystemRoots. It opens
// trustedURL, whose CA is in the system roots, and expects untrustedURL to be refused.
func openWithSystemRoots(t *testing.T, trustedURL, untrustedURL string) {
	open := func(serverURL string) error {
		parsedURL, err := url.Parse(serverURL)
		require.NoError(t, err)

		// bmclib.NewClient with bmclib.WithSecureTLS(nil) hands providers a client built this way.
		httpClient := httpclient.Build(httpclient.SecureTLSOption(nil))
		client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithHTTPClient(httpClient))

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = client.Open(ctx)
		if err == nil {
			_ = client.Close(ctx)
		}
		return err
	}

	require.NoError(t, open(trustedURL))
	require.NotEmpty(t, untrustedURL)
	assert.ErrorContains(t, open(untrustedURL), "certificate")
}
