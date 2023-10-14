package redfish

import (
	"context"
	"errors"
	"log"
	"testing"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-logr/logr"
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
	if !errors.Is(err, bmclibErrs.ErrTaskNotFound) {
		t.Fatal("err should be TaskNotFound:", err)
	}

}

func TestGetTask(t *testing.T) {

	c := New("10.247.157.31", "root", "Tb4lNjFe3ojzmA", logr.Discard())

	if err := c.Open(context.TODO()); err != nil {
		log.Fatal(err)
	}

	defer c.Close(context.TODO())

	//	_, err := c.redfishwrapper.Managers(context.TODO())
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	tasks, err := c.redfishwrapper.Tasks(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Println("hello")
	//fmt.Println("yello")

	spew.Dump(tasks)
}
