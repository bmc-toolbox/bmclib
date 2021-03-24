package bmc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type firmwareTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (f *firmwareTester) GetBMCVersion(ctx context.Context) (version string, err error) {
	if f.MakeErrorOut {
		return "", errors.New("failed to get BMC version")
	}
	if f.MakeNotOK {
		return "", nil
	}
	return "1.33.7", nil
}

func (f *firmwareTester) FirmwareUpdateBMC(ctx context.Context, fileReader io.Reader, fileSize int64) (err error) {
	if f.MakeErrorOut {
		return errors.New("failed update")
	}

	return nil
}

func (f *firmwareTester) FirmwareUpdateBIOS(ctx context.Context, fileReader io.Reader, fileSize int64) (err error) {
	if f.MakeErrorOut {
		return errors.New("failed update")
	}
	return nil
}

func (f *firmwareTester) GetBIOSVersion(ctx context.Context) (version string, err error) {
	if f.MakeErrorOut {
		return "", errors.New("failed to get BIOS version")
	}
	if f.MakeNotOK {
		return "", nil
	}
	return "1.44.7", nil
}

func TestGetBMCVersion(t *testing.T) {
	testCases := []struct {
		name       string
		version    string
		makeFail   bool
		err        error
		ctxTimeout time.Duration
	}{
		{name: "success", version: "1.33.7", err: nil},
		{name: "failure", version: "", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("failed to get BMC version"), errors.New("failed to get BMC version")}}},
		{name: "fail context timeout", version: "", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to get BMC version")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := firmwareTester{MakeErrorOut: tc.makeFail}
			expectedResult := tc.version
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := GetBMCVersion(ctx, []BMCVersionGetter{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestGetBMCVersionFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		version           string
		err               error
		badImplementation bool
		want              string
	}{
		{name: "success", version: "1.33.7", err: nil},
		{name: "no implementations found", version: "", want: "", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a BMCVersionGetter implementation: *struct {}"), errors.New("no BMCVersionGetter implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := firmwareTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.version
			result, err := GetBMCVersionFromInterfaces(context.Background(), generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestUpdateBMCFirmware(t *testing.T) {
	testCases := []struct {
		name       string
		makeFail   bool
		err        error
		ctxTimeout time.Duration
	}{
		{name: "success", err: nil},
		{name: "failure", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("failed update"), errors.New("failed to update BMC firmware")}}},
		{name: "fail context timeout", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to update BMC firmware")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := firmwareTester{MakeErrorOut: tc.makeFail}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			err := UpdateBMCFirmware(ctx, bytes.NewReader([]byte(`foo`)), 0, []BMCFirmwareUpdater{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestUpdateBMCFirmwareFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		badImplementation bool
		want              string
	}{
		{name: "success", err: nil},
		{name: "no implementations found", want: "", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a BMCFirmwareUpdater implementation: *struct {}"), errors.New("no BMCFirmwareUpdater implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := firmwareTester{}
				generic = []interface{}{&testImplementation}
			}
			err := UpdateBMCFirmwareFromInterfaces(context.Background(), bytes.NewReader([]byte(`foo`)), 0, generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestGetBIOSVersion(t *testing.T) {
	testCases := []struct {
		name       string
		version    string
		makeFail   bool
		err        error
		ctxTimeout time.Duration
	}{
		{name: "success", version: "1.44.7", err: nil},
		{name: "failure", version: "", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("failed to get BIOS version"), errors.New("failed to get BIOS version")}}},
		{name: "fail context timeout", version: "", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to get BIOS version")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := firmwareTester{MakeErrorOut: tc.makeFail}
			expectedResult := tc.version
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := GetBIOSVersion(ctx, []BIOSVersionGetter{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestGetBIOSVersionFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		version           string
		err               error
		badImplementation bool
		want              string
	}{
		{name: "success", version: "1.44.7", err: nil},
		{name: "no implementations found", version: "", want: "", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a BIOSVersionGetter implementation: *struct {}"), errors.New("no BIOSVersionGetter implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := firmwareTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.version
			result, err := GetBIOSVersionFromInterfaces(context.Background(), generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestUpdateBIOSFirmware(t *testing.T) {
	testCases := []struct {
		name       string
		makeFail   bool
		err        error
		ctxTimeout time.Duration
	}{
		{name: "success", err: nil},
		{name: "failure", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("failed update"), errors.New("failed to update BIOS firmware")}}},
		{name: "fail context timeout", makeFail: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to update BIOS firmware")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := firmwareTester{MakeErrorOut: tc.makeFail}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			err := UpdateBIOSFirmware(ctx, bytes.NewReader([]byte(`foo`)), 0, []BIOSFirmwareUpdater{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestUpdateBIOSFirmwareFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		badImplementation bool
		want              string
	}{
		{name: "success", err: nil},
		{name: "no implementations found", want: "", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a BIOSFirmwareUpdater implementation: *struct {}"), errors.New("no BIOSFirmwareUpdater implementations found")}}},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := firmwareTester{}
				generic = []interface{}{&testImplementation}
			}
			err := UpdateBIOSFirmwareFromInterfaces(context.Background(), bytes.NewReader([]byte(`foo`)), 0, generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
