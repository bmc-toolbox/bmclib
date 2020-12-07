package ipmitool

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/bmc-toolbox/bmclib/internal/ipmi"
	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/google/go-cmp/cmp"
)

func TestBmcReset(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		resetType  string
		expectedOk bool
		wantErr    error
	}{
		{"state on", nil, "cold", true, nil},
		{"err", errors.New("connection timed out"), "warm", false, errors.New("connection timed out")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var ipm *ipmi.Ipmi
			monkey.PatchInstanceMethod(reflect.TypeOf(ipm), "PowerResetBmc", func(_ *ipmi.Ipmi, _ context.Context, _ string) (ok bool, err error) {
				return tc.expectedOk, tc.err
			})
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			user := "ADMIN"
			pass := "ADMIN"
			host := "127.1.1.1"
			port := "623"
			i, _ := ipmi.New(user, pass, host+":"+port)
			c := Conn{
				Host: host,
				Port: port,
				User: user,
				Pass: pass,
				Log:  logging.DefaultLogger(),
				con:  i,
			}
			ok, err := c.BmcReset(ctx, tc.resetType)
			if err != nil {
				diff := cmp.Diff(tc.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if ok != tc.expectedOk {
				t.Fatalf("expected: %v, got: %v", tc.expectedOk, ok)
			}

		})
	}
}
