package bmc

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockSecureBootStateGetter struct {
	enabled bool
	err     error
}

func (m *mockSecureBootStateGetter) GetSecureBoot(ctx context.Context) (bool, error) {
	return m.enabled, m.err
}

func (m *mockSecureBootStateGetter) Name() string {
	return "mock"
}

type mockSecureBootSetter struct {
	err error
}

func (m *mockSecureBootSetter) SetSecureBoot(ctx context.Context, _ bool) error {
	return m.err
}

func (m *mockSecureBootSetter) Name() string {
	return "mock"
}

type mockSecureBootKeysResetter struct {
	err error
}

func (m *mockSecureBootKeysResetter) ResetSecureBootKeys(ctx context.Context, _ string) error {
	return m.err
}

func (m *mockSecureBootKeysResetter) Name() string {
	return "mock"
}

func TestGetSecureBootStateFromInterfaces(t *testing.T) {
	testCases := []struct {
		name            string
		generic         []interface{}
		errMsg          string
		expectedEnabled bool
	}{
		{
			name:            "success, enabled",
			generic:         []interface{}{&mockSecureBootStateGetter{enabled: true}},
			expectedEnabled: true,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no SecureBootStateGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no SecureBootStateGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockSecureBootStateGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			enabled, _, err := GetSecureBootStateFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedEnabled, enabled)
		})
	}
}

func TestSetSecureBootFromInterfaces(t *testing.T) {
	testCases := []struct {
		name    string
		generic []interface{}
		errMsg  string
	}{
		{
			name:    "success",
			generic: []interface{}{&mockSecureBootSetter{}},
		},
		{
			name:    "not an implementation",
			generic: []interface{}{&mockSecureBootStateGetter{}},
			errMsg:  "no SecureBootSetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no SecureBootSetter implementations found",
		},
		{
			name:    "error from setter",
			generic: []interface{}{&mockSecureBootSetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SetSecureBootFromInterfaces(context.Background(), tt.generic, true)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}

func TestResetSecureBootKeysFromInterfaces(t *testing.T) {
	testCases := []struct {
		name    string
		generic []interface{}
		errMsg  string
	}{
		{
			name:    "success",
			generic: []interface{}{&mockSecureBootKeysResetter{}},
		},
		{
			name:    "not an implementation",
			generic: []interface{}{&mockSecureBootStateGetter{}},
			errMsg:  "no SecureBootKeysResetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no SecureBootKeysResetter implementations found",
		},
		{
			name:    "error from resetter",
			generic: []interface{}{&mockSecureBootKeysResetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResetSecureBootKeysFromInterfaces(context.Background(), tt.generic, "ResetAllKeysToDefault")

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}
