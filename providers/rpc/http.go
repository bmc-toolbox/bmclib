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
func (p *Provider) createRequest(ctx context.Context, rp RequestPayload) (*http.Request, error) {
	var data []byte
	if rj := p.Opts.Experimental.CustomRequestPayload; rj != nil && p.Opts.Experimental.DotPath != "" {
		d, err := rp.embedPayload(rj, p.Opts.Experimental.DotPath)
		if err != nil {
			return nil, err
		}
		data = d
	} else {
		d, err := json.Marshal(rp)
		if err != nil {
			return nil, err
		}
		data = d
	}

	req, err := http.NewRequestWithContext(ctx, p.Opts.Request.HTTPMethod, p.listenerURL.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	for k, v := range p.Opts.Request.StaticHeaders {
		req.Header.Add(k, strings.Join(v, ","))
	}
	if p.Opts.Request.HTTPContentType != "" {
		req.Header.Set("Content-Type", p.Opts.Request.HTTPContentType)
	}
	if p.Opts.Request.TimestampHeader != "" {
		req.Header.Add(p.Opts.Request.TimestampHeader, time.Now().Format(p.Opts.Request.TimestampFormat))
	}

	return req, nil
}

func (p *Provider) handleResponse(resp *http.Response, reqKeysAndValues []interface{}) (ResponsePayload, error) {
	kvs := reqKeysAndValues
	defer func() {
		if !p.LogNotificationsDisabled {
			kvs = append(kvs, responseKVS(resp)...)
			p.Logger.Info("rpc notification details", kvs...)
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
		example, _ := json.Marshal(ResponsePayload{ID: 123, Host: p.Host, Error: &ResponseError{Code: 1, Message: "error message"}})
		return ResponsePayload{}, fmt.Errorf("failed to parse response: got: %q, error: %w, expected response json spec: %v", string(bodyBytes), err, string(example))
	}
	if resp.StatusCode != http.StatusOK {
		return ResponsePayload{}, fmt.Errorf("unexpected status code: %d, response error(optional): %v", resp.StatusCode, res.Error)
	}
	// reset the body so it can be read again by deferred functions.
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return res, nil
}
