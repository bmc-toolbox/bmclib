package bmclib

import (
	"context"
	"os"
	"testing"
	"time"

	//_ "github.com/bmc-toolbox/bmclib/providers/ipmitool"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

var (
	log logr.Logger
	/*
		equateErrorMessage = cmp.Comparer(func(x, y error) bool {
			if x == nil || y == nil {
				return x == nil && y == nil
			}
			return x.Error() == y.Error()
		})
	*/
)

func TestMain(m *testing.M) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	z, err := config.Build()
	if err != nil {
		os.Exit(1)
	}
	log = zapr.NewLogger(z)
	os.Exit(m.Run())
}

func TestBMC(t *testing.T) {
	t.Skip("needs ipmitool and real ipmi server")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	cl := NewClient(host, user, pass, WithLogger(log))
	err := cl.AddVendorSpecificToRegistry(ctx)
	if err != nil {
		t.Fatal(err)
	}

	state, err := cl.GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)

	users, err := cl.ReadUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users)

	t.Fatal()
}
