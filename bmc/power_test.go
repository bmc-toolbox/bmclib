package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

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

func (p *powerTester) Name() string {
	return "test provider"
}

func TestSetPowerState(t *testing.T) {
	testCases := map[string]struct {
		state        string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {state: "off", want: true},
		"not ok return":         {state: "off", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to set power state"), errors.New("failed to set power state")}}},
		"error":                 {state: "off", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("power set failed"), errors.New("failed to set power state")}}},
		"error context timeout": {state: "off", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to set power state")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := powerTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := SetPowerState(ctx, tc.state, []powerProviders{{"", nil, &testImplementation}})
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

func TestSetPowerStateFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		state             string
		err               error
		badImplementation bool
		want              bool
		withMetadata      bool
	}{
		"success":                  {state: "off", want: true},
		"success with metadata":    {state: "on", want: true, withMetadata: true},
		"no implementations found": {state: "on", want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a PowerSetter implementation: *struct {}"), errors.New("no PowerSetter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := powerTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			var result bool
			var err error
			var metadata Metadata
			if tc.withMetadata {
				result, err = SetPowerStateFromInterfaces(context.Background(), tc.state, generic, &metadata)
			} else {
				result, err = SetPowerStateFromInterfaces(context.Background(), tc.state, generic)
			}
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
			if tc.withMetadata {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestGetPowerState(t *testing.T) {
	testCases := map[string]struct {
		state      string
		makeFail   bool
		err        error
		ctxTimeout time.Duration
	}{
		"success":              {state: "on", err: nil},
		"failure":              {state: "on", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("power state get failed"), errors.New("failed to get power state")}}},
		"fail context timeout": {state: "on", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to get power state")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := powerTester{MakeErrorOut: tc.makeFail}
			expectedResult := tc.state
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := GetPowerState(ctx, []powerProviders{{"", &testImplementation, nil}})
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

func TestGetPowerStateFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		state             string
		err               error
		badImplementation bool
		want              string
		withMetadata      bool
	}{
		"success":                  {state: "on", want: "on"},
		"success with metadata":    {state: "on", want: "on", withMetadata: true},
		"no implementations found": {state: "on", want: "", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a PowerStateGetter implementation: *struct {}"), errors.New("no PowerStateGetter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := powerTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			var result string
			var err error
			var metadata Metadata
			if tc.withMetadata {
				result, err = GetPowerStateFromInterfaces(context.Background(), generic, &metadata)
			} else {
				result, err = GetPowerStateFromInterfaces(context.Background(), generic)
			}
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
			if tc.withMetadata {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
