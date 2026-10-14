package bmc

import "context"

// SecureKeyRepoServer is a key-repository (SKLM/TKLM) server endpoint.
type SecureKeyRepoServer struct {
	// HostName is a remote server address (IPv4, IPv6 or hostname).
	HostName string
	// Port is the remote server port (1-65535).
	Port int
}

// SecureKeyLifecycleConfig is the Secure Key Lifecycle (SKLM) configuration.
type SecureKeyLifecycleConfig struct {
	// DeviceGroup is the SKLM device group name.
	DeviceGroup string
	// KeyRepoServers is the list of configured key-repository servers.
	KeyRepoServers []SecureKeyRepoServer
}

// SecureKeyLifecycle is implemented by providers that can read and configure the
// Secure Key Lifecycle service (key-repository servers).
type SecureKeyLifecycle interface {
	// GetSecureKeyLifecycle returns the current SKLM configuration.
	GetSecureKeyLifecycle(ctx context.Context) (SecureKeyLifecycleConfig, error)
	// SetSecureKeyRepoServers replaces the configured key-repository servers.
	SetSecureKeyRepoServers(ctx context.Context, servers []SecureKeyRepoServer) error
}
