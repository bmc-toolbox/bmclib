package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type solTermTester struct {
	MakeErrorOut bool
}

func (r *solTermTester) DeactivateSOL(ctx context.Context) (err error) {
	if r.MakeErrorOut {
		return errors.New("SOL deactivation failed")
	}
	return nil
}

func (r *solTermTester) Name() string {
	return "test provider"
}

func TestDeactivateSOL(t *testing.T) {
	testCases := map[string]struct {
		makeErrorOut bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {makeErrorOut: false},
		"error":                 {makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: SOL deactivation failed"), errors.New("failed to deactivate SOL session")}}},
		"error context timeout": {makeErrorOut: false, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := solTermTester{MakeErrorOut: tc.makeErrorOut}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			_, err := deactivateSOL(ctx, 0, []deactivatorProvider{{"test provider", &testImplementation}})
			var diff string
			if err != nil && tc.err != nil {
				diff = cmp.Diff(err.Error(), tc.err.Error())
			} else {
				diff = cmp.Diff(err, tc.err)
			}
			if diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestDeactivateSOLFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		err               error
		badImplementation bool
		withName          bool
	}{
		"success":                  {},
		"success with metadata":    {withName: true},
		"no implementations found": {badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not an SOLDeactivator implementation: *struct {}"), errors.New("no SOLDeactivator implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := solTermTester{}
				generic = []interface{}{&testImplementation}
			}
			metadata, err := DeactivateSOLFromInterfaces(context.Background(), 0, generic)
			var diff string
			if err != nil && tc.err != nil {
				diff = cmp.Diff(err.Error(), tc.err.Error())
			} else {
				diff = cmp.Diff(err, tc.err)
			}
			if diff != "" {
				t.Fatal(diff)
			}
			if tc.withName {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
