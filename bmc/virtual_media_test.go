package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type virtualMediaTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (r *virtualMediaTester) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (ok bool, err error) {
	if r.MakeErrorOut {
		return ok, errors.New("setting virtual media failed")
	}
	if r.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (r *virtualMediaTester) Name() string {
	return "test provider"
}

func TestSetVirtualMedia(t *testing.T) {
	testCases := map[string]struct {
		kind         string
		mediaURL     string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {kind: "cdrom", mediaURL: "example.com/some.iso", want: true},
		"not ok return":         {kind: "cdrom", mediaURL: "example.com/some.iso", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider, failed to set virtual media"), errors.New("failed to set virtual media")}}},
		"error":                 {kind: "cdrom", mediaURL: "example.com/some.iso", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: setting virtual media failed"), errors.New("failed to set virtual media")}}},
		"error context timeout": {kind: "cdrom", mediaURL: "example.com/some.iso", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to set virtual media")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := virtualMediaTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, _, err := setVirtualMedia(ctx, tc.kind, tc.mediaURL, []virtualMediaProviders{{"test provider", &testImplementation}})
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

func TestSetVirtualMediaFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		kind              string
		mediaURL          string
		err               error
		badImplementation bool
		want              bool
		withName          bool
	}{
		"success":                  {kind: "cdrom", mediaURL: "example.com/some.iso", want: true},
		"success with metadata":    {kind: "cdrom", mediaURL: "example.com/some.iso", want: true, withName: true},
		"no implementations found": {kind: "cdrom", mediaURL: "example.com/some.iso", want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a VirtualMediaSetter implementation: *struct {}"), errors.New("no VirtualMediaSetter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := virtualMediaTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			result, metadata, err := SetVirtualMediaFromInterfaces(context.Background(), tc.kind, tc.mediaURL, generic)
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
