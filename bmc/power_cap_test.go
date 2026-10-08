package bmc

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockPowerMetricsGetter struct {
	metrics PowerMetrics
	err     error
}

func (m *mockPowerMetricsGetter) GetPowerMetrics(ctx context.Context) (PowerMetrics, error) {
	return m.metrics, m.err
}

func (m *mockPowerMetricsGetter) Name() string {
	return "mock"
}

type mockPowerCapSetter struct {
	err error
}

func (m *mockPowerCapSetter) SetPowerCap(ctx context.Context, _ *float64) error {
	return m.err
}

func (m *mockPowerCapSetter) Name() string {
	return "mock"
}

func TestGetPowerMetricsFromInterfaces(t *testing.T) {
	limit := 800.0
	testCases := []struct {
		name            string
		generic         []interface{}
		errMsg          string
		expectedMetrics PowerMetrics
	}{
		{
			name:            "success",
			generic:         []interface{}{&mockPowerMetricsGetter{metrics: PowerMetrics{ConsumedWatts: 234, CapacityWatts: 1400, LimitInWatts: &limit}}},
			expectedMetrics: PowerMetrics{ConsumedWatts: 234, CapacityWatts: 1400, LimitInWatts: &limit},
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no PowerMetricsGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no PowerMetricsGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockPowerMetricsGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			metrics, _, err := GetPowerMetricsFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedMetrics, metrics)
		})
	}
}

func TestSetPowerCapFromInterfaces(t *testing.T) {
	limit := 800.0
	testCases := []struct {
		name    string
		generic []interface{}
		limit   *float64
		errMsg  string
	}{
		{
			name:    "success, set",
			generic: []interface{}{&mockPowerCapSetter{}},
			limit:   &limit,
		},
		{
			name:    "success, clear",
			generic: []interface{}{&mockPowerCapSetter{}},
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no PowerCapSetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no PowerCapSetter implementations found",
		},
		{
			name:    "error from setter",
			generic: []interface{}{&mockPowerCapSetter{err: errors.New("foobar")}},
			limit:   &limit,
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SetPowerCapFromInterfaces(context.Background(), tt.generic, tt.limit)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}
