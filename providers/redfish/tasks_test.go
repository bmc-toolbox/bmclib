package redfish

import (
	"testing"
)

func Test_openbmcGetTask(t *testing.T) {
	var err error

	task, err := mockClient.openbmcGetTask("15")
	if err != nil {
		t.Fatal(err)
	}
	if task.TaskState != "TestState" {
		t.Fatal("Wrong test state:", task.TaskState)
	}

	// inexistent
	task, err = mockClient.openbmcGetTask("151515")
	if task != nil {
		t.Fatal("Task should be nil, but got:", task)
	}
	if err == nil {
		t.Fatal("err shouldn't be nil:", err)
	}

}
