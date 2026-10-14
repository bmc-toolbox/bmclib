package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

func networkBootBoolPtr(b bool) *bool { return &b }

type networkBootEnabledTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (r *networkBootEnabledTester) SetNetworkBootEnabled(ctx context.Context, httpEnabled, pxeEnabled *bool) (ok bool, err error) {
	if r.MakeErrorOut {
		return ok, errors.New("setting network boot enabled state failed")
	}
	if r.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (r *networkBootEnabledTester) Name() string {
	return "test provider"
}

func TestSetNetworkBootEnabled(t *testing.T) {
	httpEnabled := networkBootBoolPtr(true)
	testCases := map[string]struct {
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {want: true},
		"not ok return":         {want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider, failed to set network boot enabled state"), errors.New("failed to set network boot enabled state")}}},
		"error":                 {want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: setting network boot enabled state failed"), errors.New("failed to set network boot enabled state")}}},
		"error context timeout": {want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := networkBootEnabledTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, _, err := setNetworkBootEnabled(ctx, httpEnabled, nil, []networkBootEnabledProviders{{"test provider", &testImplementation}})
			if err != nil {
				diff := cmp.Diff(err.Error(), tc.err.Error())
				if diff != "" {
					t.Fatal(diff)
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

func TestSetNetworkBootEnabledFromInterfaces(t *testing.T) {
	httpEnabled := networkBootBoolPtr(true)
	testCases := map[string]struct {
		err               error
		badImplementation bool
		want              bool
		withName          bool
	}{
		"success":                  {want: true},
		"success with metadata":    {want: true, withName: true},
		"no implementations found": {want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a NetworkBootEnabledSetter implementation: *struct {}"), errors.New("no NetworkBootEnabledSetter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := networkBootEnabledTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			result, metadata, err := SetNetworkBootEnabledFromInterfaces(context.Background(), httpEnabled, nil, generic)
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
			if tc.withName {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
