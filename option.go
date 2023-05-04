package bmclib

import "net/http"

func WithIpmitoolCipherSuite(cipherSuite int) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.CipherSuite = cipherSuite
	}
}

func WithIpmitoolPort(port string) Option {
	return func(args *Client) {
		args.providerConfig.ipmitool.Port = port
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
