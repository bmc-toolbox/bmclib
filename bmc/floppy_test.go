package bmc

import (
	"context"
	"io"
	"testing"
	"time"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/stretchr/testify/assert"
)

type mountFloppyImageTester struct {
	returnError error
}

func (p *mountFloppyImageTester) MountFloppyImage(ctx context.Context, reader io.Reader) (err error) {
	return p.returnError
}

func (p *mountFloppyImageTester) Name() string {
	return "foo"
}

func TestMountFloppyFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName           string
		image              io.Reader
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
		badImplementation  bool
	}{
		{"success with metadata", nil, nil, 5 * time.Second, "foo", 1, false},
		{"failure with bad implementation", nil, bmclibErrs.ErrProviderImplementation, 1 * time.Nanosecond, "foo", 1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := &mountFloppyImageTester{returnError: tc.returnError}
				generic = []interface{}{testImplementation}
			}
			metadata, err := MountFloppyImageFromInterfaces(context.Background(), tc.image, generic)
			if tc.returnError != nil {
				assert.ErrorContains(t, err, tc.returnError.Error())
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.returnError, err)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
		})
	}
}

type unmountFloppyImageTester struct {
	returnError error
}

func (p *unmountFloppyImageTester) UnmountFloppyImage(ctx context.Context) (err error) {
	return p.returnError
}

func (p *unmountFloppyImageTester) Name() string {
	return "foo"
}

func TestUnmountFloppyFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName           string
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
		badImplementation  bool
	}{
		{"success with metadata", nil, 5 * time.Second, "foo", 1, false},
		{"failure with bad implementation", bmclibErrs.ErrProviderImplementation, 1 * time.Nanosecond, "foo", 1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := &unmountFloppyImageTester{returnError: tc.returnError}
				generic = []interface{}{testImplementation}
			}
			metadata, err := UnmountFloppyImageFromInterfaces(context.Background(), generic)
			if tc.returnError != nil {
				assert.ErrorContains(t, err, tc.returnError.Error())
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.returnError, err)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
		})
	}
}
