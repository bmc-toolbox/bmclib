package bmc

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type powerTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (p *powerTester) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	if p.MakeErrorOut {
		return ok, errors.New("power set failed")
	}
	if p.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (p *powerTester) PowerStateGet(ctx context.Context) (state string, err error) {
	if p.MakeErrorOut {
		return state, errors.New("power state get failed")
	}
	return "on", nil
}

func TestSetPowerState(t *testing.T) {
	testCases := []struct {
		name         string
		state        string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
	}{
		{name: "success", state: "off", want: true},
		{name: "not ok return", state: "off", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to set power state"), errors.New("failed to set power state")}}},
		{name: "error", state: "off", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("power set failed"), errors.New("failed to set power state")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := powerTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			result, err := SetPowerState(context.Background(), tc.state, []PowerSetter{&testImplementation})
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

func TestSetPowerStateFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		state             string
		err               error
		badImplementation bool
		want              bool
	}{
		{name: "success", state: "off", want: true},
		{name: "no implementations found", state: "on", want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a PowerSetter implementation: *struct {}"), errors.New("no PowerSetter implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := powerTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			result, err := SetPowerStateFromInterfaces(context.Background(), tc.state, generic)
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

func TestGetPowerState(t *testing.T) {
	testCases := []struct {
		name     string
		state    string
		makeFail bool
		err      error
	}{
		{name: "success", state: "on", err: nil},
		{name: "failure", state: "on", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("power state get failed"), errors.New("failed to get power state")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := powerTester{MakeErrorOut: tc.makeFail}
			expectedResult := tc.state
			result, err := GetPowerState(context.Background(), []PowerStateGetter{&testImplementation})
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

func TestGetPowerStateFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		state             string
		err               error
		badImplementation bool
		want              string
	}{
		{name: "success", state: "on", want: "on"},
		{name: "no implementations found", state: "on", want: "", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a PowerStateGetter implementation: *struct {}"), errors.New("no PowerStateGetter implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := powerTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			result, err := GetPowerStateFromInterfaces(context.Background(), generic)
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
