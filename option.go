package bmclib

import (
	"context"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/providers/rpc"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

// Option for setting optional Client values
type Option func(*Client)

type rpcOpts struct {
	Secrets rpc.Secrets
	// ConsumerURL is the URL where a rpc consumer/listener is running and to which we will send notifications.
	ConsumerURL string
	// BaseSignatureHeader is the header name that should contain the signature(s). Example: X-BMCLIB-Signature
	BaseSignatureHeader string
	// IncludeAlgoHeader determines whether to append the algorithm to the signature header or not.
	// Example: X-BMCLIB-Signature becomes X-BMCLIB-Signature-256
	// When set to false, a header will be added for each algorithm. Example: X-BMCLIB-Signature-256 and X-BMCLIB-Signature-512
	IncludeAlgoHeader bool
	// IncludedPayloadHeaders are headers whose values will be included in the signature payload. Example: X-BMCLIB-Timestamp
	IncludedPayloadHeaders []string
	// IncludeAlgoPrefix will prepend the algorithm and an equal sign to the signature. Example: sha256=abc123
	IncludeAlgoPrefix bool
	// Logger is the logger to use for logging.
	Logger logr.Logger
	// LogNotifications determines whether responses from rpc consumer/listeners will be logged or not.
	LogNotifications bool
	// HTTPContentType is the content type to use for the rpc request notification.
	HTTPContentType string
	// HTTPMethod is the HTTP method to use for the rpc request notification.
	HTTPMethod string
	// TimestampHeader is the header name that should contain the timestamp. Example: X-BMCLIB-Timestamp
	TimestampHeader string
}

func (w *rpcOpts) SetNonDefaults(wc *rpc.Config) {
	if w.BaseSignatureHeader != "" {
		wc.SetBaseSignatureHeader(w.BaseSignatureHeader)
	}
	if len(w.IncludedPayloadHeaders) > 0 {
		wc.SetIncludedPayloadHeaders(w.IncludedPayloadHeaders)
	}
	// by default, the algorithm is appended to the signature header.
	if !w.IncludeAlgoHeader {
		wc.SetIncludeAlgoHeader(w.IncludeAlgoHeader)
	}
	if !w.Logger.IsZero() {
		wc.Logger = w.Logger
	}
	// by default, the rpc notifications are logged.
	if !w.LogNotifications {
		wc.LogNotifications = w.LogNotifications
	}
	if w.HTTPContentType != "" {
		wc.HTTPContentType = w.HTTPContentType
	}
	if w.HTTPMethod != "" {
		wc.HTTPMethod = w.HTTPMethod
	}
	if w.TimestampHeader != "" {
		wc.SetTimestampHeader(w.TimestampHeader)
	}
}

func WithRPCConsumerURL(url string) Option {
	return func(args *Client) {
		args.providerConfig.rpc.ConsumerURL = url
	}
}

func WithRPCBaseSignatureHeader(header string) Option {
	return func(args *Client) {
		args.providerConfig.rpc.BaseSignatureHeader = header
	}
}

func WithRPCIncludeAlgoHeader(include bool) Option {
	return func(args *Client) {
		args.providerConfig.rpc.IncludeAlgoHeader = include
	}
}

func WithRPCIncludedPayloadHeaders(headers []string) Option {
	return func(args *Client) {
		args.providerConfig.rpc.IncludedPayloadHeaders = headers
	}
}

func WithRPCIncludeAlgoPrefix(include bool) Option {
	return func(args *Client) {
		args.providerConfig.rpc.IncludeAlgoPrefix = include
	}
}

func WithRPCLogNotifications(log bool) Option {
	return func(args *Client) {
		args.providerConfig.rpc.LogNotifications = log
	}
}

func WithRPCHTTPContentType(contentType string) Option {
	return func(args *Client) {
		args.providerConfig.rpc.HTTPContentType = contentType
	}
}

func WithRPCHTTPMethod(method string) Option {
	return func(args *Client) {
		args.providerConfig.rpc.HTTPMethod = method
	}
}

func WithRPCSecrets(secrets rpc.Secrets) Option {
	return func(args *Client) {
		args.providerConfig.rpc.Secrets = secrets
	}
}

func WithRPCTimestampHeader(header string) Option {
	return func(args *Client) {
		args.providerConfig.rpc.TimestampHeader = header
	}
}

// WithLogger sets the logger
func WithLogger(logger logr.Logger) Option {
	return func(args *Client) { args.Logger = logger }
}

// WithRegistry sets the Registry
func WithRegistry(registry *registrar.Registry) Option {
	return func(args *Client) { args.Registry = registry }
}

// WithSecureTLS enforces trusted TLS connections, with an optional CA certificate pool.
// Using this option with an nil pool uses the system CAs.
func WithSecureTLS(rootCAs *x509.CertPool) Option {
	return func(args *Client) {
		args.httpClientSetupFuncs = append(args.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

// WithHTTPClient sets an http client
func WithHTTPClient(c *http.Client) Option {
	return func(args *Client) {
		args.httpClient = c
	}
}

// WithPerProviderTimeout sets the timeout when interacting with a BMC.
// This timeout value is applied per provider.
// When not defined and a context with a timeout is passed to a method, the default timeout
// will be the context timeout duration divided by the number of providers in the registry,
// meaning, the len(Client.Registry.Drivers).
// If this per provider timeout is not defined and no context timeout is defined,
// the defaultConnectTimeout is used.
func WithPerProviderTimeout(timeout time.Duration) Option {
	return func(args *Client) {
		args.perProviderTimeout = func(context.Context) time.Duration { return timeout }
	}
}

func WithIpmitoolCipherSuite(cipherSuite string) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.CipherSuite = cipherSuite
	}
}

func WithIpmitoolPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.Port = port
	}
}

