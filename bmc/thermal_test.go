package bmc

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockFanSpeedsGetter struct {
	fans []ChassisFan
	err  error
}

func (m *mockFanSpeedsGetter) GetChassisFanSpeeds(ctx context.Context) ([]ChassisFan, error) {
	return m.fans, m.err
}

func (m *mockFanSpeedsGetter) Name() string {
	return "mock"
}

type mockTemperaturesGetter struct {
	temperatures []ChassisTemperature
	err          error
}

func (m *mockTemperaturesGetter) GetChassisTemperatures(ctx context.Context) ([]ChassisTemperature, error) {
	return m.temperatures, m.err
}

func (m *mockTemperaturesGetter) Name() string {
	return "mock"
}

func TestGetChassisFanSpeedsFromInterfaces(t *testing.T) {
	rpm := 6840.0
	fans := []ChassisFan{{Name: "Fan 1 Tach", SpeedRPM: &rpm, Health: "OK", State: "Enabled"}}

	testCases := []struct {
		name         string
		generic      []interface{}
		errMsg       string
		expectedFans []ChassisFan
	}{
		{
			name:         "success",
			generic:      []interface{}{&mockFanSpeedsGetter{fans: fans}},
			expectedFans: fans,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ChassisFanSpeedsGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ChassisFanSpeedsGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockFanSpeedsGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := GetChassisFanSpeedsFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedFans, got)
		})
	}
}

func TestGetChassisTemperaturesFromInterfaces(t *testing.T) {
	celsius := 23.0
	temperatures := []ChassisTemperature{{Name: "Ambient Temp", ReadingCelsius: &celsius, Health: "OK", State: "Enabled"}}

	testCases := []struct {
		name                 string
		generic              []interface{}
		errMsg               string
		expectedTemperatures []ChassisTemperature
	}{
		{
			name:                 "success",
			generic:              []interface{}{&mockTemperaturesGetter{temperatures: temperatures}},
			expectedTemperatures: temperatures,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ChassisTemperaturesGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ChassisTemperaturesGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockTemperaturesGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := GetChassisTemperaturesFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedTemperatures, got)
		})
	}
}
