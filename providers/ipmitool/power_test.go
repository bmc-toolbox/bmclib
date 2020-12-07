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

func TestPowerStateGet(t *testing.T) {
	testCases := []struct {
		name          string
		err           error
		expectedState string
		wantErr       error
	}{
		{"state on", nil, "on", nil},
		{"err", errors.New("connection timed out"), "on", errors.New("connection timed out")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var ipm *ipmi.Ipmi
			monkey.PatchInstanceMethod(reflect.TypeOf(ipm), "PowerState", func(_ *ipmi.Ipmi, _ context.Context) (state string, err error) {
				return "on", tc.err
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
			state, err := c.PowerStateGet(ctx)
			if err != nil {
				diff := cmp.Diff(tc.wantErr.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if state != tc.expectedState {
				t.Fatalf("expected: %v, got: %v", tc.expectedState, state)
			}

		})
	}
}

func TestPowerSet(t *testing.T) {
	lookup := map[string]string{
		"on":      "PowerOn",
		"off":     "PowerOff",
		"reset":   "PowerReset",
		"soft":    "PowerSoft",
		"cycle":   "PowerCycle",
		"unknown": "PowerCycle",
	}
	testCases := []struct {
		name       string
		err        error
		state      string
		expectedOk bool
		isOnOk     bool
		wantErr    error
	}{
		{"set power on - already on", nil, "on", true, false, nil},
		{"set power on", nil, "on", true, true, nil},
		{"set power off - already off", nil, "off", true, false, nil},
		{"set power off", nil, "off", true, true, nil},
		{"set power reset", nil, "reset", true, false, nil},
		{"set power soft", nil, "soft", true, false, nil},
		{"set power cycle", nil, "cycle", true, false, nil},
		{"set power unknown state type", nil, "unknown", false, false, errors.New("requested state type unknown")},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var ipm *ipmi.Ipmi
			if tc.state == "on" || tc.state == "off" {
				monkey.PatchInstanceMethod(reflect.TypeOf(ipm), "IsOn", func(_ *ipmi.Ipmi, _ context.Context) (ok bool, err error) {
					return tc.isOnOk, tc.err
				})
			}
			monkey.PatchInstanceMethod(reflect.TypeOf(ipm), lookup[tc.state], func(_ *ipmi.Ipmi, _ context.Context) (ok bool, err error) {
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

			ok, err := c.PowerSet(ctx, tc.state)
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
