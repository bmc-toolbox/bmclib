package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type bootDeviceTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (b *bootDeviceTester) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	if b.MakeErrorOut {
		return ok, errors.New("boot device set failed")
	}
	if b.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func TestSetBootDevice(t *testing.T) {
	testCases := []struct {
		name         string
		bootDevice   string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		{name: "success", bootDevice: "pxe", want: true},
		{name: "not ok return", bootDevice: "pxe", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to set boot device"), errors.New("failed to set boot device")}}},
		{name: "error", bootDevice: "pxe", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("boot device set failed"), errors.New("failed to set boot device")}}},
		{name: "error context timeout", bootDevice: "pxe", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to set boot device")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := bootDeviceTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := SetBootDevice(ctx, tc.bootDevice, false, false, []BootDeviceSetter{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestSetBootDeviceFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		bootDevice        string
		err               error
		badImplementation bool
		want              bool
	}{
		{name: "success", bootDevice: "pxe", want: true},
		{name: "no implementations found", bootDevice: "pxe", want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a BootDeviceSetter implementation: *struct {}"), errors.New("no BootDeviceSetter implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := bootDeviceTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			result, err := SetBootDeviceFromInterfaces(context.Background(), tc.bootDevice, false, false, generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}
