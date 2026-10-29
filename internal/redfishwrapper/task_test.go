package redfishwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

func TestConvertTaskState(t *testing.T) {
	testCases := []struct {
		testName string
		state    string
		expected constants.TaskState
	}{
		{"starting state", "starting", constants.Initializing},
		{"downloading state", "downloading", constants.Initializing},
		{"downloaded state", "downloaded", constants.Initializing},
		{"scheduling state", "scheduling", constants.Initializing},
		{"running state", "running", constants.Running},
		{"stopping state", "stopping", constants.Running},
		{"cancelling state", "cancelling", constants.Running}, //nolint:misspell // Redfish API uses British spelling
		{"pending state", "pending", constants.Queued},
		{"new state", "new", constants.Queued},
		{"scheduled state", "scheduled", constants.PowerCycleHost},
		{"interrupted state", "interrupted", constants.Failed},
		{"killed state", "killed", constants.Failed},
		{"exception state", "exception", constants.Failed},
		{"cancelled state", "cancelled", constants.Failed}, //nolint:misspell // Redfish API uses British spelling
		{"suspended state", "suspended", constants.Failed},
		{"failed state", "failed", constants.Failed},
		{"completed state", "completed", constants.Complete},
		{"unknown state", "unknown_state", constants.Unknown},
	}

	client := Client{}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			result := client.ConvertTaskState(tc.state)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTaskStateActive(t *testing.T) {
	testCases := []struct {
		testName  string
		taskState constants.TaskState
		expected  bool
		err       error
	}{
		{"active initializing", constants.Initializing, true, nil},
		{"active running", constants.Running, true, nil},
		{"active queued", constants.Queued, true, nil},
		{"inactive complete", constants.Complete, false, nil},
		{"inactive failed", constants.Failed, false, nil},
		{"unknown state", "foobar", false, errUnexpectedTaskState},
	}

	client := &Client{}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			active, err := client.TaskStateActive(tc.taskState)

			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.expected, active)
		})
	}
}

func TestTaskStatus(t *testing.T) {
	type hmap map[string]func(http.ResponseWriter, *http.Request)
	withHandler := func(s string, f func(http.ResponseWriter, *http.Request)) hmap {
		return hmap{
			"/redfish/v1/":                  endpointFunc(t, "serviceroot.json"),
			"/redfish/v1/Systems":           endpointFunc(t, "systems.json"),
			"/redfish/v1/TaskService":       endpointFunc(t, "taskservice.json"),
			"/redfish/v1/TaskService/Tasks": endpointFunc(t, "tasks.json"),
			//			"/redfish/v1/TaskService/Tasks/1": endpointFunc(t, "tasks_1.json"),
			//			"/redfish/v1/TaskService/Tasks/2": endpointFunc(t, "tasks_2.json"),
			s: f,
		}
	}

	tests := map[string]struct {
		hmap           hmap
		expectedState  constants.TaskState
		expectedStatus string
		expectedErr    error
	}{
		"task in Initializing state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_starting.json"),
			),
			expectedState:  constants.Initializing,
			expectedStatus: "id: 1, state: Starting, status: OK",
			expectedErr:    nil,
		},
		"task in Running state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_running.json"),
			),
			expectedState:  constants.Running,
			expectedStatus: "id: 1, state: Running, status: OK",
			expectedErr:    nil,
		},
		"task in Queued state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_pending.json"),
			),
			expectedState:  constants.Queued,
			expectedStatus: "id: 1, state: Pending, status: OK",
			expectedErr:    nil,
		},
		"task in PowerCycleHost state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_scheduled.json"),
			),
			expectedState:  constants.PowerCycleHost,
			expectedStatus: "id: 1, state: Scheduled, status: OK",
			expectedErr:    nil,
		},
		"task in Failed state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_failed.json"),
			),
			expectedState:  constants.Failed,
			expectedStatus: "id: 1, state: Failed, status: OK",
			expectedErr:    nil,
		},
		"task in Complete state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_completed.json"),
			),
			expectedState:  constants.Complete,
			expectedStatus: "id: 1, state: Completed, status: OK",
			expectedErr:    nil,
		},
		"unknown task state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_unknown.json"),
			),
			expectedState:  constants.Unknown,
			expectedStatus: "id: 1, state: foobared, status: OK",
			expectedErr:    nil,
		},
		"failure case - no task found": {
			hmap: hmap{
				"/redfish/v1/":                  endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Systems":           endpointFunc(t, "systems.json"),
				"/redfish/v1/TaskService":       endpointFunc(t, "taskservice.json"),
				"/redfish/v1/TaskService/Tasks": endpointFunc(t, "tasks.json"),
			},
			expectedState:  "",
			expectedStatus: "",
			expectedErr:    bmclibErrs.ErrTaskNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()

			for endpoint, handler := range tc.hmap {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			state, status, err := client.TaskStatus(ctx, "1")
			if tc.expectedErr != nil {
				assert.ErrorContains(t, err, tc.expectedErr.Error())
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedState, state)
			assert.Equal(t, tc.expectedStatus, status)

			client.Close(context.Background())
		})
	}
}

