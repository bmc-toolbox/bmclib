package redfish

import (
	"context"
	"testing"
)

func Test_activeTask(t *testing.T) {
	_, err := mockClient.activeTask(context.TODO())
	// Current mocking should fail
	if err == nil {
		t.Fatal(err)
	}
}

func Test_GetTask(t *testing.T) {
	var err error

	task, err := mockClient.GetTask("15")
	if err != nil {
		t.Fatal(err)
	}
	if task.TaskState != "TestState" {
		t.Fatal("Wrong test state:", task.TaskState)
	}

	// inexistent
	task, err = mockClient.GetTask("151515")
	if task != nil {
		t.Fatal("Task should be nil, but got:", task)
	}
	if err == nil {
		t.Fatal("err shouldn't be nil:", err)
	}

}
