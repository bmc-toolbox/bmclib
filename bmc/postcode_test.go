package bmc

import (
	"context"
	"testing"
	"time"

	"github.com/metal-toolbox/bmclib/constants"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/stretchr/testify/assert"
)

type postCodeGetterTester struct {
	returnStatus string
	returnCode   int
	returnError  error
}

func (p *postCodeGetterTester) PostCode(ctx context.Context) (status string, code int, err error) {
	return p.returnStatus, p.returnCode, p.returnError
}

func (p *postCodeGetterTester) Name() string {
	return "foo"
}

func TestPostCode(t *testing.T) {
	testCases := []struct {
		testName           string
		returnStatus       string
		returnCode         int
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", constants.POSTStateOS, 164, nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", constants.POSTCodeUnknown, 0, bmclibErrs.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", "", 0, context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := postCodeGetterTester{returnStatus: tc.returnStatus, returnCode: tc.returnCode, returnError: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			status, code, metadata, err := postCode(ctx, []postCodeGetterProvider{{tc.providerName, &testImplementation}})
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.returnStatus, status)
			assert.Equal(t, tc.returnCode, code)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
			assert.Equal(t, tc.providersAttempted, len(metadata.ProvidersAttempted))
		})
	}
}

func TestPostCodeFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName           string
		returnStatus       string
		returnCode         int
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
		badImplementation  bool
	}{
		{"success with metadata", constants.POSTStateOS, 164, nil, 5 * time.Second, "foo", 1, false},
		{"failure with bad implementation", "", 0, bmclibErrs.ErrProviderImplementation, 1 * time.Nanosecond, "foo", 1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := &postCodeGetterTester{returnStatus: tc.returnStatus, returnCode: tc.returnCode, returnError: tc.returnError}
				generic = []interface{}{testImplementation}
			}
			status, code, metadata, err := GetPostCodeInterfaces(context.Background(), generic)
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.returnStatus, status)
			assert.Equal(t, tc.returnCode, code)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
		})
	}
}
