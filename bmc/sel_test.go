package bmc

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockSELService struct {
	name string
	err  error
}

func (m *mockSELService) ClearSEL(ctx context.Context) error {
	return m.err
}

func (m *mockSELService) Name() string {
	return m.name
}

func TestClearSEL(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with a mock SELService that returns nil
	mockService := &mockSELService{name: "mock1", err: nil}
	metadata, err := clearSEL(ctx, timeout, []selProviders{{name: mockService.name, selProvider: mockService}})
	assert.Nil(t, err)
	assert.Equal(t, mockService.name, metadata.SuccessfulProvider)

	// Test with a mock SELService that returns an error
	mockService = &mockSELService{name: "mock2", err: errors.New("mock error")}
	metadata, err = clearSEL(ctx, timeout, []selProviders{{name: mockService.name, selProvider: mockService}})
	assert.NotNil(t, err)
	assert.NotEqual(t, mockService.name, metadata.SuccessfulProvider)
}

func TestClearSELFromInterfaces(t *testing.T) {
	ctx := context.Background()
	timeout := 1 * time.Second

	// Test with an empty slice
	metadata, err := ClearSELFromInterfaces(ctx, timeout, []interface{}{})
	assert.NotNil(t, err)
	assert.Empty(t, metadata.SuccessfulProvider)

	// Test with a slice containing a non-SELService object
	metadata, err = ClearSELFromInterfaces(ctx, timeout, []interface{}{"not a SELService"})
	assert.NotNil(t, err)
	assert.Empty(t, metadata.SuccessfulProvider)

	// Test with a slice containing a mock SELService
	mockService := &mockSELService{name: "mock1"}
	metadata, err = ClearSELFromInterfaces(ctx, timeout, []interface{}{mockService})
	assert.Nil(t, err)
	assert.Equal(t, mockService.name, metadata.SuccessfulProvider)
}
