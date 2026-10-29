package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// Task is an entry of the BMC task collection: asynchronous work the BMC tracks, such as a
// firmware install, a virtual media mount or a certificate replacement.
type Task struct {
	// ID is the Redfish Task resource ID, e.g. "14" or "JID_123456789012".
	ID string
	// Name is the task name as reported by the BMC, e.g. "InsertMedia Action".
	Name string
	// TaskState is the Redfish task state, e.g. "Running", "Completed", "Exception".
	TaskState string
	// TaskStatus is the Redfish task health, e.g. "OK", "Warning", "Critical".
	TaskStatus string
	// State is TaskState mapped onto the bmclib task states (Initializing, Running, Queued,
	// PowerCycleHost, Complete, Failed, Unknown), the same mapping FirmwareTaskStatus uses.
	State constants.TaskState
	// PercentComplete is the task progress in percent; nil when the BMC does not report it.
	PercentComplete *int
	// StartTime and EndTime are the task timestamps as reported by the BMC; empty when not reported.
	StartTime string
	EndTime   string
	// Messages are the non-empty message texts attached to the task.
	Messages []string
}

// TasksGetter provides retrieval of the BMC task collection (TaskService/Tasks).
type TasksGetter interface {
	GetTasks(ctx context.Context) (tasks []Task, err error)
}

type tasksGetterProvider struct {
	name string
	TasksGetter
}

func getTasks(ctx context.Context, generic []tasksGetterProvider) (tasks []Task, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.TasksGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			tasks, vErr := elem.GetTasks(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return tasks, metadata, nil
		}
	}

	return nil, metadata, multierror.Append(err, errors.New("failure to get tasks"))
}

// GetTasksFromInterfaces identifies implementations of the TasksGetter interface and passes
// them to the getTasks() wrapper.
func GetTasksFromInterfaces(ctx context.Context, generic []interface{}) (tasks []Task, metadata Metadata, err error) {
	implementations := make([]tasksGetterProvider, 0)
	for _, elem := range generic {
		temp := tasksGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case TasksGetter:
			temp.TasksGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a TasksGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return nil, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no TasksGetter implementations found",
			),
		)
	}

	return getTasks(ctx, implementations)
}
