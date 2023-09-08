package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func testRequest(method, reqURL string, body RequestPayload, headers http.Header) *http.Request {
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req, _ := http.NewRequestWithContext(context.Background(), method, reqURL, &buf)
	req.Header = headers
	return req
}

func TestRequestKVS(t *testing.T) {
	tests := map[string]struct {
		req      *http.Request
		expected []interface{}
	}{
		"success": {
			req: testRequest(
				http.MethodPost, "http://example.com",
				RequestPayload{ID: 1, Host: "127.0.0.1", Method: "POST", Params: nil},
				http.Header{"Content-Type": []string{"application/json"}},
			),
			expected: []interface{}{"request", requestDetails{
				Body: RequestPayload{
					ID:     1,
					Host:   "127.0.0.1",
					Method: "POST",
					Params: nil,
				},
				Headers: http.Header{"Content-Type": {"application/json"}},
				URL:     "http://example.com",
				Method:  "POST",
			}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			kvs := requestKVS(tc.req)
			if diff := cmp.Diff(kvs, tc.expected); diff != "" {
				t.Fatalf("requestKVS() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequestKVSOneOffs(t *testing.T) {
	t.Run("nil body", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://example.com", nil)
		got := requestKVS(req)
		if diff := cmp.Diff(got, []interface{}{"request", requestDetails{}}); diff != "" {
			t.Logf("got: %+v", got)
			t.Fatalf("requestKVS(req) mismatch (-want +got):\n%s", diff)
		}
	})
	t.Run("nil request", func(t *testing.T) {
		if diff := cmp.Diff(requestKVS(nil), []interface{}{"request", requestDetails{}}); diff != "" {
			t.Fatalf("requestKVS(nil) mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("failed to unmarshal body", func(t *testing.T) {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, "http://example.com", bytes.NewBufferString("invalid"))
		got := requestKVS(req)
		if diff := cmp.Diff(got, []interface{}{"request", requestDetails{URL: "http://example.com", Method: http.MethodPost, Headers: http.Header{}}}); diff != "" {
			t.Logf("got: %+v", got)
			t.Fatalf("requestKVS(req) mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestResponseKVS(t *testing.T) {
	tests := map[string]struct {
		resp     *http.Response
		expected []interface{}
	}{
		"one": {
			resp: &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":1,"host":"127.0.0.1"}`))},
			expected: []interface{}{"response", responseDetails{
				StatusCode: 200,
				Headers:    http.Header{"Content-Type": {"application/json"}},
				Body:       ResponsePayload{ID: 1, Host: "127.0.0.1"},
			}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			kvs := responseKVS(tc.resp)
			if diff := cmp.Diff(kvs, tc.expected); diff != "" {
				t.Fatalf("requestKVS() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCreateRequest(t *testing.T) {
	tests := map[string]struct {
		cfg      Provider
		body     RequestPayload
		expected *http.Request
	}{
		"success": {
			cfg: Provider{
				Opts: Opts{
					Request: RequestOpts{
						HTTPMethod:      http.MethodPost,
						HTTPContentType: "application/json",
						StaticHeaders:   http.Header{"X-Test": []string{"test"}},
					},
				},
				listenerURL: &url.URL{Scheme: "http", Host: "example.com"},
			},
			body: RequestPayload{ID: 1, Host: "127.0.0.1", Method: PowerSetMethod},
			expected: &http.Request{
				ContentLength: 52,
				Host:          "example.com",
				Proto:         "HTTP/1.1",
				ProtoMajor:    1,
				ProtoMinor:    1,
				Method:        http.MethodPost,
				URL:           &url.URL{Scheme: "http", Host: "example.com"},
				Header:        http.Header{"X-Test": {"test"}, "Content-Type": {"application/json"}},
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			data, _ := json.Marshal(tc.body)
			body := bytes.NewReader(data)
			tc.expected.Body = io.NopCloser(body)
			req, err := tc.cfg.createRequest(context.Background(), tc.body)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(req, tc.expected, cmpopts.IgnoreUnexported(http.Request{}, bytes.Reader{}), cmpopts.IgnoreFields(http.Request{}, "GetBody")); diff != "" {
				t.Fatalf("createRequest() mismatch (-got +want):\n%s", diff)
			}
		})
	}
}
