package rpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

/*
func TestMerge(t *testing.T) {
	c := New("http://example.com", "127.0.0.1", Secrets{SHA256: {"superSecret1"}})
	// control := New("http://example.com", "127.0.0.1", Secrets{SHA256: {"superSecret1"}})
	customized := &Config{LogNotifications: boolPTR(false), Opts: Opts{Signature: SignatureOpts{IncludeAlgoPrefix: false}}}
	want := &Config{
		Host:             "127.0.0.1",
		ConsumerURL:      "http://example.com",
		Logger:           logr.Discard(),
		LogNotifications: boolPTR(true),
		Opts: Opts{
			Request: RequestOpts{
				HTTPContentType: "application/json",
				HTTPMethod:      http.MethodPost,
				TimestampHeader: timestampHeader,
				TimestampFormat: time.RFC3339,
				Client:          http.DefaultClient,
			},
			Signature: SignatureOpts{
				IncludeAlgoPrefix: true,
			},
		},
		/*
			Sig: hmac.Signature{
				HeaderName:             "X-Bmclib-Signature",
				AppendAlgoToHeader:     true,
				IncludedPayloadHeaders: []string{timestampHeader},
				HMAC: &hmac.Conf{
					Hashes:    hmac.NewSHA256("superSecret1"),
					PrefixSig: true,
				},
			},

	}
	t.Log(want)

	t.Logf("before: %+v", c)
	mergo.Merge(c, customized, mergo.WithOverride, mergo.WithTransformers(&Config{}))
	t.Logf("after:  %+v", c)

	h := map[Algorithm][]hash.Hash{}

	if diff := cmp.Diff(c, want, cmpopts.IgnoreUnexported(Config{}, Signature{}), cmpopts.IgnoreTypes(logr.Logger{}, h)); diff != "" {
		t.Fatalf("mismatch (+want -got):\n%s", diff)
	}

	t.Fatal()
}
*/

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
			svr := testConsumer{rp: ResponsePayload{}}.testServer()
			defer svr.Close()

			u := svr.URL
			if tc.url != "" {
				u = tc.url
			}
			c := New(u, "127.0.1.1", Secrets{SHA256: []string{"superSecret1"}})
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
			rsp := testConsumer{
				rp: ResponsePayload{},
			}
			if tc.shouldErr {
				rsp.rp.Error = &ResponseError{Code: 500, Message: "failed"}
			}
			svr := rsp.testServer()
			defer svr.Close()

			u := svr.URL
			if tc.url != "" {
				u = tc.url
			}
			c := New(u, "127.0.1.1", Secrets{SHA256: {"superSecret1"}})
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
			rsp := testConsumer{
				rp: ResponsePayload{Result: tc.powerState},
			}
			if tc.shouldErr {
				rsp.rp.Error = &ResponseError{Code: 500, Message: "failed"}
			}
			svr := rsp.testServer()
			defer svr.Close()

			u := svr.URL
			if tc.url != "" {
				u = tc.url
			}
			c := New(u, "127.0.1.1", Secrets{SHA256: {"superSecret1"}})
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
			rsp := testConsumer{
				rp: ResponsePayload{Result: tc.powerState},
			}
			if tc.shouldErr {
				rsp.rp.Error = &ResponseError{Code: 500, Message: "failed"}
			}
			svr := rsp.testServer()
			defer svr.Close()

			u := svr.URL
			if tc.url != "" {
				u = tc.url
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			c := New(u, "127.0.1.1", Secrets{SHA256: {"superSecret1"}})
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

func TestServerErrors(t *testing.T) {
	tests := map[string]struct {
		statusCode int
		shouldErr  bool
	}{
		"bad request": {statusCode: http.StatusBadRequest, shouldErr: true},
		"not found":   {statusCode: http.StatusNotFound, shouldErr: true},
		"internal":    {statusCode: http.StatusInternalServerError, shouldErr: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			rsp := testConsumer{
				rp:         ResponsePayload{Result: "on"},
				statusCode: tc.statusCode,
			}
			svr := rsp.testServer()
			defer svr.Close()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			c := New(svr.URL, "127.0.0.1", Secrets{SHA256: {"superSecret1"}})
			if err := c.Open(ctx); err != nil {
				t.Fatal(err)
			}
			_, err := c.PowerStateGet(ctx)
			if err == nil {
				t.Fatal("expected error, got none")
			}
		})
	}
}

type testConsumer struct {
	rp         ResponsePayload
	statusCode int
}

func (t testConsumer) testServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if t.statusCode != 0 {
			w.WriteHeader(t.statusCode)
			return
		}
		b, _ := json.Marshal(t.rp)
		_, _ = w.Write(b)
	}))
}
