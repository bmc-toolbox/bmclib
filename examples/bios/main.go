package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	logrusr "github.com/bombsimon/logrusr/v2"
	bmclib "github.com/metal-toolbox/bmclib"
	"github.com/metal-toolbox/bmclib/providers"

	"github.com/sirupsen/logrus"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Command line option flag parsing
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Username to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	mode := flag.String("mode", "get", "Mode [get,set,reset]")
	dfile := flag.String("file", "", "Read data from file")

	flag.Parse()

	// Logger configuration
	l := logrus.New()
	l.Level = logrus.DebugLevel
	// l.Level = logrus.TraceLevel
	logger := logrusr.New(l)
	logger.V(9)

	// bmclib client abstraction
	clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}

	client := bmclib.NewClient(*host, *user, *pass, clientOpts...)
	client.Registry.Drivers = client.Registry.Supports(
		providers.FeatureGetBiosConfiguration,
		providers.FeatureSetBiosConfiguration,
		providers.FeatureResetBiosConfiguration,
		providers.FeatureSetBiosConfigurationFromFile)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer client.Close(ctx)

	// Operating mode selection
	switch strings.ToLower(*mode) {
	case "get":
		// retrieve bios configuration
		biosConfig, err := client.GetBiosConfiguration(ctx)
		if err != nil {
			l.Fatal(err)
		}

		fmt.Printf("biosConfig: %#v\n", biosConfig)
	case "set":
		exampleConfig := make(map[string]string)

		if *dfile != "" {
			jsonFile, err := os.Open(*dfile)
			if err != nil {
				l.Fatal(err)
			}

			defer jsonFile.Close()

			jsonData, _ := io.ReadAll(jsonFile)

			err = json.Unmarshal(jsonData, &exampleConfig)
			if err != nil {
				l.Fatal(err)
			}
		} else {
			exampleConfig["TpmSecurity"] = "Off"
		}

		fmt.Println("Attempting to set BIOS configuration:")
		fmt.Printf("exampleConfig: %+v\n", exampleConfig)

		err := client.SetBiosConfiguration(ctx, exampleConfig)
		if err != nil {
			l.Error(err)
		}
	case "setfile":
		fmt.Println("Attempting to set BIOS configuration:")

		contents, err := os.ReadFile(*dfile)
		if err != nil {
			l.Fatal(err)
		}

		err = client.SetBiosConfigurationFromFile(ctx, string(contents))
		if err != nil {
			l.Error(err)
		}
	case "reset":
		err := client.ResetBiosConfiguration(ctx)
		if err != nil {
			l.Error(err)
		}
	default:
		l.Fatal("Unknown mode: " + *mode)
	}
}
