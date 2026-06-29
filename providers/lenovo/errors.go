package lenovo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/stmcginnis/gofish/schemas"
)

// errFailedToParseResponse is returned when a response body could not be read
// or decoded.
var errFailedToParseResponse = errors.New("failed to parse XCC response")

// errResourceNotFound wraps a 404 response so callers can distinguish an absent
// resource (e.g. an OEM service a given XCC firmware level does not expose) from
// other failures via errors.Is.
var errResourceNotFound = errors.New("resource not found")

// errPersistentBootUnsupported is returned when a persistent (Continuous) boot
// override is requested but the XCC only allows a one-time (Once) override.
var errPersistentBootUnsupported = errors.New("persistent boot override unsupported")

// isNotFound reports whether err is a gofish HTTP error carrying a 404 status.
// gofish surfaces non-2xx responses as a *schemas.Error before redfishwrapper
// hands the response back, so a 404 arrives as an error rather than a response.
func isNotFound(err error) bool {
	var re *schemas.Error
	if errors.As(err, &re) {
		return re.HTTPReturnedStatusCode == http.StatusNotFound
	}
	return false
}

// checkResponse closes resp.Body and returns an error if the request errored or
// the response was not a 2xx status. It is used by the raw PATCH/POST paths that
// drop below gofish for XCC OEM operations.
func checkResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}
	if resp == nil {
		return errFailedToParseResponse
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return parseRedfishError(resp)
	}

	return nil
}

// parseRedfishError converts a non-2xx XCC Redfish HTTP response into an error.
//
// When the body carries the standard Redfish error envelope, the returned error
// includes the registry message id, the message text and (for Lenovo OEM
// "ExtendedError" messages) the suggested resolution, so callers get an
// actionable message rather than a bare status code. The caller is responsible
// for closing resp.Body.
func parseRedfishError(resp *http.Response) error {
	if resp == nil {
		return errFailedToParseResponse
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: %d: %v", errFailedToParseResponse, resp.StatusCode, err)
	}

	detail := redfishErrorDetail(body)
	if detail == "" {
		// No parseable Redfish envelope — fall back to status + raw body.
		detail = strings.TrimSpace(string(body))
	}

	if detail == "" {
		return fmt.Errorf("unexpected XCC response: HTTP %d", resp.StatusCode)
	}

	return fmt.Errorf("unexpected XCC response: HTTP %d: %s", resp.StatusCode, detail)
}

// redfishErrorDetail extracts a human-readable detail string from a Redfish
// error body. It returns an empty string when the body is not a recognisable
// Redfish error envelope.
func redfishErrorDetail(body []byte) string {
	var re redfishError
	if err := json.Unmarshal(body, &re); err != nil {
		return ""
	}

	// Prefer the most specific extended-info entry (XCC OEM messages live here).
	for _, info := range re.Error.ExtendedInfo {
		parts := make([]string, 0, 3)
		if info.MessageID != "" {
			parts = append(parts, info.MessageID)
		}
		if info.Message != "" {
			parts = append(parts, info.Message)
		}
		if info.Resolution != "" && info.Resolution != "None" {
			parts = append(parts, "resolution: "+info.Resolution)
		}

		if len(parts) > 0 {
			return strings.Join(parts, ": ")
		}
	}

	// Fall back to the top-level code/message.
	switch {
	case re.Error.Code != "" && re.Error.Message != "":
		return re.Error.Code + ": " + re.Error.Message
	case re.Error.Message != "":
		return re.Error.Message
	case re.Error.Code != "":
		return re.Error.Code
	default:
		return ""
	}
}
