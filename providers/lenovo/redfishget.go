package lenovo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
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

// decodeJSONBody reads and unmarshals an already-open HTTP response body into
// out. The caller is responsible for closing resp.Body.
func decodeJSONBody(resp *http.Response, out any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %v", errFailedToParseResponse, err)
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%w: %v", errFailedToParseResponse, err)
	}
	return nil
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

// memberIDs GETs a Redfish collection and returns the trailing-segment Id of
// each member.
func (c *Conn) memberIDs(url string) ([]string, error) {
	members, err := c.collectionMembers(url)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(members))
	for _, m := range members {
		ids = append(ids, path.Base(strings.TrimRight(m.ODataID, "/")))
	}

	return ids, nil
}
