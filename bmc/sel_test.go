package bmc

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSystemEventLogService struct {
	name string
	err  error
}

func (m *mockSystemEventLogService) ClearSystemEventLog(ctx context.Context) error {
	return m.err
}

func (m *mockSystemEventLogService) GetSystemEventLog(ctx context.Context) (entries [][]string, err error) {
	return nil, m.err
}

func (m *mockSystemEventLogService) GetSystemEventLogRaw(ctx context.Context) (eventlog string, err error) {
	return "", m.err
}

func (m *mockSystemEventLogService) Name() string {
	return m.name
}

func TestClearSystemEventLog(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with a mock SystemEventLogService that returns nil
	mockService := &mockSystemEventLogService{name: "mock1", err: nil}
	metadata, err := clearSystemEventLog(ctx, timeout, []systemEventLogProviders{{name: mockService.name, systemEventLogProvider: mockService}})
	require.NoError(t, err)
	require.Equal(t, mockService.name, metadata.SuccessfulProvider)

	// Test with a mock SystemEventLogService that returns an error
	mockService = &mockSystemEventLogService{name: "mock2", err: errors.New("mock error")}
	metadata, err = clearSystemEventLog(ctx, timeout, []systemEventLogProviders{{name: mockService.name, systemEventLogProvider: mockService}})
	require.Error(t, err)
	assert.NotEqual(t, mockService.name, metadata.SuccessfulProvider)
}

func TestClearSystemEventLogFromInterfaces(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with an empty slice
	metadata, err := ClearSystemEventLogFromInterfaces(ctx, timeout, []interface{}{})
	require.Error(t, err)
	require.Empty(t, metadata.SuccessfulProvider)

	// Test with a slice containing a non-SystemEventLog object
	metadata, err = ClearSystemEventLogFromInterfaces(ctx, timeout, []interface{}{"not a SystemEventLog Service"})
	require.Error(t, err)
	require.Empty(t, metadata.SuccessfulProvider)

	// Test with a slice containing a mock SystemEventLogService that returns nil
	mockService := &mockSystemEventLogService{name: "mock1"}
	metadata, err = ClearSystemEventLogFromInterfaces(ctx, timeout, []interface{}{mockService})
	require.NoError(t, err)
	assert.Equal(t, mockService.name, metadata.SuccessfulProvider)
}

func TestGetSystemEventLog(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with a mock SystemEventLogService that returns nil
	mockService := &mockSystemEventLogService{name: "mock1", err: nil}
	_, _, err := getSystemEventLog(ctx, timeout, []systemEventLogProviders{{name: mockService.name, systemEventLogProvider: mockService}})
	require.NoError(t, err)

	// Test with a mock SystemEventLogService that returns an error
	mockService = &mockSystemEventLogService{name: "mock2", err: errors.New("mock error")}
	_, _, err = getSystemEventLog(ctx, timeout, []systemEventLogProviders{{name: mockService.name, systemEventLogProvider: mockService}})
	assert.Error(t, err)
}

func TestGetSystemEventLogFromInterfaces(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with an empty slice
	_, _, err := GetSystemEventLogFromInterfaces(ctx, timeout, []interface{}{})
	require.Error(t, err)

	// Test with a slice containing a non-SystemEventLog object
	_, _, err = GetSystemEventLogFromInterfaces(ctx, timeout, []interface{}{"not a SystemEventLog Service"})
	require.Error(t, err)

	// Test with a slice containing a mock SystemEventLogService that returns nil
	mockService := &mockSystemEventLogService{name: "mock1"}
	_, _, err = GetSystemEventLogFromInterfaces(ctx, timeout, []interface{}{mockService})
	assert.NoError(t, err)
}

func TestGetSystemEventLogRaw(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with a mock SystemEventLogService that returns nil
	mockService := &mockSystemEventLogService{name: "mock1", err: nil}
	_, _, err := getSystemEventLogRaw(ctx, timeout, []systemEventLogProviders{{name: mockService.name, systemEventLogProvider: mockService}})
	require.NoError(t, err)

	// Test with a mock SystemEventLogService that returns an error
	mockService = &mockSystemEventLogService{name: "mock2", err: errors.New("mock error")}
	_, _, err = getSystemEventLogRaw(ctx, timeout, []systemEventLogProviders{{name: mockService.name, systemEventLogProvider: mockService}})
	assert.Error(t, err)
}

func TestGetSystemEventLogRawFromInterfaces(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with an empty slice
	_, _, err := GetSystemEventLogRawFromInterfaces(ctx, timeout, []interface{}{})
	require.Error(t, err)

	// Test with a slice containing a non-SystemEventLog object
	_, _, err = GetSystemEventLogRawFromInterfaces(ctx, timeout, []interface{}{"not a SystemEventLog Service"})
	require.Error(t, err)

	// Test with a slice containing a mock SystemEventLogService that returns nil
	mockService := &mockSystemEventLogService{name: "mock1"}
	_, _, err = GetSystemEventLogRawFromInterfaces(ctx, timeout, []interface{}{mockService})
	assert.NoError(t, err)
}
