package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/goleak"
)

type connTester1 struct {
	MakeErrorOutConn                 bool
	MakeErrorOutIncompatibleProvider bool
}

func (r *connTester1) Open(ctx context.Context) (err error) {
	if r.MakeErrorOutConn {
		return errors.New("open connection failed")
	}

	if r.MakeErrorOutIncompatibleProvider {
		return bmclibErrs.ErrIncompatibleProvider
	}

	return nil
}

func (r *connTester1) Close(ctx context.Context) (err error) {
	if r.MakeErrorOutConn {
		return errors.New("close connection failed")
	}
	return nil
}

func (p *connTester1) Name() string {
	return "test provider"
}

type connTester2 struct{}

func (r *connTester2) Open(ctx context.Context) (err error) {
	<-ctx.Done()
	return nil
}

func (r *connTester2) Close(ctx context.Context) (err error) {
	return nil
}

func (p *connTester2) Name() string {
	return "test provider 2"
}

func TestOpenConnectionFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		err                              error
		makeErrorOutConn                 bool
		makeErrorOutIncompatibleProvider bool
		badImplementation                bool
		withMetadata                     bool
		withMultipleProviders            bool
		ctxTimeout                       time.Duration
	}{
		"success":                                   {},
		"success with metadata":                     {withMetadata: true},
		"error context deadline":                    {err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("no Opener implementations found")}}, ctxTimeout: time.Nanosecond * 1},
		"error failed open conn":                    {makeErrorOutConn: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: open connection failed"), errors.New("no Opener implementations found")}}},
		"error failed open - incompatible provider": {makeErrorOutIncompatibleProvider: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: provider not compatible with device"), errors.New("no Opener implementations found")}}},
		"no implementations found":                  {badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a Opener implementation: *struct {}"), errors.New("no Opener implementations found")}}},
		"multiple providers attempted":              {withMultipleProviders: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			defer goleak.VerifyNone(t)

			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = append(generic, &badImplementation)
			} else {
				testImplementation := &connTester1{
					MakeErrorOutConn:                 tc.makeErrorOutConn,
					MakeErrorOutIncompatibleProvider: tc.makeErrorOutIncompatibleProvider,
				}
				generic = append(generic, testImplementation)
			}

			if tc.withMultipleProviders {
				generic = append(generic, &connTester2{})
			}

			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx := context.Background()
			opened, metadata, err := OpenConnectionFromInterfaces(ctx, tc.ctxTimeout, generic, tc.makeErrorOutIncompatibleProvider)
			if err != nil {
				if tc.err != nil {
					if diff := cmp.Diff(err.Error(), tc.err.Error()); diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			} else {
				expected := []interface{}{&connTester1{MakeErrorOutConn: tc.makeErrorOutConn, MakeErrorOutIncompatibleProvider: tc.makeErrorOutIncompatibleProvider}}
				if tc.withMultipleProviders {
					expected = append(expected, &connTester2{})
				}

				if diff := cmp.Diff(opened, expected); diff != "" {
					t.Fatal(diff)
				}
			}
			if tc.withMetadata {
				if diff := cmp.Diff(metadata.SuccessfulOpenConns, []string{"test provider"}); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestCloseConnection(t *testing.T) {
	testCases := map[string]struct {
		makeErrorOutConn bool
		err              error
		ctxTimeout       time.Duration
	}{
		"success":                {},
		"error context deadline": {err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded")}}, ctxTimeout: time.Nanosecond * 1},
		"error":                  {makeErrorOutConn: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: close connection failed"), errors.New("failed to close connection")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := connTester1{MakeErrorOutConn: tc.makeErrorOutConn}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			_, err := closeConnection(ctx, []connectionProviders{{"test provider", &testImplementation}})
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
	testCases := map[string]struct {
		err               error
		badImplementation bool
		withMetadata      bool
	}{
		"success":                  {},
		"success with metadata":    {withMetadata: true},
		"no implementations found": {badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a Closer implementation: *struct {}"), errors.New("no Closer implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := connTester1{}
				generic = []interface{}{&testImplementation}
			}

			metadata, err := CloseConnectionFromInterfaces(context.Background(), generic)
			if err != nil {
				diff := cmp.Diff(err.Error(), tc.err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if tc.withMetadata {
				if diff := cmp.Diff(metadata.SuccessfulCloseConns, []string{"test provider"}); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
