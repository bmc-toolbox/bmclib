package asrockrack

import (
	"bytes"
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "asrockrack"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "vendorapi"
)

var (
	// Features implemented by asrockrack https
	Features = registrar.Features{
		providers.FeatureInventoryRead,
		providers.FeatureFirmwareInstall,
		providers.FeatureFirmwareInstallStatus,
		providers.FeaturePostCodeRead,
		providers.FeatureBmcReset,
	}
)

// ASRockRack holds the status and properties of a connection to a asrockrack bmc
type ASRockRack struct {
	ip                   string
	username             string
	password             string
	loginSession         *loginSession
	httpClient           *http.Client
	resetRequired        bool // Indicates if the BMC requires a reset
	skipLogout           bool // A Close() / httpsLogout() request is ignored if the BMC was just flashed - since the sessions are terminated either way
	log                  logr.Logger
	httpClientSetupFuncs []func(*http.Client)
}

// ASRockOption is a type that can configure an *ASRockRack
type ASRockOption func(*ASRockRack)

// WithSecureTLS enforces trusted TLS connections, with an optional CA certificate pool.
// Using this option with an nil pool uses the system CAs.
func WithSecureTLS(rootCAs *x509.CertPool) ASRockOption {
	return func(r *ASRockRack) {
		r.httpClientSetupFuncs = append(r.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

// WithHTTPClient sets an HTTP client on the ASRockRack
func WithHTTPClient(c *http.Client) ASRockOption {
	return func(ar *ASRockRack) {
		ar.httpClient = c
	}
}

// New returns a new ASRockRack instance ready to be used
func New(ip string, username string, password string, log logr.Logger) (*ASRockRack, error) {
	return NewWithOptions(ip, username, password, log)
}

// NewWithOptions returns a new ASRockRack instance with options ready to be used
func NewWithOptions(ip string, username string, password string, log logr.Logger, opts ...ASRockOption) (r *ASRockRack, err error) {
	r = &ASRockRack{
		ip:           ip,
		username:     username,
		password:     password,
		log:          log,
		loginSession: &loginSession{},
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.httpClient == nil {
		r.httpClient, err = httpclient.Build(r.httpClientSetupFuncs...)
		if err != nil {
			return nil, err
		}
	} else {
		for _, setupFunc := range r.httpClientSetupFuncs {
			setupFunc(r.httpClient)
		}
	}
	return r, nil
}

// Compatible implements the registrar.Verifier interface
// returns true if the BMC is identified to be an asrockrack
func (a *ASRockRack) Compatible() bool {
	resp, statusCode, err := a.queryHTTPS(context.TODO(), "/", "GET", nil, nil, 0)
	if err != nil {
		return false
	}

	if statusCode != 200 {
		return false
	}

	return bytes.Contains(resp, []byte(`ASRockRack`))
}

// Open a connection to a BMC, implements the Opener interface
func (a *ASRockRack) Open(ctx context.Context) (err error) {
	return a.httpsLogin(ctx)
}

// Close a connection to a BMC, implements the Closer interface
func (a *ASRockRack) Close(ctx context.Context) (err error) {
	if a.skipLogout {
		return nil
	}

	return a.httpsLogout(ctx)
}

// CheckCredentials verify whether the credentials are valid or not
func (a *ASRockRack) CheckCredentials(ctx context.Context) (err error) {
	return a.httpsLogin(ctx)
}

func (a *ASRockRack) PostCode(ctx context.Context) (status string, code int, err error) {
	postInfo, err := a.postCodeInfo(ctx)
	if err != nil {
		return status, code, err
	}

	code = postInfo.PostData
	status, exists := knownPOSTCodes[code]
	if !exists {
		status = devices.POSTCodeUnknown
	}

	return status, code, nil
}
