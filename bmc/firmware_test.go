package bmc

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/constants"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/pkg/errors"
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
		{"failure with metadata", common.SlugBIOS, string(constants.OnReset), false, nil, "1234", bmclibErrs.ErrNon200Response, 5 * time.Second, "foo", 1},
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
		{"failure with metadata", common.SlugBIOS, "1.1", "1234", constants.FirmwareInstallFailed, bmclibErrs.ErrNon200Response, 5 * time.Second, "foo", 1},
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

type firmwareInstallUploadAndInitiateTester struct {
	returnTaskID string
	returnError  error
}

func (f *firmwareInstallUploadAndInitiateTester) FirmwareInstallUploadAndInitiate(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	return f.returnTaskID, f.returnError
}

func (r *firmwareInstallUploadAndInitiateTester) Name() string {
	return "foo"
}

func TestFirmwareInstallUploadAndInitiate(t *testing.T) {
	testCases := []struct {
		testName           string
		component          string
		file               *os.File
		returnTaskID       string
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", "componentA", &os.File{}, "1234", nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", "componentB", &os.File{}, "1234", errors.New("failed to upload and initiate"), 5 * time.Second, "foo", 1},
		{"failure with context timeout", "componentC", &os.File{}, "", context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := &firmwareInstallUploadAndInitiateTester{returnTaskID: tc.returnTaskID, returnError: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			taskID, metadata, err := firmwareInstallUploadAndInitiate(ctx, tc.component, tc.file, []firmwareInstallProvider{{tc.providerName, testImplementation}})
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

func TestFirmwareInstallUploadAndInitiateFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName          string
		component         string
		file              *os.File
		returnTaskID      string
		returnError       error
		providerName      string
		badImplementation bool
	}{
		{"success with metadata", "componentA", &os.File{}, "1234", nil, "foo", false},
		{"failure with bad implementation", "componentB", &os.File{}, "1234", bmclibErrs.ErrProviderImplementation, "foo", true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := &firmwareInstallUploadAndInitiateTester{returnTaskID: tc.returnTaskID, returnError: tc.returnError}
				generic = []interface{}{testImplementation}
			}
			taskID, metadata, err := FirmwareInstallUploadAndInitiateFromInterfaces(context.Background(), tc.component, tc.file, generic)
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

type firmwareInstallUploadTester struct {
	TaskID string
	Err    error
}

func (f *firmwareInstallUploadTester) FirmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (taskID string, err error) {
	return f.TaskID, f.Err
}

func (r *firmwareInstallUploadTester) Name() string {
	return "foo"
}

func TestFirmwareInstallUploaded(t *testing.T) {
	testCases := []struct {
		testName           string
		component          string
		uploadTaskID       string
		returnTaskID       string
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", common.SlugBIOS, "1234", "5678", nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", common.SlugBIOS, "1234", "", bmclibErrs.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", common.SlugBIOS, "1234", "", context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			mockImplementation := &firmwareInstallUploadTester{TaskID: tc.returnTaskID, Err: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 4
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()

			taskID, metadata, err := firmwareInstallUploaded(ctx, tc.component, tc.uploadTaskID, []firmwareInstallerWithOptionsProvider{{tc.providerName, mockImplementation}})
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

func TestFirmwareInstallerUploadedFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName          string
		component         string
		uploadTaskID      string
		returnTaskID      string
		returnError       error
		providerName      string
		badImplementation bool
	}{
		{"success with metadata", common.SlugBIOS, "1234", "5678", nil, "foo", false},
		{"failure with bad implementation", common.SlugBIOS, "1234", "", bmclibErrs.ErrProviderImplementation, "foo", true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				mockImplementation := &firmwareInstallUploadTester{TaskID: tc.returnTaskID, Err: tc.returnError}
				generic = []interface{}{mockImplementation}
			}

			installTaskID, metadata, err := FirmwareInstallerUploadedFromInterfaces(context.Background(), tc.component, tc.uploadTaskID, generic)
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.returnTaskID, installTaskID)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
		})
	}
}

type firmwareUploadTester struct {
	returnTaskID string
	returnError  error
}

func (f *firmwareUploadTester) FirmwareUpload(ctx context.Context, component string, file *os.File) (uploadVerifyTaskID string, err error) {
	return f.returnTaskID, f.returnError
}

func (r *firmwareUploadTester) Name() string {
	return "foo"
}

