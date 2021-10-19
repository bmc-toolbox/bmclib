package ibmc

import (
	"testing"

	"github.com/bmc-toolbox/bmclib/devices"
)

func TestIbmcInterface(t *testing.T) {
	bmc, err := New("127.0.0.1", "foo", "bar")
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	_ = devices.Configure(bmc)
	_ = devices.Bmc(bmc)
}
