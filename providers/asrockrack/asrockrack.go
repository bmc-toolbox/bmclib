// Package asrockrack implements a bmclib provider for ASRock Rack BMCs.
package asrockrack

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	"github.com/bmc-toolbox/common"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/pkg/errors"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/providers"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "asrockrack"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "vendorapi"

	// E3C256D4ID_NL is a supported ASRockRack board model.
	E3C256D4ID_NL = "E3C256D4ID-NL" //nolint:revive // exported const kept for backwards compatibility
	// E3C246D4ID_NL is a supported ASRockRack board model.
	E3C246D4ID_NL = "E3C246D4ID-NL" //nolint:revive // exported const kept for backwards compatibility
	// E3C246D4I_NL is a supported ASRockRack board model.
	E3C246D4I_NL = "E3C246D4I-NL" //nolint:revive // exported const kept for backwards compatibility
)

// Features implemented by asrockrack https
var Features = registrar.Features{
	providers.FeaturePostCodeRead,
	providers.FeatureBmcReset,
	providers.FeatureUserCreate,
	providers.FeatureUserUpdate,
	providers.FeatureFirmwareUpload,
	providers.FeatureFirmwareInstallUploaded,
	providers.FeatureFirmwareTaskStatus,
	providers.FeatureFirmwareInstallSteps,
	providers.FeatureInventoryRead,
	providers.FeaturePowerSet,
	providers.FeaturePowerState,
}

// ASRockRack holds the status and properties of a connection to a asrockrack bmc
type ASRockRack struct {
	loginSession         *loginSession
	httpClient           *http.Client
	log                  logr.Logger
	ip                   string
	username             string
	password             string
	deviceModel          string
	httpClientSetupFuncs []func(*http.Client)
	resetRequired        bool
	skipLogout           bool
}

// Config holds the optional configuration for an ASRockRack connection.
type Config struct {
	HttpClient *http.Client //nolint:revive // exported field kept for backwards compatibility
	Port       string
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
func New(ip, username, password string, log logr.Logger) *ASRockRack {
	return NewWithOptions(ip, username, password, log)
}

// NewWithOptions returns a new ASRockRack instance with options ready to be used
func NewWithOptions(ip, username, password string, log logr.Logger, opts ...ASRockOption) *ASRockRack {
	r := &ASRockRack{
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
		r.httpClient = httpclient.Build(r.httpClientSetupFuncs...)
	} else {
		for _, setupFunc := range r.httpClientSetupFuncs {
			setupFunc(r.httpClient)
		}
	}
	return r
}

// Name returns the name of this provider.
func (a *ASRockRack) Name() string {
	return ProviderName
}

// Open a connection to a BMC, implements the Opener interface
func (a *ASRockRack) Open(ctx context.Context) (err error) {
	if err := a.httpsLogin(ctx); err != nil {
		return err
	}

	return a.supported(ctx)
}

func (a *ASRockRack) supported(ctx context.Context) error {
	supported := []string{
		E3C256D4ID_NL,
		E3C246D4ID_NL,
		E3C246D4I_NL,
	}

	if a.deviceModel == "" {
		device := common.NewDevice()
		device.Metadata = map[string]string{}

		err := a.fruAttributes(ctx, &device)
		if err != nil {
			return errors.Wrap(err, "failed to identify device model")
		}

		if device.Model == "" {
			return errors.Wrap(err, "failed to identify device model - empty model attribute")
		}

		a.deviceModel = device.Model
	}

	for _, s := range supported {
		if strings.EqualFold(a.deviceModel, s) {
			return nil
		}
	}

	return fmt.Errorf("device model not supported: %s", a.deviceModel)
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

// PostCode returns the BIOS/UEFI post code status and value.
func (a *ASRockRack) PostCode(ctx context.Context) (status string, code int, err error) {
	postInfo, err := a.postCodeInfo(ctx)
	if err != nil {
		return status, code, err
	}

	code = postInfo.PostData
	status, exists := knownPOSTCodes[code]
	if !exists {
		status = constants.POSTCodeUnknown
	}

	return status, code, nil
}
