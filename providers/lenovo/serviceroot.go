package lenovo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// serviceRootPath is the well-known Redfish service root URI. It is the only
// resource reachable without authentication.
const serviceRootPath = "/redfish/v1/"

// ServiceRoot fetches and parses the XCC Redfish service root.
//
// Capability code should resolve resource paths (Systems, Managers,
// UpdateService, ...) from the links returned here rather than hard-coding
// URIs, so the provider adapts to XCC service-root layout differences across
// firmware versions.
func (c *Conn) ServiceRoot(ctx context.Context) (*ServiceRoot, error) {
	resp, err := c.redfishwrapper.Get(serviceRootPath)
	if err != nil {
		return nil, fmt.Errorf("fetching XCC service root: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseRedfishError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errFailedToParseResponse, err)
	}

	root := &ServiceRoot{}
	if err := json.Unmarshal(body, root); err != nil {
		return nil, fmt.Errorf("%w: %v", errFailedToParseResponse, err)
	}

	return root, nil
}
