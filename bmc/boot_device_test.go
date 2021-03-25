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

func (b *bootDeviceTester) Name() string {
	return "test provider"
}

func TestSetBootDevice(t *testing.T) {
	testCases := map[string]struct {
		bootDevice   string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {bootDevice: "pxe", want: true},
		"not ok return":         {bootDevice: "pxe", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to set boot device"), errors.New("failed to set boot device")}}},
		"error":                 {bootDevice: "pxe", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("boot device set failed"), errors.New("failed to set boot device")}}},
		"error context timeout": {bootDevice: "pxe", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to set boot device")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := bootDeviceTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := SetBootDevice(ctx, tc.bootDevice, false, false, []bootDeviceProviders{{"", &testImplementation}})
			if err != nil {
				if tc.err != nil {
					diff := cmp.Diff(err.Error(), tc.err.Error())
					if diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}

			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestSetBootDeviceFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		bootDevice        string
		err               error
		badImplementation bool
		want              bool
		withName          bool
	}{
		"success":                  {bootDevice: "pxe", want: true},
		"success with metadata":    {bootDevice: "pxe", want: true, withName: true},
		"no implementations found": {bootDevice: "pxe", want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a BootDeviceSetter implementation: *struct {}"), errors.New("no BootDeviceSetter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := bootDeviceTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			var result bool
			var err error
			var metadata Metadata
			if tc.withName {
				result, err = SetBootDeviceFromInterfaces(context.Background(), tc.bootDevice, false, false, generic, &metadata)
			} else {
				result, err = SetBootDeviceFromInterfaces(context.Background(), tc.bootDevice, false, false, generic)
			}
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if tc.withName {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