func TestFirmwareUpload(t *testing.T) {
	testCases := []struct {
		testName           string
		component          string
		file               *os.File
		returnTaskID       string
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", common.SlugBIOS, nil, "1234", nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", common.SlugBIOS, nil, "1234", bmclibErrs.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", common.SlugBIOS, nil, "1234", context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := firmwareUploadTester{returnTaskID: tc.returnTaskID, returnError: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			taskID, metadata, err := firmwareUpload(ctx, tc.component, tc.file, []firmwareUploaderProvider{{tc.providerName, &testImplementation}})
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

type firmwareInstallStepsGetterTester struct {
	Steps []constants.FirmwareInstallStep
	Err   error
}

func (m *firmwareInstallStepsGetterTester) FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error) {
	return m.Steps, m.Err
}

func (m *firmwareInstallStepsGetterTester) Name() string {
	return "foo"
}

func TestFirmwareInstallStepsFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName          string
		component         string
		returnSteps       []constants.FirmwareInstallStep
		returnError       error
		providerName      string
		badImplementation bool
	}{
		{"success with metadata", common.SlugBIOS, []constants.FirmwareInstallStep{constants.FirmwareInstallStepUpload, constants.FirmwareInstallStepInstallStatus}, nil, "foo", false},
		{"failure with bad implementation", common.SlugBIOS, nil, bmclibErrs.ErrProviderImplementation, "foo", true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				mockImplementation := &firmwareInstallStepsGetterTester{Steps: tc.returnSteps, Err: tc.returnError}
				generic = []interface{}{mockImplementation}
			}

			steps, metadata, err := FirmwareInstallStepsFromInterfaces(context.Background(), tc.component, generic)
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.returnSteps, steps)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
		})
	}
}

type firmwareInstallStepsTester struct {
	returnSteps []constants.FirmwareInstallStep
	returnError error
}

func (f *firmwareInstallStepsTester) FirmwareInstallSteps(ctx context.Context, component string) (steps []constants.FirmwareInstallStep, err error) {
	return f.returnSteps, f.returnError
}

func (r *firmwareInstallStepsTester) Name() string {
	return "foo"
}

func TestFirmwareInstallSteps(t *testing.T) {
	testCases := []struct {
		testName           string
		component          string
		returnSteps        []constants.FirmwareInstallStep
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", common.SlugBIOS, []constants.FirmwareInstallStep{constants.FirmwareInstallStepUpload, constants.FirmwareInstallStepInstallStatus}, nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", common.SlugBIOS, nil, bmclibErrs.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", common.SlugBIOS, nil, context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := firmwareInstallStepsTester{returnSteps: tc.returnSteps, returnError: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			steps, metadata, err := firmwareInstallSteps(ctx, tc.component, []firmwareInstallStepsGetterProvider{{tc.providerName, &testImplementation}})
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.returnSteps, steps)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
			assert.Equal(t, tc.providersAttempted, len(metadata.ProvidersAttempted))
		})
	}
}

type firmwareTaskStatusTester struct {
	returnState  constants.TaskState
	returnStatus string
	returnError  error
}

func (f *firmwareTaskStatusTester) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state constants.TaskState, status string, err error) {
	return f.returnState, f.returnStatus, f.returnError
}

func (r *firmwareTaskStatusTester) Name() string {
	return "foo"
}

func TestFirmwareTaskStatus(t *testing.T) {
	testCases := []struct {
		testName           string
		kind               constants.FirmwareInstallStep
		component          string
		taskID             string
		installVersion     string
		returnState        constants.TaskState
		returnStatus       string
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", constants.FirmwareInstallStepUpload, common.SlugBIOS, "1234", "1.0", constants.FirmwareInstallComplete, "Upload completed", nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", constants.FirmwareInstallStepUpload, common.SlugBIOS, "1234", "1.0", constants.FirmwareInstallFailed, "Upload failed", bmclibErrs.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", constants.FirmwareInstallStepUpload, common.SlugBIOS, "1234", "1.0", "", "", context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := firmwareTaskStatusTester{returnState: tc.returnState, returnStatus: tc.returnStatus, returnError: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			state, status, metadata, err := firmwareTaskStatus(ctx, tc.kind, tc.component, tc.taskID, tc.installVersion, []firmwareTaskVerifierProvider{{tc.providerName, &testImplementation}})
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.returnState, state)
			assert.Equal(t, tc.returnStatus, status)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
			assert.Equal(t, tc.providersAttempted, len(metadata.ProvidersAttempted))
		})
	}
}

func TestFirmwareTaskStatusFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName           string
		kind               constants.FirmwareInstallStep
		component          string
		taskID             string
		installVersion     string
		returnState        constants.TaskState
		returnStatus       string
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", constants.FirmwareInstallStepUpload, common.SlugBIOS, "1234", "1.0", constants.Complete, "uploading", nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", constants.FirmwareInstallStepUpload, common.SlugBIOS, "1234", "1.0", constants.Failed, "failed", bmclibErrs.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", constants.FirmwareInstallStepUpload, common.SlugBIOS, "1234", "1.0", "", "", context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := firmwareTaskStatusTester{
				returnState:  tc.returnState,
				returnStatus: tc.returnStatus,
				returnError:  tc.returnError,
			}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			state, status, metadata, err := FirmwareTaskStatusFromInterfaces(ctx, tc.kind, tc.component, tc.taskID, tc.installVersion, []interface{}{&testImplementation})
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.returnState, state)
			assert.Equal(t, tc.returnStatus, status)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
			assert.Equal(t, tc.providersAttempted, len(metadata.ProvidersAttempted))
		})
	}
}
