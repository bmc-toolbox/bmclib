package bmclib

import (
	"context"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/providers/homeassistant"
	"github.com/bmc-toolbox/bmclib/v2/providers/rpc"
)

// Option for setting optional Client values
type Option func(*Client)

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

// WithIPMICipherSuite sets the cipher suite for the pure-go ipmi provider.
func WithIPMICipherSuite(cipherSuite string) Option {
	return func(args *Client) {
		args.providerConfig.ipmi.CipherSuite = cipherSuite
	}
}

// WithIPMIPort sets the port for the pure-go ipmi provider.
func WithIPMIPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.ipmi.Port = port
	}
}

// WithIpmitoolCipherSuite sets the IPMI cipher suite used by the ipmitool provider.
func WithIpmitoolCipherSuite(cipherSuite string) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.CipherSuite = cipherSuite
	}
}

// WithIpmitoolPort sets the port used by the ipmitool provider.
func WithIpmitoolPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.Port = port
	}
}

// WithIpmitoolPath sets the path to the ipmitool binary used by the ipmitool provider.
func WithIpmitoolPath(path string) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.IpmitoolPath = path
	}
}

// WithAsrockrackHTTPClient sets the HTTP client used by the asrockrack provider.
func WithAsrockrackHTTPClient(httpClient *http.Client) Option {
	return func(args *Client) {
		args.providerConfig.asrock.HTTPClient = httpClient
	}
}

// WithAsrockrackPort sets the port used by the asrockrack provider.
func WithAsrockrackPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.asrock.Port = port
	}
}

// WithRedfishHTTPClient sets the HTTP client used by the redfish (gofish) provider.
func WithRedfishHTTPClient(httpClient *http.Client) Option {
	return func(args *Client) {
		args.providerConfig.gofish.HTTPClient = httpClient
	}
}

// WithRedfishPort sets the port used by the redfish (gofish) provider.
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

// WithRedfishUseBasicAuth sets HTTP Basic auth (instead of session login) for the redfish provider.
func WithRedfishUseBasicAuth(useBasicAuth bool) Option {
	return func(args *Client) {
		args.providerConfig.gofish.UseBasicAuth = useBasicAuth
	}
}

// WithRedfishEtagMatchDisabled disables use of ETag matching by the redfish provider.
func WithRedfishEtagMatchDisabled(d bool) Option {
	return func(args *Client) {
		args.providerConfig.gofish.DisableEtagMatch = d
	}
}

// WithRedfishSystemName sets the redfish system name targeted by the redfish provider.
func WithRedfishSystemName(name string) Option {
	return func(args *Client) {
		args.providerConfig.gofish.SystemName = name
	}
}

// WithIntelAMTHostScheme sets the host scheme (http/https) used by the Intel AMT provider.
func WithIntelAMTHostScheme(hostScheme string) Option {
	return func(args *Client) {
		args.providerConfig.intelamt.HostScheme = hostScheme
	}
}

// WithIntelAMTPort sets the port used by the Intel AMT provider.
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

// WithDellRedfishUseBasicAuth sets HTTP Basic auth (instead of session login) for the Dell redfish provider.
func WithDellRedfishUseBasicAuth(useBasicAuth bool) Option {
	return func(args *Client) {
		args.providerConfig.dell.UseBasicAuth = useBasicAuth
	}
}

// WithLenovoPort sets the port for the Lenovo XCC (redfish) provider.
func WithLenovoPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.lenovo.Port = port
	}
}

// WithLenovoUseBasicAuth sets HTTP Basic auth (instead of session login) for the
// Lenovo XCC provider.
func WithLenovoUseBasicAuth(useBasicAuth bool) Option {
	return func(args *Client) {
		args.providerConfig.lenovo.UseBasicAuth = useBasicAuth
	}
}

// WithLenovoVersionsNotCompatible sets the list of incompatible redfish versions
// for the Lenovo XCC provider.
//
// With this option set, the bmclib.Registry.FilterForCompatible(ctx) method will
// not proceed on devices with the given redfish version(s).
func WithLenovoVersionsNotCompatible(versions []string) Option {
	return func(args *Client) {
		args.providerConfig.lenovo.VersionsNotCompatible = append(args.providerConfig.lenovo.VersionsNotCompatible, versions...)
	}
}

// WithRPCOpt configures the rpc provider.
func WithRPCOpt(opt rpc.Provider) Option { //nolint:gocritic // functional options take their config by value by convention
	return func(args *Client) {
		args.providerConfig.rpc = opt
	}
}

// WithHomeAssistantOpt configures the Home Assistant provider.
func WithHomeAssistantOpt(opt homeassistant.Config) Option { //nolint:gocritic // functional options take their config by value by convention
	return func(args *Client) {
		args.providerConfig.homeassistant = opt
	}
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified a noop tracerprovider is used.
func WithTracerProvider(provider oteltrace.TracerProvider) Option {
	return func(args *Client) {
		if provider != nil {
			args.traceprovider = provider
		}
	}
}