func WithIpmitoolPath(path string) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.IpmitoolPath = path
	}
}

func WithAsrockrackHTTPClient(httpClient *http.Client) Option {
	return func(args *Client) {
		args.providerConfig.asrock.HttpClient = httpClient
	}
}

func WithAsrockrackPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.asrock.Port = port
	}
}

func WithRedfishHTTPClient(httpClient *http.Client) Option {
	return func(args *Client) {
		args.providerConfig.gofish.HttpClient = httpClient
	}
}

func WithRedfishPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.gofish.Port = port
	}
}

// WithRedfishVersionsNotCompatible sets the list of incompatible redfish versions.
//
// With this option set, The bmclib.Registry.FilterForCompatible(ctx) method will not proceed on
// devices with the given redfish version(s).
func WithRedfishVersionsNotCompatible(versions []string) Option {
	return func(args *Client) {
		args.providerConfig.gofish.VersionsNotCompatible = append(args.providerConfig.gofish.VersionsNotCompatible, versions...)
	}
}

func WithRedfishUseBasicAuth(useBasicAuth bool) Option {
	return func(args *Client) {
		args.providerConfig.gofish.UseBasicAuth = useBasicAuth
	}
}

func WithIntelAMTHostScheme(hostScheme string) Option {
	return func(args *Client) {
		args.providerConfig.intelamt.HostScheme = hostScheme
	}
}

func WithIntelAMTPort(port uint32) Option {
	return func(args *Client) {
		args.providerConfig.intelamt.Port = port
	}
}

// WithDellRedfishVersionsNotCompatible sets the list of incompatible redfish versions.
//
// With this option set, The bmclib.Registry.FilterForCompatible(ctx) method will not proceed on
// devices with the given redfish version(s).
func WithDellRedfishVersionsNotCompatible(versions []string) Option {
	return func(args *Client) {
		args.providerConfig.dell.VersionsNotCompatible = append(args.providerConfig.dell.VersionsNotCompatible, versions...)
	}
}

func WithDellRedfishUseBasicAuth(useBasicAuth bool) Option {
	return func(args *Client) {
		args.providerConfig.dell.UseBasicAuth = useBasicAuth
	}
}
