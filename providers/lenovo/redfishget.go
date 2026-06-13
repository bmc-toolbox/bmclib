package lenovo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// getJSON GETs a Redfish resource and unmarshals its body into out. It maps
// non-2xx responses to a descriptive error via parseRedfishError.
func (c *Conn) getJSON(url string, out any) error {
	resp, err := c.redfishwrapper.Get(url)
	if err != nil {
		// gofish returns non-2xx responses as an error; surface 404 as the
		// errResourceNotFound sentinel so callers can treat an absent resource
		// gracefully.
		if isNotFound(err) {
			return fmt.Errorf("%w: GET %s", errResourceNotFound, url)
		}
		return fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		rerr := parseRedfishError(resp)
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("%w: %v", errResourceNotFound, rerr)
		}
		return rerr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", errFailedToParseResponse, err)
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%w: %v", errFailedToParseResponse, err)
	}

	return nil
}

// singleTrailingSlashJoin joins a base path and sub-path with exactly one
// separating slash, trimming any duplicates at the boundary.
func singleTrailingSlashJoin(base, sub string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(sub, "/")
}

// collectionMembers GETs a Redfish collection and returns its member links.
func (c *Conn) collectionMembers(url string) ([]odataID, error) {
	var coll struct {
		Members []odataID `json:"Members"`
	}
	if err := c.getJSON(url, &coll); err != nil {
		return nil, err
	}

	return coll.Members, nil
}
