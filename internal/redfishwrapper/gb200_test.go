package redfishwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// The tests in this file pin how the client behaves today against a Supermicro
// GB200 NVL compute tray (BMC firmware 70.00.03.19, HGX firmware GB200Nvl-25.06-A).
//
// The tray lists two systems (System_0, then HGX_Baseboard_0), two managers
// (BMC_0, HGX_BMC_0) and 46 chassis. Several tests document behavior that is
// wrong for this hardware; those say so, and a fix should change their assertions.

// gb200FixtureRoot holds the Redfish mockup of the tray.
var gb200FixtureRoot = "fixtures/smc_gb200"

const gb200SELPath = "/redfish/v1/Systems/System_0/LogServices/SEL"

// newGB200Client opens a client against the GB200 mockup.
func newGB200Client(t *testing.T, opts ...Option) (*Client, *mockupServer) {
	t.Helper()

	return newMockupClient(t, gb200FixtureRoot, opts...)
}

// With two systems and two managers, the client needs a system name to pick one.
func TestGB200SystemSelection(t *testing.T) {
	ctx := context.Background()

	t.Run("without a system name", func(t *testing.T) {
		client, _ := newGB200Client(t)

		_, err := client.SystemPowerStatus(ctx)
		require.ErrorContains(t, err, "no matching redfish system found")

		_, err = client.Manager(ctx)
		require.ErrorContains(t, err, "no matching redfish manager found")
	})

	t.Run("with System_0", func(t *testing.T) {
		client, _ := newGB200Client(t, WithSystemName("System_0"))

		state, err := client.SystemPowerStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, "On", state)

		manager, err := client.Manager(ctx)
		require.NoError(t, err)
		assert.Equal(t, "BMC_0", manager.ID)
	})
}

// Bug: ClearSystemEventLog walks chassis log services only. The first one it meets
// is the GPU XID log, which has no ClearLog action, so the call fails before any
// write. The SEL under System_0, which does have a ClearLog action, is never touched.
func TestGB200ClearSystemEventLog(t *testing.T) {
	client, mockup := newGB200Client(t, WithSystemName("System_0"))

	err := client.ClearSystemEventLog(context.Background())
	require.ErrorContains(t, err, "no target provided")

	assert.True(t, mockup.requested("/redfish/v1/Chassis/HGX_GPU_0/LogServices/XID"))
	assert.Empty(t, mockup.writes())
	assert.False(t, mockup.requested(gb200SELPath))
}

// Bug: GetSystemEventLog reads every log service of every manager. It returns the
// 5 BMC_0 EventLog entries mixed with 5 BMC dump entries, and never reads the
// 5 entries of the SEL under System_0.
func TestGB200GetSystemEventLog(t *testing.T) {
	client, mockup := newGB200Client(t, WithSystemName("System_0"))

	entries, err := client.GetSystemEventLog(context.Background())
	require.NoError(t, err)
	assert.Len(t, entries, 10)

	assert.True(t, mockup.requested("/redfish/v1/Managers/BMC_0/LogServices/Dump/Entries"))
	assert.True(t, mockup.requested("/redfish/v1/Managers/HGX_BMC_0/LogServices/Dump/Entries"))
	assert.False(t, mockup.requested(gb200SELPath))
}

// Inventory reads the selected system and its manager.
//
// Bug: chassis are still filtered by a hardcoded allowlist that none of the tray's
// chassis are on, so the device has no enclosures, mainboard or PSUs. The tray's
// CPUs and GPUs live under HGX_Baseboard_0, which inventory does not read, so
// both lists are empty.
func TestGB200Inventory(t *testing.T) {
	ctx := context.Background()

	t.Run("failOnError false", func(t *testing.T) {
		client, _ := newGB200Client(t, WithSystemName("System_0"))

		device, err := client.Inventory(ctx, false)
		require.NoError(t, err)

		assert.Equal(t, "Supermicro", device.Vendor)
		assert.Equal(t, "GB200 NVL", device.Model)
		assert.Len(t, device.Drives, 9)

		require.NotNil(t, device.BMC)
		assert.Equal(t, "BMC_0", device.BMC.ID)
		require.NotNil(t, device.BMC.Firmware)
		assert.Equal(t, "70.00.03.19", device.BMC.Firmware.Installed)

		assert.Empty(t, device.Enclosures)
		assert.Empty(t, device.Mainboard.Model)
		assert.Empty(t, device.PSUs)
		assert.Empty(t, device.CPUs)
		assert.Empty(t, device.GPUs)
	})

	t.Run("failOnError true", func(t *testing.T) {
		client, _ := newGB200Client(t, WithSystemName("System_0"))

		_, err := client.Inventory(ctx, true)
		require.ErrorIs(t, err, bmclibErrs.ErrRedfishChassisOdataID)
	})

	// With two systems and two managers and no system name, neither the system
	// nor the BMC can be picked, so both sections stay empty.
	t.Run("without a system name", func(t *testing.T) {
		client, _ := newGB200Client(t)

		device, err := client.Inventory(ctx, false)
		require.NoError(t, err)

		assert.Empty(t, device.Vendor)
		assert.Empty(t, device.Drives)
		require.NotNil(t, device.BMC)
		assert.Empty(t, device.BMC.ID)
	})
}

