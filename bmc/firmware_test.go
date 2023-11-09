package bmc

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	"github.com/bmc-toolbox/bmclib/v2/errors"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/common"
	"github.com/stretchr/testify/assert"
)

type firmwareInstallTester struct {
	returnTaskID string
	returnError  error
}

func (f *firmwareInstallTester) FirmwareInstall(ctx context.Context, component, applyAt string, forceInstall bool, reader io.Reader) (taskID string, err error) {
	return f.returnTaskID, f.returnError
}

func (r *firmwareInstallTester) Name() string {
	return "foo"
}

func TestFirmwareInstall(t *testing.T) {
	testCases := []struct {
		testName           string
		component          string
		applyAt            string
		forceInstall       bool
		reader             io.Reader
		returnTaskID       string
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", common.SlugBIOS, string(constants.OnReset), false, nil, "1234", nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", common.SlugBIOS, string(constants.OnReset), false, nil, "1234", errors.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", common.SlugBIOS, string(constants.OnReset), false, nil, "1234", context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := firmwareInstallTester{returnTaskID: tc.returnTaskID, returnError: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			taskID, metadata, err := firmwareInstall(ctx, tc.component, tc.applyAt, tc.forceInstall, tc.reader, []firmwareInstallerProvider{{tc.providerName, &testImplementation}})
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.returnTaskID, taskID)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
			assert.Equal(t, tc.providersAttempted, len(metadata.ProvidersAttempted))
		})
	}
}
func TestFirmwareInstallFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName          string
		component         string
		applyAt           string
		forceInstall      bool
		reader            io.Reader
		returnTaskID      string
		returnError       error
		providerName      string
		badImplementation bool
	}{
		{"success with metadata", common.SlugBIOS, string(constants.OnReset), false, nil, "1234", nil, "foo", false},
		{"failure with metadata", common.SlugBIOS, string(constants.OnReset), false, nil, "1234", bmclibErrs.ErrProviderImplementation, "foo", true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := &firmwareInstallTester{returnTaskID: tc.returnTaskID, returnError: tc.returnError}
				generic = []interface{}{testImplementation}
			}
			taskID, metadata, err := FirmwareInstallFromInterfaces(context.Background(), tc.component, tc.applyAt, tc.forceInstall, tc.reader, generic)
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.returnTaskID, taskID)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
		})
	}
}

type firmwareInstallStatusTester struct {
	returnStatus string
	returnError  error
}

func (f *firmwareInstallStatusTester) FirmwareInstallStatus(ctx context.Context, installVersion, component, taskID string) (status string, err error) {
	return f.returnStatus, f.returnError
}

func (r *firmwareInstallStatusTester) Name() string {
	return "foo"
}

func TestFirmwareInstallStatus(t *testing.T) {
	testCases := []struct {
		testName           string
		component          string
		installVersion     string
		taskID             string
		returnStatus       string
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", common.SlugBIOS, "1.1", "1234", constants.FirmwareInstallComplete, nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", common.SlugBIOS, "1.1", "1234", constants.FirmwareInstallFailed, errors.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", common.SlugBIOS, "1.1", "1234", "", context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := firmwareInstallStatusTester{returnStatus: tc.returnStatus, returnError: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 4
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			taskID, metadata, err := firmwareInstallStatus(ctx, tc.installVersion, tc.component, tc.taskID, []firmwareInstallVerifierProvider{{tc.providerName, &testImplementation}})
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.returnStatus, taskID)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
			assert.Equal(t, tc.providersAttempted, len(metadata.ProvidersAttempted))
		})
	}
}

func TestFirmwareInstallStatusFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName          string
		component         string
		installVersion    string
		taskID            string
		returnStatus      string
		returnError       error
		providerName      string
		badImplementation bool
	}{
		{"success with metadata", common.SlugBIOS, "1.1", "1234", "status-done", nil, "foo", false},
		{"failure with bad implementation", common.SlugBIOS, "1.1", "1234", "status-done", bmclibErrs.ErrProviderImplementation, "foo", true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := &firmwareInstallStatusTester{returnStatus: tc.returnStatus, returnError: tc.returnError}
				generic = []interface{}{testImplementation}
			}
			status, metadata, err := FirmwareInstallStatusFromInterfaces(context.Background(), tc.component, tc.installVersion, tc.taskID, generic)
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.returnStatus, status)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
		})
	}
}
