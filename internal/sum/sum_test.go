package sum

import (
	"context"
	"os"
	"testing"

	ex "github.com/metal-toolbox/bmclib/internal/executor"
)

func newFakeSum(t *testing.T, fixtureName string) *Sum {
	e := &Sum{
		Executor: ex.NewFakeExecutor("sum"),
	}

	b, err := os.ReadFile("../../fixtures/internal/sum/" + fixtureName)
	if err != nil {
		t.Error(err)
	}

	e.Executor.SetStdout(b)

	return e
}

func TestExec_Run(t *testing.T) {
	// Create a new instance of Sum
	exec := newFakeSum(t, "GetBIOSInfo")

	// Create a new context
	ctx := context.Background()

	// Call the run function
	_, err := exec.run(ctx, "GetBIOSInfo")

	// Check the output and error
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestExec_SetBiosConfiguration(t *testing.T) {
	// Create a new context
	ctx := context.Background()

	// Define the BIOS configuration
	biosConfig := map[string]string{
		"boot_mode":   "UEFI",
		"boot_order":  "UEFI",
		"intel_sgx":   "Enabled",
		"secure_boot": "Enabled",
		"tpm":         "Enabled",
		"smt":         "Disabled",
		"sr_iov":      "Enabled",
		"raw:Menu1,SubMenu1,SubMenuMenu1,SettingName": "Value",
	}

	exec := newFakeSum(t, "SetBiosConfiguration")

	// Call the SetBiosConfiguration function
	err := exec.SetBiosConfiguration(ctx, biosConfig)

	// Check for any errors
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Additional assertions can be added to verify the behavior of the function
}

func TestExec_GetBiosConfiguration(t *testing.T) {
	// Create a new context
	ctx := context.Background()

	exec := newFakeSum(t, "GetBiosConfiguration")

	// Call the SetBiosConfiguration function
	biosConfig, err := exec.GetBiosConfiguration(ctx)

	// Check for any errors
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Confirm boot_mode exists
	_, ok := biosConfig["boot_mode"]
	if !ok {
		t.Fail()
	}
}
