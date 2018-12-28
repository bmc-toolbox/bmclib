package config

import (
	"io"
	"os"
	"testing"
)

// TestLoad tests if the Params struct was filled
// in with all required parameters.
func TestLoad(t *testing.T) {

	// source config file
	sFile := "../../samples/bmcbutler.yml.sample"

	// target config file
	tFile := "/tmp/bmcbutler.yml"

	// verify source exists
	_, err := os.Stat(sFile)
	if err != nil {
		t.Fatalf("Expected sample config not found: %s", sFile)
	}

	// get the source file handle
	sFD, err := os.Open(sFile)
	if err != nil {
		t.Fatalf("Unable to open sample config file %s", sFile)
	}
	defer sFD.Close()

	// attempt to create the target file
	tFD, err := os.Create(tFile)
	if err != nil {
		t.Fatalf("Unable to create test config file %s", tFile)
	}
	defer tFD.Close()
	// clean up file after test
	defer func() { _ = os.Remove(tFile) }()

	// copy source data to target.
	_, err = io.Copy(tFD, sFD)
	if err != nil {
		t.Fatalf("Unable to copy %s -> %s", sFile, tFile)
	}

	// Load config
	cfg := &Params{}
	cfg.Load(tFile)

	// validate the tFile is used as the config file.
	if tFile != cfg.CfgFile {
		t.Fatalf("Expected cfgFile to match %s but is %s", tFile, cfg.CfgFile)
	}

	// validate butlersToSpawn cfg param was read in.
	if cfg.ButlersToSpawn <= 0 {
		t.Fatal("Expected ButlersToSpawn config param to be > 0")
	}

	// validate Inventory source was read.
	if cfg.InventoryParams.Source == "" {
		t.Fatal("Expected InventoryParams Source to have a valid attribute.")
	}

	// validate locations cfg params was read in.
	if len(cfg.Locations) == 0 {
		t.Fatal("Expected Locations config param empty.")
	}

	// validate bmcCfgDir cfg param was read in.
	if len(cfg.Credentials) == 0 {
		t.Fatal("Expected credentials param empty.")
	}

}
