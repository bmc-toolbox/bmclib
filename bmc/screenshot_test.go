package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"gopkg.in/go-playground/assert.v1"
)

type screenshotTester struct {
	MakeErrorOut bool
}

func (r *screenshotTester) Screenshot(ctx context.Context) (img []byte, fileType string, err error) {
	if r.MakeErrorOut {
		return nil, "", errors.New("crappy bmc is crappy")
	}

	return []byte(`foobar`), "png", nil
}

func (r *screenshotTester) Name() string {
	return "test screenshot provider"
}

func TestScreenshot(t *testing.T) {
	testCases := map[string]struct {
		makeErrorOut           bool
		wantImage              []byte
		wantFileType           string
		wantSuccessfulProvider string
		wantProvidersAttempted []string
		wantErr                error
		ctxTimeout             time.Duration
	}{
		"success":               {false, []byte(`foobar`), "png", "test provider", []string{"test provider"}, nil, 1 * time.Second},
		"error":                 {true, nil, "", "", []string{"test provider"}, &multierror.Error{Errors: []error{errors.New("provider: test provider: crappy bmc is crappy"), errors.New("failed to capture screenshot")}}, 1 * time.Second},
		"error context timeout": {true, nil, "", "", nil, &multierror.Error{Errors: []error{errors.New("context deadline exceeded")}}, 1 * time.Nanosecond},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := screenshotTester{MakeErrorOut: tc.makeErrorOut}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}

			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			image, fileType, metadata, err := screenshot(ctx, []screenshotGetterProvider{{"test provider", &testImplementation}})
			if err != nil {
				if tc.wantErr == nil {
					t.Fatal(err)
				}

				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				assert.Equal(t, tc.wantImage, image)
				assert.Equal(t, tc.wantFileType, fileType)
			}

			assert.Equal(t, tc.wantProvidersAttempted, metadata.ProvidersAttempted)
			assert.Equal(t, tc.wantSuccessfulProvider, metadata.SuccessfulProvider)
		})
	}
}

func TestScreenshotFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		wantImage              []byte
		wantFileType           string
		wantSuccessfulProvider string
		wantProvidersAttempted []string
		wantErr                error
		badImplementation      bool
	}{
		"success with metadata":    {[]byte(`foobar`), "png", "test screenshot provider", []string{"test screenshot provider"}, nil, false},
		"no implementations found": {nil, "", "", nil, &multierror.Error{Errors: []error{errors.New("not a ScreenshotGetter implementation: *struct {}"), errors.New("no ScreenshotGetter implementations found: error in provider implementation")}}, true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := screenshotTester{}
				generic = []interface{}{&testImplementation}
			}
			image, fileType, metadata, err := ScreenshotFromInterfaces(context.Background(), generic)
			if err != nil {
				if tc.wantErr == nil {
					t.Fatal(err)
				}

				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				assert.Equal(t, tc.wantImage, image)
				assert.Equal(t, tc.wantFileType, fileType)
			}

			assert.Equal(t, tc.wantProvidersAttempted, metadata.ProvidersAttempted)
			assert.Equal(t, tc.wantSuccessfulProvider, metadata.SuccessfulProvider)
		})
	}
}
