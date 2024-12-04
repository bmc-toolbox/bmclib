package supermicro

import (
	"fmt"
	"strings"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/redfish"
	"golang.org/x/exp/slices"
)

// noTasksRunning returns an error if a firmware related task was found active
func noTasksRunning(component string, t *redfish.Task) error {
	if t.TaskState == "Killed" {
		return nil
	}

	errTaskActive := errors.New("A firmware task was found active for component: " + component)

	const (
		// The redfish task name when the BMC is verifies the uploaded BMC firmware.
		verifyBMCFirmware = "BMC Verify"
		// The redfish task name when the BMC is installing the uploaded BMC firmware.
		updateBMCFirmware = "BMC Update"
		// The redfish task name when the BMC is verifies the uploaded BIOS firmware.
		verifyBIOSFirmware = "BIOS Verify"
		// The redfish task name when the BMC is installing the uploaded BIOS firmware.
		updateBIOSFirmware = "BIOS Update"
	)

	var verifyTaskName, updateTaskName string

	switch strings.ToUpper(component) {
	case common.SlugBMC:
		verifyTaskName = verifyBMCFirmware
		updateTaskName = updateBMCFirmware
	case common.SlugBIOS:
		verifyTaskName = verifyBIOSFirmware
		updateTaskName = updateBIOSFirmware
	}

	taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", t.ID, t.TaskState, t.TaskStatus)

	switch t.Name {
	case verifyTaskName:
		return errors.Wrap(errTaskActive, taskInfo)
	case updateTaskName:
		return errors.Wrap(errTaskActive, taskInfo)
	default:
		return nil
	}
}

func stateFinalized(s redfish.TaskState) bool {
	finalized := []redfish.TaskState{
		redfish.CompletedTaskState,
		redfish.CancelledTaskState,
		redfish.InterruptedTaskState,
		redfish.ExceptionTaskState,
	}

	return slices.Contains(finalized, s)
}
