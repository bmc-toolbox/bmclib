package inventory

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmcbutler/pkg/asset"
	"github.com/bmc-toolbox/bmcbutler/pkg/config"
	"github.com/sirupsen/logrus"
)

// TestExecCmd tests
// - A bad command exec failure is returned as expected
// - A successful command exec returns as expected
// - Args given are returned as expected.
func TestExecCmd(t *testing.T) {

	badCMD := "/usr/bin/echofoobar"
	_, err := ExecCmd(badCMD, []string{})
	if err == nil {
		t.Fatalf("Expected error on bad command, but returned error was nil")
	}

	cmd := "/usr/bin/echo"
	args := []string{"ARG1", "ARG2", "ARG3"}
	out, err := ExecCmd(cmd, args)
	if err != nil {
		t.Fatalf("Expected successful command execution, got error: %s", err)
	}

	out1 := strings.Trim(string(out), "\n")

	if out1 != strings.Join(args, " ") {
		t.Fatalf("Expected output '%s' instead got '%s'", strings.Join(args, " "), out1)
	}

}

// TestAttributesExtrasAsMap tests the method returns a map[string]string
// which is a map of the Extras Attributes of an asset.
func TestAttributesExtrasAsMap(t *testing.T) {

	aExtras := new(AttributesExtras)
	aExtras.State = "live"
	aExtras.Company = "acme"
	aExtras.LiveAssets = &[]string{"FOO", "BAR"}

	expectedMap := map[string]string{
		"state":      "live",
		"company":    "acme",
		"liveAssets": strings.ToLower(strings.Join(*aExtras.LiveAssets, ",")),
	}

	extras := AttributesExtrasAsMap(aExtras)

	for k := range expectedMap {
		_, exists := extras[k]
		if !exists {
			t.Fatalf("Expected extras map to have key, %s", k)
		}

		if expectedMap[k] != extras[k] {
			t.Fatalf("Expected extras map key value (%s -> %s) to match expectedMap key value (%s -> %s)", k, extras[k], k, expectedMap[k])
		}
	}
}

// TestEncQueryBySerial tests assets are returned by encQueryBySerial as expected.
func TestEncQueryBySerial(t *testing.T) {

	serials := []string{"foobar", "barfoo"}
	assetLookup := "/tmp/assetlookup"
	cmd := "/usr/bin/go"

	//build asset lookup bin
	args := []string{"build", "-o", assetLookup, "../../samples/assetlookup.go"}
	_, err := ExecCmd(cmd, args)
	if err != nil {
		t.Fatalf("Expected to build assetlookup for test, but failed with error : %s", err)
	}
	defer func() { _ = os.Remove(assetLookup) }()

	enc := Enc{
		Log:    logrus.New(),
		Config: &config.Params{InventoryParams: &config.InventoryParams{EncExecutable: assetLookup}},
	}

	assets := enc.encQueryBySerial(strings.Join(serials, ","))
	if len(assets) < 2 {
		t.Fatalf("Expected two assets to be returned, got %d", len(assets))
	}

	for _, serial := range serials {
		if !assetInAssets(serial, assets) {
			t.Fatalf("Expected asset with serial %s not found in returned assets", serial)
		}
	}
}

func assetInAssets(serial string, assets []asset.Asset) bool {
	for _, asset := range assets {
		if asset.Serial == serial {
			fmt.Printf("%+v\n", asset)
			return true
		}
	}
	return false
}
