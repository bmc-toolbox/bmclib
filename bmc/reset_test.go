package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type resetTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (r *resetTester) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	if r.MakeErrorOut {
		return ok, errors.New("bmc reset failed")
	}
	if r.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (r *resetTester) Name() string {
	return "test provider"
}

func TestResetBMC(t *testing.T) {
	testCases := []struct {
		name         string
		resetType    string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		{name: "success", resetType: "cold", want: true},
		{name: "not ok return", resetType: "warm", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to reset BMC"), errors.New("failed to reset BMC")}}},
		{name: "error", resetType: "cold", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("bmc reset failed"), errors.New("failed to reset BMC")}}},
		{name: "error context timeout", resetType: "cold", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to reset BMC")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := resetTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := ResetBMC(ctx, tc.resetType, []bmcProviders{{"", &testImplementation}})
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

		})
	}
}

func TestResetBMCFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		resetType         string
		err               error
		badImplementation bool
		want              bool
		withName          bool
	}{
		{name: "success", resetType: "cold", want: true},
		{name: "success", resetType: "cold", want: true, withName: true},
		{name: "no implementations found", resetType: "warm", want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a BMCResetter implementation: *struct {}"), errors.New("no BMCResetter implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := resetTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			var result bool
			var err error
			var successfulProvider Metadata
			if tc.withName {
				result, err = ResetBMCFromInterfaces(context.Background(), tc.resetType, generic, &successfulProvider)
			} else {
				result, err = ResetBMCFromInterfaces(context.Background(), tc.resetType, generic)
			}
			if err != nil {
				if tc.err != nil {
					diff := cmp.Diff(tc.err.Error(), err.Error())
					if diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if tc.withName {
				if diff := cmp.Diff(successfulProvider.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
