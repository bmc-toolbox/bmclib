package redfishwrapper

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/stretchr/testify/assert"
)

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
		expectedState  string
		expectedStatus string
		expectedErr    error
	}{
		"task in Initializing state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_starting.json"),
			),
			expectedState:  constants.FirmwareInstallInitializing,
			expectedStatus: "id: 1, state: Starting, status: OK",
			expectedErr:    nil,
		},
		"task in Running state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_running.json"),
			),
			expectedState:  constants.FirmwareInstallRunning,
			expectedStatus: "id: 1, state: Running, status: OK",
			expectedErr:    nil,
		},
		"task in Queued state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_pending.json"),
			),
			expectedState:  constants.FirmwareInstallQueued,
			expectedStatus: "id: 1, state: Pending, status: OK",
			expectedErr:    nil,
		},
		"task in PowerCycleHost state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_scheduled.json"),
			),
			expectedState:  constants.FirmwareInstallPowerCycleHost,
			expectedStatus: "id: 1, state: Scheduled, status: OK",
			expectedErr:    nil,
		},
		"task in Failed state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_failed.json"),
			),
			expectedState:  constants.FirmwareInstallFailed,
			expectedStatus: "id: 1, state: Failed, status: OK",
			expectedErr:    nil,
		},
		"task in Complete state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_completed.json"),
			),
			expectedState:  constants.FirmwareInstallComplete,
			expectedStatus: "id: 1, state: Completed, status: OK",
			expectedErr:    nil,
		},
		"unknown task state": {
			hmap: withHandler(
				"/redfish/v1/TaskService/Tasks/1",
				endpointFunc(t, "tasks/tasks_1_unknown.json"),
			),
			expectedState:  constants.FirmwareInstallUnknown,
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

			//os.Setenv("DEBUG_BMCLIB", "true")
			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

			err = client.Open(ctx)
			if err != nil {
				t.Fatal(err)
			}

			got, err := client.Task(ctx, tc.taskID)
			if tc.err != nil {
				fmt.Println(err)
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
