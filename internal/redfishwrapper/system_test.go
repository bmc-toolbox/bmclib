package redfishwrapper

import (
	"testing"

	"github.com/stmcginnis/gofish/common"
	"github.com/stmcginnis/gofish/redfish"
	"github.com/stretchr/testify/assert"
)

func TestMatchingSystem(t *testing.T) {
	tests := map[string]struct {
		client      *Client
		systems     []*redfish.ComputerSystem
		expectCount int
		expectNames []string
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
			expectCount: 1,
			expectNames: []string{"System1"},
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
			expectCount: 0,
			expectNames: []string{},
		},
		"empty systems list": {
			client: &Client{
				systemName: "System1",
			},
			systems:     []*redfish.ComputerSystem{},
			expectCount: 0,
			expectNames: []string{},
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
			expectCount: 0,
			expectNames: []string{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := tc.client.matchingSystem(tc.systems)

			assert.Len(t, result, tc.expectCount)

			for i, system := range result {
				if i < len(tc.expectNames) {
					assert.Equal(t, tc.expectNames[i], system.Name)
				}
			}
		})
	}
}
