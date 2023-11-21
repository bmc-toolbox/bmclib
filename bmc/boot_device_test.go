package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
)

type bootDeviceTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (b *bootDeviceTester) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	if b.MakeErrorOut {
		return ok, errors.New("boot device set failed")
	}
	if b.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (b *bootDeviceTester) Name() string {
	return "test provider"
}

func TestSetBootDevice(t *testing.T) {
	testCases := map[string]struct {
		bootDevice   string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {bootDevice: "pxe", want: true},
		"not ok return":         {bootDevice: "pxe", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider, failed to set boot device"), errors.New("failed to set boot device")}}},
		"error":                 {bootDevice: "pxe", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("provider: test provider: boot device set failed"), errors.New("failed to set boot device")}}},
		"error context timeout": {bootDevice: "pxe", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := bootDeviceTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, _, err := setBootDevice(ctx, 0, tc.bootDevice, false, false, []bootDeviceProviders{{"test provider", &testImplementation}})
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
		})
	}
}

func TestSetBootDeviceFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		bootDevice        string
		err               error
		badImplementation bool
		want              bool
		withName          bool
	}{
		"success":                  {bootDevice: "pxe", want: true},
		"success with metadata":    {bootDevice: "pxe", want: true, withName: true},
		"no implementations found": {bootDevice: "pxe", want: false, badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a BootDeviceSetter implementation: *struct {}"), errors.New("no BootDeviceSetter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := bootDeviceTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			result, metadata, err := SetBootDeviceFromInterfaces(context.Background(), 0, tc.bootDevice, false, false, generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
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

type mockBootDeviceOverrideGetter struct {
	overrideReturn BootDeviceOverride
	errReturn      error
}

func (m *mockBootDeviceOverrideGetter) Name() string {
	return "Mock"
}

func (m *mockBootDeviceOverrideGetter) BootDeviceOverrideGet(_ context.Context) (BootDeviceOverride, error) {
	return m.overrideReturn, m.errReturn
}

func TestBootDeviceOverrideGet(t *testing.T) {
	successOverride := BootDeviceOverride{
		IsPersistent: false,
		IsEFIBoot:    true,
		Device:       "disk",
	}

	successMetadata := &Metadata{
		SuccessfulProvider:   "Mock",
		ProvidersAttempted:   []string{"Mock"},
		SuccessfulOpenConns:  nil,
		SuccessfulCloseConns: []string(nil),
		FailedProviderDetail: map[string]string{},
	}

	mixedMetadata := &Metadata{
		SuccessfulProvider:   "Mock",
		ProvidersAttempted:   []string{"Mock", "Mock"},
		SuccessfulOpenConns:  nil,
		SuccessfulCloseConns: []string(nil),
		FailedProviderDetail: map[string]string{"Mock": "foo-failure"},
	}

	failMetadata := &Metadata{
		SuccessfulProvider:   "",
		ProvidersAttempted:   []string{"Mock"},
		SuccessfulOpenConns:  nil,
		SuccessfulCloseConns: []string(nil),
		FailedProviderDetail: map[string]string{"Mock": "foo-failure"},
	}

	emptyMetadata := &Metadata{
		FailedProviderDetail: make(map[string]string),
	}

	testCases := []struct {
		name               string
		hasCanceledContext bool
		expectedErrorMsg   string
		expectedMetadata   *Metadata
		expectedOverride   BootDeviceOverride
		getters            []interface{}
	}{
		{
			name:             "success",
			expectedMetadata: successMetadata,
			expectedOverride: successOverride,
			getters: []interface{}{
				&mockBootDeviceOverrideGetter{overrideReturn: successOverride},
			},
		},
		{
			name:             "multiple getters",
			expectedMetadata: mixedMetadata,
			expectedOverride: successOverride,
			getters: []interface{}{
				"not a getter",
				&mockBootDeviceOverrideGetter{errReturn: fmt.Errorf("foo-failure")},
				&mockBootDeviceOverrideGetter{overrideReturn: successOverride},
			},
		},
		{
			name:             "error",
			expectedMetadata: failMetadata,
			expectedErrorMsg: "failed to get boot device override settings",
			getters: []interface{}{
				&mockBootDeviceOverrideGetter{errReturn: fmt.Errorf("foo-failure")},
			},
		},
		{
			name:             "nil BootDeviceOverrideGetters",
			expectedMetadata: emptyMetadata,
			expectedErrorMsg: "no BootDeviceOverrideGetter implementations found",
		},
		{
			name:             "nil BootDeviceOverrideGetter",
			expectedMetadata: emptyMetadata,
			expectedErrorMsg: "no BootDeviceOverrideGetter implementations found",
			getters:          []interface{}{nil},
		},
		{
			name:               "with canceled context",
			hasCanceledContext: true,
			expectedMetadata:   emptyMetadata,
			expectedErrorMsg:   "context canceled",
			getters: []interface{}{
				&mockBootDeviceOverrideGetter{},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if testCase.hasCanceledContext {
				cancel()
			}

			override, metadata, err := GetBootDeviceOverrideFromInterface(ctx, 0, testCase.getters)

			if testCase.expectedErrorMsg != "" {
				assert.ErrorContains(t, err, testCase.expectedErrorMsg)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, testCase.expectedOverride, override)
			assert.Equal(t, testCase.expectedMetadata, &metadata)
		})
	}
}
