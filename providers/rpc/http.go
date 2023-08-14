package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func requestKVS(req *http.Request) []interface{} {
	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		// TODO(jacobweinstock): either log the error or change the func signature to return it
		return nil
	}
	req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	var p RequestPayload
	if err := json.Unmarshal(reqBody, &p); err != nil {
		// TODO(jacobweinstock): either log the error or change the func signature to return it
		return nil
	}

	s := struct {
		Body    RequestPayload `json:"body"`
		Headers http.Header    `json:"headers"`
		URL     string         `json:"url"`
		Method  string         `json:"method"`
	}{
		Body:    p,
		Headers: req.Header,
		URL:     req.URL.String(),
		Method:  req.Method,
	}

	return []interface{}{"request", s}
}

func responseKVS(resp *http.Response) []interface{} {
	reqBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var p map[string]interface{}
	if err := json.Unmarshal(reqBody, &p); err != nil {
		return nil
	}

	r := struct {
		StatusCode int                    `json:"statusCode"`
		Body       map[string]interface{} `json:"body"`
		Headers    http.Header            `json:"headers"`
	}{
		StatusCode: resp.StatusCode,
		Body:       p,
		Headers:    resp.Header,
	}

	return []interface{}{"response", r}
}

func (c *Config) createRequest(ctx context.Context, p RequestPayload) (*http.Request, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, c.HTTPMethod, c.listenerURL.String(), bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", c.HTTPContentType)
	req.Header.Add(c.timestampHeader, time.Now().Format(c.timestampFormat))

	return req, nil
}

func (c *Config) signAndSend(p RequestPayload, req *http.Request) (*ResponsePayload, error) {
	if err := c.sig.AddSignature(req); err != nil {
		return nil, err
	}
	// have to copy the body out before sending the request.
	kvs := requestKVS(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.Logger.Error(err, "failed to send rpc notification", kvs...)
		return nil, err
	}
	defer func() {
		if c.LogNotifications {
			if p.Params != nil {
				kvs = append(kvs, []interface{}{"params", p.Params}...)
			}
			kvs = append(kvs, responseKVS(resp)...)
			kvs = append(kvs, []interface{}{"host", c.host, "method", p.Method, "consumerURL", c.consumerURL}...)
			c.Logger.Info("rpc notification details", kvs...)
		}
	}()
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	res := &ResponsePayload{}
	if err := json.Unmarshal(bodyBytes, res); err != nil {
		example, _ := json.Marshal(ResponsePayload{ID: 123, Host: c.host, Error: &ResponseError{Code: 1, Message: "error message"}})
		return nil, fmt.Errorf("failed to parse response: got: %q, error: %w, response json spec: %v", string(bodyBytes), err, string(example))
	}
	// reset the body so it can be read again by deferred functions.
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return res, nil
}
