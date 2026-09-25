package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type httpBootURITester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (r *httpBootURITester) SetHTTPBootURI(ctx context.Context, uri string) (ok bool, err error) {
	if r.MakeErrorOut {
		return ok, errors.New("setting http boot uri failed")
	}
	if r.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (r *httpBootURITester) Name() string {
	return "test provider"
}

func TestSetHTTPBootURI(t *testing.T) {
	testCases := map[string]struct {
		uri          string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {uri: "http://example.com/boot.efi", want: true},
		"not ok return":         {uri: "http://example.com/boot.efi", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider, failed to set http boot uri"), errors.New("failed to set http boot uri")}}},
		"error":                 {uri: "http://example.com/boot.efi", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: setting http boot uri failed"), errors.New("failed to set http boot uri")}}},
		"error context timeout": {uri: "http://example.com/boot.efi", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := httpBootURITester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, _, err := setHTTPBootURI(ctx, tc.uri, []httpBootURIProviders{{"test provider", &testImplementation}})
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

func TestSetHTTPBootURIFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		uri               string
		err               error
		badImplementation bool
		want              bool
		withName          bool
	}{
		"success":                  {uri: "http://example.com/boot.efi", want: true},
		"success with metadata":    {uri: "http://example.com/boot.efi", want: true, withName: true},
		"no implementations found": {uri: "http://example.com/boot.efi", want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a HTTPBootURISetter implementation: *struct {}"), errors.New("no HTTPBootURISetter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := httpBootURITester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			result, metadata, err := SetHTTPBootURIFromInterfaces(context.Background(), tc.uri, generic)
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

type httpBootTLSModeTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (r *httpBootTLSModeTester) SetHTTPBootTLSMode(ctx context.Context, mode HTTPBootTLSMode) (ok bool, err error) {
	if r.MakeErrorOut {
		return ok, errors.New("setting http boot tls mode failed")
	}
	if r.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (r *httpBootTLSModeTester) Name() string {
	return "test provider"
}

func TestSetHTTPBootTLSMode(t *testing.T) {
	testCases := map[string]struct {
		mode         HTTPBootTLSMode
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {mode: HTTPBootTLSModeNone, want: true},
		"not ok return":         {mode: HTTPBootTLSModeNone, want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider, failed to set http boot tls mode"), errors.New("failed to set http boot tls mode")}}},
		"error":                 {mode: HTTPBootTLSModeNone, want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: setting http boot tls mode failed"), errors.New("failed to set http boot tls mode")}}},
		"error context timeout": {mode: HTTPBootTLSModeNone, want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := httpBootTLSModeTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, _, err := setHTTPBootTLSMode(ctx, tc.mode, []httpBootTLSModeProviders{{"test provider", &testImplementation}})
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

func TestSetHTTPBootTLSModeFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		mode              HTTPBootTLSMode
		err               error
		badImplementation bool
		want              bool
		withName          bool
	}{
		"success":                  {mode: HTTPBootTLSModeNone, want: true},
		"success with metadata":    {mode: HTTPBootTLSModeNone, want: true, withName: true},
		"no implementations found": {mode: HTTPBootTLSModeNone, want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a HTTPBootTLSModeSetter implementation: *struct {}"), errors.New("no HTTPBootTLSModeSetter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := httpBootTLSModeTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			result, metadata, err := SetHTTPBootTLSModeFromInterfaces(context.Background(), tc.mode, generic)
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
