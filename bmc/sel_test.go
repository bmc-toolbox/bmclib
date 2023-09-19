package bmc

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockSystemEventLogService struct {
	name string
	err  error
}

func (m *mockSystemEventLogService) ClearSystemEventLog(ctx context.Context) error {
	return m.err
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
	assert.Nil(t, err)
	assert.Equal(t, mockService.name, metadata.SuccessfulProvider)

	// Test with a mock SystemEventLogService that returns an error
	mockService = &mockSystemEventLogService{name: "mock2", err: errors.New("mock error")}
	metadata, err = clearSystemEventLog(ctx, timeout, []systemEventLogProviders{{name: mockService.name, systemEventLogProvider: mockService}})
	assert.NotNil(t, err)
	assert.NotEqual(t, mockService.name, metadata.SuccessfulProvider)
}

func TestClearSystemEventLogFromInterfaces(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with an empty slice
	metadata, err := ClearSystemEventLogFromInterfaces(ctx, timeout, []interface{}{})
	assert.NotNil(t, err)
	assert.Empty(t, metadata.SuccessfulProvider)

	// Test with a slice containing a non-SystemEventLog object
	metadata, err = ClearSystemEventLogFromInterfaces(ctx, timeout, []interface{}{"not a SystemEventLog Service"})
	assert.NotNil(t, err)
	assert.Empty(t, metadata.SuccessfulProvider)

	// Test with a slice containing a mock SystemEventLogService that returns nil
	mockService := &mockSystemEventLogService{name: "mock1"}
	metadata, err = ClearSystemEventLogFromInterfaces(ctx, timeout, []interface{}{mockService})
	assert.Nil(t, err)
	assert.Equal(t, mockService.name, metadata.SuccessfulProvider)
}
