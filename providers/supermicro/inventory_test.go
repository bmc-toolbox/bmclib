package supermicro

import (
	"context"
	"testing"

	"github.com/bombsimon/logrusr/v2"
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

func Test_Inventory(t *testing.T) {

	l := logrus.New()
	l.Level = logrus.TraceLevel + 1
	logger := logrusr.New(l)

	//client, err := New(context.TODO(), "10.247.135.27", "ADMIN", "FJEWWZWJQZ", logger)
	client, err := New(context.TODO(), "10.247.135.27", "ADMIN", "FJEWWZWJQZ", logger)
	if err != nil {
		t.Fatal(err)
	}

	// os.Setenv("BMCLIB_LOG_LEVEL", "trace")
	// defer os.Unsetenv("BMCLIB_LOG_LEVEL")
	err = client.Open(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	defer client.Close(context.TODO())

	device, err := client.Inventory(context.TODO())
	if err != nil {
		t.Fatal(err)
	}

	spew.Dump(device)
}
