package bmc

import (
	"context"
	"testing"
	"time"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/errors"
	"github.com/stretchr/testify/assert"
)

type inventoryGetterTester struct {
	returnDevice *common.Device
	returnError  error
}

func (f *inventoryGetterTester) Inventory(ctx context.Context) (device *common.Device, err error) {
	return f.returnDevice, f.returnError
}

func (f *inventoryGetterTester) Name() string {
	return "foo"
}

func TestInventory(t *testing.T) {
	testCases := []struct {
		testName           string
		returnDevice       *common.Device
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
	}{
		{"success with metadata", &common.Device{Common: common.Common{Vendor: "foo"}}, nil, 5 * time.Second, "foo", 1},
		{"failure with metadata", nil, errors.ErrNon200Response, 5 * time.Second, "foo", 1},
		{"failure with context timeout", nil, context.DeadlineExceeded, 1 * time.Nanosecond, "foo", 1},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			testImplementation := inventoryGetterTester{returnDevice: tc.returnDevice, returnError: tc.returnError}
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			device, metadata, err := inventory(ctx, []inventoryGetterProvider{{tc.providerName, &testImplementation}})
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.returnDevice, device)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
			assert.Equal(t, tc.providersAttempted, len(metadata.ProvidersAttempted))
		})
	}
}

func TestInventoryFromInterfaces(t *testing.T) {
	testCases := []struct {
		testName           string
		returnDevice       *common.Device
		returnError        error
		ctxTimeout         time.Duration
		providerName       string
		providersAttempted int
		badImplementation  bool
	}{
		{"success with metadata", &common.Device{Common: common.Common{Vendor: "foo"}}, nil, 5 * time.Second, "foo", 1, false},
		{"failure with bad implementation", nil, errors.ErrProviderImplementation, 5 * time.Second, "foo", 1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := &inventoryGetterTester{returnDevice: tc.returnDevice, returnError: tc.returnError}
				generic = []interface{}{testImplementation}
			}
			device, metadata, err := GetInventoryFromInterfaces(context.Background(), generic)
			if tc.returnError != nil {
				assert.ErrorIs(t, err, tc.returnError)
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.returnDevice, device)
			assert.Equal(t, tc.providerName, metadata.SuccessfulProvider)
		})
	}
}
