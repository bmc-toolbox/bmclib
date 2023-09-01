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
func responseKVS(resp *http.Response) []interface{} {
	var r responseDetails
	if resp != nil && resp.Body != nil {
		var p ResponsePayload
		reqBody, err := io.ReadAll(resp.Body)
		if err == nil {
			_ = json.Unmarshal(reqBody, &p)
		}
		r = responseDetails{
			StatusCode: resp.StatusCode,
			Body:       p,
			Headers:    resp.Header,
		}
	}

	return []interface{}{"response", r}
}
