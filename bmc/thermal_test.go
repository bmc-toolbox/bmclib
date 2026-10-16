package bmc

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockThermalGetter struct {
	thermal ChassisThermal
	err     error
}

func (m *mockThermalGetter) GetChassisThermal(ctx context.Context) (ChassisThermal, error) {
	return m.thermal, m.err
}

func (m *mockThermalGetter) Name() string {
	return "mock"
}

func TestGetChassisThermalFromInterfaces(t *testing.T) {
	celsius := 23.0
	rpm := 6840
	thermal := ChassisThermal{
		Temperatures: []ChassisTemperature{{Name: "Ambient Temp", ReadingCelsius: &celsius, Health: "OK", State: "Enabled"}},
		Fans:         []ChassisFan{{Name: "Fan 1 Tach", Reading: &rpm, ReadingUnits: "RPM", Health: "OK", State: "Enabled"}},
	}

	testCases := []struct {
		name            string
		generic         []interface{}
		errMsg          string
		expectedThermal ChassisThermal
	}{
		{
			name:            "success",
			generic:         []interface{}{&mockThermalGetter{thermal: thermal}},
			expectedThermal: thermal,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ChassisThermalGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ChassisThermalGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockThermalGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := GetChassisThermalFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedThermal, got)
		})
	}
}
