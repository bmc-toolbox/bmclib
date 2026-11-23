package redfishwrapper

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockupSystem and mockupManager describe the resources of a small mockup tree.
type mockupSystem struct {
	id        string
	managedBy []string
}

type mockupManager struct {
	id          string
	managerType string
	forServers  []string
}

func links(kind string, ids []string) []map[string]string {
	out := make([]map[string]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, map[string]string{"@odata.id": "/redfish/v1/" + kind + "/" + id})
	}

	return out
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()

	b, err := json.Marshal(v)
	require.NoError(t, err)

	return string(b)
}

// managerTopology builds a mockup with the given systems and managers. Each
// system's Name equals its Id, so WithSystemName can select it by Id.
func managerTopology(t *testing.T, systems []mockupSystem, managers []mockupManager) string {
	t.Helper()

	systemIDs := make([]string, 0, len(systems))
	managerIDs := make([]string, 0, len(managers))

	resources := map[string]string{}
	for _, s := range systems {
		systemIDs = append(systemIDs, s.id)
		resources["redfish/v1/Systems/"+s.id] = mustJSON(t, map[string]any{
			"@odata.id": "/redfish/v1/Systems/" + s.id,
			"Id":        s.id,
			"Name":      s.id,
			"Links":     map[string]any{"ManagedBy": links("Managers", s.managedBy)},
		})
	}

	for _, m := range managers {
		managerIDs = append(managerIDs, m.id)
		resources["redfish/v1/Managers/"+m.id] = mustJSON(t, map[string]any{
			"@odata.id":   "/redfish/v1/Managers/" + m.id,
			"Id":          m.id,
			"Name":        m.id,
			"ManagerType": m.managerType,
			"Links":       map[string]any{"ManagerForServers": links("Systems", m.forServers)},
		})
	}

	resources["redfish/v1"] = mustJSON(t, map[string]any{
		"@odata.id":      "/redfish/v1",
		"RedfishVersion": "1.17.0",
		"Systems":        map[string]string{"@odata.id": "/redfish/v1/Systems"},
		"Managers":       map[string]string{"@odata.id": "/redfish/v1/Managers"},
	})
	resources["redfish/v1/Systems"] = mustJSON(t, map[string]any{
		"@odata.id": "/redfish/v1/Systems",
		"Members":   links("Systems", systemIDs),
	})
	resources["redfish/v1/Managers"] = mustJSON(t, map[string]any{
		"@odata.id": "/redfish/v1/Managers",
		"Members":   links("Managers", managerIDs),
	})

	return writeMockup(t, resources)
}

func TestBMCManager(t *testing.T) {
	tests := []struct {
		name       string
		systems    []mockupSystem
		managers   []mockupManager
		systemName string
		want       string
		wantErr    string
	}{
		{
			// The shape of the Lenovo XCC fixture's manager, which has no ManagerForServers.
			name:       "single manager without links, system name set",
			systems:    []mockupSystem{{id: "1"}},
			managers:   []mockupManager{{id: "1", managerType: "BMC"}},
			systemName: "1",
			want:       "1",
		},
		{
			name:     "single manager that is not a BMC",
			systems:  []mockupSystem{{id: "1"}},
			managers: []mockupManager{{id: "CMC", managerType: "EnclosureManager"}},
		},
		{
			name:    "non-BMC manager listed first in ManagedBy",
			systems: []mockupSystem{{id: "1", managedBy: []string{"CMC", "BMC"}}},
			managers: []mockupManager{
				{id: "CMC", managerType: "EnclosureManager", forServers: []string{"1"}},
				{id: "BMC", managerType: "BMC", forServers: []string{"1"}},
			},
			want: "BMC",
		},
		{
			name:    "BMC found through ManagerForServers only",
			systems: []mockupSystem{{id: "1"}, {id: "2"}},
			managers: []mockupManager{
				{id: "BMC_A", managerType: "BMC", forServers: []string{"2"}},
				{id: "BMC_B", managerType: "BMC", forServers: []string{"1"}},
			},
			systemName: "1",
			want:       "BMC_B",
		},
		{
			name:    "no BMC manages the selected system",
			systems: []mockupSystem{{id: "1"}, {id: "2"}},
			managers: []mockupManager{
				{id: "BMC_A", managerType: "BMC", forServers: []string{"2"}},
				{id: "BMC_B", managerType: "BMC", forServers: []string{"2"}},
			},
			systemName: "1",
		},
		{
			name:    "several systems without a system name",
			systems: []mockupSystem{{id: "1"}, {id: "2"}},
			managers: []mockupManager{
				{id: "BMC_A", managerType: "BMC", forServers: []string{"1"}},
				{id: "BMC_B", managerType: "BMC", forServers: []string{"2"}},
			},
			wantErr: "no matching redfish system found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := managerTopology(t, tt.systems, tt.managers)

			var opts []Option
			if tt.systemName != "" {
				opts = append(opts, WithSystemName(tt.systemName))
			}
			client, _ := newMockupClient(t, root, opts...)

			manager, err := client.bmcManager(context.Background())
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			if tt.want == "" {
				assert.Nil(t, manager)
				return
			}
			require.NotNil(t, manager)
			assert.Equal(t, tt.want, manager.ID)
		})
	}
}
