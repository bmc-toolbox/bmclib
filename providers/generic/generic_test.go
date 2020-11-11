package generic

import (
	"context"
	"fmt"
	"testing"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bombsimon/logrusr"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
)

func TestStatus(t *testing.T) {
	t.Skip("need to spin up a simulator for this")
	logger := logrus.New()
	host := "127.0.0.1"
	user := "ADMIN"
	password := "ADMIN"
	expected := &devices.DataPoint{
		Name:  "state",
		Value: "on",
		Type:  devices.SystemState,
	}
	ctx := context.Background()
	g, err := New(logrusr.NewLogger(logger), host, user, password)
	if err != nil {
		t.Fatal(err)
	}
	data, err := g.DataRequest(ctx, devices.SystemState)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("power state: %+v\n", data)
	diff := cmp.Diff(expected, data)
	if diff != "" {
		t.Fatal(diff)
	}
}

func TestPower(t *testing.T) {
	t.Skip("need to spin up a simulator for this")
	logger := logrus.New()
	host := "127.0.0.1"
	user := "ADMIN"
	password := "ADMIN"
	ctx := context.Background()
	g, err := New(logrusr.NewLogger(logger), host, user, password)
	if err != nil {
		t.Fatal(err)
	}
	data, err := g.PowerRequest(ctx, devices.PowerOn)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("power state: %+v\n", data)
	if !data {
		t.Fatalf("expected: %v, got: %v", true, data)
	}
}

func TestUser(t *testing.T) {
	t.Skip("need to spin up a simulator for this")
	logger := logrus.New()
	host := "127.0.0.1"
	user := "ADMIN"
	password := "ADMIN"
	expected := "not implemented"
	ctx := context.Background()
	g, err := New(logrusr.NewLogger(logger), host, user, password)
	if err != nil {
		t.Fatal(err)
	}
	uData := devices.Config{
		Name: "create user",
		Value: devices.UserSetting{
			Name:     "jaocb",
			Password: "jacob",
			Role:     "Administrator",
			Enabled:  true,
		},
		Type: devices.User,
	}
	_, err = g.Configure(ctx, uData)
	if err == nil {
		t.Fatal("expecting error")
	}
	if err.Error() != expected {
		t.Fatalf("expected: %v, got: %v", expected, err.Error())
	}
	/*
		fmt.Printf("config state: %+v\n", data)
		diff := cmp.Diff(expected, data)
		if diff != "" {
			t.Fatal(diff)
		}
	*/
}

func TestConnection(t *testing.T) {
	t.Skip("need to spin up a simulator for this")
	logger := logrus.New()
	host := "127.0.0.1"
	user := "ADMIN"
	password := "ADMIN"
	ctx := context.Background()
	g, err := New(logrusr.NewLogger(logger), host, user, password)
	if err != nil {
		t.Fatal(err)
	}

	err = g.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBootDevice(t *testing.T) {
	t.Skip("need to spin up a simulator for this")
	logger := logrus.New()
	host := "127.0.0.1"
	user := "ADMIN"
	password := "ADMIN"
	ctx := context.Background()
	g, err := New(logrusr.NewLogger(logger), host, user, password)
	if err != nil {
		t.Fatal(err)
	}
	ok, err := g.BootDeviceRequest(ctx, devices.BootOptions{
		Device:     devices.DiskBoot,
		Persistent: true,
		EfiBoot:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("expected: %v, got: %v", true, ok)
	}
}
