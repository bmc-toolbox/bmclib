package redfishwrapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithVersionsNotCompatible(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name     string
		versions []string
	}{
		{
			"no versions",
			[]string{},
		},
		{
			"with versions",
			[]string{"1.2.3", "4.5.6"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithVersionsNotCompatible(tt.versions))
			assert.Equal(t, tt.versions, client.versionsNotCompatible)
		})
	}
}

func TestWithBasicAuthEnabled(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			"disabled",
			false,
		},
		{
			"enabled",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithBasicAuthEnabled(tt.enabled))
			assert.Equal(t, tt.enabled, client.basicAuth)
		})
	}
}

func TestWithEtagMatchDisabled(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name     string
		disabled bool
	}{
		{
			"disabled",
			true,
		},
		{
			"enabled",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(host, "", user, pass, WithEtagMatchDisabled(tt.disabled))
			assert.Equal(t, tt.disabled, client.disableEtagMatch)
		})
	}
}
