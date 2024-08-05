package sum

import (
	"context"
	"os"
	"testing"

	ex "github.com/bmc-toolbox/bmclib/v2/internal/executor"
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
	// Create a new instance of Exec
	exec := &Sum{
		Host:     "example.com",
		Username: "user",
		Password: "password",
		SumPath:  "/path/to/sum",
	}

	// Create a new context
	ctx := context.Background()

	// Define the command and additional arguments
	command := "some-command"
	additionalArgs := []string{"arg1", "arg2"}

	// Call the run function
	output, err := exec.run(ctx, command, additionalArgs...)

	// Check the output and error
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	expectedOutput := "some output"
	if output != expectedOutput {
		t.Errorf("Expected output: %s, got: %s", expectedOutput, output)
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