// Bug: GetChassisTemperatures returns the readings of the first chassis with
// ThermalMetrics. On this tray that is BMC_0, so the result is the BMC board
// temperature alone, with no error.
func TestGB200ChassisTemperatures(t *testing.T) {
	client, _ := newGB200Client(t, WithSystemName("System_0"))

	temps, err := client.GetChassisTemperatures(context.Background())
	require.NoError(t, err)
	require.Len(t, temps, 1)

	assert.Equal(t, "BMC_0_DCSCM_Temp_0", temps[0].Name)
	require.NotNil(t, temps[0].ReadingCelsius)
	assert.InDelta(t, 23.5, *temps[0].ReadingCelsius, 0.001)
}

// BIOS reads and writes go to System_0, whose ID is not on any vendor list.
func TestGB200BiosConfiguration(t *testing.T) {
	ctx := context.Background()

	t.Run("get", func(t *testing.T) {
		client, _ := newGB200Client(t, WithSystemName("System_0"))

		config, err := client.GetBiosConfiguration(ctx)
		require.NoError(t, err)
		assert.Equal(t, map[string]string{
			"Above4GDecoding": "Enabled",
			"BusMasterEnable": "Disabled",
		}, config)
	})

	t.Run("set", func(t *testing.T) {
		client, mockup := newGB200Client(t, WithSystemName("System_0"))

		err := client.SetBiosConfiguration(ctx, map[string]string{"Above4GDecoding": "Disabled"})
		require.NoError(t, err)

		writes := mockup.writes()
		require.Len(t, writes, 1)
		assert.Equal(t, http.MethodPatch, writes[0].Method)
		assert.Equal(t, "/redfish/v1/Systems/System_0/Bios/Settings", writes[0].Path)

		var body struct {
			Attributes map[string]any `json:"Attributes"`
		}
		require.NoError(t, json.Unmarshal([]byte(writes[0].Body), &body))
		assert.Equal(t, map[string]any{"Above4GDecoding": "Disabled"}, body.Attributes)
	})

	t.Run("reset", func(t *testing.T) {
		client, mockup := newGB200Client(t, WithSystemName("System_0"))

		require.NoError(t, client.ResetBiosConfiguration(ctx))

		writes := mockup.writes()
		require.Len(t, writes, 1)
		assert.Equal(t, http.MethodPost, writes[0].Method)
		assert.Equal(t, "/redfish/v1/Systems/System_0/Bios/Actions/Bios.ResetBios", writes[0].Path)
	})

	// HGX_Baseboard_0 has no Bios resource.
	t.Run("system without a Bios resource", func(t *testing.T) {
		client, mockup := newGB200Client(t, WithSystemName("HGX_Baseboard_0"))

		_, err := client.GetBiosConfiguration(ctx)
		require.ErrorIs(t, err, bmclibErrs.ErrNoBiosAttributes)

		err = client.SetBiosConfiguration(ctx, map[string]string{"Above4GDecoding": "Disabled"})
		require.ErrorIs(t, err, bmclibErrs.ErrNoBiosAttributes)

		require.ErrorIs(t, client.ResetBiosConfiguration(ctx), bmclibErrs.ErrNoBiosAttributes)

		assert.Empty(t, mockup.writes())
	})
}

// The mockup has no SecureBoot resource, so this only shows that the request
// reaches the BMC instead of stopping at the system ID check.
func TestGB200SecureBoot(t *testing.T) {
	client, mockup := newGB200Client(t, WithSystemName("System_0"))

	_, err := client.GetSecureBoot(context.Background())
	require.Error(t, err)
	assert.NotErrorIs(t, err, bmclibErrs.ErrRedfishSystemOdataID)
	assert.True(t, mockup.requested("/redfish/v1/Systems/System_0/SecureBoot"))
}
