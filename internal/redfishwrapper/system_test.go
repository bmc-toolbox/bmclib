package redfishwrapper

import (
	"testing"

	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/stretchr/testify/assert"
)

func TestMatchingSystem(t *testing.T) {
	tests := map[string]struct {
		client       *Client
		systems      []*redfish.ComputerSystem
		expectErr    bool
		expectSystem *redfish.ComputerSystem
	}{
		"finds matching system by name": {
			client: &Client{
				systemName: "System1",
			},
			systems: []*redfish.ComputerSystem{
				{
					Entity: common.Entity{
						Name: "System1",
					},
				},
				{
					Entity: common.Entity{
						Name: "System2",
					},
				},
			},
			expectErr: false,
			expectSystem: &redfish.ComputerSystem{
				Entity: common.Entity{
					Name: "System1",
				},
			},
		},
		"no matching system found": {
			client: &Client{
				systemName: "NonExistent",
			},
			systems: []*redfish.ComputerSystem{
				{
					Entity: common.Entity{
						Name: "System1",
					},
				},
			},
			expectErr:    true,
			expectSystem: nil,
		},
		"empty systems list": {
			client: &Client{
				systemName: "System1",
			},
			systems:      []*redfish.ComputerSystem{},
			expectErr:    true,
			expectSystem: nil,
		},
		"system name empty": {
			client: &Client{
				systemName: "",
			},
			systems: []*redfish.ComputerSystem{
				{
					Entity: common.Entity{
						Name: "System1",
					},
				},
				{
					Entity: common.Entity{
						Name: "System2",
					},
				},
			},
			expectErr:    true,
			expectSystem: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := tc.client.matchingSystem(tc.systems)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectSystem, result)
			}
		})
	}
}
