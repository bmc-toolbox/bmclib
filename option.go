package bmclib

import "net/http"

func WithIpmitoolCipherSuite(cipherSuite int) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.cipherSuite = cipherSuite
	}
}

func WithIpmitoolPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.port = port
	}
}

func WithAsrockrackHTTPClient(httpClient *http.Client) Option {
	return func(args *Client) {
		args.providerConfig.asrock.httpClient = httpClient
	}
}

func WithAsrockrackPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.asrock.port = port
	}
}

func WithGofishHTTPClient(httpClient *http.Client) Option {
	return func(args *Client) {
		args.providerConfig.gofish.httpClient = httpClient
	}
}

func WithGofishPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.gofish.port = port
	}
}

// WithGofishVersionsNotCompatible sets the list of incompatible redfish versions.
//
// With this option set, The bmclib.Registry.FilterForCompatible(ctx) method will not proceed on
// devices with the given redfish version(s).
func WithGofishVersionsNotCompatible(versions []string) Option {
	return func(args *Client) {
		args.providerConfig.gofish.versionsNotCompatible = append(args.providerConfig.gofish.versionsNotCompatible, versions...)
	}
}

func WithIntelAMTHostScheme(hostScheme string) Option {
	return func(args *Client) {
		args.providerConfig.intelamt.hostScheme = hostScheme
	}
}

func WithIntelAMTPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.intelamt.port = port
	}
}
