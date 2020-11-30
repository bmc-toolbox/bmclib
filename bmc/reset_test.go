package bmc

import (
	"context"
	"errors"
	"testing"

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

func TestResetBMC(t *testing.T) {
	testCases := []struct {
		name         string
		resetType    string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
	}{
		{name: "success", resetType: "cold", want: true},
		{name: "not ok return", resetType: "warm", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to reset BMC"), errors.New("failed to reset BMC")}}},
		{name: "error", resetType: "cold", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("bmc reset failed"), errors.New("failed to reset BMC")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := resetTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			result, err := ResetBMC(context.Background(), tc.resetType, []BMCResetter{&testImplementation})
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

func TestResetBMCFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		resetType         string
		err               error
		badImplementation bool
		want              bool
	}{
		{name: "success", resetType: "cold", want: true},
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
			result, err := ResetBMCFromInterfaces(context.Background(), tc.resetType, generic)
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
