// Package main demonstrates reading, setting and clearing the chassis power cap with bmclib.
package main

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"

	logrusr "github.com/bombsimon/logrusr/v2"
	"github.com/sirupsen/logrus"

	bmclib "github.com/bmc-toolbox/bmclib/v2"
	"github.com/bmc-toolbox/bmclib/v2/providers"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Command line option flag parsing
	user := flag.String("user", "", "Username to login with")
	pass := flag.String("password", "", "Password to login with")
	host := flag.String("host", "", "BMC hostname to connect to")
	mode := flag.String("mode", "get", "Mode [get,set,clear]")
	limit := flag.Float64("limit", 0, "Power cap in watts (for set mode)")

	flag.Parse()

	// Logger configuration
	l := logrus.New()
	l.Level = logrus.DebugLevel
	logger := logrusr.New(l)

	// Validate required parameters
	if *host == "" || *user == "" || *pass == "" {
		l.Fatal("required host/user/pass parameters not defined")
	}

	// bmclib client abstraction
	clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}
	client := bmclib.NewClient(*host, *user, *pass, clientOpts...)

	// Filter to providers that support power cap operations
	client.Registry.Drivers = client.Registry.Supports(
		providers.FeatureGetPowerMetrics,
		providers.FeatureSetPowerCap,
	)

	err := client.Open(ctx)
	if err != nil {
		l.Fatal(err, "bmc login failed")
	}

	defer func() { _ = client.Close(ctx) }()

	// Operating mode selection
	switch strings.ToLower(*mode) {
	case "get":
		// Read the current power readings and cap
		metrics, err := client.GetPowerMetrics(ctx)
		if err != nil {
			l.Fatal(err)
		}

		fmt.Printf("Consumed: %.0f W\nCapacity: %.0f W\n", metrics.ConsumedWatts, metrics.CapacityWatts)
		if metrics.LimitInWatts == nil {
			fmt.Println("Power cap: not set")
		} else {
			fmt.Printf("Power cap: %.0f W\n", *metrics.LimitInWatts)
		}

	case "set":
		// Set a power cap
		if *limit <= 0 {
			l.Fatal("set mode requires a positive -limit in watts")
		}

		err := client.SetPowerCap(ctx, limit)
		if err != nil {
			l.Fatal(err)
		}

		fmt.Printf("Power cap set to %.0f W\n", *limit)

	case "clear":
		// Clear the power cap (nil limit disables capping)
		err := client.SetPowerCap(ctx, nil)
		if err != nil {
			l.Fatal(err)
		}

		fmt.Println("Power cap cleared")

	default:
		l.Fatal("unknown mode: " + *mode)
	}
}
