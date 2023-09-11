package rpc

import (
	"bytes"
	"encoding/json"
	"io"
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
func requestKVS(req *http.Request) []interface{} {
	var r requestDetails
	if req != nil && req.Body != nil {
		var p RequestPayload
		reqBody, err := io.ReadAll(req.Body)
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			_ = json.Unmarshal(reqBody, &p)
		}
		r = requestDetails{
			Body:    p,
			Headers: req.Header,
			URL:     req.URL.String(),
			Method:  req.Method,
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
