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

func (m *mockSecureBootKeysResetter) ResetSecureBootKeys(ctx context.Context, _ ResetSecureBootKeysType) error {
	return m.err
}

func (m *mockSecureBootKeysResetter) Name() string {
	return "mock"
}

type mockSecureBootDatabaseKeysResetter struct {
	err error
}

func (m *mockSecureBootDatabaseKeysResetter) ResetSecureBootDatabaseKeys(ctx context.Context, _ SecureBootDatabase, _ ResetSecureBootDatabaseKeysType) error {
	return m.err
}

func (m *mockSecureBootDatabaseKeysResetter) Name() string {
	return "mock"
}

type mockSecureBootCertificateImporter struct {
	err error
}

func (m *mockSecureBootCertificateImporter) ImportSecureBootCertificate(ctx context.Context, _ SecureBootDatabase, _ string) error {
	return m.err
}

func (m *mockSecureBootCertificateImporter) Name() string {
	return "mock"
}

type mockSecureBootKeyManagementSetter struct {
	rebootRequired bool
	err            error
}

func (m *mockSecureBootKeyManagementSetter) SetSecureBootKeyManagement(ctx context.Context, _ bool) (bool, error) {
	return m.rebootRequired, m.err
}

func (m *mockSecureBootKeyManagementSetter) Name() string {
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
			_, err := ResetSecureBootKeysFromInterfaces(context.Background(), tt.generic, ResetSecureBootKeysTypeResetAllKeysToDefault)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}

func TestResetSecureBootDatabaseKeysFromInterfaces(t *testing.T) {
	testCases := []struct {
		name    string
		generic []interface{}
		errMsg  string
	}{
		{
			name:    "success",
			generic: []interface{}{&mockSecureBootDatabaseKeysResetter{}},
		},
		{
			name:    "not an implementation",
			generic: []interface{}{&mockSecureBootStateGetter{}},
			errMsg:  "no SecureBootDatabaseKeysResetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no SecureBootDatabaseKeysResetter implementations found",
		},
		{
			name:    "error from resetter",
			generic: []interface{}{&mockSecureBootDatabaseKeysResetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ResetSecureBootDatabaseKeysFromInterfaces(context.Background(), tt.generic, SecureBootDatabaseDB, ResetSecureBootDatabaseKeysTypeResetAllKeysToDefault)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}

func TestImportSecureBootCertificateFromInterfaces(t *testing.T) {
	testCases := []struct {
		name    string
		generic []interface{}
		errMsg  string
	}{
		{
			name:    "success",
			generic: []interface{}{&mockSecureBootCertificateImporter{}},
		},
		{
			name:    "not an implementation",
			generic: []interface{}{&mockSecureBootStateGetter{}},
			errMsg:  "no SecureBootCertificateImporter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no SecureBootCertificateImporter implementations found",
		},
		{
			name:    "error from importer",
			generic: []interface{}{&mockSecureBootCertificateImporter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ImportSecureBootCertificateFromInterfaces(context.Background(), tt.generic, SecureBootDatabaseDB, "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----")

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}

func TestSetSecureBootKeyManagementFromInterfaces(t *testing.T) {
	testCases := []struct {
		name                   string
		generic                []interface{}
		errMsg                 string
		expectedRebootRequired bool
	}{
		{
			name:                   "success, reboot required",
			generic:                []interface{}{&mockSecureBootKeyManagementSetter{rebootRequired: true}},
			expectedRebootRequired: true,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{&mockSecureBootStateGetter{}},
			errMsg:  "no SecureBootKeyManagementSetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no SecureBootKeyManagementSetter implementations found",
		},
		{
			name:    "error from enabler",
			generic: []interface{}{&mockSecureBootKeyManagementSetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			rebootRequired, _, err := SetSecureBootKeyManagementFromInterfaces(context.Background(), tt.generic, true)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedRebootRequired, rebootRequired)
		})
	}
}
