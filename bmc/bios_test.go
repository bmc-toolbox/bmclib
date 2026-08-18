package bmc

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// mockBiosConfigurationMapSetter implements only SetBiosConfiguration, matching
// providers (e.g. Redfish, iDRAC, Lenovo) that have no file-import equivalent.
// Regression coverage for the BiosConfigurationSetter/BiosConfigurationFileSetter
// split: before the split this type failed to satisfy the combined interface at
// all, so SetBiosConfigurationInterfaces reported "no implementations found"
// even though SetBiosConfiguration itself worked fine.
type mockBiosConfigurationMapSetter struct {
	err error
}

func (m *mockBiosConfigurationMapSetter) SetBiosConfiguration(ctx context.Context, _ map[string]string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.err
	}
}

func (m *mockBiosConfigurationMapSetter) Name() string {
	return "mock"
}

// mockBiosConfigurationFileSetter implements only SetBiosConfigurationFromFile,
// matching providers (e.g. Dell SUM-style tools) that only support whole-config
// file import, not the generic map-based path.
type mockBiosConfigurationFileSetter struct {
	err error
}

func (m *mockBiosConfigurationFileSetter) SetBiosConfigurationFromFile(ctx context.Context, _ string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return m.err
	}
}

func (m *mockBiosConfigurationFileSetter) Name() string {
	return "mock"
}

func TestSetBiosConfigurationInterfaces(t *testing.T) {
	testCases := []struct {
		name             string
		generic          []interface{}
		biosConfig       map[string]string
		errMsg           string
		expectedMetadata Metadata
	}{
		{
			name:       "success with a map-only setter",
			generic:    []interface{}{&mockBiosConfigurationMapSetter{}},
			biosConfig: map[string]string{"NetworkStack": "Enabled"},
			expectedMetadata: Metadata{
				SuccessfulProvider:   "mock",
				ProvidersAttempted:   []string{"mock"},
				FailedProviderDetail: make(map[string]string),
			},
		},
		{
			name:             "a file-only setter does not satisfy BiosConfigurationSetter",
			generic:          []interface{}{&mockBiosConfigurationFileSetter{}},
			errMsg:           "no BiosConfigurationSetter implementations found",
			expectedMetadata: Metadata{},
		},
		{
			name:             "no setters",
			generic:          []interface{}{},
			errMsg:           "no BiosConfigurationSetter implementations found",
			expectedMetadata: Metadata{},
		},
		{
			name:    "error from setter",
			generic: []interface{}{&mockBiosConfigurationMapSetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
			expectedMetadata: Metadata{
				ProvidersAttempted:   []string{"mock"},
				FailedProviderDetail: map[string]string{},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := SetBiosConfigurationInterfaces(context.Background(), tt.generic, tt.biosConfig)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedMetadata, metadata)
		})
	}
}

func TestSetBiosConfigurationFromFileInterfaces(t *testing.T) {
	testCases := []struct {
		name             string
		generic          []interface{}
		errMsg           string
		expectedMetadata Metadata
	}{
		{
			name:    "success with a file-only setter",
			generic: []interface{}{&mockBiosConfigurationFileSetter{}},
			expectedMetadata: Metadata{
				SuccessfulProvider:   "mock",
				ProvidersAttempted:   []string{"mock"},
				FailedProviderDetail: make(map[string]string),
			},
		},
		{
			name:             "a map-only setter does not satisfy BiosConfigurationFileSetter",
			generic:          []interface{}{&mockBiosConfigurationMapSetter{}},
			errMsg:           "no BiosConfigurationFileSetter implementations found",
			expectedMetadata: Metadata{},
		},
		{
			name:             "no setters",
			generic:          []interface{}{},
			errMsg:           "no BiosConfigurationFileSetter implementations found",
			expectedMetadata: Metadata{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := SetBiosConfigurationFromFileInterfaces(context.Background(), tt.generic, "cfg")

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedMetadata, metadata)
		})
	}
}

// mockBiosConfigurationSetter implements both SetBiosConfiguration and
// SetBiosConfigurationFromFile, matching providers (e.g. Supermicro) that
// support both paths — proves the split doesn't break dual-implementers.
type mockBiosConfigurationSetter struct {
	mockBiosConfigurationMapSetter
	mockBiosConfigurationFileSetter
}

func (m *mockBiosConfigurationSetter) Name() string {
	return "mock"
}

func TestSetBiosConfigurationInterfaces_dualImplementer(t *testing.T) {
	dual := &mockBiosConfigurationSetter{}

	_, err := SetBiosConfigurationInterfaces(context.Background(), []interface{}{dual}, map[string]string{})
	assert.NoError(t, err)

	_, err = SetBiosConfigurationFromFileInterfaces(context.Background(), []interface{}{dual}, "cfg")
	assert.NoError(t, err)
}
