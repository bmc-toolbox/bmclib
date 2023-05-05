package bmclib

import (
	"context"
	"crypto/x509"
	"net/http"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
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

func WithGofishHTTPClient(httpClient *http.Client) Option {
	return func(args *Client) {
		args.providerConfig.gofish.HttpClient = httpClient
	}
}

func WithGofishPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.gofish.Port = port
	}
}

// WithGofishVersionsNotCompatible sets the list of incompatible redfish versions.
//
// With this option set, The bmclib.Registry.FilterForCompatible(ctx) method will not proceed on
// devices with the given redfish version(s).
func WithGofishVersionsNotCompatible(versions []string) Option {
	return func(args *Client) {
		args.providerConfig.gofish.VersionsNotCompatible = append(args.providerConfig.gofish.VersionsNotCompatible, versions...)
	}
}

func WithGofishUseBasicAuth(useBasicAuth bool) Option {
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
