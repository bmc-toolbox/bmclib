package redfishwrapper

import (
	"testing"

	"github.com/stmcginnis/gofish/schemas"
	"github.com/stretchr/testify/assert"
)

func TestMatchingSystem(t *testing.T) {
	tests := map[string]struct {
		client       *Client
		systems      []*schemas.ComputerSystem
		expectErr    bool
		expectSystem *schemas.ComputerSystem
	}{
		"finds matching system by name": {
			client: &Client{
				systemName: "System1",
			},
			systems: []*schemas.ComputerSystem{
				{
					Entity: schemas.Entity{
						Name: "System1",
					},
				},
				{
					Entity: schemas.Entity{
						Name: "System2",
					},
				},
			},
			expectErr: false,
			expectSystem: &schemas.ComputerSystem{
				Entity: schemas.Entity{
					Name: "System1",
				},
			},
		},
		"no matching system found": {
			client: &Client{
				systemName: "NonExistent",
			},
			systems: []*schemas.ComputerSystem{
				{
					Entity: schemas.Entity{
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
			systems:      []*schemas.ComputerSystem{},
			expectErr:    true,
			expectSystem: nil,
		},
		"system name empty": {
			client: &Client{
				systemName: "",
			},
			systems: []*schemas.ComputerSystem{
				{
					Entity: schemas.Entity{
						Name: "System1",
					},
				},
				{
					Entity: schemas.Entity{
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
