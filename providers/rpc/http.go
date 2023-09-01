package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// createRequest
func (c *Config) createRequest(ctx context.Context, p RequestPayload) (*http.Request, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, c.Opts.Request.HTTPMethod, c.listenerURL.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	for k, v := range c.Opts.Request.StaticHeaders {
		req.Header.Add(k, strings.Join(v, ","))
	}
	if c.Opts.Request.HTTPContentType != "" {
		req.Header.Set("Content-Type", c.Opts.Request.HTTPContentType)
	}
	if c.Opts.Request.TimestampHeader != "" {
		req.Header.Add(c.Opts.Request.TimestampHeader, time.Now().Format(c.Opts.Request.TimestampFormat))
	}

	return req, nil
}

func (c *Config) handleResponse(resp *http.Response, reqKeysAndValues []interface{}) (ResponsePayload, error) {
	kvs := reqKeysAndValues
	defer func() {
		if !c.LogNotificationsDisabled {
			kvs = append(kvs, responseKVS(resp)...)
			c.Logger.Info("rpc notification details", kvs...)
		}
	}()
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponsePayload{}, fmt.Errorf("failed to read response body: %v", err)
	}

	res := ResponsePayload{}
	if err := json.Unmarshal(bodyBytes, &res); err != nil {
		if resp.StatusCode != http.StatusOK {
			return ResponsePayload{}, fmt.Errorf("unexpected status code: %d, response error(optional): %v", resp.StatusCode, res.Error)
		}
		example, _ := json.Marshal(ResponsePayload{ID: 123, Host: c.Host, Error: &ResponseError{Code: 1, Message: "error message"}})
		return ResponsePayload{}, fmt.Errorf("failed to parse response: got: %q, error: %w, expected response json spec: %v", string(bodyBytes), err, string(example))
	}
	if resp.StatusCode != http.StatusOK {
		return ResponsePayload{}, fmt.Errorf("unexpected status code: %d, response error(optional): %v", resp.StatusCode, res.Error)
	}
	// reset the body so it can be read again by deferred functions.
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return res, nil
}
