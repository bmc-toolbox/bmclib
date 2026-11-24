package bmclib

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// fakeIpmitool returns the path of an empty file, so the ipmitool provider registers without
// the ipmitool binary installed. It is never run: the tests don't open the client.
func fakeIpmitool(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ipmitool")
	if err := os.WriteFile(path, nil, 0o755); err != nil {
		t.Fatal(err)
	}

	return path
}

// TestProviderNames checks that ProviderNames lists the providers NewClient registers by default,
// in the same order.
func TestProviderNames(t *testing.T) {
	cl := NewClient("127.0.0.1", "user", "pass", WithIpmitoolPath(fakeIpmitool(t)))
	if diff := cmp.Diff(ProviderNames(), registryNames(cl.Registry.Drivers)); diff != "" {
		t.Errorf("ProviderNames() and the registered providers differ (-want +got):\n%s", diff)
	}
}

func TestWithProviders(t *testing.T) {
	tests := map[string]struct {
		providers []string
		want      []string
	}{
		"no providers named registers all": {
			want: ProviderNames(),
		},
		"one provider": {
			providers: []string{"gofish"},
			want:      []string{"gofish"},
		},
		"names compared without regard to case, registration order kept": {
			providers: []string{"INTELAMT", "GoFish"},
			want:      []string{"gofish", "IntelAMT"},
		},
		"unknown names are ignored": {
			providers: []string{"gofish", "nonexistent"},
			want:      []string{"gofish"},
		},
		"only unknown names registers none": {
			providers: []string{"nonexistent"},
			want:      []string{},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cl := NewClient("127.0.0.1", "user", "pass", WithIpmitoolPath(fakeIpmitool(t)), WithProviders(tc.providers...))
			if diff := cmp.Diff(tc.want, registryNames(cl.Registry.Drivers)); diff != "" {
				t.Errorf("registered providers mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
