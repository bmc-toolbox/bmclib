package lenovo

import (
	"context"
	"net/url"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// sklmSubPath is the XCC OEM Secure Key Lifecycle service, relative to the
// Manager resource.
const sklmSubPath = "Oem/Lenovo/SecureKeyLifecycleService"

// compile-time assertion that the provider implements the interface.
var _ bmc.SecureKeyLifecycle = (*Conn)(nil)

// keyRepoServer mirrors an XCC SecureKeyLifecycleService KeyRepoServers entry.
type keyRepoServer struct {
	HostName string `json:"HostName"`
	Port     int    `json:"Port"`
}

// secureKeyLifecycleDoc is a partial model of the XCC SecureKeyLifecycleService.
type secureKeyLifecycleDoc struct {
	DeviceGroup    string          `json:"DeviceGroup"`
	KeyRepoServers []keyRepoServer `json:"KeyRepoServers"`
}

// GetSecureKeyLifecycle reads the XCC Secure Key Lifecycle configuration.
//
// Implements bmc.SecureKeyLifecycle.
func (c *Conn) GetSecureKeyLifecycle(ctx context.Context) (bmc.SecureKeyLifecycleConfig, error) {
	path, err := c.secureKeyLifecyclePath(ctx)
	if err != nil {
		return bmc.SecureKeyLifecycleConfig{}, err
	}

	var doc secureKeyLifecycleDoc
	if err := c.getJSON(path, &doc); err != nil {
		return bmc.SecureKeyLifecycleConfig{}, err
	}

	cfg := bmc.SecureKeyLifecycleConfig{DeviceGroup: doc.DeviceGroup}
	for _, s := range doc.KeyRepoServers {
		cfg.KeyRepoServers = append(cfg.KeyRepoServers, bmc.SecureKeyRepoServer{
			HostName: s.HostName,
			Port:     s.Port,
		})
	}

	return cfg, nil
}

// SetSecureKeyRepoServers replaces the configured key-repository servers by
// PATCHing the SecureKeyLifecycleService.
//
// Implements bmc.SecureKeyLifecycle.
func (c *Conn) SetSecureKeyRepoServers(ctx context.Context, servers []bmc.SecureKeyRepoServer) error {
	path, err := c.secureKeyLifecyclePath(ctx)
	if err != nil {
		return err
	}

	repo := make([]keyRepoServer, 0, len(servers))
	for _, s := range servers {
		repo = append(repo, keyRepoServer{HostName: s.HostName, Port: s.Port})
	}

	payload := map[string]any{"KeyRepoServers": repo}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, path, payload, nil))
}

// secureKeyLifecyclePath resolves the SecureKeyLifecycleService path from the
// managed Manager resource.
func (c *Conn) secureKeyLifecyclePath(ctx context.Context) (string, error) {
	manager, err := c.redfishwrapper.Manager(ctx)
	if err != nil {
		return "", err
	}

	return url.JoinPath(manager.ODataID, sklmSubPath)
}
