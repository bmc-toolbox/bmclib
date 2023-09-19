package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type requestDetails struct {
	Body    RequestPayload `json:"body"`
	Headers http.Header    `json:"headers"`
	URL     string         `json:"url"`
	Method  string         `json:"method"`
}

type responseDetails struct {
	StatusCode int             `json:"statusCode"`
	Body       ResponsePayload `json:"body"`
	Headers    http.Header     `json:"headers"`
}

// requestKVS returns a slice of key, value sets. Used for logging.
func requestKVS(method, url string, headers http.Header, body *bytes.Buffer) []interface{} {
	var r requestDetails
	if body.Len() > 0 {
		var p RequestPayload
		_ = json.Unmarshal(body.Bytes(), &p)

		r = requestDetails{
			Body:    p,
			Headers: headers,
			URL:     url,
			Method:  method,
		}
	}

	return []interface{}{"request", r}
}

// responseKVS returns a slice of key, value sets. Used for logging.
func responseKVS(statusCode int, headers http.Header, body *bytes.Buffer) []interface{} {
	var r responseDetails
	if body.Len() > 0 {
		var p ResponsePayload
		_ = json.Unmarshal(body.Bytes(), &p)
		r = responseDetails{
			StatusCode: statusCode,
			Body:       p,
			Headers:    headers,
		}
	}

	return []interface{}{"response", r}
}
