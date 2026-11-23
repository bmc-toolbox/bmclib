package redfishwrapper

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockupRequest struct {
	Method string
	Path   string
	// Body is recorded for requests other than GET.
	Body string
}

// mockupServer serves a Redfish-Mockup-Creator tree, where the resource at
// /redfish/v1/X lives in <root>/redfish/v1/X/index.json, and records every request.
// Requests other than GET succeed with 204 so tests can assert on what was sent.
type mockupServer struct {
	root     string
	mu       sync.Mutex
	requests []mockupRequest
}

func (m *mockupServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p := path.Clean(r.URL.Path)

	req := mockupRequest{Method: r.Method, Path: p}
	if r.Method != http.MethodGet {
		body, _ := io.ReadAll(r.Body)
		req.Body = string(body)
	}

	m.mu.Lock()
	m.requests = append(m.requests, req)
	m.mu.Unlock()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	body, err := os.ReadFile(filepath.Join(m.root, filepath.FromSlash(p), "index.json"))
	if err != nil {
		notFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(body)
}

func (m *mockupServer) Requests() []mockupRequest {
	m.mu.Lock()
	defer m.mu.Unlock()

	return append([]mockupRequest(nil), m.requests...)
}

// requested reports whether any request was made for p or a resource under it.
func (m *mockupServer) requested(p string) bool {
	for _, r := range m.Requests() {
		if r.Path == p || strings.HasPrefix(r.Path, p+"/") {
			return true
		}
	}

	return false
}

// writes returns the requests that were not GETs.
func (m *mockupServer) writes() []mockupRequest {
	var out []mockupRequest
	for _, r := range m.Requests() {
		if r.Method != http.MethodGet {
			out = append(out, r)
		}
	}

	return out
}

// newMockupClient opens a client against the mockup tree at root.
func newMockupClient(t *testing.T, root string, opts ...Option) (*Client, *mockupServer) {
	t.Helper()

	mockup := &mockupServer{root: root}
	server := httptest.NewTLSServer(mockup)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	opts = append([]Option{WithBasicAuthEnabled(true)}, opts...)
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", opts...)

	ctx := context.Background()
	require.NoError(t, client.Open(ctx))
	t.Cleanup(func() { _ = client.Close(ctx) })

	t.Cleanup(func() {
		if !t.Failed() {
			return
		}
		for _, r := range mockup.Requests() {
			t.Logf("mockup request: %s %s", r.Method, r.Path)
		}
	})

	return client, mockup
}

// writeMockup builds a mockup tree in a temporary directory from resources keyed
// by their Redfish path, and returns its root.
func writeMockup(t *testing.T, resources map[string]string) string {
	t.Helper()

	root := t.TempDir()
	for p, body := range resources {
		dir := filepath.Join(root, filepath.FromSlash(p))
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "index.json"), []byte(body), 0o600))
	}

	return root
}
