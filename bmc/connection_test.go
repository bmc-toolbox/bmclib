package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type connTester struct {
	MakeErrorOut bool
}

func (r *connTester) Open(ctx context.Context) (err error) {
	if r.MakeErrorOut {
		return errors.New("open connection failed")
	}

	return nil
}

func (r *connTester) Close(ctx context.Context) (err error) {
	if r.MakeErrorOut {
		return errors.New("close connection failed")
	}

	return nil
}

func TestOpenConnection(t *testing.T) {
	testCases := []struct {
		name         string
		makeErrorOut bool
		err          error
		ctxTimeout   time.Duration
	}{
		{name: "success"},
		{name: "error context deadline", err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to open connection")}}, ctxTimeout: time.Nanosecond * 1},
		{name: "error", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("open connection failed"), errors.New("failed to open connection")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := connTester{MakeErrorOut: tc.makeErrorOut}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			err := OpenConnection(ctx, []Opener{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestOpenConnectionFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		badImplementation bool
	}{
		{name: "success"},
		{name: "no implementations found", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a Opener implementation: *struct {}"), errors.New("no Opener implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := connTester{}
				generic = []interface{}{&testImplementation}
			}
			err := OpenConnectionFromInterfaces(context.Background(), generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestCloseConnection(t *testing.T) {
	testCases := []struct {
		name         string
		makeErrorOut bool
		err          error
		ctxTimeout   time.Duration
	}{
		{name: "success"},
		{name: "error context deadline", err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to close connection")}}, ctxTimeout: time.Nanosecond * 1},
		{name: "error", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("close connection failed"), errors.New("failed to close connection")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := connTester{MakeErrorOut: tc.makeErrorOut}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			err := CloseConnection(ctx, []Closer{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestCloseConnectionFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		badImplementation bool
	}{
		{name: "success"},
		{name: "no implementations found", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a Closer implementation: *struct {}"), errors.New("no Closer implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := connTester{}
				generic = []interface{}{&testImplementation}
			}
			err := CloseConnectionFromInterfaces(context.Background(), generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
