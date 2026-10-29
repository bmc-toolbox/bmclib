package bmc

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/bmc-toolbox/bmclib/v2/constants"
)

type mockTasksGetter struct {
	tasks []Task
	err   error
}

func (m *mockTasksGetter) GetTasks(ctx context.Context) ([]Task, error) {
	return m.tasks, m.err
}

func (m *mockTasksGetter) Name() string {
	return "mock"
}

type mockJobsGetter struct {
	jobs []Job
	err  error
}

func (m *mockJobsGetter) GetJobs(ctx context.Context) ([]Job, error) {
	return m.jobs, m.err
}

func (m *mockJobsGetter) Name() string {
	return "mock"
}

func TestGetTasksFromInterfaces(t *testing.T) {
	percent := 100
	tasks := []Task{{ID: "2", Name: "BIOS Update", TaskState: "Completed", TaskStatus: "OK", State: constants.Complete, PercentComplete: &percent}}

	testCases := []struct {
		name          string
		generic       []interface{}
		errMsg        string
		expectedTasks []Task
	}{
		{
			name:          "success",
			generic:       []interface{}{&mockTasksGetter{tasks: tasks}},
			expectedTasks: tasks,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no TasksGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no TasksGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockTasksGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, metadata, err := GetTasksFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
				assert.Equal(t, "mock", metadata.SuccessfulProvider)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedTasks, got)
		})
	}
}

func TestGetJobsFromInterfaces(t *testing.T) {
	jobs := []Job{{ID: "JID_1", Name: "Configure: BIOS", JobState: "Pending"}}

	testCases := []struct {
		name         string
		generic      []interface{}
		errMsg       string
		expectedJobs []Job
	}{
		{
			name:         "success",
			generic:      []interface{}{&mockJobsGetter{jobs: jobs}},
			expectedJobs: jobs,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no JobsGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no JobsGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockJobsGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, metadata, err := GetJobsFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
				assert.Equal(t, "mock", metadata.SuccessfulProvider)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedJobs, got)
		})
	}
}
