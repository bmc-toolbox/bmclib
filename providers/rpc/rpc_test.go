package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestOpen(t *testing.T) {
	tests := map[string]struct {
		url       string
		shouldErr bool
	}{
		"success":        {},
		"bad url":        {url: "%", shouldErr: true},
		"failed request": {url: "127.1.1.1", shouldErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			svr := ResponsePayload{}.testConsumer()
			defer svr.Close()

			u := svr.URL
			if tc.url != "" {
				u = tc.url
			}
			c := New(u, "127.0.1.1", map[Algorithm][]string{SHA256: []string{"superSecret1"}})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if err := c.Open(ctx); err != nil && !tc.shouldErr {
				t.Fatal(err)
			}
			c.Close(ctx)
		})
	}
}

func TestBootDeviceSet(t *testing.T) {
	tests := map[string]struct {
		url       string
		shouldErr bool
	}{
		"success":               {},
		"failure from consumer": {shouldErr: true},
		"failed request":        {url: "127.1.1.1", shouldErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rsp := ResponsePayload{}
			if tc.shouldErr {
				rsp.Error = &ResponseError{Code: 500, Message: "failed"}
			}
			svr := rsp.testConsumer()
			defer svr.Close()

			u := svr.URL
			if tc.url != "" {
				u = tc.url
			}
			c := New(u, "127.0.1.1", map[Algorithm][]string{SHA256: {"superSecret1"}})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_ = c.Open(ctx)
			if _, err := c.BootDeviceSet(ctx, "pxe", false, false); err != nil && !tc.shouldErr {
				t.Fatal(err)
			} else if err == nil && tc.shouldErr {
				t.Fatal("expected error, got none")
			}
		})
	}
}

func TestPowerSet(t *testing.T) {
	tests := map[string]struct {
		url        string
		powerState string
		shouldErr  bool
	}{
		"success":               {powerState: "on"},
		"failed request":        {powerState: "on", url: "127.1.1.1", shouldErr: true},
		"unknown state":         {powerState: "unknown", shouldErr: true},
		"failure from consumer": {powerState: "on", shouldErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rsp := ResponsePayload{Result: tc.powerState}
			if tc.shouldErr {
				rsp.Error = &ResponseError{Code: 500, Message: "failed"}
			}
			svr := rsp.testConsumer()
			defer svr.Close()

			u := svr.URL
			if tc.url != "" {
				u = tc.url
			}
			c := New(u, "127.0.1.1", map[Algorithm][]string{SHA256: {"superSecret1"}})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			_ = c.Open(ctx)

			_, err := c.PowerSet(ctx, tc.powerState)
			if err != nil && !tc.shouldErr {
				t.Fatal(err)
			}

		})
	}
}

func TestPowerStateGet(t *testing.T) {
	tests := map[string]struct {
		powerState string
		shouldErr  bool
		url        string
	}{
		"success":       {},
		"unknown state": {shouldErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rsp := ResponsePayload{Result: tc.powerState}
			if tc.shouldErr {
				rsp.Error = &ResponseError{Code: 500, Message: "failed"}
			}
			svr := rsp.testConsumer()
			defer svr.Close()

			u := svr.URL
			if tc.url != "" {
				u = tc.url
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			c := New(u, "127.0.1.1", map[Algorithm][]string{SHA256: {"superSecret1"}})
			_ = c.Open(ctx)
			gotState, err := c.PowerStateGet(ctx)
			if err != nil && !tc.shouldErr {
				t.Fatal(err)
			}
			if diff := cmp.Diff(gotState, tc.powerState); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func (rs ResponsePayload) testConsumer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := json.Marshal(rs)
		_, _ = w.Write(b)
	}))
}
