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

func TestBootDeviceSet(t *testing.T) {
	testCases := []struct {
		name       string
		err        error
		expectedOk bool
		wantErr    error
	}{
		{"success", nil, true, nil},
		{"err", errors.New("connection timed out"), false, errors.New("connection timed out")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var ipm *ipmi.Ipmi
			monkey.PatchInstanceMethod(reflect.TypeOf(ipm), "BootDeviceSet", func(_ *ipmi.Ipmi, _ context.Context, _ string, _, _ bool) (ok bool, err error) {
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
			ok, err := c.BootDeviceSet(ctx, "pxe", true, true)
			if err != nil {
				diff := cmp.Diff(tc.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
			diff := cmp.Diff(ok, tc.expectedOk)
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