func TestTask(t *testing.T) {
	type hmap map[string]func(http.ResponseWriter, *http.Request)
	handlers := func() hmap {
		return hmap{
			"/redfish/v1/":                    endpointFunc(t, "serviceroot.json"),
			"/redfish/v1/Systems":             endpointFunc(t, "systems.json"),
			"/redfish/v1/TaskService":         endpointFunc(t, "taskservice.json"),
			"/redfish/v1/TaskService/Tasks":   endpointFunc(t, "tasks.json"),
			"/redfish/v1/TaskService/Tasks/1": endpointFunc(t, "/tasks/tasks_1_completed.json"),
			"/redfish/v1/TaskService/Tasks/2": endpointFunc(t, "/tasks/tasks_2.json"),
		}
	}

	tests := map[string]struct {
		handlers         hmap
		taskID           string
		expectTaskStatus string
		expectTaskState  string
		err              error
	}{
		"happy case - task 1": {
			handlers:         handlers(),
			taskID:           "1",
			expectTaskStatus: "OK",
			expectTaskState:  "Completed",
			err:              nil,
		},
		"happy case - task 2": {
			handlers:         handlers(),
			taskID:           "2",
			expectTaskStatus: "OK",
			expectTaskState:  "Completed",
			err:              nil,
		},
		"failure case - no task found": {
			handlers: handlers(),
			taskID:   "3",
			err:      bmclibErrs.ErrTaskNotFound,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()

			for endpoint, handler := range tc.handlers {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()

			// os.Setenv("DEBUG_BMCLIB", "true")
			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			got, err := client.Task(ctx, tc.taskID)
			if tc.err != nil {
				assert.ErrorContains(t, err, tc.err.Error())
				return
			}

			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tc.expectTaskStatus, string(got.TaskStatus))
			assert.Equal(t, tc.expectTaskState, string(got.TaskState))

			client.Close(context.Background())
		})
	}
}

func TestGetTasks(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "systems.json"))
	mux.HandleFunc("/redfish/v1/TaskService", endpointFunc(t, "taskservice.json"))
	mux.HandleFunc("/redfish/v1/TaskService/Tasks", endpointFunc(t, "tasks.json"))
	mux.HandleFunc("/redfish/v1/TaskService/Tasks/1", endpointFunc(t, "tasks/tasks_1_running.json"))
	mux.HandleFunc("/redfish/v1/TaskService/Tasks/2", endpointFunc(t, "tasks/tasks_2.json"))

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)
	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(context.Background()))
	t.Cleanup(func() { _ = client.Close(context.Background()) })

	tasks, err := client.GetTasks(context.Background())
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	percent := 100
	assert.Equal(t, bmc.Task{
		ID:              "1",
		Name:            "BIOS Verify",
		TaskState:       "Running",
		TaskStatus:      "OK",
		State:           constants.Running,
		PercentComplete: &percent,
		StartTime:       "2023-11-06T12:04:16+00:00",
		EndTime:         "2023-11-06T12:05:31+00:00",
	}, tasks[0], "empty fixture messages must be dropped")
	assert.Equal(t, "2", tasks[1].ID)
	assert.Equal(t, constants.Complete, tasks[1].State)
}

func TestGetTasksNoTaskService(t *testing.T) {
	mux := http.NewServeMux()
	// A service root without TaskService.
	mux.HandleFunc("/redfish/v1/", func(w http.ResponseWriter, _ *http.Request) {
		var root map[string]any
		require.NoError(t, json.Unmarshal(mustReadFile(t, "serviceroot.json"), &root))
		delete(root, "Tasks")
		delete(root, "TaskService")
		_ = json.NewEncoder(w).Encode(root)
	})
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "systems.json"))

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)
	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(context.Background()))
	t.Cleanup(func() { _ = client.Close(context.Background()) })

	_, err = client.GetTasks(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrTaskServiceNotFound)
}
